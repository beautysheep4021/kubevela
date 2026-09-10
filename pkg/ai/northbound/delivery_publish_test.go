package northbound

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oam-dev/kubevela/pkg/ai/observe"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type deliveryPublishReader struct {
	*recordingApplicationReader
	summary         *observe.Summary
	lookupErr       error
	lookupNamespace string
	lookupName      string
}

func (r *deliveryPublishReader) SummarizeApplication(_ context.Context, namespace, name string) (*observe.Summary, error) {
	r.lookupNamespace, r.lookupName = namespace, name
	if r.summary != nil || r.lookupErr != nil {
		return r.summary, r.lookupErr
	}
	return nil, apierrors.NewNotFound(schema.GroupResource{Group: "core.oam.dev", Resource: "applications"}, name)
}

func TestPublishDeliveryServiceChecksTargetBeforeApply(t *testing.T) {
	application := schema.GroupResource{Group: "core.oam.dev", Resource: "applications"}
	for _, tc := range []struct {
		name       string
		body       string
		target     string
		summary    *observe.Summary
		lookupErr  error
		wantStatus int
	}{
		{name: "existing explicit target", body: `{"serviceName":"existing-app"}`, target: "existing-app", summary: &observe.Summary{}, wantStatus: http.StatusConflict},
		{name: "existing default target", body: `{}`, target: "train-demo-service", summary: &observe.Summary{}, wantStatus: http.StatusConflict},
		{name: "target is source job", body: `{"serviceName":"train-demo"}`, target: "train-demo", summary: &observe.Summary{}, wantStatus: http.StatusConflict},
		{name: "missing explicit target", body: `{"serviceName":"new-service"}`, target: "new-service", wantStatus: http.StatusOK},
		{name: "missing default target", body: `{}`, target: "train-demo-service", wantStatus: http.StatusOK},
		{name: "wrapped application not found", body: `{"serviceName":"new-service"}`, target: "new-service", lookupErr: fmt.Errorf("get Application: %w", apierrors.NewNotFound(application, "new-service")), wantStatus: http.StatusOK},
		{name: "lookup unavailable", body: `{}`, target: "train-demo-service", lookupErr: apierrors.NewServiceUnavailable("cluster unavailable"), wantStatus: http.StatusBadGateway},
		{name: "lookup forbidden", body: `{}`, target: "train-demo-service", lookupErr: apierrors.NewForbidden(application, "train-demo-service", fmt.Errorf("denied")), wantStatus: http.StatusBadGateway},
		{name: "untyped not found", body: `{}`, target: "train-demo-service", lookupErr: fmt.Errorf("not found"), wantStatus: http.StatusBadGateway},
		{name: "dependent resource not found", body: `{}`, target: "train-demo-service", lookupErr: fmt.Errorf("list pods: %w", apierrors.NewNotFound(schema.GroupResource{Resource: "pods"}, "")), wantStatus: http.StatusBadGateway},
		{name: "different application not found", body: `{}`, target: "train-demo-service", lookupErr: apierrors.NewNotFound(application, "another-app"), wantStatus: http.StatusBadGateway},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reader := &deliveryPublishReader{
				recordingApplicationReader: &recordingApplicationReader{logs: map[string]*observe.Logs{
					"sock-shop/train-demo": {Logs: `AI_RESULT_JSON={"modelURI":"inline://models/train-demo/v1"}`},
				}},
				summary: tc.summary, lookupErr: tc.lookupErr,
			}
			applier := &recordingApplicationApplier{}
			artifacts := &recordingArtifactStore{}
			audits := &recordingAuditStore{}
			server := authenticatedServer(t, NewServerWithOptions(Options{Reader: reader, Applier: applier, Artifacts: artifacts, Audits: audits}))
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/ai/deliveries/sock-shop/train-demo/publish-service", strings.NewReader(tc.body)))
			if recorder.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d: %s", recorder.Code, tc.wantStatus, recorder.Body.String())
			}
			if reader.lookupNamespace != "sock-shop" || reader.lookupName != tc.target {
				t.Fatalf("lookup = %s/%s, want sock-shop/%s", reader.lookupNamespace, reader.lookupName, tc.target)
			}
			if tc.wantStatus != http.StatusOK {
				if applier.called || len(artifacts.saved) != 0 || len(audits.events) != 0 {
					t.Fatal("rejected publish must not apply, record an artifact, or record success")
				}
				var response errorResponse
				if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil || response.Error == "" {
					t.Fatalf("missing JSON error: %s", recorder.Body.String())
				}
				if tc.wantStatus == http.StatusConflict && !strings.Contains(response.Error, tc.target) {
					t.Fatalf("conflict must identify target: %s", response.Error)
				}
			} else if !applier.called || applier.namespace != "sock-shop" || applier.name != tc.target || len(artifacts.saved) != 1 || len(audits.events) != 1 {
				t.Fatal("missing target must proceed through normal publication")
			}
		})
	}
}
