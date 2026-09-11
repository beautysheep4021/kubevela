package northbound

import (
	"context"
	"encoding/json"
	"errors"
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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

const maxRequestBodyBytes = 1 << 20

type Options struct {
	TenantResources TenantResourceReader
	Applier         domainapply.ApplicationApplier
	Reader          ApplicationReader
	Manager         ApplicationManager
	Prober          ApplicationProber
	Artifacts       ArtifactStore
	Datasets        DatasetStore
	Audits          AuditStore
	Auth            *authConfig
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

type ApplicationProber interface {
	ProbeApplication(ctx context.Context, options observe.ProbeOptions) (*observe.ProbeResult, error)
}

type AuditStore interface {
	RecordAudit(ctx context.Context, event observe.AuditEvent) error
	ListAudits(ctx context.Context, namespace string) ([]observe.AuditEvent, error)
}

type ArtifactStore interface {
	SaveArtifact(ctx context.Context, artifact observe.ModelArtifact) error
	ListArtifacts(ctx context.Context, namespace string) ([]observe.ModelArtifact, error)
}

type DatasetStore interface {
	SaveDataset(ctx context.Context, dataset observe.DatasetArtifact) error
	ListDatasets(ctx context.Context, namespace string) ([]observe.DatasetArtifact, error)
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

type ModelEvaluationRequest struct {
	EvaluationDatasetURI string  `json:"evaluationDatasetURI"`
	EvaluationType       string  `json:"evaluationType"`
	PassThreshold        float64 `json:"passThreshold"`
	Image                string  `json:"image"`
	JobName              string  `json:"jobName"`
}

type ModelEvaluationResponse struct {
	Namespace   string             `json:"namespace"`
	ModelName   string             `json:"modelName"`
	JobName     string             `json:"jobName"`
	ModelURI    string             `json:"modelURI"`
	Application domainapply.Result `json:"application"`
}

type EvaluationSyncRequest struct {
	EvaluationJobName string `json:"evaluationJobName"`
}

type EvaluationSyncResponse struct {
	Namespace string                `json:"namespace"`
	ModelName string                `json:"modelName"`
	JobName   string                `json:"jobName"`
	Result    DeliveryResult        `json:"result"`
	Asset     observe.ModelArtifact `json:"asset"`
}

func NewServer() http.Handler {
	return NewServerWithOptions(Options{})
}

func NewServerWithOptions(options Options) http.Handler {
	auth := newConsoleAuthWithConfig(options.Auth)
	mux := http.NewServeMux()
	mux.HandleFunc("/", auth.handleRoot)
	mux.HandleFunc("/login", auth.handleLogin)
	mux.HandleFunc("/logout", auth.handleLogout)
	mux.HandleFunc("/user", auth.handleConsolePage(consoleRoleUser))
	mux.HandleFunc("/monitor", auth.handleConsolePage(consoleRoleMonitor))
	mux.HandleFunc("/healthz", healthz)
	mux.HandleFunc("/console/", auth.requireSession(serveConsoleAsset))
	mux.HandleFunc("/api/v1/ai/session", auth.requireSession(currentConsoleSession))
	mux.HandleFunc("/api/v1/ai/validate", auth.requireSession(validate))
	mux.HandleFunc("/api/v1/ai/normalize", auth.requireSession(normalize))
	mux.HandleFunc("/api/v1/ai/applications", auth.requireSession(applications(options)))
	mux.HandleFunc("/api/v1/ai/applications/", auth.requireSession(applicationDetail(options)))
	mux.HandleFunc("/api/v1/ai/audits", auth.requireSession(audits(options.Audits)))
	mux.HandleFunc("/api/v1/ai/tenant-resources", auth.requireSession(tenantResources(options.TenantResources)))
	mux.HandleFunc("/api/v1/ai/artifacts", auth.requireSession(artifacts(options.Artifacts)))
	mux.HandleFunc("/api/v1/ai/models", auth.requireSession(models(options.Artifacts)))
	mux.HandleFunc("/api/v1/ai/models/", auth.requireSession(modelDetail(options)))
	mux.HandleFunc("/api/v1/ai/datasets", auth.requireSession(datasets(options.Datasets)))
	mux.HandleFunc("/api/v1/ai/deliveries/", auth.requireSession(deliveries(options)))
	return mux
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
	namespace, err := namespaceForListRequest(r, r.URL.Query().Get("namespace"))
	if err != nil {
		writeError(w, http.StatusForbidden, err.Error())
		return
	}
	items, err := reader.ListApplications(r.Context(), namespace)
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
		if err := authorizeNamespaceRequest(r, namespace); err != nil {
			writeError(w, http.StatusForbidden, err.Error())
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
		if r.Method == http.MethodPost && action == "probe" {
			probeApplication(options.Prober, w, r, namespace, name)
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
			statusCode := http.StatusBadGateway
			var status apierrors.APIStatus
			if apierrors.IsNotFound(err) && errors.As(err, &status) {
				details := status.Status().Details
				if details != nil && details.Group == "core.oam.dev" && details.Kind == "applications" && details.Name == name {
					statusCode = http.StatusNotFound
				}
			}
			writeError(w, statusCode, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, summary)
	}
}

func parseApplicationDetailPath(path string) (string, string, string, bool) {
	rest := strings.TrimPrefix(path, "/api/v1/ai/applications/")
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	action := ""
	if len(parts) == 3 && (parts[2] == "status" || parts[2] == "logs" || parts[2] == "restart" || parts[2] == "rerun" || parts[2] == "probe") {
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

type probeRequest struct {
	Path           string `json:"path"`
	TimeoutSeconds int64  `json:"timeoutSeconds"`
}

func probeApplication(prober ApplicationProber, w http.ResponseWriter, r *http.Request, namespace, name string) {
	if prober == nil {
		writeError(w, http.StatusServiceUnavailable, "application prober is not configured")
		return
	}
	var request probeRequest
	if r.Body != nil {
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)).Decode(&request); err != nil && err != io.EOF {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("decode probe request: %v", err))
			return
		}
	}
	result, err := prober.ProbeApplication(r.Context(), observe.ProbeOptions{
		Namespace:      namespace,
		Name:           name,
		Path:           request.Path,
		TimeoutSeconds: request.TimeoutSeconds,
	})
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
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
		namespace, err := namespaceForListRequest(r, r.URL.Query().Get("namespace"))
		if err != nil {
			writeError(w, http.StatusForbidden, err.Error())
			return
		}
		items, err := store.ListAudits(r.Context(), namespace)
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, observe.AuditList{Items: items})
	}
}

func artifacts(store ArtifactStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if store == nil {
			writeError(w, http.StatusServiceUnavailable, "artifact store is not configured")
			return
		}
		namespace, err := namespaceForListRequest(r, r.URL.Query().Get("namespace"))
		if err != nil {
			writeError(w, http.StatusForbidden, err.Error())
			return
		}
		items, err := store.ListArtifacts(r.Context(), namespace)
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, observe.ArtifactList{Items: items})
	}
}

