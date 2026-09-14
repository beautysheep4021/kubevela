package datasetstorage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func newRepository(t *testing.T) (*Repository, string) {
	t.Helper()
	root := filepath.Join(t.TempDir(), "datasets")
	r, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	return r, root
}

func TestCreatePersistsOriginalBytesAndMetadata(t *testing.T) {
	r, root := newRepository(t)
	data := "\n{\"x\": 1}\r\n \t\n{\"nested\":{\"ok\":true}}"
	before := time.Now()
	a, err := r.Create(context.Background(), "tenant-1", "train", "alice", strings.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256([]byte(data))
	wantHash := hex.EncodeToString(hash[:])
	if a.Namespace != "tenant-1" || a.Name != "train" || a.Owner != "alice" || a.Version != wantHash || a.SHA256 != wantHash || a.Bytes != int64(len(data)) || a.Rows != 2 || a.Format != "jsonl" || a.Status != "validated" || a.RuntimeReady || a.Source != "upload" {
		t.Fatalf("unexpected asset: %+v", a)
	}
	if a.CreatedAt.Before(before) || a.CreatedAt.After(time.Now()) {
		t.Fatalf("invalid timestamp: %v", a.CreatedAt)
	}
	if len(a.Preview) != 2 || !json.Valid(a.Preview[0]) {
		t.Fatalf("preview: %s", a.Preview)
	}
	p := filepath.Join(root, a.Namespace, a.Name, a.Version)
	stored, err := os.ReadFile(filepath.Join(p, "data.jsonl"))
	if err != nil || string(stored) != data {
		t.Fatalf("stored data %q, %v", stored, err)
	}
	metadata, err := os.ReadFile(filepath.Join(p, "metadata.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(metadata, &fields); err != nil {
		t.Fatal(err)
	}
	keys := []string{"namespace", "name", "owner", "version", "format", "status", "runtimeReady", "bytes", "rows", "sha256", "createdAt", "preview", "source"}
	if len(fields) != len(keys) {
		t.Fatalf("unexpected JSON fields: %s", metadata)
	}
	for _, key := range keys {
		if _, ok := fields[key]; !ok {
			t.Errorf("missing %s", key)
		}
	}
	r, err = New(root)
	if err != nil {
		t.Fatal(err)
	}
	got, err := r.Get(context.Background(), a.Namespace, a.Name, a.Version)
	if err != nil {
		t.Fatal(err)
	}
	wantJSON, _ := json.Marshal(a)
	gotJSON, _ := json.Marshal(got)
	if string(gotJSON) != string(wantJSON) {
		t.Fatalf("reload mismatch: %s != %s", gotJSON, wantJSON)
	}
}

func TestValidationAndCleanup(t *testing.T) {
	cases := map[string]string{
		"empty": "", "blank": "\n \t\r\n", "invalid": "{", "array": "[]\n", "null": "null", "string": "\"hi\"", "number": "1", "boolean": "true", "multiple": "{} {}", "later invalid": "{}\nno", "long line": `{"x":"` + strings.Repeat("x", 1<<20) + `"}`, "long blank": strings.Repeat(" ", (1<<20)+1),
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			r, root := newRepository(t)
			_, err := r.Create(context.Background(), "ns", "data", "owner", strings.NewReader(data))
			if !errors.Is(err, ErrInvalid) {
				t.Fatalf("expected ErrInvalid, got %v", err)
			}
			assertNoFiles(t, root)
		})
	}
}

