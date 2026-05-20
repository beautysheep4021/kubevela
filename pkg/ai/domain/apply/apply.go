package apply

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"
)

var ApplicationGVR = schema.GroupVersionResource{Group: "core.oam.dev", Version: "v1beta1", Resource: "applications"}

type Options struct {
	DryRun bool
}

type Result struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	DryRun    bool   `json:"dryRun"`
}

type ApplicationApplier interface {
	ApplyApplication(ctx context.Context, namespace, name string, content []byte, opts metav1.PatchOptions) (*unstructured.Unstructured, error)
}

type DynamicApplicationApplier struct {
	Client dynamic.Interface
}

func (a DynamicApplicationApplier) ApplyApplication(ctx context.Context, namespace, name string, content []byte, opts metav1.PatchOptions) (*unstructured.Unstructured, error) {
	return a.Client.Resource(ApplicationGVR).Namespace(namespace).Patch(ctx, name, types.ApplyPatchType, content, opts)
}

func ApplyYAML(ctx context.Context, client dynamic.Interface, content []byte, opts Options) (Result, error) {
	return ApplyYAMLWithApplier(ctx, DynamicApplicationApplier{Client: client}, content, opts)
}

func ApplyYAMLWithApplier(ctx context.Context, applier ApplicationApplier, content []byte, opts Options) (Result, error) {
	obj := &unstructured.Unstructured{}
	if err := yaml.Unmarshal(content, obj); err != nil {
		return Result{}, fmt.Errorf("decode Application YAML: %w", err)
	}
	if obj.GetAPIVersion() != "core.oam.dev/v1beta1" || obj.GetKind() != "Application" {
		return Result{}, fmt.Errorf("expected core.oam.dev/v1beta1 Application, got %s %s", obj.GetAPIVersion(), obj.GetKind())
	}
	namespace := obj.GetNamespace()
	if namespace == "" {
		namespace = "default"
		obj.SetNamespace(namespace)
	}
	applyOptions := metav1.PatchOptions{FieldManager: "ai-domain"}
	if opts.DryRun {
		applyOptions.DryRun = []string{metav1.DryRunAll}
	}
	_, err := applier.ApplyApplication(ctx, namespace, obj.GetName(), content, applyOptions)
	if err != nil {
		return Result{}, fmt.Errorf("apply Application %s/%s: %w", namespace, obj.GetName(), err)
	}
	return Result{Name: obj.GetName(), Namespace: namespace, DryRun: opts.DryRun}, nil
}
