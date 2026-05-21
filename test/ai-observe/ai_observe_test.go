package aiobserve

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oam-dev/kubevela/pkg/ai/observe"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
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

func TestSummarizePodContainerWaitingAndTerminatedStates(t *testing.T) {
	summary, err := observe.SummarizeObjects([]*unstructured.Unstructured{
		mustObject(t, aiServiceApplicationYAML),
		mustObject(t, deploymentYAML),
		mustObject(t, unhealthyServicePodYAML),
	})
	if err != nil {
		t.Fatalf("SummarizeObjects returned error: %v", err)
	}

	pod := summary.Components[0].Pods[0]
	if len(pod.Containers) != 2 {
		t.Fatalf("expected two container summaries, got %#v", pod.Containers)
	}
	main := pod.Containers[0]
	if main.Name != "main" || main.State != "waiting" || main.Reason != "ImagePullBackOff" {
		t.Fatalf("unexpected waiting container summary: %#v", main)
	}
	if main.Message != "Back-off pulling image" || main.RestartCount != 3 {
		t.Fatalf("unexpected waiting detail: %#v", main)
	}
	sidecar := pod.Containers[1]
	if sidecar.Name != "sidecar" || sidecar.State != "terminated" || sidecar.Reason != "Error" {
		t.Fatalf("unexpected terminated container summary: %#v", sidecar)
	}
	if sidecar.ExitCode != 137 || sidecar.RestartCount != 1 {
		t.Fatalf("unexpected terminated detail: %#v", sidecar)
	}
}

func TestSummarizePodEventsForDiagnostics(t *testing.T) {
	summary, err := observe.SummarizeObjects([]*unstructured.Unstructured{
		mustObject(t, aiServiceApplicationYAML),
		mustObject(t, deploymentYAML),
		mustObject(t, unhealthyServicePodYAML),
		mustObject(t, imagePullBackOffEventYAML),
	})
	if err != nil {
		t.Fatalf("SummarizeObjects returned error: %v", err)
	}

	pod := summary.Components[0].Pods[0]
	if len(pod.Events) != 1 {
		t.Fatalf("expected one diagnostic event, got %#v", pod.Events)
	}
	event := pod.Events[0]
	if event.Type != "Warning" || event.Reason != "Failed" {
		t.Fatalf("unexpected event summary: %#v", event)
	}
	if !strings.Contains(event.Message, "Failed to pull image") {
		t.Fatalf("expected image pull message, got %#v", event)
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

func TestClientListsOnlyAIApplications(t *testing.T) {
	client := newObserveFakeClient(
		mustObject(t, aiServiceListApplicationYAML),
		mustObject(t, nonAIApplicationYAML),
	)
	reader := observe.NewClient(client)

	items, err := reader.ListApplications(context.Background(), "sock-shop")
	if err != nil {
		t.Fatalf("ListApplications returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 AI application, got %#v", items)
	}
	item := items[0]
	if item.Name != "ai-service-domain-demo" || item.Namespace != "sock-shop" {
		t.Fatalf("unexpected item identity: %#v", item)
	}
	if item.Phase != "running" || !item.Healthy || item.Message != "Ready:1/1" {
		t.Fatalf("unexpected item status: %#v", item)
	}
	if len(item.WorkloadTypes) != 1 || item.WorkloadTypes[0] != "service" {
		t.Fatalf("unexpected workload types: %#v", item.WorkloadTypes)
	}
	if item.AIMetadata["ai.oam.dev/tenant"] != "demo-tenant" {
		t.Fatalf("expected tenant metadata, got %#v", item.AIMetadata)
	}
}

func TestClientSummarizesApplicationWithDynamicClient(t *testing.T) {
	client := newObserveFakeClient(
		mustObject(t, aiServiceApplicationYAML),
		mustObject(t, deploymentYAML),
		mustObject(t, serviceYAML),
		mustObject(t, servicePodYAML),
	)
	reader := observe.NewClient(client)

	summary, err := reader.SummarizeApplication(context.Background(), "sock-shop", "ai-service-domain-demo")
	if err != nil {
		t.Fatalf("SummarizeApplication returned error: %v", err)
	}
	if summary.Name != "ai-service-domain-demo" || summary.Components[0].Workload.ReadyReplicas != 1 {
		t.Fatalf("unexpected summary: %#v", summary)
	}
}

func newObserveFakeClient(objects ...*unstructured.Unstructured) *dynamicfake.FakeDynamicClient {
	runtimeObjects := make([]runtime.Object, 0, len(objects))
	for _, object := range objects {
		runtimeObjects = append(runtimeObjects, object)
	}
	return dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), map[schema.GroupVersionResource]string{
		{Group: "core.oam.dev", Version: "v1beta1", Resource: "applications"}: "ApplicationList",
		{Group: "apps", Version: "v1", Resource: "deployments"}:               "DeploymentList",
		{Group: "batch", Version: "v1", Resource: "jobs"}:                     "JobList",
		{Group: "", Version: "v1", Resource: "services"}:                      "ServiceList",
		{Group: "", Version: "v1", Resource: "pods"}:                          "PodList",
		{Group: "", Version: "v1", Resource: "events"}:                        "EventList",
	}, runtimeObjects...)
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

const aiServiceListApplicationYAML = `
apiVersion: core.oam.dev/v1beta1
kind: Application
metadata:
  name: ai-service-domain-demo
  namespace: sock-shop
spec:
  components:
  - name: sentiment-domain-api
    type: ai-service
    traits:
    - type: ai-runtime
      properties:
        tenant: demo-tenant
        project: sentiment
        environment: poc
status:
  status: running
  services:
  - healthy: true
    message: Ready:1/1
    name: sentiment-domain-api
    workloadDefinition:
      kind: Deployment
`

const nonAIApplicationYAML = `
apiVersion: core.oam.dev/v1beta1
kind: Application
metadata:
  name: web-demo
  namespace: sock-shop
spec:
  components:
  - name: web
    type: webservice
status:
  status: running
`

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

const unhealthyServicePodYAML = `
apiVersion: v1
kind: Pod
metadata:
  name: sentiment-domain-api-failed
  namespace: sock-shop
  labels:
    app.oam.dev/name: ai-service-domain-demo
    app.oam.dev/component: sentiment-domain-api
    ai.oam.dev/model: sentiment
status:
  phase: Pending
  containerStatuses:
  - name: main
    ready: false
    restartCount: 3
    state:
      waiting:
        reason: ImagePullBackOff
        message: Back-off pulling image
  - name: sidecar
    ready: false
    restartCount: 1
    state:
      terminated:
        exitCode: 137
        reason: Error
        message: container killed
`

const imagePullBackOffEventYAML = `
apiVersion: v1
kind: Event
metadata:
  name: sentiment-domain-api-failed.1817f
  namespace: sock-shop
involvedObject:
  kind: Pod
  namespace: sock-shop
  name: sentiment-domain-api-failed
type: Warning
reason: Failed
message: Failed to pull image "ghcr.io/example/not-exist:bad"
count: 3
lastTimestamp: "2026-05-19T10:04:21Z"
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
