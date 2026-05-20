package domain

import (
	"fmt"
	"strconv"

	yamlv3 "go.yaml.in/yaml/v3"
	"sigs.k8s.io/yaml"
)

const (
	domainAPIVersion = "ai.oam.dev/v1alpha1"
	oamAPIVersion    = "core.oam.dev/v1beta1"
)

type ValidationResult struct {
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

type NormalizedObject struct {
	APIVersion       string           `json:"apiVersion"`
	Kind             string           `json:"kind"`
	Name             string           `json:"name"`
	Namespace        string           `json:"namespace,omitempty"`
	ComponentName    string           `json:"componentName"`
	WorkloadType     string           `json:"workloadType"`
	Image            string           `json:"image"`
	Runtime          string           `json:"runtime,omitempty"`
	GovernanceIntent GovernanceIntent `json:"governanceIntent"`
	WorkloadIntent   WorkloadIntent   `json:"workloadIntent"`
}

type GovernanceIntent struct {
	Framework   string          `json:"framework,omitempty"`
	Tenant      string          `json:"tenant,omitempty"`
	Project     string          `json:"project,omitempty"`
	Environment string          `json:"environment,omitempty"`
	Owner       string          `json:"owner,omitempty"`
	ModelURI    string          `json:"modelURI,omitempty"`
	DatasetURI  string          `json:"datasetURI,omitempty"`
	Placement   PlacementIntent `json:"placement,omitempty"`
}

type PlacementIntent struct {
	Namespace string   `json:"namespace,omitempty"`
	Clusters  []string `json:"clusters,omitempty"`
}

type WorkloadIntent struct {
	Service *ServiceIntent `json:"service,omitempty"`
	Job     *JobIntent     `json:"job,omitempty"`
}

type ServiceIntent struct {
	Replicas *int64         `json:"replicas,omitempty"`
	Model    ModelIntent    `json:"model,omitempty"`
	Endpoint EndpointIntent `json:"endpoint,omitempty"`
}

type ModelIntent struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
	URI     string `json:"uri,omitempty"`
}

type EndpointIntent struct {
	Port        int64  `json:"port,omitempty"`
	ServicePort int64  `json:"servicePort,omitempty"`
	TargetPort  int64  `json:"targetPort,omitempty"`
	Type        string `json:"type,omitempty"`
}

type JobIntent struct {
	JobKind                 string     `json:"jobKind,omitempty"`
	Dataset                 DatasetRef `json:"dataset,omitempty"`
	Output                  OutputRef  `json:"output,omitempty"`
	TTLSecondsAfterFinished *int64     `json:"ttlSecondsAfterFinished,omitempty"`
	Parallelism             *int64     `json:"parallelism,omitempty"`
	Completions             *int64     `json:"completions,omitempty"`
	BackoffLimit            *int64     `json:"backoffLimit,omitempty"`
}

type DatasetRef struct {
	Name string `json:"name,omitempty"`
	URI  string `json:"uri,omitempty"`
}

type OutputRef struct {
	URI string `json:"uri,omitempty"`
}

// TranslateYAML converts one AI domain document into one native KubeVela Application.
func TranslateYAML(in []byte) ([]byte, error) {
	doc, err := decodeDomainDocument(in)
	if err != nil {
		return nil, err
	}
	result := validateDocument(doc)
	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("invalid AI domain YAML: %s", result.Errors[0])
	}

	componentType, err := componentTypeForKind(doc.Kind)
	if err != nil {
		return nil, err
	}
	return translateDocument(doc, componentType)
}

func ValidateYAML(in []byte) (ValidationResult, error) {
	doc, err := decodeDomainDocument(in)
	if err != nil {
		return ValidationResult{}, err
	}
	return validateDocument(doc), nil
}

func NormalizeYAML(in []byte) (NormalizedObject, error) {
	doc, err := decodeDomainDocument(in)
	if err != nil {
		return NormalizedObject{}, err
	}
	result := validateDocument(doc)
	if len(result.Errors) > 0 {
		return NormalizedObject{}, fmt.Errorf("invalid AI domain YAML: %s", result.Errors[0])
	}
	componentType, err := componentTypeForKind(doc.Kind)
	if err != nil {
		return NormalizedObject{}, err
	}
	componentName := doc.Spec.ComponentName
	if componentName == "" {
		componentName = doc.Metadata.Name
	}
	normalized := NormalizedObject{
		APIVersion:       doc.APIVersion,
		Kind:             doc.Kind,
		Name:             doc.Metadata.Name,
		Namespace:        doc.Metadata.Namespace,
		ComponentName:    componentName,
		WorkloadType:     workloadType(componentType),
		Image:            stringValue(doc.Spec.Properties, "image"),
		Runtime:          stringValue(doc.Spec.Runtime, "runtime"),
		GovernanceIntent: governanceIntent(doc),
		WorkloadIntent:   workloadIntent(doc, componentType),
	}
	return normalized, nil
}

