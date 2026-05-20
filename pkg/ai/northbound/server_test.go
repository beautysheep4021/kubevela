package northbound

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oam-dev/kubevela/pkg/ai/domain"
	domainapply "github.com/oam-dev/kubevela/pkg/ai/domain/apply"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestServerHealthz(t *testing.T) {
	server := NewServer()
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if strings.TrimSpace(recorder.Body.String()) != "ok" {
		t.Fatalf("body = %q, want ok", recorder.Body.String())
	}
}

func TestServerServesConsolePage(t *testing.T) {
	server := NewServer()
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	contentType := recorder.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Fatalf("content-type = %q, want text/html", contentType)
	}
	body := recorder.Body.String()
	for _, expected := range []string{
		"使用方工作台",
		"提交意图",
		"领域 YAML",
		"治理意图",
		"工作负载意图",
		"任务类型",
		"模型名称",
		"数据集 URI",
		"生成 YAML",
		"提交部署",
		"服务端 DryRun",
		"/api/v1/ai/validate",
		"/api/v1/ai/normalize",
		"/api/v1/ai/applications",
		"governanceIntent",
		"workloadIntent",
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("console page missing %q", expected)
		}
	}
}

func TestServerDeploysDomainYAMLWithDryRun(t *testing.T) {
	applier := &recordingApplicationApplier{}
	server := NewServerWithOptions(Options{Applier: applier})
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/ai/applications?dryRun=true", bytes.NewReader([]byte(validAIServiceYAML()))))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var result DeployResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if result.Application.Name != "sentiment-demo" || result.Application.Namespace != "ai-demo" || !result.Application.DryRun {
		t.Fatalf("unexpected application result: %#v", result.Application)
	}
	if result.Normalized.Kind != "AIService" || result.Normalized.GovernanceIntent.Tenant != "demo-tenant" {
		t.Fatalf("unexpected normalized payload: %#v", result.Normalized)
	}
	if !applier.called || applier.namespace != "ai-demo" || applier.name != "sentiment-demo" {
		t.Fatalf("unexpected applier call: %#v", applier)
	}
	if len(applier.options.DryRun) != 1 || applier.options.DryRun[0] != metav1.DryRunAll {
		t.Fatalf("expected server dry-run option, got %#v", applier.options.DryRun)
	}
	if !strings.Contains(string(applier.content), "kind: Application") || !strings.Contains(string(applier.content), "type: ai-service") {
		t.Fatalf("expected translated Application YAML, got:\n%s", string(applier.content))
	}
}

func TestServerDeployReturnsUnavailableWhenApplyIsNotConfigured(t *testing.T) {
	server := NewServer()
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/ai/applications", bytes.NewReader([]byte(validAIServiceYAML()))))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusServiceUnavailable, recorder.Body.String())
	}
	var payload errorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if !strings.Contains(payload.Error, "deployment is not configured") {
		t.Fatalf("unexpected error: %#v", payload)
	}
}

func TestServerValidatesDomainYAML(t *testing.T) {
	server := NewServer()
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/ai/validate", bytes.NewReader([]byte(validAIJobYAML()))))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var result domain.ValidationResult
	if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if len(result.Errors) != 0 {
		t.Fatalf("expected no validation errors, got %#v", result.Errors)
	}
}

func TestServerReturnsBadRequestForInvalidDomainYAML(t *testing.T) {
	server := NewServer()
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/ai/validate", bytes.NewReader([]byte(`
apiVersion: ai.oam.dev/v1alpha1
kind: AIJob
metadata:
  name: invalid
spec:
  properties:
    image: busybox:1.36
`))))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	var payload errorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if !strings.Contains(payload.Error, "spec.properties.jobKind is required") {
		t.Fatalf("expected jobKind error, got %#v", payload)
	}
}

func TestServerNormalizesDomainYAML(t *testing.T) {
	server := NewServer()
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/ai/normalize", bytes.NewReader([]byte(validAIServiceYAML()))))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var normalized domain.NormalizedObject
	if err := json.Unmarshal(recorder.Body.Bytes(), &normalized); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if normalized.Kind != "AIService" || normalized.WorkloadType != "service" {
		t.Fatalf("unexpected normalized identity: %#v", normalized)
	}
	if normalized.GovernanceIntent.Tenant != "demo-tenant" || normalized.WorkloadIntent.Service.Model.Name != "sentiment" {
		t.Fatalf("unexpected normalized intent: %#v", normalized)
	}
}

func TestServerRejectsUnsupportedMethod(t *testing.T) {
	server := NewServer()
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/ai/normalize", nil))

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusMethodNotAllowed)
	}
}

func validAIServiceYAML() string {
	return `
apiVersion: ai.oam.dev/v1alpha1
kind: AIService
metadata:
  name: sentiment-demo
  namespace: ai-demo
spec:
  componentName: sentiment-api
  properties:
    image: hashicorp/http-echo:0.2.3
    model:
      name: sentiment
  runtime:
    runtime: http
    tenant: demo-tenant
`
}

func validAIJobYAML() string {
	return `
apiVersion: ai.oam.dev/v1alpha1
kind: AIJob
metadata:
  name: evaluator-demo
spec:
  componentName: batch-evaluator
  properties:
    image: busybox:1.36
    jobKind: evaluation
  runtime:
    runtime: batch
`
}

type recordingApplicationApplier struct {
	called    bool
	namespace string
	name      string
	content   []byte
	options   metav1.PatchOptions
}

func (a *recordingApplicationApplier) ApplyApplication(_ context.Context, namespace, name string, content []byte, opts metav1.PatchOptions) (*unstructured.Unstructured, error) {
	a.called = true
	a.namespace = namespace
	a.name = name
	a.content = append([]byte(nil), content...)
	a.options = opts
	return &unstructured.Unstructured{}, nil
}

var _ domainapply.ApplicationApplier = (*recordingApplicationApplier)(nil)
