// Package datasetstorage persists validated JSONL uploads under a trusted local
// root. Authorization belongs to callers; local persistence is not runtime delivery.
package datasetstorage

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

var (
	ErrInvalid  = errors.New("invalid dataset")
	ErrConflict = errors.New("dataset version already exists")
	ErrNotFound = errors.New("dataset version not found")
	ErrQuota    = errors.New("dataset namespace quota exceeded")
)

// MaxNamespaceBytes limits committed upload bytes in each namespace.
const MaxNamespaceBytes int64 = 100 << 20

// MaxNamespaceVersions bounds committed file counts and metadata scans.
const MaxNamespaceVersions = 1000

const (
	maxFileBytes    = 10 << 20
	maxLineBytes    = 1 << 20
	maxPreviewBytes = 16 << 10
)

var (
	labelPattern   = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)
	versionPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)
)

// Asset is the public metadata of a locally persisted upload, without a file URI.
type Asset struct {
	Namespace    string            `json:"namespace"`
	Name         string            `json:"name"`
	Owner        string            `json:"owner"`
	Version      string            `json:"version"`
	Format       string            `json:"format"`
	Status       string            `json:"status"`
	RuntimeReady bool              `json:"runtimeReady"`
	Bytes        int64             `json:"bytes"`
	Rows         int64             `json:"rows"`
	SHA256       string            `json:"sha256"`
	CreatedAt    time.Time         `json:"createdAt"`
	Preview      []json.RawMessage `json:"preview"`
	Source       string            `json:"source"`
}

// Repository stores immutable versions on a trusted filesystem. Use one instance
// per root in a single process to serialize publication and namespace quota checks.
// Files beneath the root must not be concurrently changed by external writers.
type Repository struct {
	root    string
	mu      sync.Mutex
	syncDir func(string) error
}

// New opens or creates a persistent storage root. The root is server-configured,
// never a user-provided upload path. Existing uploads are retained.
func New(root string) (*Repository, error) {
	if root == "" {
		return nil, fmt.Errorf("%w: root is required", ErrInvalid)
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve dataset root: %w", err)
	}
	if err := createRoot(abs); err != nil {
		return nil, fmt.Errorf("create dataset root: %w", err)
	}
	// A trusted configured root may itself be a symlink; pin its resolved path.
	abs, err = filepath.EvalSymlinks(abs)
	if err != nil {
		return nil, fmt.Errorf("resolve dataset root: %w", err)
	}
	if err := checkDirectory(abs); err != nil {
		return nil, err
	}
	return &Repository{root: abs, syncDir: syncDirectory}, nil
}

