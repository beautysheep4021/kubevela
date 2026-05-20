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
	Name            string             `json:"name"`
	Phase           string             `json:"phase,omitempty"`
	ReadyContainers int64              `json:"readyContainers"`
	TotalContainers int64              `json:"totalContainers"`
	Containers      []ContainerSummary `json:"containers,omitempty"`
	Events          []EventSummary     `json:"events,omitempty"`
}

type ContainerSummary struct {
	Name         string `json:"name"`
	Ready        bool   `json:"ready"`
	RestartCount int64  `json:"restartCount"`
	State        string `json:"state,omitempty"`
	Reason       string `json:"reason,omitempty"`
	Message      string `json:"message,omitempty"`
	ExitCode     int64  `json:"exitCode,omitempty"`
}

type EventSummary struct {
	Type    string `json:"type,omitempty"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
	Count   int64  `json:"count,omitempty"`
	LastAt  string `json:"lastAt,omitempty"`
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
		workload := findWorkload(objects, summary.Name, component.Name, component.WorkloadKind)
		component.Workload = summarizeWorkload(workload)
		component.Service = summarizeService(objects, summary.Name, component.Name)
		component.Pods = summarizePods(objects, summary.Name, component.Name, component.WorkloadKind, workload, &summary.Warnings)
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

func findWorkload(objects []*unstructured.Unstructured, appName, componentName, workloadKind string) *unstructured.Unstructured {
	for _, obj := range objects {
		if obj.GetKind() != workloadKind || !belongsToComponent(obj, appName, componentName) {
			continue
		}
		return obj
	}
	return nil
}

func summarizeWorkload(obj *unstructured.Unstructured) WorkloadSummary {
	if obj == nil {
		return WorkloadSummary{}
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

func shouldIncludePod(obj *unstructured.Unstructured, workloadKind string, workload *unstructured.Unstructured) (bool, string) {
	if workloadKind != "Job" {
		return true, ""
	}
	if workload == nil {
		if len(obj.GetOwnerReferences()) == 0 {
			return false, "Job pod has empty ownerReferences and may be historical residue because the current Job object is not present"
		}
		return false, "Job pod is not counted because the current Job object is not present"
	}
	for _, owner := range obj.GetOwnerReferences() {
		if owner.Kind == "Job" && owner.UID == workload.GetUID() {
			return true, ""
		}
	}
	if len(obj.GetOwnerReferences()) == 0 {
		return false, "Job pod has empty ownerReferences and is not counted as part of the current Job"
	}
	return false, "Job pod ownerReferences do not point to the current Job"
}

func orphanWarningMessage(phase string) string {
	if phase == "Succeeded" || phase == "Failed" {
		return "completed or failed Job pod has empty ownerReferences and may remain after Job cleanup"
	}
	return "Job pod has empty ownerReferences and may be historical residue from an earlier Job run"
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

func summarizePods(objects []*unstructured.Unstructured, appName, componentName, workloadKind string, workload *unstructured.Unstructured, warnings *[]Warning) []PodSummary {
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
			if !ok {
				continue
			}
			container := summarizeContainer(statusMap)
			pod.Containers = append(pod.Containers, container)
			if container.Ready {
				pod.ReadyContainers++
			}
		}
		include, warning := shouldIncludePod(obj, workloadKind, workload)
		if !include {
			*warnings = append(*warnings, Warning{
				Resource: "Pod/" + obj.GetName(),
				Message:  warning,
			})
			continue
		}
		if (pod.Phase == "Succeeded" || pod.Phase == "Failed") && len(obj.GetOwnerReferences()) == 0 {
			*warnings = append(*warnings, Warning{
				Resource: "Pod/" + obj.GetName(),
				Message:  orphanWarningMessage(pod.Phase),
			})
		}
		pod.Events = summarizeEvents(objects, obj)
		pods = append(pods, pod)
	}
	sort.Slice(pods, func(i, j int) bool {
		return pods[i].Name < pods[j].Name
	})
	return pods
}

func summarizeEvents(objects []*unstructured.Unstructured, pod *unstructured.Unstructured) []EventSummary {
	var events []EventSummary
	for _, obj := range objects {
		if obj.GetKind() != "Event" || nestedString(obj.Object, "involvedObject", "kind") != "Pod" {
			continue
		}
		if nestedString(obj.Object, "involvedObject", "namespace") != pod.GetNamespace() || nestedString(obj.Object, "involvedObject", "name") != pod.GetName() {
			continue
		}
		events = append(events, EventSummary{
			Type:    nestedString(obj.Object, "type"),
			Reason:  nestedString(obj.Object, "reason"),
			Message: nestedString(obj.Object, "message"),
			Count:   nestedInt64(obj.Object, "count"),
			LastAt:  eventLastAt(obj),
		})
	}
	sort.Slice(events, func(i, j int) bool {
		return events[i].LastAt < events[j].LastAt
	})
	return events
}

func eventLastAt(obj *unstructured.Unstructured) string {
	for _, field := range []string{"lastTimestamp", "eventTime", "metadata.creationTimestamp"} {
		if field == "metadata.creationTimestamp" {
			if value := nestedString(obj.Object, "metadata", "creationTimestamp"); value != "" {
				return value
			}
			continue
		}
		if value := nestedString(obj.Object, field); value != "" {
			return value
		}
	}
	return ""
}

func summarizeContainer(status map[string]interface{}) ContainerSummary {
	container := ContainerSummary{
		Name:         stringFromMap(status, "name"),
		Ready:        boolFromMap(status, "ready"),
		RestartCount: intFromMap(status, "restartCount"),
	}
	if waiting, ok := nestedMap(status, "state", "waiting"); ok {
		container.State = "waiting"
		container.Reason = stringFromMap(waiting, "reason")
		container.Message = stringFromMap(waiting, "message")
		return container
	}
	if terminated, ok := nestedMap(status, "state", "terminated"); ok {
		container.State = "terminated"
		container.Reason = stringFromMap(terminated, "reason")
		container.Message = stringFromMap(terminated, "message")
		container.ExitCode = intFromMap(terminated, "exitCode")
		return container
	}
	if _, ok := nestedMap(status, "state", "running"); ok {
		container.State = "running"
		return container
	}
	return container
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

func nestedMap(root map[string]interface{}, fields ...string) (map[string]interface{}, bool) {
	value, ok, _ := unstructured.NestedMap(root, fields...)
	return value, ok
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