func decodeDomainDocument(in []byte) (domainDocument, error) {
	var doc domainDocument
	if err := yamlv3.Unmarshal(in, &doc); err != nil {
		return domainDocument{}, fmt.Errorf("decode AI domain YAML: %w", err)
	}
	return doc, nil
}

func validateDocument(doc domainDocument) ValidationResult {
	var result ValidationResult
	if doc.APIVersion != domainAPIVersion {
		result.Errors = append(result.Errors, fmt.Sprintf("unsupported apiVersion %q", doc.APIVersion))
	}
	if doc.Metadata.Name == "" {
		result.Errors = append(result.Errors, "metadata.name is required")
	}
	componentType, err := componentTypeForKind(doc.Kind)
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result
	}
	if doc.Spec.ComponentName == "" {
		result.Warnings = append(result.Warnings, "spec.componentName is empty; metadata.name will be used as the component name")
	}
	if _, ok := doc.Spec.Properties["image"]; !ok {
		result.Errors = append(result.Errors, "spec.properties.image is required")
	}
	if componentType == "ai-job" {
		if _, ok := doc.Spec.Properties["jobKind"]; !ok {
			result.Errors = append(result.Errors, "spec.properties.jobKind is required for AIJob")
		}
	}
	if len(doc.Spec.Runtime) > 0 {
		if _, ok := doc.Spec.Runtime["runtime"]; !ok {
			result.Errors = append(result.Errors, "spec.runtime.runtime is required when spec.runtime is set")
		}
	}
	return result
}

func translateDocument(doc domainDocument, componentType string) ([]byte, error) {
	componentName := doc.Spec.ComponentName
	if componentName == "" {
		componentName = doc.Metadata.Name
	}

	component := map[string]interface{}{
		"name":       componentName,
		"type":       componentType,
		"properties": mapOrEmpty(doc.Spec.Properties),
	}
	if len(doc.Spec.Runtime) > 0 {
		component["traits"] = []interface{}{
			map[string]interface{}{
				"type":       "ai-runtime",
				"properties": doc.Spec.Runtime,
			},
		}
	}

	app := map[string]interface{}{
		"apiVersion": oamAPIVersion,
		"kind":       "Application",
		"metadata": map[string]interface{}{
			"name": doc.Metadata.Name,
		},
		"spec": map[string]interface{}{
			"components": []interface{}{component},
		},
	}
	if doc.Metadata.Namespace != "" {
		app["metadata"].(map[string]interface{})["namespace"] = doc.Metadata.Namespace
	}

	if doc.Spec.Placement.hasPlacement() {
		appSpec := app["spec"].(map[string]interface{})
		appSpec["policies"] = []interface{}{topologyPolicy(doc.Spec.Placement)}
		appSpec["workflow"] = map[string]interface{}{
			"steps": []interface{}{
				map[string]interface{}{
					"name": "deploy-ai-workload",
					"type": "deploy",
					"properties": map[string]interface{}{
						"policies": []interface{}{"local-topology"},
					},
				},
			},
		}
	}

	out, err := yaml.Marshal(app)
	if err != nil {
		return nil, fmt.Errorf("encode KubeVela Application YAML: %w", err)
	}
	return out, nil
}

type domainDocument struct {
	APIVersion string     `yaml:"apiVersion"`
	Kind       string     `yaml:"kind"`
	Metadata   metadata   `yaml:"metadata"`
	Spec       domainSpec `yaml:"spec"`
}

type metadata struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

type domainSpec struct {
	ComponentName string                 `yaml:"componentName"`
	Properties    map[string]interface{} `yaml:"properties"`
	Runtime       map[string]interface{} `yaml:"runtime"`
	Placement     placement              `yaml:"placement"`
}

type placement struct {
	Namespace string   `yaml:"namespace"`
	Clusters  []string `yaml:"clusters"`
}

func componentTypeForKind(kind string) (string, error) {
	switch kind {
	case "AIService":
		return "ai-service", nil
	case "AIJob":
		return "ai-job", nil
	case "AIWorkflow":
		return "", fmt.Errorf("unsupported kind AIWorkflow: AIWorkflow is out of first PoC scope")
	default:
		return "", fmt.Errorf("unsupported kind %s", kind)
	}
}

func mapOrEmpty(values map[string]interface{}) map[string]interface{} {
	if values == nil {
		return map[string]interface{}{}
	}
	return values
}

