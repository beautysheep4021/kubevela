package observe

import (
	"strings"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

func TestAuditConfigMapNameIsBounded(t *testing.T) {
	name := auditConfigMapName("2026-05-22T08:09:00.000000000Z-delete-sock-shop-ai-job-demo-with-extra-long-name")

	if len(name) > 63 {
		t.Fatalf("name length = %d, want <= 63", len(name))
	}
	if !strings.HasPrefix(name, "ai-audit-") {
		t.Fatalf("name = %q, want ai-audit prefix", name)
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