func models(store ArtifactStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if store == nil {
			writeError(w, http.StatusServiceUnavailable, "model asset store is not configured")
			return
		}
		switch r.Method {
		case http.MethodGet:
			namespace, err := namespaceForListRequest(r, r.URL.Query().Get("namespace"))
			if err != nil {
				writeError(w, http.StatusForbidden, err.Error())
				return
			}
			items, err := store.ListArtifacts(r.Context(), namespace)
			if err != nil {
				writeError(w, http.StatusBadGateway, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, observe.ArtifactList{Items: items})
		case http.MethodPost:
			var asset observe.ModelArtifact
			if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)).Decode(&asset); err != nil {
				writeError(w, http.StatusBadRequest, fmt.Sprintf("decode model asset request: %v", err))
				return
			}
			if asset.Namespace == "" {
				writeError(w, http.StatusBadRequest, "namespace is required")
				return
			}
			if err := authorizeNamespaceRequest(r, asset.Namespace); err != nil {
				writeError(w, http.StatusForbidden, err.Error())
				return
			}
			if asset.ModelURI == "" {
				writeError(w, http.StatusBadRequest, "modelURI is required")
				return
			}
			if asset.Name == "" {
				asset.Name = modelAssetNameFromURI(asset.ModelURI)
			}
			if asset.JobName == "" {
				asset.JobName = asset.Name
			}
			if asset.Status == "" {
				asset.Status = "trained"
			}
			if asset.Visibility == "" {
				asset.Visibility = "private"
			}
			if asset.EvaluationStatus == "" {
				asset.EvaluationStatus = "pending"
			}
			if asset.Owner == "" {
				asset.Owner = actorFromRequest(r)
			}
			if asset.CreatedAt == "" {
				asset.CreatedAt = time.Now().UTC().Format(time.RFC3339)
			}
			if err := store.SaveArtifact(r.Context(), asset); err != nil {
				writeError(w, http.StatusBadGateway, err.Error())
				return
			}
			writeJSON(w, http.StatusCreated, asset)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	}
}

