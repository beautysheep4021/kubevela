package aidomain

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/oam-dev/kubevela/pkg/ai/domain"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func TestTranslateAIServiceToNativeApplication(t *testing.T) {
	out := translate(t, []byte(`
apiVersion: ai.oam.dev/v1alpha1
kind: AIService
metadata:
  name: sentiment-demo
  namespace: ai-demo
spec:
  componentName: sentiment-api
  properties:
    image: hashicorp/http-echo:0.2.3
    replicas: 2
    model:
      name: sentiment
      version: v1
  runtime:
    runtime: http
    tenant: demo-tenant
    project: sentiment
    environment: poc
  placement:
    namespace: ai-runtime
    clusters:
    - local
`))

	assertStringField(t, out, "core.oam.dev/v1beta1", "apiVersion")
	assertStringField(t, out, "Application", "kind")
	assertStringField(t, out, "sentiment-demo", "metadata", "name")
	assertStringField(t, out, "ai-demo", "metadata", "namespace")
	assertStringField(t, out, "sentiment-api", "spec", "components", "0", "name")
	assertStringField(t, out, "ai-service", "spec", "components", "0", "type")
	assertStringField(t, out, "hashicorp/http-echo:0.2.3", "spec", "components", "0", "properties", "image")
	assertInt64Field(t, out, 2, "spec", "components", "0", "properties", "replicas")
	assertStringField(t, out, "sentiment", "spec", "components", "0", "properties", "model", "name")
	assertStringField(t, out, "v1", "spec", "components", "0", "properties", "model", "version")
	assertStringField(t, out, "ai-runtime", "spec", "components", "0", "traits", "0", "type")
	assertStringField(t, out, "demo-tenant", "spec", "components", "0", "traits", "0", "properties", "tenant")
	assertStringField(t, out, "topology", "spec", "policies", "0", "type")
	assertStringField(t, out, "ai-runtime", "spec", "policies", "0", "properties", "namespace")
	assertStringField(t, out, "local", "spec", "policies", "0", "properties", "clusters", "0")
	assertStringField(t, out, "deploy", "spec", "workflow", "steps", "0", "type")
	assertStringField(t, out, "local-topology", "spec", "workflow", "steps", "0", "properties", "policies", "0")
}

func TestTranslateAIJobToNativeApplication(t *testing.T) {
	out := translate(t, []byte(`
apiVersion: ai.oam.dev/v1alpha1
kind: AIJob
metadata:
  name: evaluator-demo
spec:
  componentName: batch-evaluator
  properties:
    image: busybox:1.36
    jobKind: evaluation
    dataset:
      uri: s3://datasets/eval
    output:
      uri: s3://outputs/eval
  runtime:
    runtime: batch
    tenant: demo-tenant
    project: evaluator
    datasetURI: s3://datasets/eval
  placement:
    namespace: ai-jobs
    clusters:
    - local
`))

	assertStringField(t, out, "core.oam.dev/v1beta1", "apiVersion")
	assertStringField(t, out, "Application", "kind")
	assertStringField(t, out, "batch-evaluator", "spec", "components", "0", "name")
	assertStringField(t, out, "ai-job", "spec", "components", "0", "type")
	assertStringField(t, out, "busybox:1.36", "spec", "components", "0", "properties", "image")
	assertStringField(t, out, "evaluation", "spec", "components", "0", "properties", "jobKind")
	assertStringField(t, out, "s3://datasets/eval", "spec", "components", "0", "properties", "dataset", "uri")
	assertStringField(t, out, "s3://outputs/eval", "spec", "components", "0", "properties", "output", "uri")
	assertStringField(t, out, "ai-runtime", "spec", "components", "0", "traits", "0", "type")
	assertStringField(t, out, "s3://datasets/eval", "spec", "components", "0", "traits", "0", "properties", "datasetURI")
	assertStringField(t, out, "topology", "spec", "policies", "0", "type")
	assertStringField(t, out, "ai-jobs", "spec", "policies", "0", "properties", "namespace")
	assertStringField(t, out, "deploy", "spec", "workflow", "steps", "0", "type")
}

