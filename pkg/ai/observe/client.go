package observe

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
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
	Kube    kubernetes.Interface
}

type ApplicationListItem struct {
	Name            string            `json:"name"`
	CreatedAt       string            `json:"createdAt,omitempty"`
	Namespace       string            `json:"namespace,omitempty"`
	Phase           string            `json:"phase,omitempty"`
	Healthy         bool              `json:"healthy"`
	Message         string            `json:"message,omitempty"`
	Components      []ComponentRef    `json:"components,omitempty"`
	WorkloadTypes   []string          `json:"workloadTypes,omitempty"`
	ResourceSummary ResourceSummary   `json:"resourceSummary,omitempty"`
	AIMetadata      map[string]string `json:"aiMetadata,omitempty"`
	Warnings        []Warning         `json:"warnings,omitempty"`
}

type ComponentRef struct {
	Name         string `json:"name"`
	Type         string `json:"type,omitempty"`
	WorkloadType string `json:"workloadType,omitempty"`
}

func NewClient(client dynamic.Interface) Client {
	return Client{Dynamic: client}
}

func NewClientWithKube(dynamicClient dynamic.Interface, kubeClient kubernetes.Interface) Client {
	return Client{Dynamic: dynamicClient, Kube: kubeClient}
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

func (c Client) GetApplicationLogs(ctx context.Context, options LogOptions) (*Logs, error) {
	if c.Dynamic == nil {
		return nil, fmt.Errorf("dynamic client is required")
	}
	if c.Kube == nil {
		return nil, fmt.Errorf("kubernetes client is required for logs")
	}
	if options.Namespace == "" {
		return nil, fmt.Errorf("namespace is required")
	}
	if options.Name == "" {
		return nil, fmt.Errorf("application name is required")
	}
	if options.TailLines <= 0 {
		options.TailLines = 200
	}
	if options.TailLines > 2000 {
		options.TailLines = 2000
	}

	pods, err := c.Dynamic.Resource(podGVR).Namespace(options.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "app.oam.dev/name=" + options.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("list pods for Application %s/%s: %w", options.Namespace, options.Name, err)
	}
	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("no pods found for Application %s/%s", options.Namespace, options.Name)
	}

	logPods := make([]LogPod, 0, len(pods.Items))
	for i := range pods.Items {
		logPods = append(logPods, logPodFromUnstructured(&pods.Items[i]))
	}
	sort.Slice(logPods, func(i, j int) bool {
		return logPods[i].Name < logPods[j].Name
	})

	selected := selectLogPod(pods.Items, options.Pod)
	if selected == nil {
		return nil, fmt.Errorf("pod %q was not found for Application %s/%s", options.Pod, options.Namespace, options.Name)
	}
	container := options.Container
	if container == "" {
		container = firstContainerName(selected)
	}
	if container == "" {
		return nil, fmt.Errorf("pod %s has no containers", selected.GetName())
	}

	stream, err := c.Kube.CoreV1().Pods(options.Namespace).GetLogs(selected.GetName(), &corev1.PodLogOptions{
		Container: container,
		TailLines: &options.TailLines,
	}).Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("read logs for Pod %s/%s container %s: %w", options.Namespace, selected.GetName(), container, err)
	}
	defer stream.Close()
	content, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("read log stream for Pod %s/%s container %s: %w", options.Namespace, selected.GetName(), container, err)
	}

	return &Logs{
		Namespace:   options.Namespace,
		Application: options.Name,
		Pod:         selected.GetName(),
		Container:   container,
		TailLines:   options.TailLines,
		Logs:        string(content),
		Pods:        logPods,
	}, nil
}

