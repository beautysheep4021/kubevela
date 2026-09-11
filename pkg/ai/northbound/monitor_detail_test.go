package northbound

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/oam-dev/kubevela/pkg/ai/observe"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type missingMonitorApplication struct{ recordingApplicationReader }

func (r *missingMonitorApplication) SummarizeApplication(_ context.Context, _, name string) (*observe.Summary, error) {
	return nil, apierrors.NewNotFound(schema.GroupResource{Group: "core.oam.dev", Resource: "applications"}, name)
}

func TestMonitorDeletedApplicationDetailReturnsNotFound(t *testing.T) {
	server := authenticatedServer(t, NewServerWithOptions(Options{Reader: &missingMonitorApplication{}}))
	w := httptest.NewRecorder()
	server.ServeHTTP(w, httptest.NewRequest("GET", "/api/v1/ai/applications/ai-tenant-a/deleted/status", nil))
	if w.Code != 404 {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}
