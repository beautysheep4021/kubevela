package observe

import (
	"context"
	"fmt"
	"sort"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TenantResources struct {
	Namespace string        `json:"namespace"`
	Quotas    []TenantQuota `json:"quotas"`
}

type TenantQuota struct {
	Name          string                      `json:"name"`
	Hard          map[string]string           `json:"hard"`
	Used          map[string]string           `json:"used"`
	DesiredHard   map[string]string           `json:"desiredHard"`
	Scopes        []corev1.ResourceQuotaScope `json:"scopes,omitempty"`
	ScopeSelector *corev1.ScopeSelector       `json:"scopeSelector,omitempty"`
}

// ListTenantResources includes namespaces without quotas and only reports usage
// recorded by the quota controller. Two LIST calls avoid per-namespace fan-out.
func (c Client) ListTenantResources(ctx context.Context, namespace string) ([]TenantResources, error) {
	if c.Kube == nil {
		return nil, fmt.Errorf("kubernetes client is required for tenant resources")
	}
	namespaces, err := c.Kube.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list namespaces: %w", err)
	}
	quotas, err := c.Kube.CoreV1().ResourceQuotas(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list resource quotas: %w", err)
	}
	byNamespace := make(map[string][]TenantQuota)
	for _, quota := range quotas.Items {
		byNamespace[quota.Namespace] = append(byNamespace[quota.Namespace], TenantQuota{
			Name: quota.Name, Hard: quotaQuantityStrings(quota.Status.Hard), Used: quotaQuantityStrings(quota.Status.Used),
			DesiredHard: quotaQuantityStrings(quota.Spec.Hard),
			Scopes:      quota.Spec.Scopes, ScopeSelector: quota.Spec.ScopeSelector,
		})
	}
	items := make([]TenantResources, 0, len(namespaces.Items))
	for _, ns := range namespaces.Items {
		if namespace != "" && ns.Name != namespace {
			continue
		}
		nsQuotas := byNamespace[ns.Name]
		if nsQuotas == nil {
			nsQuotas = []TenantQuota{}
		}
		sort.Slice(nsQuotas, func(i, j int) bool { return nsQuotas[i].Name < nsQuotas[j].Name })
		items = append(items, TenantResources{Namespace: ns.Name, Quotas: nsQuotas})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Namespace < items[j].Namespace })
	return items, nil
}

func quotaQuantityStrings(resources corev1.ResourceList) map[string]string {
	values := make(map[string]string, len(resources))
	for name, quantity := range resources {
		values[string(name)] = quantity.String()
	}
	return values
}
