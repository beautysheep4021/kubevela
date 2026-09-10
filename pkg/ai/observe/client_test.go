package observe

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
)

func TestSelectLogPodUsesNewestPodByDefault(t *testing.T) {
	oldPod := logPodObject("demo-old", "main", time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC))
	newPod := logPodObject("demo-new", "main", time.Date(2026, 5, 21, 10, 1, 0, 0, time.UTC))

	selected := selectLogPod([]unstructured.Unstructured{oldPod, newPod}, "")

	if selected == nil || selected.GetName() != "demo-new" {
		t.Fatalf("selected pod = %#v, want demo-new", selected)
	}
}

func TestLogPodFromUnstructuredIncludesContainers(t *testing.T) {
	pod := logPodObject("demo-pod", "trainer", time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC))

	result := logPodFromUnstructured(&pod)

	if result.Name != "demo-pod" || result.Phase != "Running" || len(result.Containers) != 1 || result.Containers[0] != "trainer" {
		t.Fatalf("unexpected log pod: %#v", result)
	}
}

func TestRerunApplicationNameIsDNSFriendly(t *testing.T) {
	name := rerunApplicationName("ai-job-demo", time.Date(2026, 5, 22, 8, 9, 0, 0, time.UTC))

	if name != "ai-job-demo-rerun-0522080900" {
		t.Fatalf("name = %q, want ai-job-demo-rerun-0522080900", name)
	}
	if len(name) > 63 {
		t.Fatalf("name length = %d, want <= 63", len(name))
	}
}

func TestApplicationHasWorkloadTypeFindsAIJob(t *testing.T) {
	app := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "core.oam.dev/v1beta1",
			"kind":       "Application",
			"spec": map[string]interface{}{
				"components": []interface{}{
					map[string]interface{}{"name": "trainer", "type": "ai-job"},
				},
			},
		},
	}

	if !applicationHasWorkloadType(&app, "ai-job") {
		t.Fatalf("expected application to contain ai-job component")
	}
	if applicationHasWorkloadType(&app, "ai-service") {
		t.Fatalf("did not expect application to contain ai-service component")
	}
}

func TestApplicationListItemIncludesResourceSummary(t *testing.T) {
	app := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "core.oam.dev/v1beta1",
			"kind":       "Application",
			"metadata": map[string]interface{}{
				"name":      "tenant-a-training",
				"namespace": "sock-shop",
			},
			"spec": map[string]interface{}{
				"components": []interface{}{
					map[string]interface{}{
						"name": "serving",
						"type": "ai-service",
						"properties": map[string]interface{}{
							"replicas": int64(2),
							"resources": map[string]interface{}{
								"requests": map[string]interface{}{
									"cpu":            "4",
									"memory":         "8Gi",
									"nvidia.com/gpu": "1",
								},
							},
						},
						"traits": []interface{}{
							map[string]interface{}{
								"type": "ai-runtime",
								"properties": map[string]interface{}{
									"tenant":      "tenant-a",
									"environment": "prod",
								},
							},
						},
					},
					map[string]interface{}{
						"name": "trainer",
						"type": "ai-job",
						"properties": map[string]interface{}{
							"parallelism": int64(3),
							"resources": map[string]interface{}{
								"requests": map[string]interface{}{
									"cpu":    "500m",
									"memory": "1Gi",
								},
							},
						},
					},
				},
			},
		},
	}

	_ = unstructured.SetNestedField(app.Object, "2026-09-10T03:00:00Z", "metadata", "creationTimestamp")
	item, ok := applicationListItem(&app)

	if !ok {
		t.Fatalf("expected application list item")
	}
	encoded, err := json.Marshal(item)
	if err != nil || !strings.Contains(string(encoded), `"createdAt":"2026-09-10T03:00:00Z"`) {
		t.Fatalf("list must expose the actual creation time: %s (%v)", encoded, err)
	}
	if item.ResourceSummary.CPUMilli != 9500 {
		t.Fatalf("cpu milli = %d, want 9500", item.ResourceSummary.CPUMilli)
	}
	if item.ResourceSummary.MemoryMi != 19456 {
		t.Fatalf("memory Mi = %d, want 19456", item.ResourceSummary.MemoryMi)
	}
	if item.ResourceSummary.GPU != 2 {
		t.Fatalf("gpu = %d, want 2", item.ResourceSummary.GPU)
	}
	if item.AIMetadata["ai.oam.dev/tenant"] != "tenant-a" {
		t.Fatalf("tenant metadata = %q, want tenant-a", item.AIMetadata["ai.oam.dev/tenant"])
	}
}

