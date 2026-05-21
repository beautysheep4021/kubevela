package northbound

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/oam-dev/kubevela/pkg/ai/domain"
	domainapply "github.com/oam-dev/kubevela/pkg/ai/domain/apply"
	"github.com/oam-dev/kubevela/pkg/ai/observe"
)

const maxRequestBodyBytes = 1 << 20

type Options struct {
	Applier domainapply.ApplicationApplier
	Reader  ApplicationReader
}

type errorResponse struct {
	Error string `json:"error"`
}

type ApplicationReader interface {
	ListApplications(ctx context.Context, namespace string) ([]observe.ApplicationListItem, error)
	SummarizeApplication(ctx context.Context, namespace, name string) (*observe.Summary, error)
	GetApplicationLogs(ctx context.Context, options observe.LogOptions) (*observe.Logs, error)
}

type DeployResponse struct {
	Normalized  domain.NormalizedObject `json:"normalized"`
	Application domainapply.Result      `json:"application"`
}

type ApplicationsResponse struct {
	Items []observe.ApplicationListItem `json:"items"`
}

func NewServer() http.Handler {
	return NewServerWithOptions(Options{})
}

func NewServerWithOptions(options Options) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", console)
	mux.HandleFunc("/healthz", healthz)
	mux.HandleFunc("/api/v1/ai/validate", validate)
	mux.HandleFunc("/api/v1/ai/normalize", normalize)
	mux.HandleFunc("/api/v1/ai/applications", applications(options))
	mux.HandleFunc("/api/v1/ai/applications/", applicationDetail(options.Reader))
	return mux
}

func console(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(consoleHTML))
}

func healthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("ok\n"))
}

func validate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	content, err := readBody(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	result, err := domain.ValidateYAML(content)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(result.Errors) > 0 {
		writeError(w, http.StatusBadRequest, result.Errors[0])
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func normalize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	content, err := readBody(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	normalized, err := domain.NormalizeYAML(content)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, normalized)
}

func applications(options Options) http.HandlerFunc {
	deployHandler := deploy(options.Applier)
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listApplications(options.Reader, w, r)
		case http.MethodPost:
			deployHandler(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	}
}

func listApplications(reader ApplicationReader, w http.ResponseWriter, r *http.Request) {
	if reader == nil {
		writeError(w, http.StatusServiceUnavailable, "application reader is not configured")
		return
	}
	items, err := reader.ListApplications(r.Context(), r.URL.Query().Get("namespace"))
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ApplicationsResponse{Items: items})
}

func applicationDetail(reader ApplicationReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if reader == nil {
			writeError(w, http.StatusServiceUnavailable, "application reader is not configured")
			return
		}
		namespace, name, action, ok := parseApplicationDetailPath(r.URL.Path)
		if !ok {
			http.NotFound(w, r)
			return
		}
		if action == "logs" {
			logs, err := reader.GetApplicationLogs(r.Context(), observe.LogOptions{
				Namespace: namespace,
				Name:      name,
				Pod:       r.URL.Query().Get("pod"),
				Container: r.URL.Query().Get("container"),
				TailLines: parseTailLines(r),
			})
			if err != nil {
				writeError(w, http.StatusBadGateway, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, logs)
			return
		}
		summary, err := reader.SummarizeApplication(r.Context(), namespace, name)
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, summary)
	}
}

func parseApplicationDetailPath(path string) (string, string, string, bool) {
	rest := strings.TrimPrefix(path, "/api/v1/ai/applications/")
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	action := ""
	if len(parts) == 3 && (parts[2] == "status" || parts[2] == "logs") {
		action = parts[2]
		parts = parts[:2]
	}
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", "", false
	}
	namespace, err := url.PathUnescape(parts[0])
	if err != nil {
		return "", "", "", false
	}
	name, err := url.PathUnescape(parts[1])
	if err != nil {
		return "", "", "", false
	}
	return namespace, name, action, true
}

func deploy(applier domainapply.ApplicationApplier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if applier == nil {
			writeError(w, http.StatusServiceUnavailable, "deployment is not configured")
			return
		}
		content, err := readBody(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		normalized, err := domain.NormalizeYAML(content)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		appYAML, err := domain.TranslateYAML(content)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		result, err := domainapply.ApplyYAMLWithApplier(r.Context(), applier, appYAML, domainapply.Options{
			DryRun: parseDryRun(r),
		})
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, DeployResponse{
			Normalized:  normalized,
			Application: result,
		})
	}
}

func parseDryRun(r *http.Request) bool {
	raw := r.URL.Query().Get("dryRun")
	if raw == "" {
		return false
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false
	}
	return value
}

func parseTailLines(r *http.Request) int64 {
	raw := r.URL.Query().Get("tailLines")
	if raw == "" {
		return 200
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return 200
	}
	if value > 2000 {
		return 2000
	}
	return value
}

func readBody(r *http.Request) ([]byte, error) {
	defer r.Body.Close()
	content, err := io.ReadAll(http.MaxBytesReader(nil, r.Body, maxRequestBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("read request body: %w", err)
	}
	if len(content) == 0 {
		return nil, fmt.Errorf("request body is required")
	}
	return content, nil
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}
