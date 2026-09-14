package northbound

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	domainapply "github.com/oam-dev/kubevela/pkg/ai/domain/apply"
	"github.com/oam-dev/kubevela/pkg/ai/observe"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// NumericExample creates tenant-local files and native Applications using create-only writes.
type NumericExample struct {
	Kube    kubernetes.Interface
	Dynamic dynamic.Interface
	Image   string
}

type NumericExampleResponse struct {
	Namespace       string `json:"namespace"`
	ModelName       string `json:"modelName"`
	DatasetName     string `json:"datasetName"`
	ApplicationName string `json:"applicationName,omitempty"`
	Kind            string `json:"kind,omitempty"`
}

func numericExample(n *NumericExample) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		session, ok := requestSession(r)
		if !ok || session.Role != consoleRoleUser || session.Namespace == "" || session.Username == "" {
			writeError(w, 403, "a tenant user session is required")
			return
		}
		if r.Method != http.MethodPost {
			writeError(w, 405, "method not allowed")
			return
		}
		if len(r.URL.Query()) != 0 {
			writeError(w, 403, "scope must come from the session")
			return
		}
		if origin := r.Header.Get("Origin"); origin != "" {
			u, err := url.Parse(origin)
			if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host != r.Host || u.User != nil {
				writeError(w, 403, "cross-origin request is not allowed")
				return
			}
		}
		if r.Header.Get("Sec-Fetch-Site") == "cross-site" {
			writeError(w, 403, "cross-site request is not allowed")
			return
		}
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxRequestBodyBytes))
		if err != nil {
			var limit *http.MaxBytesError
			if errors.As(err, &limit) {
				writeError(w, 413, "request body is too large")
			} else {
				writeError(w, 400, "cannot read request")
			}
			return
		}
		var input struct {
			Action string   `json:"action"`
			X      *float64 `json:"x"`
		}
		decoder := json.NewDecoder(strings.NewReader(string(body)))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			writeError(w, 400, "expected action and optional numeric x fields")
			return
		}
		if err := decoder.Decode(new(interface{})); err != io.EOF {
			writeError(w, 400, "expected one JSON object")
			return
		}
		if input.Action != "initialize" && input.Action != "job" && input.Action != "service" && input.Action != "predict" {
			writeError(w, 400, "action must be initialize, job, service or predict")
			return
		}
		if input.Action == "predict" && (input.X == nil || math.IsNaN(*input.X) || math.IsInf(*input.X, 0)) {
			writeError(w, 400, "predict requires a finite numeric x")
			return
		}
		if input.Action != "predict" && input.X != nil {
			writeError(w, 400, "x is only supported for predict")
			return
		}
		if n == nil || n.Kube == nil || (input.Action != "initialize" && n.Dynamic == nil) {
			writeError(w, 503, "numeric example Kubernetes clients are not configured")
			return
		}
		if input.Action == "predict" {
			n.predict(w, r, session, *input.X)
			return
		}
		result, err := n.execute(r.Context(), session, input.Action)
		if err != nil {
			switch {
			case apierrors.IsAlreadyExists(err), apierrors.IsConflict(err):
				writeError(w, 409, "numeric example object already exists with conflicting ownership or content")
			case apierrors.IsNotFound(err):
				writeError(w, 404, "tenant namespace or numeric example files not found; initialize first")
			default:
				writeError(w, 500, "numeric example operation failed")
			}
			return
		}
		status := http.StatusCreated
		if input.Action == "initialize" {
			status = http.StatusOK
		}
		writeJSON(w, status, result)
	}
}