func TestAuditConfigMapNameIsBounded(t *testing.T) {
	name := auditConfigMapName("2026-05-22T08:09:00.000000000Z-delete-sock-shop-ai-job-demo-with-extra-long-name")

	if len(name) > 63 {
		t.Fatalf("name length = %d, want <= 63", len(name))
	}
	if !strings.HasPrefix(name, "ai-audit-") {
		t.Fatalf("name = %q, want ai-audit prefix", name)
	}
}

func TestClientSavesAndListsArtifacts(t *testing.T) {
	client := Client{Kube: fake.NewSimpleClientset()}

	if err := client.SaveArtifact(context.Background(), ModelArtifact{
		Namespace:         "sock-shop",
		Name:              "train-demo",
		JobName:           "train-demo",
		ModelURI:          "inline://models/train-demo/v1",
		PublishedServices: []string{"train-demo-service"},
	}); err != nil {
		t.Fatalf("save artifact: %v", err)
	}
	items, err := client.ListArtifacts(context.Background(), "sock-shop")
	if err != nil {
		t.Fatalf("list artifacts: %v", err)
	}
	if len(items) != 1 || items[0].ModelURI != "inline://models/train-demo/v1" || items[0].PublishedServices[0] != "train-demo-service" {
		t.Fatalf("unexpected artifacts: %#v", items)
	}
}

func TestClientSavesAndListsDatasets(t *testing.T) {
	client := Client{Kube: fake.NewSimpleClientset()}

	if err := client.SaveDataset(context.Background(), DatasetArtifact{
		Namespace:   "sock-shop",
		Name:        "customer-sft",
		DisplayName: "客服问答 SFT 数据集",
		DatasetURI:  "dataset://sock-shop/customer-sft/v1",
		Format:      "sharegpt-jsonl",
		Purpose:     "sft",
		Status:      "validated",
	}); err != nil {
		t.Fatalf("save dataset: %v", err)
	}
	items, err := client.ListDatasets(context.Background(), "sock-shop")
	if err != nil {
		t.Fatalf("list datasets: %v", err)
	}
	if len(items) != 1 || items[0].DatasetURI != "dataset://sock-shop/customer-sft/v1" || items[0].Format != "sharegpt-jsonl" {
		t.Fatalf("unexpected datasets: %#v", items)
	}
}

func logPodObject(name, container string, createdAt time.Time) unstructured.Unstructured {
	return unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name":              name,
				"creationTimestamp": createdAt.Format(time.RFC3339),
			},
			"spec": map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{"name": container},
				},
			},
			"status": map[string]interface{}{
				"phase": "Running",
			},
		},
	}
}