func (c Client) ProbeApplication(ctx context.Context, options ProbeOptions) (*ProbeResult, error) {
	if c.Dynamic == nil {
		return nil, fmt.Errorf("dynamic client is required")
	}
	if options.Namespace == "" {
		return nil, fmt.Errorf("namespace is required")
	}
	if options.Name == "" {
		return nil, fmt.Errorf("application name is required")
	}
	if options.Path == "" {
		options.Path = "/healthz"
	}
	if !strings.HasPrefix(options.Path, "/") {
		options.Path = "/" + options.Path
	}
	if options.TimeoutSeconds <= 0 {
		options.TimeoutSeconds = 5
	}
	if options.TimeoutSeconds > 30 {
		options.TimeoutSeconds = 30
	}
	services, err := c.Dynamic.Resource(serviceGVR).Namespace(options.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "app.oam.dev/name=" + options.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("list services for Application %s/%s: %w", options.Namespace, options.Name, err)
	}
	if len(services.Items) == 0 {
		return nil, fmt.Errorf("no services found for Application %s/%s", options.Namespace, options.Name)
	}
	service := selectProbeService(services.Items)
	port := firstServicePort(&service)
	if port == 0 {
		return nil, fmt.Errorf("service %s/%s has no ports", options.Namespace, service.GetName())
	}
	host := fmt.Sprintf("%s.%s.svc.cluster.local", service.GetName(), options.Namespace)
	if clusterIP := nestedString(service.Object, "spec", "clusterIP"); clusterIP != "" && clusterIP != "None" {
		host = clusterIP
	}
	target := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", host, port),
		Path:   options.Path,
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create probe request: %w", err)
	}
	client := &http.Client{Timeout: time.Duration(options.TimeoutSeconds) * time.Second}
	response, err := client.Do(request)
	result := &ProbeResult{
		Namespace:   options.Namespace,
		Application: options.Name,
		ServiceName: service.GetName(),
		URL:         target.String(),
		Path:        options.Path,
	}
	if err != nil {
		result.Error = err.Error()
		return result, nil
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 64*1024))
	if err != nil {
		return nil, fmt.Errorf("read probe response: %w", err)
	}
	result.StatusCode = response.StatusCode
	result.Healthy = response.StatusCode >= 200 && response.StatusCode < 400
	result.Body = string(body)
	return result, nil
}

func (c Client) DeleteApplication(ctx context.Context, namespace, name string) (*LifecycleResult, error) {
	if c.Dynamic == nil {
		return nil, fmt.Errorf("dynamic client is required")
	}
	if namespace == "" {
		return nil, fmt.Errorf("namespace is required")
	}
	if name == "" {
		return nil, fmt.Errorf("application name is required")
	}
	if err := c.Dynamic.Resource(applicationGVR).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return nil, fmt.Errorf("delete Application %s/%s: %w", namespace, name, err)
	}
	return &LifecycleResult{
		Action:    "delete",
		Namespace: namespace,
		Name:      name,
		Message:   "delete requested",
	}, nil
}

func (c Client) RestartApplication(ctx context.Context, namespace, name string) (*LifecycleResult, error) {
	if c.Dynamic == nil {
		return nil, fmt.Errorf("dynamic client is required")
	}
	if namespace == "" {
		return nil, fmt.Errorf("namespace is required")
	}
	if name == "" {
		return nil, fmt.Errorf("application name is required")
	}
	deployments, err := c.Dynamic.Resource(deploymentGVR).Namespace(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "app.oam.dev/name=" + name,
	})
	if err != nil {
		return nil, fmt.Errorf("list deployments for Application %s/%s: %w", namespace, name, err)
	}
	if len(deployments.Items) == 0 {
		return nil, fmt.Errorf("no deployments found for Application %s/%s", namespace, name)
	}
	restartedAt := time.Now().UTC().Format(time.RFC3339)
	for i := range deployments.Items {
		deploy := &deployments.Items[i]
		annotations, _, _ := unstructured.NestedStringMap(deploy.Object, "spec", "template", "metadata", "annotations")
		if annotations == nil {
			annotations = map[string]string{}
		}
		annotations["kubectl.kubernetes.io/restartedAt"] = restartedAt
		if err := unstructured.SetNestedStringMap(deploy.Object, annotations, "spec", "template", "metadata", "annotations"); err != nil {
			return nil, fmt.Errorf("set restart annotation for Deployment %s/%s: %w", namespace, deploy.GetName(), err)
		}
		if _, err := c.Dynamic.Resource(deploymentGVR).Namespace(namespace).Update(ctx, deploy, metav1.UpdateOptions{}); err != nil {
			return nil, fmt.Errorf("update Deployment %s/%s for restart: %w", namespace, deploy.GetName(), err)
		}
	}
	return &LifecycleResult{
		Action:    "restart",
		Namespace: namespace,
		Name:      name,
		Message:   fmt.Sprintf("restart requested for %d deployment(s)", len(deployments.Items)),
	}, nil
}

