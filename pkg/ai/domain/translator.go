package domain

import (
	"fmt"

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
