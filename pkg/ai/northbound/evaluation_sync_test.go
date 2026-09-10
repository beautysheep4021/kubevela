package northbound

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/oam-dev/kubevela/pkg/ai/observe"
)

func TestSyncEvaluationModelAssociation(t *testing.T) {
	const modelURI = "inline://models/model-a/v1"
	for _, tc := range []struct {
		name   string
		target string
		result interface{}
		want   int
	}{
		{"matching", modelURI, modelURI, http.StatusOK},
		{"matching object store", "oss://models/model-a/v1", "oss://models/model-a/v1", http.StatusOK},
		{"different model", modelURI, "inline://models/model-b/v1", http.StatusBadRequest},
		{"different version", modelURI, "inline://models/model-a/v2", http.StatusBadRequest},
		{"missing target", "", modelURI, http.StatusBadRequest},
		{"blank target", "  ", modelURI, http.StatusBadRequest},
		{"blank result", modelURI, "  ", http.StatusBadRequest},
		{"invalid matching URI", "inline://models/%zz", "inline://models/%zz", http.StatusBadRequest},
		{"relative matching URI", "models/model-a", "models/model-a", http.StatusBadRequest},
		{"empty URI location", "inline:", "inline:", http.StatusBadRequest},
		{"whitespace matching URI", "inline://models/model a", "inline://models/model a", http.StatusBadRequest},
		{"missing result", modelURI, nil, http.StatusBadGateway},
		{"nonstring result", modelURI, 42, http.StatusBadGateway},
	} {
		t.Run(tc.name, func(t *testing.T) {
			original := observe.ModelArtifact{Namespace: "sock-shop", Name: "model-a", ModelURI: tc.target,
				Status: "registered", EvaluationStatus: "pending", Metrics: map[string]float64{"accuracy": 0.5}, Summary: "original"}
			store := &recordingArtifactStore{list: []observe.ModelArtifact{original}}
			payload := map[string]interface{}{"metrics": map[string]float64{"accuracy": 0.91}, "summary": "evaluation-passed"}
			if tc.result != nil {
				payload["modelURI"] = tc.result
			}
			encoded, err := json.Marshal(payload)
			if err != nil {
				t.Fatal(err)
			}
			reader := &recordingApplicationReader{logs: map[string]*observe.Logs{
				"sock-shop/model-a-eval": {Pod: "eval-pod", Logs: "AI_RESULT_JSON=" + string(encoded)},
			}}
			server := authenticatedServer(t, NewServerWithOptions(Options{Reader: reader, Artifacts: store}))
			response := httptest.NewRecorder()
			server.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/ai/models/sock-shop/model-a/sync-evaluation", strings.NewReader(`{"evaluationJobName":"model-a-eval"}`)))
			if response.Code != tc.want {
				t.Fatalf("status = %d, want %d: %s", response.Code, tc.want, response.Body.String())
			}
			if tc.want != http.StatusOK {
				if len(store.saved) != 0 {
					t.Fatalf("rejected evaluation saved %d artifacts", len(store.saved))
				}
				if !reflect.DeepEqual(store.list[0], original) {
					t.Fatal("rejected evaluation mutated the existing asset")
				}
				return
			}
			if len(store.saved) != 1 {
				t.Fatalf("saved = %d, want 1", len(store.saved))
			}
			asset := store.saved[0]
			if asset.ModelURI != tc.target || asset.Metrics["accuracy"] != 0.91 || asset.EvaluationStatus != "passed" || asset.SourceApplication != "model-a-eval" {
				t.Fatalf("unexpected saved asset: %#v", asset)
			}
		})
	}
}