func (c Client) RerunApplication(ctx context.Context, namespace, name string) (*LifecycleResult, error) {
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
	if !applicationHasWorkloadType(app, "ai-job") {
		return nil, fmt.Errorf("Application %s/%s is not an AIJob", namespace, name)
	}
	rerunName := rerunApplicationName(name, time.Now().UTC())
	copy := app.DeepCopy()
	copy.SetResourceVersion("")
	copy.SetUID("")
	copy.SetGeneration(0)
	copy.SetManagedFields(nil)
	copy.SetFinalizers(nil)
	copy.SetName(rerunName)
	copy.SetCreationTimestamp(metav1.Time{})
	copy.SetAnnotations(withString(copy.GetAnnotations(), "ai.oam.dev/rerun-from", name))
	unstructured.RemoveNestedField(copy.Object, "status")
	if _, err := c.Dynamic.Resource(applicationGVR).Namespace(namespace).Create(ctx, copy, metav1.CreateOptions{}); err != nil {
		return nil, fmt.Errorf("create rerun Application %s/%s from %s: %w", namespace, rerunName, name, err)
	}
	return &LifecycleResult{
		Action:    "rerun",
		Namespace: namespace,
		Name:      rerunName,
		Message:   "rerun application created from " + name,
	}, nil
}

func (c Client) RecordAudit(ctx context.Context, event AuditEvent) error {
	if c.Kube == nil {
		return fmt.Errorf("kubernetes client is required for audit")
	}
	if event.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}
	if event.ID == "" {
		event.ID = auditEventID(event)
	}
	if event.Time == "" {
		event.Time = time.Now().UTC().Format(time.RFC3339)
	}
	content, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("encode audit event: %w", err)
	}
	name := auditConfigMapName(event.ID)
	_, err = c.Kube.CoreV1().ConfigMaps(event.Namespace).Create(ctx, &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: event.Namespace,
			Labels: map[string]string{
				"ai.oam.dev/audit":  "true",
				"ai.oam.dev/action": event.Action,
				"ai.oam.dev/name":   event.Name,
			},
		},
		Data: map[string]string{"event.json": string(content)},
	}, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("create audit ConfigMap %s/%s: %w", event.Namespace, name, err)
	}
	return nil
}

func (c Client) ListAudits(ctx context.Context, namespace string) ([]AuditEvent, error) {
	if c.Kube == nil {
		return nil, fmt.Errorf("kubernetes client is required for audit")
	}
	list, err := c.Kube.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "ai.oam.dev/audit=true",
	})
	if err != nil {
		return nil, fmt.Errorf("list audit ConfigMaps: %w", err)
	}
	items := make([]AuditEvent, 0, len(list.Items))
	for i := range list.Items {
		raw := list.Items[i].Data["event.json"]
		if raw == "" {
			continue
		}
		var event AuditEvent
		if err := json.Unmarshal([]byte(raw), &event); err != nil {
			continue
		}
		items = append(items, event)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Time > items[j].Time
	})
	return items, nil
}

func (c Client) SaveArtifact(ctx context.Context, artifact ModelArtifact) error {
	if c.Kube == nil {
		return fmt.Errorf("kubernetes client is required for artifacts")
	}
	if artifact.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}
	if artifact.JobName == "" {
		return fmt.Errorf("job name is required")
	}
	if artifact.Name == "" {
		artifact.Name = artifact.JobName
	}
	if artifact.CreatedAt == "" {
		artifact.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	content, err := json.Marshal(artifact)
	if err != nil {
		return fmt.Errorf("encode model artifact: %w", err)
	}
	name := artifactConfigMapName(artifact.Name)
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: artifact.Namespace,
			Labels: map[string]string{
				"ai.oam.dev/artifact": "true",
				"ai.oam.dev/job":      artifact.JobName,
			},
		},
		Data: map[string]string{"artifact.json": string(content)},
	}
	existing, err := c.Kube.CoreV1().ConfigMaps(artifact.Namespace).Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		configMap.ResourceVersion = existing.ResourceVersion
		if _, err := c.Kube.CoreV1().ConfigMaps(artifact.Namespace).Update(ctx, configMap, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("update artifact ConfigMap %s/%s: %w", artifact.Namespace, name, err)
		}
		return nil
	}
	if _, err := c.Kube.CoreV1().ConfigMaps(artifact.Namespace).Create(ctx, configMap, metav1.CreateOptions{}); err != nil {
		return fmt.Errorf("create artifact ConfigMap %s/%s: %w", artifact.Namespace, name, err)
	}
	return nil
}