func (n *NumericExample) predict(w http.ResponseWriter, r *http.Request, session sessionRecord, x float64) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	name := numericName("numeric-service-", session.Namespace, 30)
	app, err := n.Dynamic.Resource(domainapply.ApplicationGVR).Namespace(session.Namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			writeError(w, 404, "numeric service application not found")
		} else {
			writeError(w, 502, "cannot read numeric service application")
		}
		return
	}
	components, _, err := unstructured.NestedSlice(app.Object, "spec", "components")
	owned := app.GetNamespace() == session.Namespace && app.GetName() == name && app.GetLabels()["ai.oam.dev/source"] == "numeric-example" && app.GetAnnotations()["ai.oam.dev/owner"] == session.Username && err == nil && len(components) == 1
	if owned {
		component, ok := components[0].(map[string]interface{})
		owned = ok && component["name"] == name && component["type"] == "ai-service"
	}
	if !owned {
		writeError(w, 409, "numeric service application ownership or component conflicts")
		return
	}
	client := n.Kube.CoreV1().RESTClient()
	// The generated fake client may return an interface containing a nil RESTClient pointer.
	if client == nil || (reflect.ValueOf(client).Kind() == reflect.Ptr && reflect.ValueOf(client).IsNil()) {
		writeError(w, 503, "numeric service proxy is not configured")
		return
	}
	stream, err := client.Get().Namespace(session.Namespace).Resource("services").Name("http:"+name+"-"+name+":80").SubResource("proxy").Suffix("predict").Param("x", strconv.FormatFloat(x, 'g', -1, 64)).Stream(ctx)
	if err != nil {
		writeError(w, 502, "numeric service proxy request failed")
		return
	}
	defer stream.Close()
	const maxPredictionBytes = 64 << 10
	body, err := io.ReadAll(io.LimitReader(stream, maxPredictionBytes+1))
	if err != nil || len(body) > maxPredictionBytes || !json.Valid(body) {
		writeError(w, 502, "numeric service returned an invalid or oversized response")
		return
	}
	writeJSON(w, http.StatusOK, json.RawMessage(body))
}

