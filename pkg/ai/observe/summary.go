package observe

import (
	"fmt"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const aiMetadataPrefix = "ai.oam.dev/"

type Summary struct {
	Name       string             `json:"name"`
	Namespace  string             `json:"namespace,omitempty"`
	Phase      string             `json:"phase,omitempty"`
	Healthy    bool               `json:"healthy"`
	Message    string             `json:"message,omitempty"`
	Components []ComponentSummary `json:"components,omitempty"`
	AIMetadata map[string]string  `json:"aiMetadata,omitempty"`
	Warnings   []Warning          `json:"warnings,omitempty"`
}

type ComponentSummary struct {
	Name         string          `json:"name"`
	Type         string          `json:"type,omitempty"`
	WorkloadKind string          `json:"workloadKind,omitempty"`
	Healthy      bool            `json:"healthy"`
	Message      string          `json:"message,omitempty"`
	Workload     WorkloadSummary `json:"workload,omitempty"`
	Service      ServiceSummary  `json:"service,omitempty"`
	Pods         []PodSummary    `json:"pods,omitempty"`
}

type WorkloadSummary struct {
	Name                    string `json:"name,omitempty"`
	Kind                    string `json:"kind,omitempty"`
	DesiredReplicas         int64  `json:"desiredReplicas,omitempty"`
	ReadyReplicas           int64  `json:"readyReplicas,omitempty"`
	Active                  int64  `json:"active,omitempty"`
	Succeeded               int64  `json:"succeeded,omitempty"`
	Failed                  int64  `json:"failed,omitempty"`
	TTLSecondsAfterFinished int64  `json:"ttlSecondsAfterFinished,omitempty"`
}

type ServiceSummary struct {
	Name      string        `json:"name,omitempty"`
	ClusterIP string        `json:"clusterIP,omitempty"`
	Ports     []ServicePort `json:"ports,omitempty"`
}

type ServicePort struct {
	Port       int64 `json:"port,omitempty"`
	TargetPort int64 `json:"targetPort,omitempty"`
}

type PodSummary struct {
	Name            string `json:"name"`
	Phase           string `json:"phase,omitempty"`
	ReadyContainers int64  `json:"readyContainers"`
	TotalContainers int64  `json:"totalContainers"`
}

type Warning struct {
	Resource string `json:"resource,omitempty"`
	Message  string `json:"message"`
}

// SummarizeObjects builds a readonly AI workload summary from already-fetched Kubernetes objects.
func SummarizeObjects(objects []*unstructured.Unstructured) (*Summary, error) {
	app := findApplication(objects)
	if app == nil {
		return nil, fmt.Errorf("Application object is required")
	}

	summary := &Summary{
		Name:       app.GetName(),
		Namespace:  app.GetNamespace(),
		Phase:      nestedString(app.Object, "status", "status"),
		AIMetadata: map[string]string{},
	}
	mergeAIMetadata(summary.AIMetadata, app)

	services, _, _ := unstructured.NestedSlice(app.Object, "status", "services")
	for _, item := range services {
		serviceMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		component := ComponentSummary{
			Name:    stringFromMap(serviceMap, "name"),
			Healthy: boolFromMap(serviceMap, "healthy"),
			Message: stringFromMap(serviceMap, "message"),
		}
		if component.Name == "" {
			continue
		}
		component.Type = componentType(objects, summary.Name, component.Name)
		component.WorkloadKind = nestedString(serviceMap, "workloadDefinition", "kind")
		if component.Type == "" {
			component.Type = componentTypeFromWorkloadKind(component.WorkloadKind)
		}
		component.Workload = summarizeWorkload(objects, summary.Name, component.Name, component.WorkloadKind)
		component.Service = summarizeService(objects, summary.Name, component.Name)
		component.Pods = summarizePods(objects, summary.Name, component.Name, &summary.Warnings)
		summary.Components = append(summary.Components, component)
		summary.Healthy = summary.Healthy || component.Healthy
		if summary.Message == "" && component.Message != "" {
			summary.Message = component.Message
		}
	}

	for _, obj := range objects {
		mergeAIMetadata(summary.AIMetadata, obj)
	}
	sort.Slice(summary.Components, func(i, j int) bool {
		return summary.Components[i].Name < summary.Components[j].Name
	})
	return summary, nil
}

func findApplication(objects []*unstructured.Unstructured) *unstructured.Unstructured {
	for _, obj := range objects {
		if obj.GetKind() == "Application" {
			return obj
		}
	}
	return nil
}

func componentType(objects []*unstructured.Unstructured, appName, componentName string) string {
	for _, obj := range objects {
		if !belongsToComponent(obj, appName, componentName) {
			continue
		}
		if value := obj.GetLabels()["workload.oam.dev/type"]; value != "" {
			return value
		}
		if value := obj.GetLabels()["ai.oam.dev/workload-kind"]; value == "service" {
			return "ai-service"
		}
		if value := obj.GetLabels()["ai.oam.dev/workload-kind"]; value == "job" {
			return "ai-job"
		}
	}
	return ""
}

func componentTypeFromWorkloadKind(kind string) string {
	switch kind {
	case "Deployment":
		return "ai-service"
	case "Job":
		return "ai-job"
	default:
		return ""
	}
}

func summarizeWorkload(objects []*unstructured.Unstructured, appName, componentName, workloadKind string) WorkloadSummary {
	for _, obj := range objects {
		if obj.GetKind() != workloadKind || !belongsToComponent(obj, appName, componentName) {
			continue
		}
		summary := WorkloadSummary{Name: obj.GetName(), Kind: obj.GetKind()}
		switch obj.GetKind() {
		case "Deployment":
			summary.DesiredReplicas = nestedInt64(obj.Object, "spec", "replicas")
			summary.ReadyReplicas = nestedInt64(obj.Object, "status", "readyReplicas")
		case "Job":
			summary.Active = nestedInt64(obj.Object, "status", "active")
			summary.Succeeded = nestedInt64(obj.Object, "status", "succeeded")
			summary.Failed = nestedInt64(obj.Object, "status", "failed")
			summary.TTLSecondsAfterFinished = nestedInt64(obj.Object, "spec", "ttlSecondsAfterFinished")
		}
		return summary
	}
	return WorkloadSummary{}
}

func summarizeService(objects []*unstructured.Unstructured, appName, componentName string) ServiceSummary {
	for _, obj := range objects {
		if obj.GetKind() != "Service" || !belongsToComponent(obj, appName, componentName) {
			continue
		}
		summary := ServiceSummary{
			Name:      obj.GetName(),
			ClusterIP: nestedString(obj.Object, "spec", "clusterIP"),
		}
		ports, _, _ := unstructured.NestedSlice(obj.Object, "spec", "ports")
		for _, item := range ports {
			portMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			summary.Ports = append(summary.Ports, ServicePort{
				Port:       intFromMap(portMap, "port"),
				TargetPort: intFromMap(portMap, "targetPort"),
			})
		}
		return summary
	}
	return ServiceSummary{}
}

func summarizePods(objects []*unstructured.Unstructured, appName, componentName string, warnings *[]Warning) []PodSummary {
	var pods []PodSummary
	for _, obj := range objects {
		if obj.GetKind() != "Pod" || !belongsToComponent(obj, appName, componentName) {
			continue
		}
		pod := PodSummary{
			Name:  obj.GetName(),
			Phase: nestedString(obj.Object, "status", "phase"),
		}
		statuses, _, _ := unstructured.NestedSlice(obj.Object, "status", "containerStatuses")
		pod.TotalContainers = int64(len(statuses))
		for _, item := range statuses {
			statusMap, ok := item.(map[string]interface{})
			if ok && boolFromMap(statusMap, "ready") {
				pod.ReadyContainers++
			}
		}
		if (pod.Phase == "Succeeded" || pod.Phase == "Failed") && len(obj.GetOwnerReferences()) == 0 {
			*warnings = append(*warnings, Warning{
				Resource: "Pod/" + obj.GetName(),
				Message:  "completed or failed Job pod has empty ownerReferences and may remain after Job cleanup",
			})
		}
		pods = append(pods, pod)
	}
	sort.Slice(pods, func(i, j int) bool {
		return pods[i].Name < pods[j].Name
	})
	return pods
}

func belongsToComponent(obj *unstructured.Unstructured, appName, componentName string) bool {
	labels := obj.GetLabels()
	return labels["app.oam.dev/name"] == appName && labels["app.oam.dev/component"] == componentName
}

func mergeAIMetadata(target map[string]string, obj *unstructured.Unstructured) {
	for key, value := range obj.GetLabels() {
		if strings.HasPrefix(key, aiMetadataPrefix) {
			target[key] = value
		}
	}
	for key, value := range obj.GetAnnotations() {
		if strings.HasPrefix(key, aiMetadataPrefix) {
			target[key] = value
		}
	}
	if template, ok, _ := unstructured.NestedMap(obj.Object, "spec", "template", "metadata", "labels"); ok {
		mergeStringMap(target, template)
	}
	if template, ok, _ := unstructured.NestedMap(obj.Object, "spec", "template", "metadata", "annotations"); ok {
		mergeStringMap(target, template)
	}
}

func mergeStringMap(target map[string]string, values map[string]interface{}) {
	for key, value := range values {
		if strings.HasPrefix(key, aiMetadataPrefix) {
			target[key] = fmt.Sprint(value)
		}
	}
}

func nestedString(root map[string]interface{}, fields ...string) string {
	value, _, _ := unstructured.NestedString(root, fields...)
	return value
}

func nestedInt64(root map[string]interface{}, fields ...string) int64 {
	value, _, _ := unstructured.NestedInt64(root, fields...)
	return value
}

func stringFromMap(values map[string]interface{}, key string) string {
	value, _ := values[key].(string)
	return value
}

func boolFromMap(values map[string]interface{}, key string) bool {
	value, _ := values[key].(bool)
	return value
}

func intFromMap(values map[string]interface{}, key string) int64 {
	switch value := values[key].(type) {
	case int64:
		return value
	case int:
		return int64(value)
	case float64:
		return int64(value)
	default:
		return 0
	}
}