func (c Client) ListArtifacts(ctx context.Context, namespace string) ([]ModelArtifact, error) {
	if c.Kube == nil {
		return nil, fmt.Errorf("kubernetes client is required for artifacts")
	}
	list, err := c.Kube.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "ai.oam.dev/artifact=true",
	})
	if err != nil {
		return nil, fmt.Errorf("list artifact ConfigMaps: %w", err)
	}
	items := make([]ModelArtifact, 0, len(list.Items))
	for i := range list.Items {
		raw := list.Items[i].Data["artifact.json"]
		if raw == "" {
			continue
		}
		var artifact ModelArtifact
		if err := json.Unmarshal([]byte(raw), &artifact); err != nil {
			continue
		}
		if artifact.Namespace == "" {
			artifact.Namespace = list.Items[i].Namespace
		}
		items = append(items, artifact)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt > items[j].CreatedAt
	})
	return items, nil
}

func (c Client) SaveDataset(ctx context.Context, dataset DatasetArtifact) error {
	if c.Kube == nil {
		return fmt.Errorf("kubernetes client is required for datasets")
	}
	if dataset.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}
	if dataset.Name == "" {
		return fmt.Errorf("dataset name is required")
	}
	if dataset.DatasetURI == "" {
		return fmt.Errorf("dataset URI is required")
	}
	if dataset.CreatedAt == "" {
		dataset.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	content, err := json.Marshal(dataset)
	if err != nil {
		return fmt.Errorf("encode dataset artifact: %w", err)
	}
	name := datasetConfigMapName(dataset.Name)
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: dataset.Namespace,
			Labels: map[string]string{
				"ai.oam.dev/dataset": "true",
			},
		},
		Data: map[string]string{"dataset.json": string(content)},
	}
	existing, err := c.Kube.CoreV1().ConfigMaps(dataset.Namespace).Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		configMap.ResourceVersion = existing.ResourceVersion
		if _, err := c.Kube.CoreV1().ConfigMaps(dataset.Namespace).Update(ctx, configMap, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("update dataset ConfigMap %s/%s: %w", dataset.Namespace, name, err)
		}
		return nil
	}
	if _, err := c.Kube.CoreV1().ConfigMaps(dataset.Namespace).Create(ctx, configMap, metav1.CreateOptions{}); err != nil {
		return fmt.Errorf("create dataset ConfigMap %s/%s: %w", dataset.Namespace, name, err)
	}
	return nil
}

func (c Client) ListDatasets(ctx context.Context, namespace string) ([]DatasetArtifact, error) {
	if c.Kube == nil {
		return nil, fmt.Errorf("kubernetes client is required for datasets")
	}
	list, err := c.Kube.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "ai.oam.dev/dataset=true",
	})
	if err != nil {
		return nil, fmt.Errorf("list dataset ConfigMaps: %w", err)
	}
	items := make([]DatasetArtifact, 0, len(list.Items))
	for i := range list.Items {
		raw := list.Items[i].Data["dataset.json"]
		if raw == "" {
			continue
		}
		var dataset DatasetArtifact
		if err := json.Unmarshal([]byte(raw), &dataset); err != nil {
			continue
		}
		if dataset.Namespace == "" {
			dataset.Namespace = list.Items[i].Namespace
		}
		items = append(items, dataset)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt > items[j].CreatedAt
	})
	return items, nil
}

func selectProbeService(services []unstructured.Unstructured) unstructured.Unstructured {
	sort.Slice(services, func(i, j int) bool {
		return services[i].GetName() < services[j].GetName()
	})
	return services[0]
}

func firstServicePort(service *unstructured.Unstructured) int64 {
	ports, _, _ := unstructured.NestedSlice(service.Object, "spec", "ports")
	for _, raw := range ports {
		portMap, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		if port := intFromMap(portMap, "port"); port > 0 {
			return port
		}
	}
	return 0
}

