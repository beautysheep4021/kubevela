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
	"time"

	"github.com/oam-dev/kubevela/pkg/ai/domain"
	domainapply "github.com/oam-dev/kubevela/pkg/ai/domain/apply"
	"github.com/oam-dev/kubevela/pkg/ai/observe"
)

const maxRequestBodyBytes = 1 << 20

type Options struct {
	Applier domainapply.ApplicationApplier
	Reader  ApplicationReader
	Manager ApplicationManager
	Audits  AuditStore
}

type errorResponse struct {
	Error string `json:"error"`
}

type ApplicationReader interface {
	ListApplications(ctx context.Context, namespace string) ([]observe.ApplicationListItem, error)
	SummarizeApplication(ctx context.Context, namespace, name string) (*observe.Summary, error)
	GetApplicationLogs(ctx context.Context, options observe.LogOptions) (*observe.Logs, error)
}

type ApplicationManager interface {
	DeleteApplication(ctx context.Context, namespace, name string) (*observe.LifecycleResult, error)
	RestartApplication(ctx context.Context, namespace, name string) (*observe.LifecycleResult, error)
	RerunApplication(ctx context.Context, namespace, name string) (*observe.LifecycleResult, error)
}

type AuditStore interface {
	RecordAudit(ctx context.Context, event observe.AuditEvent) error
	ListAudits(ctx context.Context, namespace string) ([]observe.AuditEvent, error)
}

type DeployResponse struct {
	Normalized  domain.NormalizedObject `json:"normalized"`
	Application domainapply.Result      `json:"application"`
}

type ApplicationsResponse struct {
	Items []observe.ApplicationListItem `json:"items"`
}

type DeliveryResult struct {
	ModelURI string                 `json:"modelURI"`
	Metrics  map[string]float64     `json:"metrics,omitempty"`
	Summary  string                 `json:"summary,omitempty"`
	Raw      map[string]interface{} `json:"raw,omitempty"`
}

type DeliveryResultResponse struct {
	Namespace string         `json:"namespace"`
	JobName   string         `json:"jobName"`
	Result    DeliveryResult `json:"result"`
}

type DeliveryPublishRequest struct {
	ServiceName string `json:"serviceName"`
	Image       string `json:"image"`
	Port        int64  `json:"port"`
	ServicePort int64  `json:"servicePort"`
}

