package observe

import (
	"context"
	"fmt"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	applicationGVR = schema.GroupVersionResource{Group: "core.oam.dev", Version: "v1beta1", Resource: "applications"}
	deploymentGVR  = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	jobGVR         = schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "jobs"}
	serviceGVR     = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
	podGVR         = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	eventGVR       = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "events"}
)

type Client struct {
	Dynamic dynamic.Interface
}

type ApplicationListItem struct {
	Name          string            `json:"name"`
	Namespace     string            `json:"namespace,omitempty"`
	Phase         string            `json:"phase,omitempty"`
	Healthy       bool              `json:"healthy"`
	Message       string            `json:"message,omitempty"`
	Components    []ComponentRef    `json:"components,omitempty"`
	WorkloadTypes []string          `json:"workloadTypes,omitempty"`
	AIMetadata    map[string]string `json:"aiMetadata,omitempty"`
	Warnings      []Warning         `json:"warnings,omitempty"`
}

type ComponentRef struct {
	Name         string `json:"name"`
	Type         string `json:"type,omitempty"`
	WorkloadType string `json:"workloadType,omitempty"`
}

func NewClient(client dynamic.Interface) Client {
	return Client{Dynamic: client}
}

// SummarizeApplication reads a KubeVela Application and related native resources using GET/LIST only.
func SummarizeApplication(ctx context.Context, kubeconfig, namespace, name string) (*Summary, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace is required")
	}
	if name == "" {
		return nil, fmt.Errorf("application name is required")
	}

	config, err := restConfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("create dynamic client: %w", err)
	}
	return NewClient(client).SummarizeApplication(ctx, namespace, name)
}

func (c Client) ListApplications(ctx context.Context, namespace string) ([]ApplicationListItem, error) {
	if c.Dynamic == nil {
		return nil, fmt.Errorf("dynamic client is required")
	}
	list, err := c.Dynamic.Resource(applicationGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list Applications: %w", err)
	}
	var items []ApplicationListItem
	for i := range list.Items {
		item, ok := applicationListItem(&list.Items[i])
		if ok {
			items = append(items, item)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Namespace == items[j].Namespace {
			return items[i].Name < items[j].Name
		}
		return items[i].Namespace < items[j].Namespace
	})
	return items, nil
}

func (c Client) SummarizeApplication(ctx context.Context, namespace, name string) (*Summary, error) {
	if c.Dynamic == nil {
		return nil, fmt.Errorf("dynamic client is required")
	}
	if namespace == "" {
		return nil, fmt.Errorf("namespace is required")
	}
	if name == "" {
		return nil, fmt.Errorf("application name is required")
	}

	app, err := c.Dynamic.Resource(applicationGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get Application %s/%s: %w", namespace, name, err)
	}

	objects := []*unstructured.Unstructured{app}
	for _, gvr := range []schema.GroupVersionResource{deploymentGVR, jobGVR, serviceGVR, podGVR} {
		list, err := c.Dynamic.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: "app.oam.dev/name=" + name,
		})
		if err != nil {
			return nil, fmt.Errorf("list %s for Application %s/%s: %w", gvr.Resource, namespace, name, err)
		}
		for i := range list.Items {
			item := list.Items[i]
			objects = append(objects, &item)
		}
	}
	eventList, err := c.Dynamic.Resource(eventGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list events for Application %s/%s: %w", namespace, name, err)
	}
	for i := range eventList.Items {
		item := eventList.Items[i]
		objects = append(objects, &item)
	}
	return SummarizeObjects(objects)
}

func applicationListItem(app *unstructured.Unstructured) (ApplicationListItem, bool) {
	components, workloadTypes := applicationComponents(app)
	if len(components) == 0 {
		return ApplicationListItem{}, false
	}
	item := ApplicationListItem{
		Name:          app.GetName(),
		Namespace:     app.GetNamespace(),
		Phase:         nestedString(app.Object, "status", "status"),
		Components:    components,
		WorkloadTypes: workloadTypes,
		AIMetadata:    map[string]string{},
	}
	mergeAIMetadata(item.AIMetadata, app)
	services, _, _ := unstructured.NestedSlice(app.Object, "status", "services")
	for _, service := range services {
		serviceMap, ok := service.(map[string]interface{})
		if !ok {
			continue
		}
		if boolFromMap(serviceMap, "healthy") {
			item.Healthy = true
		}
		if item.Message == "" {
			item.Message = stringFromMap(serviceMap, "message")
		}
	}
	return item, true
}

func applicationComponents(app *unstructured.Unstructured) ([]ComponentRef, []string) {
	rawComponents, _, _ := unstructured.NestedSlice(app.Object, "spec", "components")
	var components []ComponentRef
	seenWorkloadTypes := map[string]bool{}
	for _, raw := range rawComponents {
		componentMap, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		componentType := stringFromMap(componentMap, "type")
		workloadType := listWorkloadType(componentType)
		if workloadType == "" {
			continue
		}
		components = append(components, ComponentRef{
			Name:         stringFromMap(componentMap, "name"),
			Type:         componentType,
			WorkloadType: workloadType,
		})
		seenWorkloadTypes[workloadType] = true
	}
	var workloadTypes []string
	for workloadType := range seenWorkloadTypes {
		workloadTypes = append(workloadTypes, workloadType)
	}
	sort.Strings(workloadTypes)
	return components, workloadTypes
}

func listWorkloadType(componentType string) string {
	switch componentType {
	case "ai-service":
		return "service"
	case "ai-job":
		return "job"
	default:
		return ""
	}
}

func restConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("load kubeconfig %s: %w", kubeconfig, err)
		}
		return config, nil
	}
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}
	config, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("load kubeconfig: %w", err)
	}
	return config, nil
}