func TestApplicationListJobLifecycle(t *testing.T) {
	for _, tt := range []struct {
		name    string
		modify  func(*unstructured.Unstructured)
		phase   string
		healthy bool
	}{
		{name: "completed", phase: "succeeded", healthy: true},
		{name: "unhealthy is not failed", modify: func(app *unstructured.Unstructured) {
			_ = unstructured.SetNestedSlice(app.Object, []interface{}{map[string]interface{}{"name": "worker", "healthy": false}}, "status", "services")
		}, phase: "running"},
		{name: "missing status", modify: func(app *unstructured.Unstructured) { unstructured.RemoveNestedField(app.Object, "status", "services") }, phase: "running"},
		{name: "unrelated healthy status", modify: func(app *unstructured.Unstructured) {
			_ = unstructured.SetNestedSlice(app.Object, []interface{}{map[string]interface{}{"name": "other", "healthy": true}}, "status", "services")
		}, phase: "running"},
		{name: "stale generation", modify: func(app *unstructured.Unstructured) { app.SetGeneration(2) }, phase: "running"},
		{name: "failed application", modify: func(app *unstructured.Unstructured) {
			_ = unstructured.SetNestedField(app.Object, "failed", "status", "status")
		}, phase: "failed", healthy: true},
		{name: "workload completed with unhealthy trait", modify: func(app *unstructured.Unstructured) {
			_ = unstructured.SetNestedSlice(app.Object, []interface{}{map[string]interface{}{"name": "worker", "healthy": false, "workloadHealthy": true}}, "status", "services")
		}, phase: "succeeded"},
		{name: "partial multi job", modify: func(app *unstructured.Unstructured) {
			components, _, _ := unstructured.NestedSlice(app.Object, "spec", "components")
			_ = unstructured.SetNestedSlice(app.Object, append(components, map[string]interface{}{"name": "second", "type": "ai-job"}), "spec", "components")
		}, phase: "running"},
		{name: "mixed job and service", modify: func(app *unstructured.Unstructured) {
			components, _, _ := unstructured.NestedSlice(app.Object, "spec", "components")
			_ = unstructured.SetNestedSlice(app.Object, append(components, map[string]interface{}{"name": "service", "type": "ai-service"}), "spec", "components")
		}, phase: "running"},
		{name: "service health is readiness", modify: func(app *unstructured.Unstructured) {
			_ = unstructured.SetNestedSlice(app.Object, []interface{}{map[string]interface{}{"name": "worker", "type": "ai-service"}}, "spec", "components")
		}, phase: "running", healthy: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			app := listJobFixture()
			if tt.modify != nil {
				tt.modify(app)
			}
			item, ok := applicationListItem(app)
			if !ok || item.Phase != tt.phase || item.Healthy != tt.healthy {
				t.Fatalf("list status = %#v, want phase %q healthy %v", item, tt.phase, tt.healthy)
			}
		})
	}
}

func TestListApplicationsExposesPurposeWithoutAdditionalReads(t *testing.T) {
	app := listJobFixture()
	client := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), map[schema.GroupVersionResource]string{applicationGVR: "ApplicationList"}, app)
	items, err := NewClient(client).ListApplications(context.Background(), "tenant-a")
	if err != nil || len(items) != 1 {
		t.Fatalf("list = %#v, %v", items, err)
	}
	if items[0].Phase != "succeeded" || items[0].AIMetadata["ai.oam.dev/job-kind"] != "evaluation" {
		t.Fatalf("missing lifecycle or purpose: %#v", items[0])
	}
	if actions := client.Actions(); len(actions) != 1 || actions[0].GetVerb() != "list" || actions[0].GetResource() != applicationGVR {
		t.Fatalf("unexpected Kubernetes reads: %#v", actions)
	}
}

func TestApplicationListPurposeAnnotations(t *testing.T) {
	app := listJobFixture()
	_ = unstructured.SetNestedSlice(app.Object, []interface{}{map[string]interface{}{
		"name": "worker", "type": "ai-service", "properties": map[string]interface{}{
			"annotations": map[string]interface{}{"ai.oam.dev/purpose": "agent"},
		},
	}}, "spec", "components")
	item, _ := applicationListItem(app)
	if item.AIMetadata["ai.oam.dev/purpose"] != "agent" {
		t.Fatalf("missing annotation purpose: %#v", item)
	}
	app.SetAnnotations(map[string]string{"ai.oam.dev/purpose": "service"})
	item, _ = applicationListItem(app)
	if item.AIMetadata["ai.oam.dev/purpose"] != "service" {
		t.Fatalf("application annotation must take precedence: %#v", item)
	}
}

func listJobFixture() *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "core.oam.dev/v1beta1", "kind": "Application",
		"metadata": map[string]interface{}{"name": "eval", "namespace": "tenant-a", "generation": int64(1)},
		"spec": map[string]interface{}{"components": []interface{}{map[string]interface{}{
			"name": "worker", "type": "ai-job", "properties": map[string]interface{}{"jobKind": "evaluation", "completions": int64(3)},
		}}},
		"status": map[string]interface{}{"status": "running", "observedGeneration": int64(1),
			"services": []interface{}{map[string]interface{}{"name": "worker", "healthy": true}},
		},
	}}
}