func datasets(store DatasetStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if store == nil {
			writeError(w, http.StatusServiceUnavailable, "dataset asset store is not configured")
			return
		}
		switch r.Method {
		case http.MethodGet:
			namespace, err := namespaceForListRequest(r, r.URL.Query().Get("namespace"))
			if err != nil {
				writeError(w, http.StatusForbidden, err.Error())
				return
			}
			items, err := store.ListDatasets(r.Context(), namespace)
			if err != nil {
				writeError(w, http.StatusBadGateway, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, observe.DatasetList{Items: items})
		case http.MethodPost:
			var dataset observe.DatasetArtifact
			if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)).Decode(&dataset); err != nil {
				writeError(w, http.StatusBadRequest, fmt.Sprintf("decode dataset asset request: %v", err))
				return
			}
			if dataset.Namespace == "" {
				writeError(w, http.StatusBadRequest, "namespace is required")
				return
			}
			if err := authorizeNamespaceRequest(r, dataset.Namespace); err != nil {
				writeError(w, http.StatusForbidden, err.Error())
				return
			}
			if dataset.DatasetURI == "" {
				writeError(w, http.StatusBadRequest, "datasetURI is required")
				return
			}
			if dataset.Name == "" {
				dataset.Name = datasetAssetNameFromURI(dataset.DatasetURI)
			}
			if dataset.DisplayName == "" {
				dataset.DisplayName = dataset.Name
			}
			if dataset.Status == "" {
				dataset.Status = "registered"
			}
			if dataset.Visibility == "" {
				dataset.Visibility = "private"
			}
			if dataset.Owner == "" {
				dataset.Owner = actorFromRequest(r)
			}
			if dataset.CreatedAt == "" {
				dataset.CreatedAt = time.Now().UTC().Format(time.RFC3339)
			}
			if err := store.SaveDataset(r.Context(), dataset); err != nil {
				writeError(w, http.StatusBadGateway, err.Error())
				return
			}
			writeJSON(w, http.StatusCreated, dataset)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	}
}

func datasetAssetNameFromURI(datasetURI string) string {
	clean := strings.TrimRight(strings.Split(datasetURI, "?")[0], "/")
	parts := strings.Split(clean, "/")
	name := "dataset"
	if len(parts) > 0 && strings.TrimSpace(parts[len(parts)-1]) != "" {
		name = parts[len(parts)-1]
	}
	name = strings.ToLower(strings.NewReplacer("_", "-", ".", "-", ":", "-", "/", "-").Replace(name))
	name = strings.Trim(name, "-")
	if name == "" {
		return "dataset"
	}
	if len(name) > 50 {
		name = name[:50]
	}
	return name
}

func modelAssetNameFromURI(modelURI string) string {
	clean := strings.TrimRight(strings.Split(modelURI, "?")[0], "/")
	parts := strings.Split(clean, "/")
	name := "model"
	if len(parts) > 0 && strings.TrimSpace(parts[len(parts)-1]) != "" {
		name = parts[len(parts)-1]
	}
	if looksLikeModelVersion(name) && len(parts) > 1 && strings.TrimSpace(parts[len(parts)-2]) != "" {
		name = parts[len(parts)-2]
	}
	name = strings.ToLower(strings.NewReplacer("_", "-", ".", "-", ":", "-", "/", "-").Replace(name))
	name = strings.Trim(name, "-")
	if name == "" {
		return "model"
	}
	if len(name) > 50 {
		name = name[:50]
	}
	return name
}

func looksLikeModelVersion(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "latest" {
		return true
	}
	if len(value) < 2 || value[0] != 'v' {
		return false
	}
	for _, char := range value[1:] {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func modelDetail(options Options) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace, name, action, ok := parseModelDetailPath(r.URL.Path)
		if !ok {
			http.NotFound(w, r)
			return
		}
		if err := authorizeNamespaceRequest(r, namespace); err != nil {
			writeError(w, http.StatusForbidden, err.Error())
			return
		}
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		switch action {
		case "evaluate":
			startModelEvaluation(options, w, r, namespace, name)
		case "sync-evaluation":
			syncModelEvaluation(options, w, r, namespace, name)
		default:
			http.NotFound(w, r)
		}
	}
}

