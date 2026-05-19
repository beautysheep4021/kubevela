package aiobserve

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oam-dev/kubevela/pkg/ai/observe"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func TestSummarizeAIServiceApplication(t *testing.T) {
	summary, err := observe.SummarizeObjects([]*unstructured.Unstructured{
		mustObject(t, aiServiceApplicationYAML),
		mustObject(t, deploymentYAML),
		mustObject(t, serviceYAML),
		mustObject(t, servicePodYAML),
	})
	if err != nil {
		t.Fatalf("SummarizeObjects returned error: %v", err)
	}

	if summary.Name != "ai-service-domain-demo" || summary.Namespace != "sock-shop" {
		t.Fatalf("unexpected app identity: %#v", summary)
	}
	if summary.Phase != "running" || !summary.Healthy || summary.Message != "Ready:1/1" {
		t.Fatalf("unexpected app status: %#v", summary)
	}
	if len(summary.Components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(summary.Components))
	}
	component := summary.Components[0]
	if component.Name != "sentiment-domain-api" || component.Type != "ai-service" || component.WorkloadKind != "Deployment" {
		t.Fatalf("unexpected component: %#v", component)
	}
	if component.Workload.DesiredReplicas != 1 || component.Workload.ReadyReplicas != 1 {
		t.Fatalf("unexpected deployment readiness: %#v", component.Workload)
	}
	if component.Service.Name != "ai-service-domain-demo-sentiment-domain-api" || component.Service.ClusterIP != "10.109.168.130" {
		t.Fatalf("unexpected service summary: %#v", component.Service)
	}
	if len(component.Pods) != 1 || component.Pods[0].Phase != "Running" || component.Pods[0].ReadyContainers != 1 {
		t.Fatalf("unexpected pod summary: %#v", component.Pods)
	}
	if summary.AIMetadata["ai.oam.dev/model"] != "sentiment" {
		t.Fatalf("expected AI model metadata, got %#v", summary.AIMetadata)
	}
	if summary.AIMetadata["ai.oam.dev/model-uri"] != "oss://models/sentiment/v1" {
		t.Fatalf("expected AI model URI metadata, got %#v", summary.AIMetadata)
	}
}

func TestSummarizeAIJobApplicationWarnsForCompletedOrphanPod(t *testing.T) {
	summary, err := observe.SummarizeObjects([]*unstructured.Unstructured{
		mustObject(t, aiJobApplicationYAML),
		mustObject(t, jobYAML),
		mustObject(t, completedOrphanJobPodYAML),
	})
	if err != nil {
		t.Fatalf("SummarizeObjects returned error: %v", err)
	}

	if summary.Name != "ai-job-domain-demo" || !summary.Healthy {
		t.Fatalf("unexpected summary: %#v", summary)
	}
	component := summary.Components[0]
	if component.Name != "batch-evaluator" || component.Type != "ai-job" || component.WorkloadKind != "Job" {
		t.Fatalf("unexpected component: %#v", component)
	}
	if component.Workload.Succeeded != 1 || component.Workload.TTLSecondsAfterFinished != 300 {
		t.Fatalf("unexpected job status: %#v", component.Workload)
	}
	if len(component.Pods) != 0 {
		t.Fatalf("expected orphan pod to be excluded from current Job pods, got %#v", component.Pods)
	}
	if len(summary.Warnings) != 1 {
		t.Fatalf("expected orphan pod warning, got %#v", summary.Warnings)
	}
	if !strings.Contains(summary.Warnings[0].Message, "ownerReferences") {
		t.Fatalf("expected ownerReferences warning, got %#v", summary.Warnings[0])
	}
}