func TestLimitsAndPreview(t *testing.T) {
	for _, size := range []int{(1 << 20) - 1, 1 << 20, (1 << 20) + 1} {
		t.Run(stringSize(size), func(t *testing.T) {
			r, _ := newRepository(t)
			line := `{"x":"` + strings.Repeat("x", size-8) + `"}`
			_, err := r.Create(context.Background(), "ns", "data", "owner", strings.NewReader(line+"\n"))
			if size <= 1<<20 && err != nil {
				t.Fatal(err)
			}
			if size > 1<<20 && !errors.Is(err, ErrInvalid) {
				t.Fatalf("oversized line: %v", err)
			}
		})
	}
	t.Run("file boundary", func(t *testing.T) {
		r, root := newRepository(t)
		line := `{"x":"` + strings.Repeat("x", (1<<20)-9) + "\"}\n"
		data := strings.Repeat(line, 10)
		if len(data) != 10<<20 {
			t.Fatal("bad test fixture")
		}
		if _, err := r.Create(context.Background(), "ns", "exact", "o", strings.NewReader(data)); err != nil {
			t.Fatal(err)
		}
		if _, err := r.Create(context.Background(), "ns", "over", "o", strings.NewReader(data+"\n")); !errors.Is(err, ErrInvalid) {
			t.Fatalf("file limit: %v", err)
		}
		assertNoFiles(t, filepath.Join(root, "ns", "over"))
	})
	t.Run("bounded first five", func(t *testing.T) {
		r, _ := newRepository(t)
		data := "{\"large\":\"" + strings.Repeat("x", 17000) + "\"}\n"
		data += "{\"escaped\":\"" + strings.Repeat("<", 4000) + "\"}\n"
		data += "{\"n\":3}\n{\"n\":4}\n{\"n\":5}\n{\"n\":6}\n"
		a, err := r.Create(context.Background(), "ns", "data", "o", strings.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		preview, err := json.Marshal(a.Preview)
		if err != nil || len(preview) > 16<<10 || len(a.Preview) != 3 || strings.Contains(string(preview), `"n":6`) {
			t.Fatalf("preview size=%d rows=%d err=%v", len(preview), len(a.Preview), err)
		}
	})
}

func stringSize(n int) string { return strconv.Itoa(n) }

func TestSafePathsAndMissing(t *testing.T) {
	r, _ := newRepository(t)
	ctx := context.Background()
	for _, bad := range []string{"", ".", "..", "../x", "a/b", "A", "-a", "a-", "a_b", strings.Repeat("a", 64)} {
		if _, err := r.Create(ctx, bad, "data", "o", strings.NewReader("{}")); !errors.Is(err, ErrInvalid) {
			t.Errorf("namespace %q: %v", bad, err)
		}
		if _, err := r.Create(ctx, "ns", bad, "o", strings.NewReader("{}")); !errors.Is(err, ErrInvalid) {
			t.Errorf("name %q: %v", bad, err)
		}
		if _, err := r.List(ctx, bad); !errors.Is(err, ErrInvalid) {
			t.Errorf("list %q: %v", bad, err)
		}
		if _, err := r.Get(ctx, bad, "data", strings.Repeat("a", 64)); !errors.Is(err, ErrInvalid) {
			t.Errorf("get namespace %q: %v", bad, err)
		}
		if _, err := r.Get(ctx, "ns", bad, strings.Repeat("a", 64)); !errors.Is(err, ErrInvalid) {
			t.Errorf("get name %q: %v", bad, err)
		}
	}
	for _, bad := range []string{"", "../x", strings.Repeat("g", 64), strings.Repeat("a", 63), strings.Repeat("a", 65), strings.Repeat("A", 64)} {
		if _, err := r.Get(ctx, "ns", "data", bad); !errors.Is(err, ErrInvalid) {
			t.Errorf("version %q: %v", bad, err)
		}
	}
	if _, err := r.Get(ctx, "ns", "missing", strings.Repeat("a", 64)); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing: %v", err)
	}
	if list, err := r.List(ctx, "missing"); err != nil || list == nil || len(list) != 0 {
		t.Fatalf("empty list: %v %v", list, err)
	}
	if _, err := r.Create(ctx, "1", strings.Repeat("a", 63), "o", strings.NewReader("{}")); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Create(ctx, "ns", "nil", "o", nil); !errors.Is(err, ErrInvalid) {
		t.Fatalf("nil reader: %v", err)
	}
}

