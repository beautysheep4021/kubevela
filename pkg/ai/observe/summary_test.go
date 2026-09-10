package observe

import (
	"encoding/json"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestSummarizeObjectsJobCompletionRequiresCompleteCondition(t *testing.T) {
	for _, tt := range []struct {
		name       string
		succeeded  int64
		conditions []interface{}
		completed  bool
	}{
		{name: "partial success", succeeded: 1},
		{name: "all pods succeeded without terminal condition", succeeded: 3},
		{name: "complete false", succeeded: 1, conditions: []interface{}{map[string]interface{}{"type": "Complete", "status": "False"}}},
		{name: "complete unknown", succeeded: 1, conditions: []interface{}{map[string]interface{}{"type": "Complete", "status": "Unknown"}}},
		{name: "success criteria not terminal", succeeded: 3, conditions: []interface{}{map[string]interface{}{"type": "SuccessCriteriaMet", "status": "True"}}},
		{name: "failed condition", succeeded: 1, conditions: []interface{}{map[string]interface{}{"type": "Failed", "status": "True"}}},
		{name: "confirmed completion", succeeded: 3, completed: true, conditions: []interface{}{
			map[string]interface{}{"type": "Suspended", "status": "False"},
			map[string]interface{}{"type": "Complete", "status": "True"},
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			job := &unstructured.Unstructured{Object: map[string]interface{}{
				"apiVersion": "batch/v1", "kind": "Job",
				"metadata": map[string]interface{}{"name": "demo-trainer", "namespace": "sock-shop",
					"labels": map[string]interface{}{"app.oam.dev/name": "demo", "app.oam.dev/component": "trainer"}},
				"spec":   map[string]interface{}{"completions": int64(3)},
				"status": map[string]interface{}{"succeeded": tt.succeeded},
			}}
			if tt.conditions != nil {
				if err := unstructured.SetNestedSlice(job.Object, tt.conditions, "status", "conditions"); err != nil {
					t.Fatal(err)
				}
			}
			summary, err := SummarizeObjects([]*unstructured.Unstructured{
				applicationObject("demo", "sock-shop", "trainer", "ai-job", "Job"), job,
			})
			if err != nil || len(summary.Components) != 1 {
				t.Fatalf("summary = %#v, error = %v", summary, err)
			}
			workload := summary.Components[0].Workload
			if workload.Succeeded != tt.succeeded {
				t.Fatalf("succeeded = %d, want %d", workload.Succeeded, tt.succeeded)
			}
			encoded, err := json.Marshal(workload)
			if err != nil {
				t.Fatal(err)
			}
			var payload map[string]interface{}
			if err := json.Unmarshal(encoded, &payload); err != nil {
				t.Fatal(err)
			}
			if payload["completed"] != tt.completed {
				t.Fatalf("workload JSON = %s, want completed=%v", encoded, tt.completed)
			}
		})
	}
}

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
