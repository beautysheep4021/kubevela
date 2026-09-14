package northbound

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/oam-dev/kubevela/pkg/ai/datasetstorage"
)

type DatasetUploadStore interface {
	Create(context.Context, string, string, string, io.Reader) (datasetstorage.Asset, error)
	List(context.Context, string) ([]datasetstorage.Asset, error)
	Get(context.Context, string, string, string) (datasetstorage.Asset, error)
}

func datasetUploads(store DatasetUploadStore) http.HandlerFunc {
	uploadSlots := make(chan struct{}, 2)
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		session, ok := requestSession(r)
		if !ok || session.Role != consoleRoleUser || session.Namespace == "" {
			writeError(w, http.StatusForbidden, "a tenant user session is required")
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if store == nil {
			writeError(w, http.StatusServiceUnavailable, "dataset upload storage is not configured")
			return
		}
		const base = "/api/v1/ai/dataset-uploads"
		if r.URL.Path != base {
			parts := strings.Split(strings.TrimPrefix(r.URL.Path, base+"/"), "/")
			if len(parts) != 3 || r.Method != http.MethodGet {
				writeError(w, http.StatusNotFound, "dataset upload route not found")
				return
			}
			if parts[0] != session.Namespace {
				writeError(w, http.StatusForbidden, "dataset is outside account scope")
				return
			}
			asset, err := store.Get(r.Context(), parts[0], parts[1], parts[2])
			if err != nil {
				writeDatasetUploadError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, asset)
			return
		}
		if r.Method == http.MethodGet {
			requested := r.URL.Query().Get("namespace")
			if requested != "" && requested != session.Namespace {
				writeError(w, http.StatusForbidden, "dataset is outside account scope")
				return
			}
			items, err := store.List(r.Context(), session.Namespace)
			if err != nil {
				writeDatasetUploadError(w, err)
				return
			}
			if items == nil {
				items = []datasetstorage.Asset{}
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"items": items})
			return
		}
		// Multipart uploads are form-compatible, so reject cross-origin browser writes.
		if origin := r.Header.Get("Origin"); origin != "" {
			u, err := url.Parse(origin)
			if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host != r.Host || u.User != nil {
				writeError(w, http.StatusForbidden, "cross-origin upload is not allowed")
				return
			}
		}
		if r.Header.Get("Sec-Fetch-Site") == "cross-site" {
			writeError(w, http.StatusForbidden, "cross-site upload is not allowed")
			return
		}
		select {
		case uploadSlots <- struct{}{}:
			defer func() { <-uploadSlots }()
		default:
			writeError(w, http.StatusTooManyRequests, "too many concurrent uploads; retry later")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, (10<<20)+(64<<10))
		if err := r.ParseMultipartForm(512 << 10); err != nil {
			if r.MultipartForm != nil {
				_ = r.MultipartForm.RemoveAll()
			}
			var limit *http.MaxBytesError
			if errors.As(err, &limit) {
				writeError(w, http.StatusRequestEntityTooLarge, "upload exceeds 10 MiB limit")
			} else {
				writeError(w, http.StatusBadRequest, "invalid multipart upload")
			}
			return
		}
		defer r.MultipartForm.RemoveAll()
		values := r.MultipartForm.Value
		if len(values["namespace"]) != 1 || values["namespace"][0] != session.Namespace {
			writeError(w, http.StatusForbidden, "dataset is outside account scope")
			return
		}
		if len(values) != 2 || len(values["name"]) != 1 || len(r.MultipartForm.File) != 1 || len(r.MultipartForm.File["file"]) != 1 {
			writeError(w, http.StatusBadRequest, "expected namespace, name and exactly one file")
			return
		}
		file, err := r.MultipartForm.File["file"][0].Open()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "cannot read uploaded file")
			return
		}
		defer file.Close()
		asset, err := store.Create(r.Context(), session.Namespace, values["name"][0], session.Username, file)
		if err != nil {
			writeDatasetUploadError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, asset)
	}
}

func writeDatasetUploadError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, datasetstorage.ErrInvalid):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, datasetstorage.ErrConflict):
		writeError(w, http.StatusConflict, "this dataset version already exists")
	case errors.Is(err, datasetstorage.ErrQuota):
		writeError(w, http.StatusConflict, "dataset storage quota exceeded (100 MiB or 1000 versions per namespace)")
	case errors.Is(err, datasetstorage.ErrNotFound):
		writeError(w, http.StatusNotFound, "dataset version not found")
	default:
		writeError(w, http.StatusInternalServerError, "dataset storage operation failed")
	}
}
