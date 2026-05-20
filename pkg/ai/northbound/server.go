package northbound

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/oam-dev/kubevela/pkg/ai/domain"
	domainapply "github.com/oam-dev/kubevela/pkg/ai/domain/apply"
)

const maxRequestBodyBytes = 1 << 20

type Options struct {
	Applier domainapply.ApplicationApplier
}

type errorResponse struct {
	Error string `json:"error"`
}

type DeployResponse struct {
	Normalized  domain.NormalizedObject `json:"normalized"`
	Application domainapply.Result      `json:"application"`
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
	mux.HandleFunc("/api/v1/ai/applications", deploy(options.Applier))
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