// Create validates JSON objects in nonempty lines and hashes the original bytes.
// The line limit excludes LF (and a preceding CR for CRLF). Only the first five
// records are considered for preview; records exceeding its JSON budget are omitted.
// Cancellation is checked between reads and before publication. A blocking Read
// must be interrupted by the caller, for example by closing its request body.
// A directory-sync failure after rename returns an error but preserves the
// committed version; a retry can therefore return ErrConflict.
func (r *Repository) Create(ctx context.Context, namespace, name, owner string, reader io.Reader) (Asset, error) {
	if err := ctx.Err(); err != nil {
		return Asset{}, err
	}
	if err := validateLabels(namespace, name); err != nil {
		return Asset{}, err
	}
	if reader == nil {
		return Asset{}, fmt.Errorf("%w: reader is required", ErrInvalid)
	}
	ns, err := r.assetDirectory(namespace, "", true)
	if err != nil {
		return Asset{}, err
	}
	tmp, err := os.MkdirTemp(ns, ".upload-")
	if err != nil {
		return Asset{}, fmt.Errorf("create upload staging directory: %w", err)
	}
	defer os.RemoveAll(tmp)
	data, err := os.OpenFile(filepath.Join(tmp, "data.jsonl"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return Asset{}, err
	}
	defer data.Close()
	a := Asset{Namespace: namespace, Name: name, Owner: owner, Format: "jsonl", Status: "validated", Source: "upload", Preview: []json.RawMessage{}}
	if err := validateUpload(ctx, reader, data, &a); err != nil {
		return Asset{}, err
	}
	if err := data.Sync(); err != nil {
		return Asset{}, fmt.Errorf("sync dataset: %w", err)
	}
	if err := data.Close(); err != nil {
		return Asset{}, err
	}
	a.CreatedAt = time.Now().UTC()
	metadata, err := json.Marshal(a)
	if err != nil {
		return Asset{}, fmt.Errorf("encode dataset metadata: %w", err)
	}
	if err := writeMetadata(filepath.Join(tmp, "metadata.json"), metadata); err != nil {
		return Asset{}, err
	}
	if err := r.syncDir(tmp); err != nil {
		return Asset{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return Asset{}, err
	}
	if _, err := r.assetDirectory(namespace, name, false); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Asset{}, err
	}
	parent := filepath.Join(ns, name)
	target := filepath.Join(parent, a.Version)
	// Rename may replace an empty directory on Unix, so explicitly reject every
	// existing target while holding the same lock as all repository publications.
	if _, err := os.Lstat(target); err == nil {
		return Asset{}, fmt.Errorf("%w: %s/%s/%s", ErrConflict, namespace, name, a.Version)
	} else if !errors.Is(err, os.ErrNotExist) {
		return Asset{}, err
	}
	assets, err := r.List(ctx, namespace)
	if err != nil {
		return Asset{}, err
	}
	if len(assets) >= MaxNamespaceVersions {
		return Asset{}, fmt.Errorf("%w: %s has reached %d versions", ErrQuota, namespace, MaxNamespaceVersions)
	}
	used := int64(0)
	for _, existing := range assets {
		if existing.Bytes > MaxNamespaceBytes-used {
			return Asset{}, fmt.Errorf("%w: %s", ErrQuota, namespace)
		}
		used += existing.Bytes
	}
	if a.Bytes > MaxNamespaceBytes-used {
		return Asset{}, fmt.Errorf("%w: %s", ErrQuota, namespace)
	}
	if err := ctx.Err(); err != nil {
		return Asset{}, err
	}
	// Staging does not allocate user-chosen names. Only create a name after
	// validation and quotas succeed, while publication is serialized.
	if err := os.Mkdir(parent, 0700); err == nil {
		// Remove only an empty directory on publication failure; never a version.
		defer os.Remove(parent)
	} else if !errors.Is(err, os.ErrExist) {
		return Asset{}, err
	}
	if err := checkDirectory(parent); err != nil {
		return Asset{}, err
	}
	if err := os.Rename(tmp, target); err != nil {
		return Asset{}, fmt.Errorf("publish dataset: %w", err)
	}
	// Persist the destination version entry, both sides of the staging rename,
	// and the namespace entry in the root, in child-to-parent order.
	for _, dir := range []string{parent, ns, r.root} {
		if err := r.syncDir(dir); err != nil {
			return Asset{}, fmt.Errorf("sync published dataset directory: %w", err)
		}
	}
	return a, nil
}

func validateUpload(ctx context.Context, input io.Reader, output io.Writer, a *Asset) error {
	hash := sha256.New()
	limited := &io.LimitedReader{R: contextReader{ctx: ctx, reader: input}, N: maxFileBytes + 1}
	stream := bufio.NewReaderSize(io.TeeReader(limited, io.MultiWriter(output, hash)), maxLineBytes+2)
	previewBytes := 2 // JSON array brackets count toward the preview budget.
	for lineNumber := int64(1); ; lineNumber++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		line, err := stream.ReadSlice('\n')
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if limited.N == 0 {
			return fmt.Errorf("%w: file exceeds 10 MiB", ErrInvalid)
		}
		if errors.Is(err, bufio.ErrBufferFull) {
			return fmt.Errorf("%w: line %d exceeds 1 MiB", ErrInvalid, lineNumber)
		}
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("read upload: %w", err)
		}
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
		}
		if len(line) > maxLineBytes {
			return fmt.Errorf("%w: line %d exceeds 1 MiB", ErrInvalid, lineNumber)
		}
		line = bytes.Trim(line, " \t\r")
		if len(line) > 0 {
			if !utf8.Valid(line) || line[0] != '{' || !json.Valid(line) {
				return fmt.Errorf("%w: line %d must be a JSON object", ErrInvalid, lineNumber)
			}
			a.Rows++
			if a.Rows <= 5 {
				// Marshal RawMessage to count escaping exactly as public JSON does.
				preview, marshalErr := json.Marshal(json.RawMessage(line))
				if marshalErr != nil {
					return fmt.Errorf("%w: preview: %v", ErrInvalid, marshalErr)
				}
				extra := len(preview)
				if len(a.Preview) > 0 {
					extra++
				}
				if previewBytes+extra <= maxPreviewBytes {
					a.Preview = append(a.Preview, json.RawMessage(preview))
					previewBytes += extra
				}
			}
		}
		if errors.Is(err, io.EOF) {
			break
		}
	}
	if a.Rows == 0 {
		return fmt.Errorf("%w: dataset has no records", ErrInvalid)
	}
	a.Bytes = maxFileBytes + 1 - limited.N
	a.SHA256 = hex.EncodeToString(hash.Sum(nil))
	a.Version = a.SHA256
	return nil
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	n, err := r.reader.Read(p)
	if canceled := r.ctx.Err(); canceled != nil {
		return n, canceled
	}
	return n, err
}

// List returns only the requested namespace, sorted by name then version.
// Missing namespaces return an empty array; corrupt committed entries are errors.
func (r *Repository) List(ctx context.Context, namespace string) ([]Asset, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := validateLabels(namespace); err != nil {
		return nil, err
	}
	assets := []Asset{}
	ns, err := r.assetDirectory(namespace, "", false)
	if errors.Is(err, os.ErrNotExist) {
		return assets, nil
	}
	if err != nil {
		return nil, err
	}
	names, err := os.ReadDir(ns)
	if err != nil {
		return nil, err
	}
	for _, entry := range names {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".upload-") {
			continue
		}
		if err := validateLabels(name); err != nil {
			return nil, err
		}
		parent, err := r.assetDirectory(namespace, name, false)
		if err != nil {
			return nil, err
		}
		versions, err := os.ReadDir(parent)
		if err != nil {
			return nil, err
		}
		for _, version := range versions {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			if strings.HasPrefix(version.Name(), ".upload-") {
				continue
			}
			a, err := r.Get(ctx, namespace, name, version.Name())
			if err != nil {
				return nil, fmt.Errorf("list dataset %s/%s: %w", name, version.Name(), err)
			}
			assets = append(assets, a)
		}
	}
	sort.Slice(assets, func(i, j int) bool {
		if assets[i].Name != assets[j].Name {
			return assets[i].Name < assets[j].Name
		}
		return assets[i].Version < assets[j].Version
	})
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return assets, nil
}

