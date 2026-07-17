package observe

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const mebibyte = int64(1024 * 1024)

type ResourceSummary struct {
	CPUMilli int64 `json:"cpuMilli,omitempty"`
	MemoryMi int64 `json:"memoryMi,omitempty"`
	GPU      int64 `json:"gpu,omitempty"`
}

func applicationResourceSummary(app *unstructured.Unstructured) ResourceSummary {
	var total ResourceSummary
	rawComponents, _, _ := unstructured.NestedSlice(app.Object, "spec", "components")
	for _, rawComponent := range rawComponents {
		component, ok := rawComponent.(map[string]interface{})
		if !ok {
			continue
		}
		multiplier := componentResourceMultiplier(component)
		if multiplier <= 0 {
			continue
		}
		summary := componentResourceSummary(component)
		total.CPUMilli += summary.CPUMilli * multiplier
		total.MemoryMi += summary.MemoryMi * multiplier
		total.GPU += summary.GPU * multiplier
	}
	return total
}

func componentResourceSummary(component map[string]interface{}) ResourceSummary {
	properties, ok := nestedMap(component, "properties")
	if !ok {
		return ResourceSummary{}
	}
	requests, ok := nestedMap(properties, "resources", "requests")
	if !ok {
		return ResourceSummary{}
	}
	return ResourceSummary{
		CPUMilli: parseQuantityMilli(resourceValue(requests, "cpu")),
		MemoryMi: parseMemoryMi(resourceValue(requests, "memory")),
		GPU:      parseQuantityValue(resourceValue(requests, "nvidia.com/gpu", "gpu")),
	}
}

func componentResourceMultiplier(component map[string]interface{}) int64 {
	componentType := stringFromMap(component, "type")
	properties, ok := nestedMap(component, "properties")
	if !ok {
		return 1
	}
	switch componentType {
	case "ai-service":
		if replicas := intFromMap(properties, "replicas"); replicas > 0 {
			return replicas
		}
	case "ai-job":
		if parallelism := intFromMap(properties, "parallelism"); parallelism > 0 {
			return parallelism
		}
	}
	return 1
}

func resourceValue(values map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		raw, ok := values[key]
		if !ok || raw == nil {
			continue
		}
		value := strings.TrimSpace(fmt.Sprint(raw))
		if value != "" && value != "<nil>" {
			return value
		}
	}
	return ""
}

func parseQuantityMilli(raw string) int64 {
	q, ok := parseQuantity(raw)
	if !ok {
		return 0
	}
	return q.MilliValue()
}

func parseMemoryMi(raw string) int64 {
	q, ok := parseQuantity(raw)
	if !ok {
		return 0
	}
	bytes := q.Value()
	if bytes <= 0 {
		return 0
	}
	return (bytes + mebibyte - 1) / mebibyte
}

func parseQuantityValue(raw string) int64 {
	q, ok := parseQuantity(raw)
	if !ok {
		return 0
	}
	return q.Value()
}

func parseQuantity(raw string) (resource.Quantity, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return resource.Quantity{}, false
	}
	q, err := resource.ParseQuantity(raw)
	if err != nil {
		return resource.Quantity{}, false
	}
	return q, true
}