func (p placement) hasPlacement() bool {
	return p.Namespace != "" || len(p.Clusters) > 0
}

func topologyPolicy(p placement) map[string]interface{} {
	properties := map[string]interface{}{}
	if p.Namespace != "" {
		properties["namespace"] = p.Namespace
	}
	if len(p.Clusters) > 0 {
		clusters := make([]interface{}, 0, len(p.Clusters))
		for _, cluster := range p.Clusters {
			clusters = append(clusters, cluster)
		}
		properties["clusters"] = clusters
	}
	return map[string]interface{}{
		"name":       "local-topology",
		"type":       "topology",
		"properties": properties,
	}
}

func workloadType(componentType string) string {
	switch componentType {
	case "ai-service":
		return "service"
	case "ai-job":
		return "job"
	default:
		return componentType
	}
}

func governanceIntent(doc domainDocument) GovernanceIntent {
	return GovernanceIntent{
		Framework:   stringValue(doc.Spec.Runtime, "framework"),
		Tenant:      stringValue(doc.Spec.Runtime, "tenant"),
		Project:     stringValue(doc.Spec.Runtime, "project"),
		Environment: stringValue(doc.Spec.Runtime, "environment"),
		Owner:       stringValue(doc.Spec.Runtime, "owner"),
		ModelURI:    stringValue(doc.Spec.Runtime, "modelURI"),
		DatasetURI:  stringValue(doc.Spec.Runtime, "datasetURI"),
		Placement: PlacementIntent{
			Namespace: doc.Spec.Placement.Namespace,
			Clusters:  append([]string(nil), doc.Spec.Placement.Clusters...),
		},
	}
}

func workloadIntent(doc domainDocument, componentType string) WorkloadIntent {
	switch componentType {
	case "ai-service":
		return WorkloadIntent{Service: &ServiceIntent{
			Replicas: intPtr(doc.Spec.Properties, "replicas"),
			Model: ModelIntent{
				Name:    stringValue(nestedMap(doc.Spec.Properties, "model"), "name"),
				Version: stringValue(nestedMap(doc.Spec.Properties, "model"), "version"),
				URI:     stringValue(nestedMap(doc.Spec.Properties, "model"), "uri"),
			},
			Endpoint: EndpointIntent{
				Port:        intValue(nestedMap(doc.Spec.Properties, "endpoint"), "port"),
				ServicePort: intValue(nestedMap(doc.Spec.Properties, "endpoint"), "servicePort"),
				TargetPort:  intValue(nestedMap(doc.Spec.Properties, "endpoint"), "targetPort"),
				Type:        stringValue(nestedMap(doc.Spec.Properties, "endpoint"), "type"),
			},
		}}
	case "ai-job":
		return WorkloadIntent{Job: &JobIntent{
			JobKind:                 stringValue(doc.Spec.Properties, "jobKind"),
			Dataset:                 DatasetRef{Name: stringValue(nestedMap(doc.Spec.Properties, "dataset"), "name"), URI: stringValue(nestedMap(doc.Spec.Properties, "dataset"), "uri")},
			Output:                  OutputRef{URI: stringValue(nestedMap(doc.Spec.Properties, "output"), "uri")},
			TTLSecondsAfterFinished: intPtr(doc.Spec.Properties, "ttlSecondsAfterFinished"),
			Parallelism:             intPtr(doc.Spec.Properties, "parallelism"),
			Completions:             intPtr(doc.Spec.Properties, "completions"),
			BackoffLimit:            intPtr(doc.Spec.Properties, "backoffLimit"),
		}}
	default:
		return WorkloadIntent{}
	}
}

func nestedMap(values map[string]interface{}, key string) map[string]interface{} {
	if values == nil {
		return nil
	}
	nested, _ := values[key].(map[string]interface{})
	return nested
}

func stringValue(values map[string]interface{}, key string) string {
	if values == nil {
		return ""
	}
	value, ok := values[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return fmt.Sprint(typed)
	}
}

func intValue(values map[string]interface{}, key string) int64 {
	if values == nil {
		return 0
	}
	value, ok := values[key]
	if !ok || value == nil {
		return 0
	}
	switch typed := value.(type) {
	case int:
		return int64(typed)
	case int64:
		return typed
	case uint64:
		return int64(typed)
	case float64:
		return int64(typed)
	case string:
		parsed, _ := strconv.ParseInt(typed, 10, 64)
		return parsed
	default:
		return 0
	}
}

func intPtr(values map[string]interface{}, key string) *int64 {
	if values == nil {
		return nil
	}
	if _, ok := values[key]; !ok {
		return nil
	}
	value := intValue(values, key)
	return &value
}