func TestNormalizeAIServiceExtractsGovernanceIntent(t *testing.T) {
	normalized := normalize(t, []byte(`
apiVersion: ai.oam.dev/v1alpha1
kind: AIService
metadata:
  name: sentiment-demo
  namespace: ai-demo
spec:
  componentName: sentiment-api
  properties:
    image: hashicorp/http-echo:0.2.3
    replicas: 2
    model:
      name: sentiment
      version: v1
      uri: oss://models/sentiment/v1
    endpoint:
      port: 5678
      servicePort: 80
      type: ClusterIP
  runtime:
    runtime: http
    framework: demo
    tenant: demo-tenant
    project: sentiment
    environment: poc
    modelURI: oss://models/sentiment/v1
    owner: ai-platform
  placement:
    namespace: ai-runtime
    clusters:
    - local
`))

	if normalized.Kind != "AIService" || normalized.WorkloadType != "service" {
		t.Fatalf("unexpected normalized service identity: %#v", normalized)
	}
	if normalized.Name != "sentiment-demo" || normalized.Namespace != "ai-demo" || normalized.ComponentName != "sentiment-api" {
		t.Fatalf("unexpected normalized names: %#v", normalized)
	}
	if normalized.Image != "hashicorp/http-echo:0.2.3" || normalized.Runtime != "http" {
		t.Fatalf("unexpected image/runtime: %#v", normalized)
	}
	if normalized.GovernanceIntent.Tenant != "demo-tenant" ||
		normalized.GovernanceIntent.Project != "sentiment" ||
		normalized.GovernanceIntent.Environment != "poc" ||
		normalized.GovernanceIntent.Owner != "ai-platform" ||
		normalized.GovernanceIntent.ModelURI != "oss://models/sentiment/v1" ||
		normalized.GovernanceIntent.Placement.Namespace != "ai-runtime" ||
		len(normalized.GovernanceIntent.Placement.Clusters) != 1 ||
		normalized.GovernanceIntent.Placement.Clusters[0] != "local" {
		t.Fatalf("unexpected governance intent: %#v", normalized.GovernanceIntent)
	}
	if normalized.WorkloadIntent.Service.Model.Name != "sentiment" ||
		normalized.WorkloadIntent.Service.Model.Version != "v1" ||
		normalized.WorkloadIntent.Service.Endpoint.Type != "ClusterIP" ||
		normalized.WorkloadIntent.Service.Endpoint.Port != 5678 ||
		normalized.WorkloadIntent.Service.Endpoint.ServicePort != 80 {
		t.Fatalf("unexpected service workload intent: %#v", normalized.WorkloadIntent.Service)
	}
}

func TestNormalizeAIJobExtractsWorkloadIntent(t *testing.T) {
	normalized := normalize(t, []byte(`
apiVersion: ai.oam.dev/v1alpha1
kind: AIJob
metadata:
  name: evaluator-demo
spec:
  componentName: batch-evaluator
  properties:
    image: busybox:1.36
    jobKind: evaluation
    dataset:
      name: eval-set
      uri: s3://datasets/eval
    output:
      uri: s3://outputs/eval
    ttlSecondsAfterFinished: 300
  runtime:
    runtime: batch
    framework: shell
    tenant: demo-tenant
    project: evaluator
    datasetURI: s3://datasets/eval
`))

	if normalized.Kind != "AIJob" || normalized.WorkloadType != "job" {
		t.Fatalf("unexpected normalized job identity: %#v", normalized)
	}
	if normalized.ComponentName != "batch-evaluator" || normalized.Image != "busybox:1.36" || normalized.Runtime != "batch" {
		t.Fatalf("unexpected normalized job fields: %#v", normalized)
	}
	if normalized.GovernanceIntent.Tenant != "demo-tenant" ||
		normalized.GovernanceIntent.Project != "evaluator" ||
		normalized.GovernanceIntent.DatasetURI != "s3://datasets/eval" {
		t.Fatalf("unexpected governance intent: %#v", normalized.GovernanceIntent)
	}
	if normalized.WorkloadIntent.Job.JobKind != "evaluation" ||
		normalized.WorkloadIntent.Job.Dataset.Name != "eval-set" ||
		normalized.WorkloadIntent.Job.Dataset.URI != "s3://datasets/eval" ||
		normalized.WorkloadIntent.Job.Output.URI != "s3://outputs/eval" ||
		normalized.WorkloadIntent.Job.TTLSecondsAfterFinished == nil ||
		*normalized.WorkloadIntent.Job.TTLSecondsAfterFinished != 300 {
		t.Fatalf("unexpected job workload intent: %#v", normalized.WorkloadIntent.Job)
	}
}

func TestTranslateRejectsAIWorkflow(t *testing.T) {
	_, err := domain.TranslateYAML([]byte(`
apiVersion: ai.oam.dev/v1alpha1
kind: AIWorkflow
metadata:
  name: unsupported-workflow
spec: {}
`))
	if err == nil {
		t.Fatal("expected AIWorkflow to be rejected")
	}
	if !strings.Contains(err.Error(), "unsupported kind AIWorkflow") {
		t.Fatalf("expected unsupported-kind error, got %v", err)
	}
}