func (n *NumericExample) execute(ctx context.Context, session sessionRecord, action string) (NumericExampleResponse, error) {
	ns := session.Namespace
	result := NumericExampleResponse{Namespace: ns, ModelName: numericName("numeric-model-", ns, 50), DatasetName: numericName("numeric-data-", ns, 50)}
	if _, err := n.Kube.CoreV1().Namespaces().Get(ctx, ns, metav1.GetOptions{}); err != nil {
		return result, err
	}
	files := numericConfigMap(ns, "numeric-files-"+ns, session.Username, map[string]string{"model.json": numericModelJSON(), "data.jsonl": "{\"x\":0,\"y\":1}\n{\"x\":1,\"y\":3}\n{\"x\":2,\"y\":5}\n{\"x\":3,\"y\":7}\n{\"x\":4,\"y\":9}\n", "runtime.py": numericRuntimePython()})
	modelURI := "configmap://" + ns + "/" + files.Name + "/model.json"
	dataURI := "configmap://" + ns + "/" + files.Name + "/data.jsonl"
	model := observe.ModelArtifact{Namespace: ns, Name: result.ModelName, ModelURI: modelURI, DatasetURI: dataURI, Version: "v1", Status: "ready", Visibility: "private", Owner: session.Username, Source: "numeric-example", Summary: "Linear numeric model y=2x+1; JSON weights evaluated by Python on CPU"}
	dataset := observe.DatasetArtifact{Namespace: ns, Name: result.DatasetName, DisplayName: result.DatasetName, DatasetURI: dataURI, Format: "jsonl", Purpose: "evaluation", Status: "ready", Visibility: "private", Owner: session.Username, Size: "5 rows", Source: "numeric-example", Summary: "Five numeric x/y records for evaluating the linear model"}
	modelJSON, err := json.Marshal(model)
	if err != nil {
		return result, err
	}
	dataJSON, err := json.Marshal(dataset)
	if err != nil {
		return result, err
	}
	modelCM := numericConfigMap(ns, "ai-artifact-"+result.ModelName, session.Username, map[string]string{"artifact.json": string(modelJSON)})
	modelCM.Labels["ai.oam.dev/artifact"] = "true"
	dataCM := numericConfigMap(ns, "ai-dataset-"+result.DatasetName, session.Username, map[string]string{"dataset.json": string(dataJSON)})
	dataCM.Labels["ai.oam.dev/dataset"] = "true"
	if action == "initialize" {
		// Preflight all names before creating anything. Atomic create + verify handles racing requests.
		for _, cm := range []*corev1.ConfigMap{files, modelCM, dataCM} {
			existing, err := n.Kube.CoreV1().ConfigMaps(ns).Get(ctx, cm.Name, metav1.GetOptions{})
			if err == nil {
				if err = verifyNumericConfigMap(existing, cm); err != nil {
					return result, err
				}
			} else if !apierrors.IsNotFound(err) {
				return result, err
			}
		}
		for _, cm := range []*corev1.ConfigMap{files, modelCM, dataCM} {
			if err = n.ensureConfigMap(ctx, cm); err != nil {
				return result, err
			}
		}
		return result, nil
	}
	for _, cm := range []*corev1.ConfigMap{files, modelCM, dataCM} {
		existing, err := n.Kube.CoreV1().ConfigMaps(ns).Get(ctx, cm.Name, metav1.GetOptions{})
		if err != nil {
			return result, err
		}
		if err = verifyNumericConfigMap(existing, cm); err != nil {
			return result, err
		}
	}
	result.ApplicationName = numericName("numeric-"+action+"-", ns, 30)
	result.Kind = "AIJob"
	if action == "service" {
		result.Kind = "AIService"
	}
	image := n.Image
	if image == "" {
		image = "python:3.11-slim"
	}
	properties := map[string]interface{}{
		"image": image, "imagePullPolicy": "IfNotPresent", "cmd": []interface{}{"python3", "/numeric/runtime.py"}, "args": []interface{}{action},
		"automountServiceAccountToken": false,
		"env":                          []interface{}{map[string]interface{}{"name": "MODEL_URI", "value": modelURI}},
		"volumes":                      []interface{}{map[string]interface{}{"name": "numeric-files", "configMap": map[string]interface{}{"name": files.Name}}},
		"volumeMounts":                 []interface{}{map[string]interface{}{"name": "numeric-files", "mountPath": "/numeric", "readOnly": true}},
		"resources":                    map[string]interface{}{"requests": map[string]interface{}{"cpu": "100m", "memory": "64Mi"}, "limits": map[string]interface{}{"cpu": "500m", "memory": "128Mi"}},
	}
	purpose := action
	if action == "job" {
		purpose = "evaluation"
	}
	properties["annotations"] = map[string]interface{}{"ai.oam.dev/purpose": purpose, "ai.oam.dev/source": "numeric-example", "ai.oam.dev/owner": session.Username}
	if action == "job" {
		properties["jobKind"] = "evaluation"
		properties["dataset"] = map[string]interface{}{"uri": dataURI}
		properties["output"] = map[string]interface{}{"uri": modelURI}
		properties["activeDeadlineSeconds"] = int64(120)
		properties["backoffLimit"] = int64(0)
	} else {
		properties["model"] = map[string]interface{}{"name": result.ModelName, "uri": modelURI, "version": "v1"}
		properties["endpoint"] = map[string]interface{}{"port": int64(8080), "servicePort": int64(80), "type": "ClusterIP"}
		properties["readinessProbe"] = map[string]interface{}{"httpGet": map[string]interface{}{"path": "/healthz", "port": int64(8080)}, "initialDelaySeconds": int64(1)}
	}
	app := &unstructured.Unstructured{Object: map[string]interface{}{"apiVersion": "core.oam.dev/v1beta1", "kind": "Application", "metadata": map[string]interface{}{"name": result.ApplicationName, "namespace": ns, "labels": map[string]interface{}{"ai.oam.dev/source": "numeric-example"}, "annotations": map[string]interface{}{"ai.oam.dev/owner": session.Username}}, "spec": map[string]interface{}{"components": []interface{}{map[string]interface{}{"name": result.ApplicationName, "type": "ai-" + action, "properties": properties}}}}}
	_, err = n.Dynamic.Resource(domainapply.ApplicationGVR).Namespace(ns).Create(ctx, app, metav1.CreateOptions{})
	return result, err
}