func TestSummarizeAIJobIgnoresHistoricalPodWithoutCurrentJobOwner(t *testing.T) {
	summary, err := observe.SummarizeObjects([]*unstructured.Unstructured{
		mustObject(t, aiJobApplicationYAML),
		mustObject(t, jobYAML),
		mustObject(t, currentJobPodYAML),
		mustObject(t, historicalJobPodYAML),
	})
	if err != nil {
		t.Fatalf("SummarizeObjects returned error: %v", err)
	}

	component := summary.Components[0]
	if len(component.Pods) != 1 {
		t.Fatalf("expected only the current Job-owned pod in component pods, got %#v", component.Pods)
	}
	if component.Pods[0].Name != "ai-job-domain-demo-batch-evaluator-current" {
		t.Fatalf("expected current pod, got %#v", component.Pods[0])
	}
	if len(summary.Warnings) != 1 {
		t.Fatalf("expected historical orphan pod warning, got %#v", summary.Warnings)
	}
	if !strings.Contains(summary.Warnings[0].Resource, "ai-job-domain-demo-batch-evaluator-old") {
		t.Fatalf("expected old pod warning, got %#v", summary.Warnings[0])
	}
}

func TestSummarizeAIJobWithoutCurrentJobExcludesOrphanHistoricalPod(t *testing.T) {
	summary, err := observe.SummarizeObjects([]*unstructured.Unstructured{
		mustObject(t, aiJobApplicationYAML),
		mustObject(t, historicalJobPodYAML),
	})
	if err != nil {
		t.Fatalf("SummarizeObjects returned error: %v", err)
	}

	component := summary.Components[0]
	if len(component.Pods) != 0 {
		t.Fatalf("expected orphan historical pod to be excluded when current Job is gone, got %#v", component.Pods)
	}
	if len(summary.Warnings) != 1 {
		t.Fatalf("expected orphan pod warning, got %#v", summary.Warnings)
	}
	if !strings.Contains(summary.Warnings[0].Message, "historical") {
		t.Fatalf("expected historical warning, got %#v", summary.Warnings[0])
	}
}

func TestStatusCommandFromFiles(t *testing.T) {
	root := projectRoot(t)
	dir := t.TempDir()
	files := []string{
		writeFixture(t, dir, "app.yaml", aiServiceApplicationYAML),
		writeFixture(t, dir, "deploy.yaml", deploymentYAML),
		writeFixture(t, dir, "svc.yaml", serviceYAML),
		writeFixture(t, dir, "pod.yaml", servicePodYAML),
	}

	cmd := exec.Command("go", "run", "./references/cmd/ai-domain", "status", "--from-files", strings.Join(files, ","))
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GOCACHE=/tmp/kubevela-go-build-cache")

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("status command failed: %v\n%s", err, string(out))
	}
	var summary observe.Summary
	if err := json.Unmarshal(out, &summary); err != nil {
		t.Fatalf("decode status output: %v\n%s", err, string(out))
	}
	if summary.Name != "ai-service-domain-demo" || summary.Components[0].Workload.ReadyReplicas != 1 {
		t.Fatalf("unexpected status output: %#v", summary)
	}
}

func mustObject(t *testing.T, content string) *unstructured.Unstructured {
	t.Helper()
	obj := &unstructured.Unstructured{}
	if err := yaml.Unmarshal([]byte(content), obj); err != nil {
		t.Fatalf("decode object: %v\n%s", err, content)
	}
	return obj
}

func writeFixture(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("write fixture %s: %v", path, err)
	}
	return path
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

const aiServiceApplicationYAML = `
apiVersion: core.oam.dev/v1beta1
kind: Application
metadata:
  name: ai-service-domain-demo
  namespace: sock-shop
status:
  status: running
  services:
  - healthy: true
    message: Ready:1/1
    name: sentiment-domain-api
    namespace: sock-shop
    workloadDefinition:
      apiVersion: apps/v1
      kind: Deployment
    traits:
    - healthy: true
      type: ai-runtime
`