func parseModelDetailPath(path string) (string, string, string, bool) {
	rest := strings.TrimPrefix(path, "/api/v1/ai/models/")
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

func startModelEvaluation(options Options, w http.ResponseWriter, r *http.Request, namespace, name string) {
	if options.Applier == nil {
		writeError(w, http.StatusServiceUnavailable, "deployment is not configured")
		return
	}
	asset, err := findModelAsset(r.Context(), options.Artifacts, namespace, name)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	var request ModelEvaluationRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("decode evaluation request: %v", err))
		return
	}
	if request.EvaluationDatasetURI == "" {
		writeError(w, http.StatusBadRequest, "evaluationDatasetURI is required")
		return
	}
	if request.EvaluationType == "" {
		request.EvaluationType = "accuracy"
	}
	if request.PassThreshold == 0 {
		request.PassThreshold = 0.8
	}
	if request.Image == "" {
		request.Image = "busybox:1.36"
	}
	if request.JobName == "" {
		request.JobName = name + "-eval"
	}
	domainYAML := modelEvaluationYAML(namespace, request.JobName, asset, request)
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
	writeJSON(w, http.StatusOK, ModelEvaluationResponse{
		Namespace:   namespace,
		ModelName:   name,
		JobName:     request.JobName,
		ModelURI:    asset.ModelURI,
		Application: applied,
	})
}

func syncModelEvaluation(options Options, w http.ResponseWriter, r *http.Request, namespace, name string) {
	asset, err := findModelAsset(r.Context(), options.Artifacts, namespace, name)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	var request EvaluationSyncRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("decode evaluation sync request: %v", err))
		return
	}
	if request.EvaluationJobName == "" {
		request.EvaluationJobName = name + "-eval"
	}
	result, logs, err := deliveryResultAndLogsFromJobLogs(r.Context(), options.Reader, namespace, request.EvaluationJobName)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if err := validateEvaluationModelAssociation(asset.ModelURI, result.ModelURI); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	asset.Metrics = result.Metrics
	asset.Summary = result.Summary
	asset.Status = "evaluated"
	asset.EvaluationStatus = evaluationStatus(result)
	asset.SourcePod = logPodName(logs)
	asset.SourceApplication = request.EvaluationJobName
	asset.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := options.Artifacts.SaveArtifact(r.Context(), asset); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, EvaluationSyncResponse{
		Namespace: namespace,
		ModelName: name,
		JobName:   request.EvaluationJobName,
		Result:    result,
		Asset:     asset,
	})
}

func findModelAsset(ctx context.Context, store ArtifactStore, namespace, name string) (observe.ModelArtifact, error) {
	if store == nil {
		return observe.ModelArtifact{}, fmt.Errorf("model asset store is not configured")
	}
	items, err := store.ListArtifacts(ctx, namespace)
	if err != nil {
		return observe.ModelArtifact{}, err
	}
	for _, item := range items {
		if item.Name == name {
			return item, nil
		}
	}
	return observe.ModelArtifact{}, fmt.Errorf("model asset %s/%s was not found", namespace, name)
}

func modelEvaluationYAML(namespace, jobName string, asset observe.ModelArtifact, request ModelEvaluationRequest) string {
	return fmt.Sprintf(`apiVersion: ai.oam.dev/v1alpha1
kind: AIJob
metadata:
  name: %s
  namespace: %s
spec:
  componentName: %s
  properties:
    image: %s
    imagePullPolicy: IfNotPresent
    jobKind: evaluation
    cmd:
      - sh
      - -c
    args:
      - |
        echo evaluation-start
        echo model_uri=%s
        echo evaluation_dataset=%s
        echo evaluation_type=%s
        echo pass_threshold=%g
        echo metric accuracy=0.91
        echo 'AI_RESULT_JSON={"modelURI":"%s","metrics":{"accuracy":0.91},"summary":"evaluation-passed"}'
        echo evaluation-complete
    dataset:
      name: evaluation-dataset
      uri: %s
    output:
      uri: inline://outputs/%s
    backoffLimit: 0
    ttlSecondsAfterFinished: 3600
  runtime:
    runtime: batch
    framework: evaluation-poc
    tenant: demo-tenant
    project: model-evaluation
    environment: poc
    owner: ai-platform
    modelURI: %s
    datasetURI: %s
  placement:
    namespace: %s
    clusters:
      - local
`, jobName, namespace, jobName, request.Image, asset.ModelURI, request.EvaluationDatasetURI, request.EvaluationType, request.PassThreshold, asset.ModelURI, request.EvaluationDatasetURI, jobName, asset.ModelURI, request.EvaluationDatasetURI, namespace)
}

func deliveries(options Options) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace, jobName, action, ok := parseDeliveryPath(r.URL.Path)
		if !ok {
			http.NotFound(w, r)
			return
		}
		if err := authorizeNamespaceRequest(r, namespace); err != nil {
			writeError(w, http.StatusForbidden, err.Error())
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
	result, _, err := deliveryResultAndLogsFromJobLogs(ctx, reader, namespace, jobName)
	return result, err
}