func numericConfigMap(ns, name, owner string, data map[string]string) *corev1.ConfigMap {
	immutable := true
	return &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns, Labels: map[string]string{"ai.oam.dev/source": "numeric-example"}, Annotations: map[string]string{"ai.oam.dev/owner": owner}}, Immutable: &immutable, Data: data}
}

// Preserve readable tenant names, with a stable suffix when Kubernetes label limits require shortening.
func numericName(prefix, namespace string, limit int) string {
	name := prefix + namespace
	if len(name) <= limit {
		return name
	}
	hash := sha256.Sum256([]byte(name))
	return strings.TrimRight(name[:limit-13], "-") + fmt.Sprintf("-%x", hash[:6])
}

func verifyNumericConfigMap(existing, want *corev1.ConfigMap) error {
	if existing.Immutable == nil || !*existing.Immutable || existing.Annotations["ai.oam.dev/owner"] != want.Annotations["ai.oam.dev/owner"] || !reflect.DeepEqual(existing.Data, want.Data) || len(existing.BinaryData) != 0 {
		return apierrors.NewAlreadyExists(schema.GroupResource{Resource: "configmaps"}, want.Name)
	}
	for k, v := range want.Labels {
		if existing.Labels[k] != v {
			return apierrors.NewAlreadyExists(schema.GroupResource{Resource: "configmaps"}, want.Name)
		}
	}
	return nil
}

func (n *NumericExample) ensureConfigMap(ctx context.Context, want *corev1.ConfigMap) error {
	_, err := n.Kube.CoreV1().ConfigMaps(want.Namespace).Create(ctx, want, metav1.CreateOptions{})
	if !apierrors.IsAlreadyExists(err) {
		return err
	}
	existing, err := n.Kube.CoreV1().ConfigMaps(want.Namespace).Get(ctx, want.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	return verifyNumericConfigMap(existing, want)
}

func numericModelJSON() string { return `{"type":"linear-regression","slope":2,"intercept":1}` }

func numericRuntimePython() string {
	return `import json
import math
import os
import sys
from http.server import BaseHTTPRequestHandler, HTTPServer
from pathlib import Path
from urllib.parse import parse_qs, urlparse

root = Path(os.environ.get("NUMERIC_DIR", "/numeric"))
model = json.loads((root / "model.json").read_text())
model_uri = os.environ["MODEL_URI"]

def predict(x):
    return float(model["slope"]) * x + float(model["intercept"])

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        path = urlparse(self.path)
        status = 200
        if path.path in ("/", "/healthz"):
            result = {"status": "ok", "modelURI": model_uri, "model": model}
        elif path.path == "/predict":
            try:
                values = parse_qs(path.query).get("x", [])
                if len(values) != 1:
                    raise ValueError("expected one x")
                x = float(values[0])
                y = predict(x)
                if not math.isfinite(x) or not math.isfinite(y):
                    raise ValueError("x and prediction must be finite")
                result = {"x": x, "y": y, "prediction": y, "modelURI": model_uri}
            except (ValueError, OverflowError):
                status, result = 400, {"error": "x must be one finite number"}
        else:
            status, result = 404, {"error": "not found"}
        body = json.dumps(result, allow_nan=False).encode()
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

if __name__ == "__main__":
    if sys.argv[1] == "job":
        rows = [json.loads(line) for line in (root / "data.jsonl").read_text().splitlines() if line.strip()]
        predictions = [{"x": row["x"], "expected": row["y"], "prediction": predict(row["x"])} for row in rows]
        mae = sum(abs(row["prediction"] - row["expected"]) for row in predictions) / len(predictions)
        result = {"modelURI": model_uri, "metrics": {"mae": mae, "samples": len(rows)}, "predictions": predictions, "summary": "CPU linear model evaluation"}
        print("AI_RESULT_JSON=" + json.dumps(result, allow_nan=False), flush=True)
    elif sys.argv[1] == "service":
        HTTPServer(("0.0.0.0", int(os.environ.get("PORT", "8080"))), Handler).serve_forever()
    else:
        raise SystemExit("expected job or service")
`
}
