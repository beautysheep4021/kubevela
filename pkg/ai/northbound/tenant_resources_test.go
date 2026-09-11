package northbound

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oam-dev/kubevela/pkg/ai/observe"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

type recordingTenantResourceReader struct {
	namespace string
	calls     int
	items     []observe.TenantResources
	err       error
}

func TestTenantResourcesEndpointQuotaContract(t *testing.T) {
	kube := fake.NewSimpleClientset(
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "alpha"}},
		&corev1.ResourceQuota{ObjectMeta: metav1.ObjectMeta{Name: "compute", Namespace: "alpha"},
			Spec: corev1.ResourceQuotaSpec{
				Hard:          corev1.ResourceList{"requests.cpu": resource.MustParse("2")},
				Scopes:        []corev1.ResourceQuotaScope{corev1.ResourceQuotaScopeNotTerminating},
				ScopeSelector: &corev1.ScopeSelector{MatchExpressions: []corev1.ScopedResourceSelectorRequirement{{ScopeName: corev1.ResourceQuotaScopePriorityClass, Operator: corev1.ScopeSelectorOpIn, Values: []string{"high"}}}},
			},
			Status: corev1.ResourceQuotaStatus{Hard: corev1.ResourceList{"requests.cpu": resource.MustParse("1500m")}, Used: corev1.ResourceList{"requests.cpu": resource.MustParse("250m")}},
		},
	)
	server := authenticatedServer(t, NewServerWithOptions(Options{TenantResources: observe.Client{Kube: kube}}))
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest("GET", "/api/v1/ai/tenant-resources", nil))
	want := `{"items":[{"namespace":"alpha","quotas":[{"name":"compute","hard":{"requests.cpu":"1500m"},"used":{"requests.cpu":"250m"},"desiredHard":{"requests.cpu":"2"},"scopes":["NotTerminating"],"scopeSelector":{"matchExpressions":[{"scopeName":"PriorityClass","operator":"In","values":["high"]}]}}]}]}`
	if recorder.Code != 200 || strings.TrimSpace(recorder.Body.String()) != want {
		t.Fatalf("status=%d body=%s\nwant=%s", recorder.Code, recorder.Body.String(), want)
	}
}

func (r *recordingTenantResourceReader) ListTenantResources(_ context.Context, namespace string) ([]observe.TenantResources, error) {
	r.calls++
	r.namespace = namespace
	return r.items, r.err
}

func TestTenantResourcesEndpoint(t *testing.T) {
	for _, tc := range []struct {
		name, role, method string
		nilReader, fail    bool
		status             int
	}{
		{"anonymous", "", "GET", false, false, 401},
		{"user", "user", "GET", false, false, 403},
		{"user without reader", "user", "GET", true, false, 403},
		{"monitor without reader", "monitor", "GET", true, false, 503},
		{"monitor read failure", "monitor", "GET", false, true, 502},
		{"monitor", "monitor", "GET", false, false, 200},
		{"read only", "monitor", "POST", false, false, 405},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reader := &recordingTenantResourceReader{items: []observe.TenantResources{{Namespace: "alpha", Quotas: []observe.TenantQuota{{Name: "compute", Hard: map[string]string{"requests.cpu": "1500m"}, Used: map[string]string{}, DesiredHard: map[string]string{"requests.cpu": "2"}}}}}}
			if tc.fail {
				reader.err = errors.New("quota access denied")
			}
			options := Options{}
			if !tc.nilReader {
				options.TenantResources = reader
			}
			server := NewServerWithOptions(options)
			if tc.role == "monitor" {
				server = authenticatedServer(t, server)
			}
			if tc.role == "user" {
				server = authenticatedServerAs(t, server, "admin", "shiyong")
			}
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, httptest.NewRequest(tc.method, "/api/v1/ai/tenant-resources?namespace=alpha", nil))
			if recorder.Code != tc.status {
				t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
			}
			wantCalls := 0
			if tc.status == 200 || tc.fail {
				wantCalls = 1
			}
			if reader.calls != wantCalls {
				t.Fatalf("reader calls=%d, want %d", reader.calls, wantCalls)
			}
			if wantCalls == 1 && reader.namespace != "alpha" {
				t.Fatalf("namespace=%q", reader.namespace)
			}
			if tc.status == 200 && strings.TrimSpace(recorder.Body.String()) != `{"items":[{"namespace":"alpha","quotas":[{"name":"compute","hard":{"requests.cpu":"1500m"},"used":{},"desiredHard":{"requests.cpu":"2"}}]}]}` {
				t.Fatalf("body=%s", recorder.Body.String())
			}
		})
	}
}

func TestTenantResourcesEndpointEmpty(t *testing.T) {
	reader := &recordingTenantResourceReader{}
	server := authenticatedServer(t, NewServerWithOptions(Options{TenantResources: reader}))
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest("GET", "/api/v1/ai/tenant-resources", nil))
	if recorder.Code != 200 || strings.TrimSpace(recorder.Body.String()) != `{"items":[]}` || reader.namespace != "" {
		t.Fatalf("status=%d body=%s namespace=%q", recorder.Code, recorder.Body.String(), reader.namespace)
	}
}