const deploymentYAML = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sentiment-domain-api
  namespace: sock-shop
  labels:
    app.oam.dev/name: ai-service-domain-demo
    app.oam.dev/component: sentiment-domain-api
    ai.oam.dev/model: sentiment
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app.oam.dev/name: ai-service-domain-demo
        app.oam.dev/component: sentiment-domain-api
        ai.oam.dev/tenant: demo-tenant
        ai.oam.dev/runtime: http
      annotations:
        ai.oam.dev/model-uri: oss://models/sentiment/v1
status:
  readyReplicas: 1
  replicas: 1
`

const serviceYAML = `
apiVersion: v1
kind: Service
metadata:
  name: ai-service-domain-demo-sentiment-domain-api
  namespace: sock-shop
  labels:
    app.oam.dev/name: ai-service-domain-demo
    app.oam.dev/component: sentiment-domain-api
spec:
  clusterIP: 10.109.168.130
  ports:
  - port: 80
    targetPort: 5678
`

const servicePodYAML = `
apiVersion: v1
kind: Pod
metadata:
  name: sentiment-domain-api-dd9cbd987-wftdl
  namespace: sock-shop
  labels:
    app.oam.dev/name: ai-service-domain-demo
    app.oam.dev/component: sentiment-domain-api
    ai.oam.dev/model: sentiment
status:
  phase: Running
  containerStatuses:
  - name: main
    ready: true
`

const aiJobApplicationYAML = `
apiVersion: core.oam.dev/v1beta1
kind: Application
metadata:
  name: ai-job-domain-demo
  namespace: sock-shop
status:
  status: running
  services:
  - healthy: true
    message: Active/Failed/Succeeded:0/0/1
    name: batch-evaluator
    namespace: sock-shop
    workloadDefinition:
      apiVersion: batch/v1
      kind: Job
    traits:
    - healthy: true
      type: ai-runtime
`

const jobYAML = `
apiVersion: batch/v1
kind: Job
metadata:
  name: ai-job-domain-demo-batch-evaluator
  namespace: sock-shop
  uid: job-uid-current
  labels:
    app.oam.dev/name: ai-job-domain-demo
    app.oam.dev/component: batch-evaluator
spec:
  ttlSecondsAfterFinished: 300
  template:
    metadata:
      labels:
        app.oam.dev/name: ai-job-domain-demo
        app.oam.dev/component: batch-evaluator
        ai.oam.dev/job-kind: evaluation
      annotations:
        ai.oam.dev/dataset-uri: oss://datasets/eval-set/v1
status:
  succeeded: 1
`

const currentJobPodYAML = `
apiVersion: v1
kind: Pod
metadata:
  name: ai-job-domain-demo-batch-evaluator-current
  namespace: sock-shop
  labels:
    app.oam.dev/name: ai-job-domain-demo
    app.oam.dev/component: batch-evaluator
    ai.oam.dev/job-kind: evaluation
  ownerReferences:
  - apiVersion: batch/v1
    kind: Job
    name: ai-job-domain-demo-batch-evaluator
    uid: job-uid-current
    controller: true
status:
  phase: Succeeded
  containerStatuses:
  - name: main
    ready: false
`

const historicalJobPodYAML = `
apiVersion: v1
kind: Pod
metadata:
  name: ai-job-domain-demo-batch-evaluator-old
  namespace: sock-shop
  labels:
    app.oam.dev/name: ai-job-domain-demo
    app.oam.dev/component: batch-evaluator
    ai.oam.dev/job-kind: evaluation
status:
  phase: Pending
  containerStatuses:
  - name: main
    ready: false
    state:
      waiting:
        reason: ImagePullBackOff
`

const completedOrphanJobPodYAML = `
apiVersion: v1
kind: Pod
metadata:
  name: ai-job-domain-demo-batch-evaluator-x8kqt
  namespace: sock-shop
  labels:
    app.oam.dev/name: ai-job-domain-demo
    app.oam.dev/component: batch-evaluator
    ai.oam.dev/job-kind: evaluation
status:
  phase: Succeeded
  containerStatuses:
  - name: main
    ready: false
`