func deliveryResultAndLogsFromJobLogs(ctx context.Context, reader ApplicationReader, namespace, jobName string) (DeliveryResult, *observe.Logs, error) {
	if reader == nil {
		return DeliveryResult{}, nil, fmt.Errorf("application reader is not configured")
	}
	logs, err := reader.GetApplicationLogs(ctx, observe.LogOptions{
		Namespace: namespace,
		Name:      jobName,
		TailLines: 500,
	})
	if err != nil {
		return DeliveryResult{}, nil, err
	}
	result, err := parseDeliveryResult(logs.Logs)
	if err != nil {
		return DeliveryResult{}, logs, err
	}
	return result, logs, nil
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
	result, logs, err := deliveryResultAndLogsFromJobLogs(r.Context(), options.Reader, namespace, jobName)
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
	// This preflight is not atomic with apply; a concurrent creator can still race it.
	_, lookupErr := options.Reader.SummarizeApplication(r.Context(), namespace, request.ServiceName)
	if lookupErr == nil {
		writeError(w, http.StatusConflict, fmt.Sprintf("application %s/%s already exists", namespace, request.ServiceName))
		return
	}
	// Summary also reads related resources. Only a missing target Application
	// permits publication, not a NotFound error from those additional reads.
	var status apierrors.APIStatus
	missingTarget := false
	if apierrors.IsNotFound(lookupErr) && errors.As(lookupErr, &status) {
		details := status.Status().Details
		missingTarget = details != nil && details.Group == "core.oam.dev" && details.Kind == "applications" && details.Name == request.ServiceName
	}
	if !missingTarget {
		writeError(w, http.StatusBadGateway, fmt.Sprintf("check publish target: %v", lookupErr))
		return
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
	recordArtifact(r.Context(), options.Artifacts, observe.ModelArtifact{
		Namespace:         namespace,
		Name:              jobName,
		JobName:           jobName,
		ModelURI:          result.ModelURI,
		Status:            "published",
		Visibility:        "private",
		EvaluationStatus:  evaluationStatus(result),
		Owner:             actorFromRequest(r),
		Metrics:           result.Metrics,
		Summary:           result.Summary,
		CreatedAt:         time.Now().UTC().Format(time.RFC3339),
		SourcePod:         logPodName(logs),
		SourceApplication: jobName,
		PublishedServices: []string{request.ServiceName},
	})
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

func evaluationStatus(result DeliveryResult) string {
	summary := strings.ToLower(result.Summary)
	if strings.Contains(summary, "failed") || strings.Contains(summary, "fail") {
		return "failed"
	}
	if strings.Contains(summary, "passed") || strings.Contains(summary, "pass") {
		return "passed"
	}
	if len(result.Metrics) > 0 {
		return "passed"
	}
	return "pending"
}

func logPodName(logs *observe.Logs) string {
	if logs == nil {
		return ""
	}
	return logs.Pod
}

func recordArtifact(ctx context.Context, store ArtifactStore, artifact observe.ModelArtifact) {
	if store == nil {
		return
	}
	_ = store.SaveArtifact(ctx, artifact)
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
    cmd:
      - python
      - -c
    args:
      - |
        import http.server, json, socketserver, urllib.parse
        model_uri = %q
        port = %d
        print("serve model " + model_uri + " on :" + str(port), flush=True)
        class Handler(http.server.BaseHTTPRequestHandler):
          def do_GET(self):
            parsed = urllib.parse.urlparse(self.path)
            if parsed.path == "/healthz":
              payload = {"status": "ok", "modelURI": model_uri}
            else:
              payload = {"modelURI": model_uri, "path": parsed.path}
            body = json.dumps(payload).encode()
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            self.wfile.write(body)
        socketserver.TCPServer(("", port), Handler).serve_forever()
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
`, request.ServiceName, namespace, request.ServiceName, request.Image, result.ModelURI, request.Port, jobName, result.ModelURI, request.Port, request.ServicePort, result.ModelURI, namespace)
}

func actorFromRequest(r *http.Request) string {
	if value := requestIdentityUser(r, defaultAuthConfig()); value != "" {
		return value
	}
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
		if err := authorizeNamespaceRequest(r, normalized.Namespace); err != nil {
			writeError(w, http.StatusForbidden, err.Error())
			return
		}
		if err := authorizeNamespaceRequest(r, normalized.GovernanceIntent.Placement.Namespace); err != nil {
			writeError(w, http.StatusForbidden, err.Error())
			return
		}
		if err := authorizeNamespaceRequest(r, normalized.GovernanceIntent.Isolation.TenantNamespace); err != nil {
			writeError(w, http.StatusForbidden, err.Error())
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