func TestDuplicateConcurrentAndNoOverwrite(t *testing.T) {
	r, root := newRepository(t)
	ctx := context.Background()
	var wg sync.WaitGroup
	errs := make(chan error, 12)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := r.Create(ctx, "ns", "data", "o", strings.NewReader("{}\n"))
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	success := 0
	for err := range errs {
		if err == nil {
			success++
		} else if !errors.Is(err, ErrConflict) {
			t.Fatal(err)
		}
	}
	if success != 1 {
		t.Fatalf("successes: %d", success)
	}
	list, err := r.List(ctx, "ns")
	if err != nil || len(list) != 1 {
		t.Fatalf("list: %v %v", list, err)
	}
	entries, err := os.ReadDir(filepath.Join(root, "ns", "data"))
	if err != nil || len(entries) != 1 {
		t.Fatalf("temporary files: %v %v", entries, err)
	}
	a := list[0]
	if _, err := r.Create(ctx, "ns", "data", "different", strings.NewReader("{}\n")); !errors.Is(err, ErrConflict) {
		t.Fatal(err)
	}
	got, err := r.Get(ctx, "ns", "data", a.Version)
	if err != nil || got.Owner != "o" {
		t.Fatalf("overwritten: %+v %v", got, err)
	}
	hash := sha256.Sum256([]byte("{}"))
	versionDir := filepath.Join(root, "ns", "data", hex.EncodeToString(hash[:]))
	if err := os.Mkdir(versionDir, 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Create(ctx, "ns", "data", "o", strings.NewReader("{}")); !errors.Is(err, ErrConflict) {
		t.Fatalf("empty destination replaced: %v", err)
	}
	entries, err = os.ReadDir(versionDir)
	if err != nil || len(entries) != 0 {
		t.Fatalf("destination modified: %v %v", entries, err)
	}
}

func TestListSortAndNamespaceIsolation(t *testing.T) {
	r, _ := newRepository(t)
	ctx := context.Background()
	for _, name := range []string{"z", "a", "a"} {
		data := "{}"
		if name == "a" {
			data = `{"a":1}`
		}
		if _, err := r.Create(ctx, "ns", name, "o", strings.NewReader(data)); err != nil && !errors.Is(err, ErrConflict) {
			t.Fatal(err)
		}
	}
	if _, err := r.Create(ctx, "ns", "a", "o", strings.NewReader("{}")); err != nil {
		t.Fatal(err)
	}
	other, err := r.Create(ctx, "other", "private", "o", strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	list, err := r.List(ctx, "ns")
	if err != nil || len(list) != 3 {
		t.Fatalf("list: %v %v", list, err)
	}
	if list[0].Name != "a" || list[1].Name != "a" || list[2].Name != "z" || list[0].Version >= list[1].Version {
		t.Fatalf("unsorted: %+v", list)
	}
	if _, err := r.Get(ctx, "ns", other.Name, other.Version); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
}

func TestCorruptMetadataIsReported(t *testing.T) {
	for _, mutation := range []string{"syntax", "namespace", "name", "version", "sha256", "trailing", "missing"} {
		t.Run(mutation, func(t *testing.T) {
			r, root := newRepository(t)
			a, err := r.Create(context.Background(), "ns", "data", "o", strings.NewReader("{}"))
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(root, "ns", "data", a.Version, "metadata.json")
			var fields map[string]any
			b, _ := json.Marshal(a)
			if err := json.Unmarshal(b, &fields); err != nil {
				t.Fatal(err)
			}
			switch mutation {
			case "syntax":
				b = []byte("{")
			case "trailing":
				b = append(b, []byte(" {}")...)
			case "missing":
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
			default:
				fields[mutation] = "wrong"
				b, _ = json.Marshal(fields)
			}
			if mutation != "missing" {
				if err := os.WriteFile(path, b, 0600); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := r.Get(context.Background(), "ns", "data", a.Version); err == nil {
				t.Fatal("Get accepted corrupt metadata")
			}
			if _, err := r.List(context.Background(), "ns"); err == nil {
				t.Fatal("List hid corrupt metadata")
			}
		})
	}
}

type cancelReader struct{ cancel context.CancelFunc }

func (r cancelReader) Read(p []byte) (int, error) { r.cancel(); return copy(p, "{}\n"), nil }

type failedReader struct{}

func (failedReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

func TestCancellationReadErrorsAndCleanup(t *testing.T) {
	r, root := newRepository(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := r.Create(ctx, "ns", "data", "o", strings.NewReader("{}")); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := r.List(ctx, "ns"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := r.Get(ctx, "ns", "data", strings.Repeat("a", 64)); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	if _, err := r.Create(ctx, "ns", "data", "o", cancelReader{cancel}); !errors.Is(err, context.Canceled) {
		t.Fatalf("mid-stream cancellation: %v", err)
	}
	if _, err := r.Create(context.Background(), "ns", "data", "o", failedReader{}); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("reader failure: %v", err)
	}
	assertNoFiles(t, root)
}

func TestNewRejectsInvalidRoot(t *testing.T) {
	if _, err := New(""); !errors.Is(err, ErrInvalid) {
		t.Fatalf("empty root: %v", err)
	}
	path := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(path, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := New(path); err == nil {
		t.Fatal("accepted file root")
	}
}

func assertNoFiles(t *testing.T, root string) {
	t.Helper()
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return err
		}
		if !d.IsDir() || strings.HasPrefix(d.Name(), ".upload-") {
			t.Errorf("leftover: %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestNamespaceQuotaConcurrent(t *testing.T) {
	r, root := newRepository(t)
	ctx := context.Background()
	// Seed committed metadata near quota without writing 100 MiB per test run.
	for i := 0; i < 10; i++ {
		a, err := r.Create(ctx, "ns", "seed-"+strconv.Itoa(i), "o", strings.NewReader("{}"))
		if err != nil {
			t.Fatal(err)
		}
		a.Bytes = MaxNamespaceBytes / 10
		if i == 9 {
			a.Bytes -= 3
		}
		b, err := json.Marshal(a)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "ns", a.Name, a.Version, "metadata.json"), b, 0600); err != nil {
			t.Fatal(err)
		}
	}
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for _, name := range []string{"one", "two"} {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			_, err := r.Create(ctx, "ns", name, "o", strings.NewReader("{}\n"))
			errs <- err
		}(name)
	}
	wg.Wait()
	close(errs)
	var successes, quotas int
	for err := range errs {
		if err == nil {
			successes++
		} else if errors.Is(err, ErrQuota) {
			quotas++
		} else {
			t.Fatal(err)
		}
	}
	if successes != 1 || quotas != 1 {
		t.Fatalf("success=%d quota=%d", successes, quotas)
	}
	if _, err := r.Create(ctx, "other", "data", "o", strings.NewReader("{}")); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Create(ctx, "ns", "seed-0", "o", strings.NewReader("{}")); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate at quota: %v", err)
	}
	for _, name := range []string{"one", "two"} {
		entries, err := os.ReadDir(filepath.Join(root, "ns", name))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".upload-") {
				t.Fatal("quota rejection left temporary files")
			}
		}
	}
}

func TestRejectSymlinks(t *testing.T) {
	for _, level := range []string{"namespace", "name", "version", "metadata", "data"} {
		t.Run(level, func(t *testing.T) {
			r, root := newRepository(t)
			ctx := context.Background()
			a, err := r.Create(ctx, "ns", "data", "o", strings.NewReader("{}"))
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(root, "ns")
			switch level {
			case "name":
				path = filepath.Join(path, "data")
			case "version":
				path = filepath.Join(path, "data", a.Version)
			case "metadata":
				path = filepath.Join(path, "data", a.Version, "metadata.json")
			case "data":
				path = filepath.Join(path, "data", a.Version, "data.jsonl")
			}
			outside := filepath.Join(t.TempDir(), "target")
			if err := os.Rename(path, outside); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(outside, path); err != nil {
				t.Fatal(err)
			}
			if _, err := r.Get(ctx, "ns", "data", a.Version); err == nil {
				t.Fatal("Get followed symlink")
			}
			if _, err := r.List(ctx, "ns"); err == nil {
				t.Fatal("List followed or hid symlink")
			}
			if _, err := r.Create(ctx, "ns", "data", "o", strings.NewReader("{}\n")); err == nil {
				t.Fatal("Create accepted symlink hierarchy")
			}
		})
	}
}

func TestNamespaceVersionQuotaConcurrent(t *testing.T) {
	r, root := newRepository(t)
	ctx := context.Background()
	// Build persisted fixtures directly to avoid quadratic setup scans.
	for i := 0; i < MaxNamespaceVersions-1; i++ {
		data := `{"n":` + strconv.Itoa(i) + `}`
		hash := sha256.Sum256([]byte(data))
		version := hex.EncodeToString(hash[:])
		a := Asset{Namespace: "ns", Name: "seed", Owner: "o", Version: version, SHA256: version, Format: "jsonl", Status: "validated", Source: "upload", Bytes: int64(len(data)), Rows: 1, CreatedAt: time.Now().UTC(), Preview: []json.RawMessage{json.RawMessage(data)}}
		dir := filepath.Join(root, "ns", "seed", version)
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
		b, err := json.Marshal(a)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "metadata.json"), b, 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "data.jsonl"), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for _, name := range []string{"one", "two"} {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			_, err := r.Create(ctx, "ns", name, "o", strings.NewReader("{}"))
			errs <- err
		}(name)
	}
	wg.Wait()
	close(errs)
	var successes, quotas int
	for err := range errs {
		if err == nil {
			successes++
		} else if errors.Is(err, ErrQuota) {
			quotas++
		} else {
			t.Fatal(err)
		}
	}
	if successes != 1 || quotas != 1 {
		t.Fatalf("success=%d quota=%d", successes, quotas)
	}
	r, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Create(ctx, "ns", "third", "o", strings.NewReader("{}")); !errors.Is(err, ErrQuota) {
		t.Fatalf("quota after restart: %v", err)
	}
	assertNoFiles(t, filepath.Join(root, "ns", "third"))
	if _, err := os.Lstat(filepath.Join(root, "ns", "third")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("quota rejection left a name directory: %v", err)
	}
	if _, err := r.Create(ctx, "ns", "seed", "o", strings.NewReader(`{"n":0}`)); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate at quota: %v", err)
	}
	if _, err := r.Create(ctx, "other", "data", "o", strings.NewReader("{}")); err != nil {
		t.Fatal(err)
	}
}