func applicationListItem(app *unstructured.Unstructured) (ApplicationListItem, bool) {
	components, workloadTypes := applicationComponents(app)
	if len(components) == 0 {
		return ApplicationListItem{}, false
	}
	item := ApplicationListItem{
		Name:            app.GetName(),
		CreatedAt:       nestedString(app.Object, "metadata", "creationTimestamp"),
		Namespace:       app.GetNamespace(),
		Phase:           nestedString(app.Object, "status", "status"),
		Components:      components,
		WorkloadTypes:   workloadTypes,
		ResourceSummary: applicationResourceSummary(app),
		AIMetadata:      map[string]string{},
	}
	mergeAIMetadata(item.AIMetadata, app)
	mergeRuntimeTraitMetadata(item.AIMetadata, app)
	mergeComponentPurposeMetadata(item.AIMetadata, app)
	healthy := map[string]bool{}
	completed := map[string]bool{}
	for _, component := range components {
		healthy[component.Name] = true
		completed[component.Name] = component.Type == "ai-job"
	}
	seen := map[string]bool{}
	services, _, _ := unstructured.NestedSlice(app.Object, "status", "services")
	for _, service := range services {
		serviceMap, ok := service.(map[string]interface{})
		if !ok {
			continue
		}
		name := stringFromMap(serviceMap, "name")
		if _, relevant := healthy[name]; !relevant {
			continue
		}
		seen[name] = true
		healthy[name] = healthy[name] && boolFromMap(serviceMap, "healthy")
		workloadHealthy, present := serviceMap["workloadHealthy"].(bool)
		if !present {
			workloadHealthy = boolFromMap(serviceMap, "healthy")
		}
		completed[name] = completed[name] && workloadHealthy
		if item.Message == "" {
			item.Message = stringFromMap(serviceMap, "message")
		}
	}
	observedGeneration, _, _ := unstructured.NestedInt64(app.Object, "status", "observedGeneration")
	current := observedGeneration == app.GetGeneration()
	item.Healthy = current
	rawComponents, _, _ := unstructured.NestedSlice(app.Object, "spec", "components")
	allCompleted := current && len(rawComponents) == len(components)
	for _, component := range components {
		item.Healthy = item.Healthy && seen[component.Name] && healthy[component.Name]
		allCompleted = allCompleted && seen[component.Name] && completed[component.Name]
	}
	// ai-job health means succeeded == spec.completions, not merely a running pod.
	// Preserve other Application phases, including explicit failure and in-progress workflows.
	if allCompleted && item.Phase == "running" && app.GetDeletionTimestamp() == nil {
		item.Phase = "succeeded"
	}
	return item, true
}

func mergeComponentPurposeMetadata(target map[string]string, app *unstructured.Unstructured) {
	values := map[string]string{}
	conflicts := map[string]bool{}
	components, _, _ := unstructured.NestedSlice(app.Object, "spec", "components")
	for _, raw := range components {
		component, ok := raw.(map[string]interface{})
		if !ok || listWorkloadType(stringFromMap(component, "type")) == "" {
			continue
		}
		for _, key := range []string{"purpose", "job-kind", "task-type"} {
			metadataKey := aiMetadataPrefix + key
			value := nestedString(component, "properties", "annotations", metadataKey)
			if value == "" && key == "job-kind" && stringFromMap(component, "type") == "ai-job" {
				value = nestedString(component, "properties", "jobKind")
			}
			if value == "" {
				continue
			}
			if previous := values[metadataKey]; previous != "" && previous != value {
				conflicts[metadataKey] = true
			}
			values[metadataKey] = value
		}
	}
	for key, value := range values {
		if target[key] == "" && !conflicts[key] {
			target[key] = value
		}
	}
}

func logPodFromUnstructured(pod *unstructured.Unstructured) LogPod {
	item := LogPod{
		Name:  pod.GetName(),
		Phase: nestedString(pod.Object, "status", "phase"),
	}
	containers, _, _ := unstructured.NestedSlice(pod.Object, "spec", "containers")
	for _, raw := range containers {
		container, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		if name := stringFromMap(container, "name"); name != "" {
			item.Containers = append(item.Containers, name)
		}
	}
	return item
}

