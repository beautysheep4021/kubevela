package observe

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

func TestTenantResourcesQuotaSemantics(t *testing.T) {
	for _, tc := range []struct {
		name     string
		status   corev1.ResourceQuotaStatus
		scopes   []corev1.ResourceQuotaScope
		selector *corev1.ScopeSelector
		want     string
	}{
		{name: "desired differs from enforced", status: corev1.ResourceQuotaStatus{
			Hard: corev1.ResourceList{"requests.cpu": resource.MustParse("1500m")},
			Used: corev1.ResourceList{"requests.cpu": resource.MustParse("250m")},
		}, want: `{"name":"compute","hard":{"requests.cpu":"1500m"},"used":{"requests.cpu":"250m"},"desiredHard":{"requests.cpu":"2"}}`},
		{name: "status absent", want: `{"name":"compute","hard":{},"used":{},"desiredHard":{"requests.cpu":"2"}}`},
		{name: "scoped quota", scopes: []corev1.ResourceQuotaScope{corev1.ResourceQuotaScopeNotTerminating},
			selector: &corev1.ScopeSelector{MatchExpressions: []corev1.ScopedResourceSelectorRequirement{{
				ScopeName: corev1.ResourceQuotaScopePriorityClass, Operator: corev1.ScopeSelectorOpIn, Values: []string{"high", "critical"},
			}}},
			want: `{"name":"compute","hard":{},"used":{},"desiredHard":{"requests.cpu":"2"},"scopes":["NotTerminating"],"scopeSelector":{"matchExpressions":[{"scopeName":"PriorityClass","operator":"In","values":["high","critical"]}]}}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			kube := fake.NewSimpleClientset(
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "alpha"}},
				&corev1.ResourceQuota{ObjectMeta: metav1.ObjectMeta{Name: "compute", Namespace: "alpha"},
					Spec: corev1.ResourceQuotaSpec{Hard: corev1.ResourceList{"requests.cpu": resource.MustParse("2")}, Scopes: tc.scopes, ScopeSelector: tc.selector}, Status: tc.status},
			)
			items, err := (Client{Kube: kube}).ListTenantResources(context.Background(), "alpha")
			if err != nil {
				t.Fatal(err)
			}
			if len(items) != 1 || len(items[0].Quotas) != 1 {
				t.Fatalf("items=%#v", items)
			}
			got, err := json.Marshal(items[0].Quotas[0])
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tc.want {
				t.Fatalf("quota=%s\nwant=%s", got, tc.want)
			}
		})
	}
}

func TestTenantResourcesList(t *testing.T) {
	for _, namespace := range []string{"", "alpha", "empty", "missing"} {
		t.Run("namespace="+namespace, func(t *testing.T) {
			kube := fake.NewSimpleClientset(
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "empty"}},
				&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "alpha"}},
				&corev1.ResourceQuota{ObjectMeta: metav1.ObjectMeta{Name: "z", Namespace: "alpha"}},
				&corev1.ResourceQuota{ObjectMeta: metav1.ObjectMeta{Name: "compute", Namespace: "alpha"},
					Spec: corev1.ResourceQuotaSpec{Hard: corev1.ResourceList{
						"requests.cpu": resource.MustParse("1500m"), "requests.memory": resource.MustParse("2Gi"),
						"requests.nvidia.com/gpu": resource.MustParse("4"),
					}},
					Status: corev1.ResourceQuotaStatus{Used: corev1.ResourceList{"requests.cpu": resource.MustParse("250m")}}},
			)
			items, err := (Client{Kube: kube}).ListTenantResources(context.Background(), namespace)
			if err != nil {
				t.Fatal(err)
			}
			want := []TenantResources{}
			if namespace == "" || namespace == "alpha" {
				want = append(want, TenantResources{Namespace: "alpha", Quotas: []TenantQuota{
					{Name: "compute", Hard: map[string]string{}, DesiredHard: map[string]string{"requests.cpu": "1500m", "requests.memory": "2Gi", "requests.nvidia.com/gpu": "4"}, Used: map[string]string{"requests.cpu": "250m"}},
					{Name: "z", Hard: map[string]string{}, Used: map[string]string{}, DesiredHard: map[string]string{}},
				}})
			}
			if namespace == "" || namespace == "empty" {
				want = append(want, TenantResources{Namespace: "empty", Quotas: []TenantQuota{}})
			}
			if !reflect.DeepEqual(items, want) {
				t.Fatalf("items = %#v, want %#v", items, want)
			}
			actions := kube.Actions()
			if len(actions) != 2 || !actions[0].Matches("list", "namespaces") || !actions[1].Matches("list", "resourcequotas") || actions[1].GetNamespace() != namespace {
				t.Fatalf("expected namespace LIST then scoped quota LIST, got %#v", actions)
			}
		})
	}
}

func TestTenantResourcesReadErrors(t *testing.T) {
	if _, err := (Client{}).ListTenantResources(context.Background(), ""); err == nil {
		t.Fatal("nil client should fail")
	}
	for _, resourceName := range []string{"namespaces", "resourcequotas"} {
		t.Run(resourceName, func(t *testing.T) {
			kube := fake.NewSimpleClientset()
			failure := errors.New("read denied")
			kube.PrependReactor("list", resourceName, func(ktesting.Action) (bool, runtime.Object, error) { return true, nil, failure })
			items, err := (Client{Kube: kube}).ListTenantResources(context.Background(), "")
			if !errors.Is(err, failure) || items != nil {
				t.Fatalf("items=%#v error=%v", items, err)
			}
			if resourceName == "namespaces" && len(kube.Actions()) != 1 {
				t.Fatal("quota read after namespace failure")
			}
		})
	}
}
