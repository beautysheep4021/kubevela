package observe

import (
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestSummarizeObjectsAddsDiagnosticsForFailedScheduling(t *testing.T) {
	summary, err := SummarizeObjects([]*unstructured.Unstructured{
		applicationObject("demo", "sock-shop", "trainer", "ai-service", "Deployment"),
		podObjectWithReason("demo-trainer-pod", "sock-shop", "demo", "trainer", "Pending", "trainer", "Waiting", "ErrImagePull", "pull image failed"),
		eventObject("demo-trainer-pod.1", "sock-shop", "Pod", "demo-trainer-pod", "Warning", "FailedScheduling", "0/1 nodes are available"),
	})
	if err != nil {
		t.Fatalf("summarize: %v", err)
	}

	if len(summary.Diagnostics) == 0 {
		t.Fatalf("expected diagnostics, got none: %#v", summary)
	}
	joined := diagnosticsText(summary.Diagnostics)
	if !strings.Contains(joined, "FailedScheduling") || !strings.Contains(joined, "ErrImagePull") {
		t.Fatalf("expected scheduling and image diagnostics, got %s", joined)
	}
}

func diagnosticsText(items []Diagnostic) string {
	var parts []string
	for _, item := range items {
		parts = append(parts, item.Reason+" "+item.Message+" "+item.Evidence)
	}
	return strings.Join(parts, "\n")
}

func applicationObject(name, namespace, componentName, componentType, workloadKind string) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "core.oam.dev/v1beta1",
		"kind":       "Application",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{
			"components": []interface{}{
				map[string]interface{}{"name": componentName, "type": componentType},
			},
		},
		"status": map[string]interface{}{
			"status": "running",
			"services": []interface{}{
				map[string]interface{}{
					"name":    componentName,
					"healthy": false,
					"workloadDefinition": map[string]interface{}{
						"kind": workloadKind,
					},
				},
			},
		},
	}}
}

func podObjectWithReason(name, namespace, appName, componentName, phase, containerName, state, reason, message string) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Pod",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
			"labels": map[string]interface{}{
				"app.oam.dev/name":      appName,
				"app.oam.dev/component": componentName,
			},
		},
		"spec": map[string]interface{}{
			"containers": []interface{}{map[string]interface{}{"name": containerName}},
		},
		"status": map[string]interface{}{
			"phase": phase,
			"containerStatuses": []interface{}{
				map[string]interface{}{
					"name":  containerName,
					"ready": false,
					"state": map[string]interface{}{
						strings.ToLower(state): map[string]interface{}{
							"reason":  reason,
							"message": message,
						},
					},
				},
			},
		},
	}}
}

func eventObject(name, namespace, involvedKind, involvedName, eventType, reason, message string) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Event",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"type":    eventType,
		"reason":  reason,
		"message": message,
		"involvedObject": map[string]interface{}{
			"kind":      involvedKind,
			"name":      involvedName,
			"namespace": namespace,
		},
	}}
}