type DeliveryPublishResponse struct {
	Namespace   string             `json:"namespace"`
	JobName     string             `json:"jobName"`
	ServiceName string             `json:"serviceName"`
	ModelURI    string             `json:"modelURI"`
	Application domainapply.Result `json:"application"`
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
	mux.HandleFunc("/api/v1/ai/applications/", applicationDetail(options))
	mux.HandleFunc("/api/v1/ai/audits", audits(options.Audits))
	mux.HandleFunc("/api/v1/ai/deliveries/", deliveries(options))
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

func applicationDetail(options Options) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace, name, action, ok := parseApplicationDetailPath(r.URL.Path)
		if !ok {
			http.NotFound(w, r)
			return
		}
		if r.Method == http.MethodDelete && action == "" {
			runLifecycleAction(options, w, r, "delete", namespace, name)
			return
		}
		if r.Method == http.MethodPost && (action == "restart" || action == "rerun") {
			runLifecycleAction(options, w, r, action, namespace, name)
			return
		}
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if options.Reader == nil {
			writeError(w, http.StatusServiceUnavailable, "application reader is not configured")
			return
		}
		if action == "logs" {
			logs, err := options.Reader.GetApplicationLogs(r.Context(), observe.LogOptions{
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
		summary, err := options.Reader.SummarizeApplication(r.Context(), namespace, name)
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
	if len(parts) == 3 && (parts[2] == "status" || parts[2] == "logs" || parts[2] == "restart" || parts[2] == "rerun") {
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

func runLifecycleAction(options Options, w http.ResponseWriter, r *http.Request, action, namespace, name string) {
	if options.Manager == nil {
		writeError(w, http.StatusServiceUnavailable, "application manager is not configured")
		return
	}
	actor := actorFromRequest(r)
	event := observe.AuditEvent{
		ID:        auditID(action, namespace, name),
		Time:      time.Now().UTC().Format(time.RFC3339),
		Namespace: namespace,
		Name:      name,
		Action:    action,
		Actor:     actor,
	}
	var (
		result *observe.LifecycleResult
		err    error
	)
	switch action {
	case "delete":
		result, err = options.Manager.DeleteApplication(r.Context(), namespace, name)
	case "restart":
		result, err = options.Manager.RestartApplication(r.Context(), namespace, name)
	case "rerun":
		result, err = options.Manager.RerunApplication(r.Context(), namespace, name)
	default:
		http.NotFound(w, r)
		return
	}
	if err != nil {
		event.Success = false
		event.Error = err.Error()
		recordAudit(r.Context(), options.Audits, event)
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	event.Success = true
	if result != nil {
		event.Message = result.Message
	}
	recordAudit(r.Context(), options.Audits, event)
	writeJSON(w, http.StatusOK, result)
}

func audits(store AuditStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if store == nil {
			writeError(w, http.StatusServiceUnavailable, "audit store is not configured")
			return
		}
		items, err := store.ListAudits(r.Context(), r.URL.Query().Get("namespace"))
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, observe.AuditList{Items: items})
	}
}

func deliveries(options Options) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace, jobName, action, ok := parseDeliveryPath(r.URL.Path)
		if !ok {
			http.NotFound(w, r)
			return
		}
		switch {
		case r.Method == http.MethodGet && action == "result":
			result, err := deliveryResultFromJobLogs(r.Context(), options.Reader, namespace, jobName)
			if err != nil {
				writeError(w, http.StatusBadGateway, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, DeliveryResultResponse{
				Namespace: namespace,
				JobName:   jobName,
				Result:    result,
			})
		case r.Method == http.MethodPost && action == "publish-service":
			publishDeliveryService(options, w, r, namespace, jobName)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	}
}

func parseDeliveryPath(path string) (string, string, string, bool) {
	rest := strings.TrimPrefix(path, "/api/v1/ai/deliveries/")
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
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
	return namespace, name, parts[2], true
}

func deliveryResultFromJobLogs(ctx context.Context, reader ApplicationReader, namespace, jobName string) (DeliveryResult, error) {
	if reader == nil {
		return DeliveryResult{}, fmt.Errorf("application reader is not configured")
	}
	logs, err := reader.GetApplicationLogs(ctx, observe.LogOptions{
		Namespace: namespace,
		Name:      jobName,
		TailLines: 500,
	})
	if err != nil {
		return DeliveryResult{}, err
	}
	return parseDeliveryResult(logs.Logs)
}

func parseDeliveryResult(logs string) (DeliveryResult, error) {
	const marker = "AI_RESULT_JSON="
	lines := strings.Split(logs, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(line, marker) {
			continue
		}
		rawJSON := strings.TrimSpace(strings.TrimPrefix(line, marker))
		var raw map[string]interface{}
		if err := json.Unmarshal([]byte(rawJSON), &raw); err != nil {
			return DeliveryResult{}, fmt.Errorf("decode AI_RESULT_JSON: %w", err)
		}
		modelURI, _ := raw["modelURI"].(string)
		if modelURI == "" {
			return DeliveryResult{}, fmt.Errorf("AI_RESULT_JSON.modelURI is required")
		}
		result := DeliveryResult{
			ModelURI: modelURI,
			Metrics:  map[string]float64{},
			Raw:      raw,
		}
		if summary, ok := raw["summary"].(string); ok {
			result.Summary = summary
		}
		if metrics, ok := raw["metrics"].(map[string]interface{}); ok {
			for key, value := range metrics {
				if number, ok := value.(float64); ok {
					result.Metrics[key] = number
				}
			}
		}
		if len(result.Metrics) == 0 {
			result.Metrics = nil
		}
		return result, nil
	}
	return DeliveryResult{}, fmt.Errorf("AI_RESULT_JSON marker was not found in job logs")
}

func publishDeliveryService(options Options, w http.ResponseWriter, r *http.Request, namespace, jobName string) {
	if options.Applier == nil {
		writeError(w, http.StatusServiceUnavailable, "deployment is not configured")
		return
	}
	result, err := deliveryResultFromJobLogs(r.Context(), options.Reader, namespace, jobName)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	var request DeliveryPublishRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("decode publish request: %v", err))
		return
	}
	if request.ServiceName == "" {
		request.ServiceName = jobName + "-service"
	}
	if request.Image == "" {
		request.Image = "python:3.11-slim"
	}
	if request.Port == 0 {
		request.Port = 8080
	}
	if request.ServicePort == 0 {
		request.ServicePort = 80
	}
	domainYAML := deliveryServiceYAML(namespace, jobName, request, result)
	appYAML, err := domain.TranslateYAML([]byte(domainYAML))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	applied, err := domainapply.ApplyYAMLWithApplier(r.Context(), options.Applier, appYAML, domainapply.Options{})
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	recordAudit(r.Context(), options.Audits, observe.AuditEvent{
		ID:        auditID("publish-service", namespace, request.ServiceName),
		Time:      time.Now().UTC().Format(time.RFC3339),
		Namespace: namespace,
		Name:      request.ServiceName,
		Action:    "publish-service",
		Actor:     actorFromRequest(r),
		Success:   true,
		Message:   "published service from " + jobName,
	})
	writeJSON(w, http.StatusOK, DeliveryPublishResponse{
		Namespace:   namespace,
		JobName:     jobName,
		ServiceName: request.ServiceName,
		ModelURI:    result.ModelURI,
		Application: applied,
	})
}

func deliveryServiceYAML(namespace, jobName string, request DeliveryPublishRequest, result DeliveryResult) string {
	return fmt.Sprintf(`apiVersion: ai.oam.dev/v1alpha1
kind: AIService
metadata:
  name: %s
  namespace: %s
spec:
  componentName: %s
  properties:
    image: %s
    replicas: 1
    model:
      name: %s
      uri: %s
    endpoint:
      port: %d
      servicePort: %d
      type: ClusterIP
  runtime:
    runtime: http
    framework: delivery-poc
    tenant: demo-tenant
    project: delivery
    environment: poc
    owner: ai-platform
    modelURI: %s
  placement:
    namespace: %s
    clusters:
      - local
`, request.ServiceName, namespace, request.ServiceName, request.Image, jobName, result.ModelURI, request.Port, request.ServicePort, result.ModelURI, namespace)
}

func actorFromRequest(r *http.Request) string {
	for _, header := range []string{"X-AI-User", "X-User", "X-Forwarded-User"} {
		if value := strings.TrimSpace(r.Header.Get(header)); value != "" {
			return value
		}
	}
	return "anonymous"
}

func auditID(action, namespace, name string) string {
	value := strings.Join([]string{time.Now().UTC().Format("20060102T150405.000000000Z"), action, namespace, name}, "-")
	return strings.NewReplacer("/", "-", " ", "-").Replace(value)
}

func recordAudit(ctx context.Context, store AuditStore, event observe.AuditEvent) {
	if store == nil {
		return
	}
	_ = store.RecordAudit(ctx, event)
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