func selectLogPod(pods []unstructured.Unstructured, podName string) *unstructured.Unstructured {
	var selected *unstructured.Unstructured
	for i := range pods {
		pod := &pods[i]
		if podName != "" {
			if pod.GetName() == podName {
				return pod
			}
			continue
		}
		if selected == nil {
			selected = pod
			continue
		}
		podCreatedAt := pod.GetCreationTimestamp().Time
		selectedCreatedAt := selected.GetCreationTimestamp().Time
		if podCreatedAt.After(selectedCreatedAt) || (podCreatedAt.Equal(selectedCreatedAt) && pod.GetName() > selected.GetName()) {
			selected = pod
		}
	}
	return selected
}

func firstContainerName(pod *unstructured.Unstructured) string {
	containers, _, _ := unstructured.NestedSlice(pod.Object, "spec", "containers")
	for _, raw := range containers {
		container, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		if name := stringFromMap(container, "name"); name != "" {
			return name
		}
	}
	return ""
}

func applicationHasWorkloadType(app *unstructured.Unstructured, componentType string) bool {
	components, _, _ := unstructured.NestedSlice(app.Object, "spec", "components")
	for _, raw := range components {
		component, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		if stringFromMap(component, "type") == componentType {
			return true
		}
	}
	return false
}

func rerunApplicationName(name string, at time.Time) string {
	base := name
	maxBaseLen := 44
	if len(base) > maxBaseLen {
		base = base[:maxBaseLen]
	}
	return strings.ToLower(base + "-rerun-" + at.Format("0102150405"))
}

func withString(values map[string]string, key, value string) map[string]string {
	if values == nil {
		values = map[string]string{}
	}
	values[key] = value
	return values
}

func auditConfigMapName(id string) string {
	name := "ai-audit-" + strings.ToLower(strings.NewReplacer("_", "-", ".", "-", ":", "-", "/", "-").Replace(id))
	if len(name) > 63 {
		name = name[:63]
	}
	return strings.Trim(name, "-")
}

func auditEventID(event AuditEvent) string {
	base := strings.Join([]string{event.Time, event.Action, event.Namespace, event.Name}, "-")
	base = strings.NewReplacer(":", "-", ".", "-", "/", "-", " ", "-").Replace(base)
	if base == "" {
		return time.Now().UTC().Format("20060102T150405.000000000Z")
	}
	return base
}

func artifactConfigMapName(name string) string {
	value := "ai-artifact-" + strings.ToLower(strings.NewReplacer("_", "-", ".", "-", ":", "-", "/", "-").Replace(name))
	if len(value) > 63 {
		value = value[:63]
	}
	return strings.Trim(value, "-")
}

func datasetConfigMapName(name string) string {
	value := "ai-dataset-" + strings.ToLower(strings.NewReplacer("_", "-", ".", "-", ":", "-", "/", "-").Replace(name))
	if len(value) > 63 {
		value = value[:63]
	}
	return strings.Trim(value, "-")
}

func mergeRuntimeTraitMetadata(target map[string]string, app *unstructured.Unstructured) {
	rawComponents, _, _ := unstructured.NestedSlice(app.Object, "spec", "components")
	for _, rawComponent := range rawComponents {
		component, ok := rawComponent.(map[string]interface{})
		if !ok {
			continue
		}
		rawTraits, _, _ := unstructured.NestedSlice(component, "traits")
		for _, rawTrait := range rawTraits {
			trait, ok := rawTrait.(map[string]interface{})
			if !ok || stringFromMap(trait, "type") != "ai-runtime" {
				continue
			}
			properties, ok := nestedMap(trait, "properties")
			if !ok {
				continue
			}
			for _, key := range []string{"runtime", "framework", "tenant", "project", "environment", "owner", "modelURI", "datasetURI"} {
				if value := stringFromMap(properties, key); value != "" {
					target[aiMetadataPrefix+metadataKey(key)] = value
				}
			}
		}
	}
}

func metadataKey(key string) string {
	switch key {
	case "modelURI":
		return "model-uri"
	case "datasetURI":
		return "dataset-uri"
	default:
		return key
	}
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
