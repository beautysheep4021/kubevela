package observe

import (
	"context"
	"fmt"

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

	app, err := client.Resource(applicationGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get Application %s/%s: %w", namespace, name, err)
	}

	objects := []*unstructured.Unstructured{app}
	for _, gvr := range []schema.GroupVersionResource{deploymentGVR, jobGVR, serviceGVR, podGVR} {
		list, err := client.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{
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
	eventList, err := client.Resource(eventGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list events for Application %s/%s: %w", namespace, name, err)
	}
	for i := range eventList.Items {
		item := eventList.Items[i]
		objects = append(objects, &item)
	}
	return SummarizeObjects(objects)
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