func TestValidateAIServiceRequiresImage(t *testing.T) {
	result, err := domain.ValidateYAML([]byte(`
apiVersion: ai.oam.dev/v1alpha1
kind: AIService
metadata:
  name: no-image
spec:
  properties:
    replicas: 1
`))
	if err != nil {
		t.Fatalf("ValidateYAML returned error: %v", err)
	}
	if len(result.Errors) != 1 || !strings.Contains(result.Errors[0], "spec.properties.image is required") {
		t.Fatalf("expected missing image validation error, got %#v", result.Errors)
	}
}

func TestValidateAIJobRequiresJobKind(t *testing.T) {
	result, err := domain.ValidateYAML([]byte(`
apiVersion: ai.oam.dev/v1alpha1
kind: AIJob
metadata:
  name: no-job-kind
spec:
  properties:
    image: busybox:1.36
`))
	if err != nil {
		t.Fatalf("ValidateYAML returned error: %v", err)
	}
	if len(result.Errors) != 1 || !strings.Contains(result.Errors[0], "spec.properties.jobKind is required") {
		t.Fatalf("expected missing jobKind validation error, got %#v", result.Errors)
	}
}

func TestValidateAIRuntimeRequiresRuntimeWhenTraitIsPresent(t *testing.T) {
	result, err := domain.ValidateYAML([]byte(`
apiVersion: ai.oam.dev/v1alpha1
kind: AIService
metadata:
  name: invalid-runtime
spec:
  properties:
    image: hashicorp/http-echo:0.2.3
  runtime:
    framework: demo
`))
	if err != nil {
		t.Fatalf("ValidateYAML returned error: %v", err)
	}
	if len(result.Errors) != 1 || !strings.Contains(result.Errors[0], "spec.runtime.runtime is required when spec.runtime is set") {
		t.Fatalf("expected missing runtime validation error, got %#v", result.Errors)
	}
}

func TestValidateWarnsWhenComponentNameDefaultsToApplicationName(t *testing.T) {
	result, err := domain.ValidateYAML([]byte(`
apiVersion: ai.oam.dev/v1alpha1
kind: AIService
metadata:
  name: sentiment
spec:
  properties:
    image: hashicorp/http-echo:0.2.3
`))
	if err != nil {
		t.Fatalf("ValidateYAML returned error: %v", err)
	}
	if len(result.Errors) != 0 {
		t.Fatalf("expected no validation errors, got %#v", result.Errors)
	}
	if len(result.Warnings) != 1 || !strings.Contains(result.Warnings[0], "spec.componentName is empty") {
		t.Fatalf("expected default componentName warning, got %#v", result.Warnings)
	}
}

func TestDomainExamplesTranslateToNativeApplications(t *testing.T) {
	root := projectRoot(t)
	for _, path := range []string{
		"docs/examples/ai-platform/domain/ai-service.yaml",
		"docs/examples/ai-platform/domain/ai-job.yaml",
	} {
		path := path
		t.Run(path, func(t *testing.T) {
			obj := translate(t, readFile(t, filepath.Join(root, path)))
			assertStringField(t, obj, "core.oam.dev/v1beta1", "apiVersion")
			assertStringField(t, obj, "Application", "kind")
			assertStringField(t, obj, "ai-runtime", "spec", "components", "0", "traits", "0", "type")
			assertStringField(t, obj, "topology", "spec", "policies", "0", "type")
			assertStringField(t, obj, "deploy", "spec", "workflow", "steps", "0", "type")
		})
	}
}

func TestDomainCommandTranslatesExampleFile(t *testing.T) {
	root := projectRoot(t)
	cmd := exec.Command("go", "run", "./references/cmd/ai-domain", "-f", "docs/examples/ai-platform/domain/ai-service.yaml")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GOCACHE=/tmp/kubevela-go-build-cache")

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ai-domain command failed: %v\n%s", err, string(out))
	}
	obj := decodeApplication(t, out)
	assertStringField(t, obj, "core.oam.dev/v1beta1", "apiVersion")
	assertStringField(t, obj, "Application", "kind")
	assertStringField(t, obj, "ai-service", "spec", "components", "0", "type")
}

func TestDomainCommandValidatesInvalidFile(t *testing.T) {
	root := projectRoot(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.yaml")
	if err := os.WriteFile(path, []byte(`
apiVersion: ai.oam.dev/v1alpha1
kind: AIJob
metadata:
  name: invalid
spec:
  properties:
    image: busybox:1.36
`), 0600); err != nil {
		t.Fatalf("write invalid fixture: %v", err)
	}

	cmd := exec.Command("go", "run", "./references/cmd/ai-domain", "validate", "-f", path)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GOCACHE=/tmp/kubevela-go-build-cache")

	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected validate command to fail for invalid file, output:\n%s", string(out))
	}
	if !strings.Contains(string(out), "spec.properties.jobKind is required") {
		t.Fatalf("expected jobKind validation output, got:\n%s", string(out))
	}
}

