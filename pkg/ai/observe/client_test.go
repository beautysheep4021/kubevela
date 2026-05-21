package observe

import (
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