// Get reads metadata for an immutable version within the explicit namespace.
func (r *Repository) Get(ctx context.Context, namespace, name, version string) (Asset, error) {
	if err := ctx.Err(); err != nil {
		return Asset{}, err
	}
	if err := validateLabels(namespace, name); err != nil {
		return Asset{}, err
	}
	if !versionPattern.MatchString(version) {
		return Asset{}, fmt.Errorf("%w: version must be lowercase SHA256 hex", ErrInvalid)
	}
	parent, err := r.assetDirectory(namespace, name, false)
	if errors.Is(err, os.ErrNotExist) {
		return Asset{}, fmt.Errorf("%w: %s/%s/%s", ErrNotFound, namespace, name, version)
	}
	if err != nil {
		return Asset{}, err
	}
	dir := filepath.Join(parent, version)
	if err := checkDirectory(dir); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Asset{}, fmt.Errorf("%w: %s/%s/%s", ErrNotFound, namespace, name, version)
		}
		return Asset{}, err
	}
	for _, file := range []string{"metadata.json", "data.jsonl"} {
		info, err := os.Lstat(filepath.Join(dir, file))
		if err != nil {
			return Asset{}, fmt.Errorf("inspect committed dataset %s: %w", file, err)
		}
		if !info.Mode().IsRegular() {
			return Asset{}, fmt.Errorf("%w: %s must be a regular file", ErrInvalid, file)
		}
	}
	b, err := os.ReadFile(filepath.Join(dir, "metadata.json"))
	if err != nil {
		return Asset{}, fmt.Errorf("read dataset metadata: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return Asset{}, err
	}
	var a Asset
	if err := json.Unmarshal(b, &a); err != nil {
		return Asset{}, fmt.Errorf("%w: malformed metadata: %v", ErrInvalid, err)
	}
	if a.Namespace != namespace || a.Name != name || a.Version != version || a.SHA256 != version {
		return Asset{}, fmt.Errorf("%w: metadata identity does not match path", ErrInvalid)
	}
	if a.Format != "jsonl" || a.Status != "validated" || a.RuntimeReady || a.Source != "upload" || a.Bytes <= 0 || a.Bytes > maxFileBytes || a.Rows <= 0 || a.CreatedAt.IsZero() || a.Preview == nil || len(a.Preview) > 5 || int64(len(a.Preview)) > a.Rows {
		return Asset{}, fmt.Errorf("%w: inconsistent dataset metadata", ErrInvalid)
	}
	preview, err := json.Marshal(a.Preview)
	if err != nil || len(preview) > maxPreviewBytes {
		return Asset{}, fmt.Errorf("%w: invalid metadata preview", ErrInvalid)
	}
	for _, row := range a.Preview {
		trimmed := bytes.TrimSpace(row)
		if len(trimmed) == 0 || trimmed[0] != '{' || !json.Valid(trimmed) {
			return Asset{}, fmt.Errorf("%w: preview must contain JSON objects", ErrInvalid)
		}
	}
	return a, nil
}

func validateLabels(labels ...string) error {
	for _, label := range labels {
		if !labelPattern.MatchString(label) {
			return fmt.Errorf("%w: namespace and name must be DNS labels of 1-63 characters", ErrInvalid)
		}
	}
	return nil
}

func checkDirectory(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%w: storage path must be a directory, not a symlink or file", ErrInvalid)
	}
	return nil
}

func (r *Repository) assetDirectory(namespace, name string, create bool) (string, error) {
	path := r.root
	if err := checkDirectory(path); err != nil {
		return "", err
	}
	parts := []string{namespace}
	if name != "" {
		parts = append(parts, name)
	}
	for _, part := range parts {
		path = filepath.Join(path, part)
		if create {
			if err := os.Mkdir(path, 0700); err != nil && !errors.Is(err, os.ErrExist) {
				return "", err
			}
		}
		if err := checkDirectory(path); err != nil {
			return "", err
		}
	}
	return path, nil
}

func writeMetadata(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	return f.Close()
}

func syncDirectory(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Sync()
}

// Persist newly created root ancestors as well as the root directory entry.
// Symlinks are permitted only here, in the trusted configured root path.
func createRoot(path string) error {
	if info, err := os.Stat(path); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("%w: root must be a directory", ErrInvalid)
		}
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	parent := filepath.Dir(path)
	if err := createRoot(parent); err != nil {
		return err
	}
	if err := os.Mkdir(path, 0700); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	return syncDirectory(parent)
}