func TestInvalidUniqueNamesLeaveNoDirectories(t *testing.T) {
	r, root := newRepository(t)
	for i := 0; i < 20; i++ {
		_, err := r.Create(context.Background(), "ns", "invalid-"+strconv.Itoa(i), "o", strings.NewReader("bad JSON"))
		if !errors.Is(err, ErrInvalid) {
			t.Fatal(err)
		}
	}
	entries, err := os.ReadDir(filepath.Join(root, "ns"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("failed uploads left %d namespace entries", len(entries))
	}
}

func TestPublicationSyncsAncestorsAndReportsFailure(t *testing.T) {
	for _, failAt := range []string{"", "name", "namespace", "root"} {
		t.Run("fail-"+failAt, func(t *testing.T) {
			r, root := newRepository(t)
			root = r.root
			hash := sha256.Sum256([]byte("{}"))
			version := hex.EncodeToString(hash[:])
			name := filepath.Join(root, "ns", "data")
			paths := map[string]string{"name": name, "namespace": filepath.Dir(name), "root": root}
			injected := errors.New("directory sync failed")
			var calls []string
			r.syncDir = func(path string) error {
				calls = append(calls, path)
				if !strings.HasPrefix(filepath.Base(path), ".upload-") {
					if _, err := os.Stat(filepath.Join(name, version, "metadata.json")); err != nil {
						t.Errorf("sync before publication: %s: %v", path, err)
					}
				}
				if path == paths[failAt] {
					return injected
				}
				return syncDirectory(path)
			}
			_, err := r.Create(context.Background(), "ns", "data", "o", strings.NewReader("{}"))
			if failAt != "" {
				if !errors.Is(err, injected) {
					t.Fatalf("lost sync failure: %v", err)
				}
				// Rename committed the version before the durability error; preserve it.
				if _, err := r.Get(context.Background(), "ns", "data", version); err != nil {
					t.Fatal(err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(calls) != 4 || calls[1] != name || calls[2] != filepath.Dir(name) || calls[3] != root {
				t.Fatalf("missing ordered publication barriers: %v", calls)
			}
		})
	}
}

type pausedReader struct {
	started chan struct{}
	release chan struct{}
	data    string
	once    sync.Once
}

func (r *pausedReader) Read(p []byte) (int, error) {
	r.once.Do(func() { close(r.started); <-r.release })
	if r.data == "" {
		return 0, io.EOF
	}
	n := copy(p, r.data)
	r.data = r.data[n:]
	return n, nil
}

func TestStagingIsHiddenAndFailureCannotRemoveConcurrentUpload(t *testing.T) {
	r, root := newRepository(t)
	reader := &pausedReader{started: make(chan struct{}), release: make(chan struct{}), data: "invalid"}
	result := make(chan error, 1)
	go func() { _, err := r.Create(context.Background(), "ns", "same", "o", reader); result <- err }()
	<-reader.started
	defer func() {
		close(reader.release)
		if err := <-result; !errors.Is(err, ErrInvalid) {
			t.Errorf("invalid upload: %v", err)
		}
	}()
	if _, err := os.Lstat(filepath.Join(root, "ns", "same")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("name created before validation: %v", err)
	}
	list, err := r.List(context.Background(), "ns")
	if err != nil || len(list) != 0 {
		t.Fatalf("staging leaked into List: %v %v", list, err)
	}
	a, err := r.Create(context.Background(), "ns", "same", "o", strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := r.Get(context.Background(), "ns", "same", a.Version); err != nil {
			t.Errorf("concurrent upload was removed: %v", err)
		}
	})
}