func TestDomainCommandNormalizesExampleFile(t *testing.T) {
	root := projectRoot(t)
	cmd := exec.Command("go", "run", "./references/cmd/ai-domain", "normalize", "-f", "docs/examples/ai-platform/domain/ai-service.yaml")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GOCACHE=/tmp/kubevela-go-build-cache")

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("normalize command failed: %v\n%s", err, string(out))
	}
	var normalized domain.NormalizedObject
	if err := json.Unmarshal(out, &normalized); err != nil {
		t.Fatalf("decode normalize output: %v\n%s", err, string(out))
	}
	if normalized.Kind != "AIService" || normalized.WorkloadType != "service" || normalized.GovernanceIntent.Tenant != "demo-tenant" {
		t.Fatalf("unexpected normalize output: %#v", normalized)
	}
}

func TestDomainLayerScopeDoesNotAddRuntimeControlPlane(t *testing.T) {
	root := projectRoot(t)
	for _, path := range []string{
		"pkg/ai/domain/translator.go",
		"references/cmd/ai-domain/main.go",
		"docs/examples/ai-platform/domain/ai-service.yaml",
		"docs/examples/ai-platform/domain/ai-job.yaml",
	} {
		path := path
		t.Run(path, func(t *testing.T) {
			content := string(readFile(t, filepath.Join(root, path)))
			for _, forbidden := range []string{"Reconcile", "Controller", "database", "Database", "SQL"} {
				if strings.Contains(content, forbidden) {
					t.Fatalf("expected %s to avoid %q in the minimal domain layer", path, forbidden)
				}
			}
		})
	}
}

func translate(t *testing.T, in []byte) *unstructured.Unstructured {
	t.Helper()
	out, err := domain.TranslateYAML(in)
	if err != nil {
		t.Fatalf("TranslateYAML returned error: %v", err)
	}
	return decodeApplication(t, out)
}

func normalize(t *testing.T, in []byte) domain.NormalizedObject {
	t.Helper()
	normalized, err := domain.NormalizeYAML(in)
	if err != nil {
		t.Fatalf("NormalizeYAML returned error: %v", err)
	}
	return normalized
}

func decodeApplication(t *testing.T, out []byte) *unstructured.Unstructured {
	t.Helper()
	obj := &unstructured.Unstructured{}
	if err := yaml.Unmarshal(out, obj); err != nil {
		t.Fatalf("failed to decode translated YAML: %v\n%s", err, string(out))
	}
	return obj
}

func projectRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			t.Fatal("could not find project root containing go.mod")
		}
		wd = parent
	}
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return content
}

func assertStringField(t *testing.T, obj *unstructured.Unstructured, expected string, fields ...string) {
	t.Helper()
	value := nestedValue(t, obj.Object, fields...)
	actual, ok := value.(string)
	if !ok {
		t.Fatalf("field %v = %T(%v), want string", fields, value, value)
	}
	if actual != expected {
		t.Fatalf("field %v = %q, want %q", fields, actual, expected)
	}
}

func assertInt64Field(t *testing.T, obj *unstructured.Unstructured, expected int64, fields ...string) {
	t.Helper()
	value := nestedValue(t, obj.Object, fields...)
	actual, ok := value.(int64)
	if !ok {
		t.Fatalf("field %v = %T(%v), want int64", fields, value, value)
	}
	if actual != expected {
		t.Fatalf("field %v = %d, want %d", fields, actual, expected)
	}
}

func nestedValue(t *testing.T, root interface{}, fields ...string) interface{} {
	t.Helper()
	current := root
	for _, field := range fields {
		switch typed := current.(type) {
		case map[string]interface{}:
			next, ok := typed[field]
			if !ok {
				t.Fatalf("field %v not found at %q", fields, field)
			}
			current = next
		case []interface{}:
			index, err := strconv.Atoi(field)
			if err != nil {
				t.Fatalf("field %v expected numeric list index at %q: %v", fields, field, err)
			}
			if index < 0 || index >= len(typed) {
				t.Fatalf("field %v index %d out of range", fields, index)
			}
			current = typed[index]
		default:
			t.Fatalf("field %v cannot descend into %s", fields, fmt.Sprintf("%T", current))
		}
	}
	return current
}
