package northbound

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"

	domainapply "github.com/oam-dev/kubevela/pkg/ai/domain/apply"
	"github.com/oam-dev/kubevela/pkg/ai/observe"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

func TestNumericExampleRequiresTenantUserAndSessionNamespace(t *testing.T) {
	h := NewServerWithOptions(Options{})
	r := httptest.NewRequest("POST", "/api/v1/ai/examples/numeric", strings.NewReader(`{"action":"initialize","namespace":"other"}`))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatalf("unauthenticated status=%d", w.Code)
	}
}

func TestNumericModelJSONIsActualLinearModel(t *testing.T) {
	if got := numericModelJSON(); !strings.Contains(got, `"slope":2`) || !strings.Contains(got, `"intercept":1`) {
		t.Fatalf("model=%s", got)
	}
}

func numericTestServer(t *testing.T) (*NumericExample, http.Handler) {
	t.Helper()
	k := kubefake.NewSimpleClientset(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ai-tenant-a"}}, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ai-tenant-b"}})
	d := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())
	n := &NumericExample{Kube: k, Dynamic: d}
	return n, NewServerWithOptions(Options{NumericExample: n})
}

func numericPost(h http.Handler, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("POST", "/api/v1/ai/examples/numeric", strings.NewReader(body)))
	return w
}

func TestNumericPredictProxyAndOwnership(t *testing.T) {
	n, h := numericTestServer(t)
	a := authenticatedServerAs(t, h, "admin", "shiyong")
	for _, action := range []string{"initialize", "service"} {
		w := numericPost(a, `{"action":"`+action+`"}`)
		if w.Code >= 300 {
			t.Fatal(w.Body.String())
		}
	}
	mode := "ok"
	calls := 0
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != "GET" || r.URL.Path != "/api/v1/namespaces/ai-tenant-a/services/http:numeric-service-ai-tenant-a-numeric-service-ai-tenant-a:80/proxy/predict" || r.URL.Query().Get("x") != "4" {
			t.Errorf("unexpected proxy request %s %s", r.Method, r.URL.String())
		}
		switch mode {
		case "error":
			http.Error(w, "unavailable", 503)
		case "large":
			_, _ = w.Write([]byte(strings.Repeat(" ", 65537)))
		case "invalid":
			_, _ = w.Write([]byte("not json"))
		default:
			_, _ = w.Write([]byte(`{"x":4,"y":9,"prediction":9,"modelURI":"configmap://ai-tenant-a/numeric-files-ai-tenant-a/model.json"}`))
		}
	}))
	defer proxy.Close()
	k, err := kubernetes.NewForConfig(&rest.Config{Host: proxy.URL})
	if err != nil {
		t.Fatal(err)
	}
	n.Kube = k
	w := numericPost(a, `{"action":"predict","x":4}`)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"prediction":9`) {
		t.Fatalf("predict: %d %s", w.Code, w.Body.String())
	}
	for _, value := range []string{"error", "large", "invalid"} {
		mode = value
		w = numericPost(a, `{"action":"predict","x":4}`)
		if w.Code != 502 {
			t.Fatalf("%s: %d %s", value, w.Code, w.Body.String())
		}
	}
	before := calls
	b := authenticatedServerAs(t, h, "tenant-b", "tenant-b-123456")
	if w = numericPost(b, `{"action":"predict","x":4}`); w.Code != 404 {
		t.Fatalf("other tenant: %d", w.Code)
	}
	app, err := n.Dynamic.Resource(domainapply.ApplicationGVR).Namespace("ai-tenant-a").Get(context.Background(), "numeric-service-ai-tenant-a", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	app.SetAnnotations(map[string]string{"ai.oam.dev/owner": "foreign"})
	if _, err = n.Dynamic.Resource(domainapply.ApplicationGVR).Namespace("ai-tenant-a").Update(context.Background(), app, metav1.UpdateOptions{}); err != nil {
		t.Fatal(err)
	}
	if w = numericPost(a, `{"action":"predict","x":4}`); w.Code != 409 {
		t.Fatalf("foreign app: %d", w.Code)
	}
	if calls != before {
		t.Fatal("proxied an unauthorized request")
	}
}

func TestNumericPredictValidationAndUnavailableBackend(t *testing.T) {
	n, h := numericTestServer(t)
	a := authenticatedServerAs(t, h, "admin", "shiyong")
	for _, body := range []string{`{"action":"predict"}`, `{"action":"predict","x":null}`, `{"action":"predict","x":"4"}`, `{"action":"predict","x":1e999}`, `{"action":"predict","x":4,"namespace":"other"}`, `{"action":"predict","x":4,"serviceName":"other"}`} {
		w := numericPost(a, body)
		if w.Code != 400 {
			t.Fatalf("%s: %d", body, w.Code)
		}
	}
	if w := numericPost(h, `{"action":"predict","x":4}`); w.Code != 401 {
		t.Fatalf("unauthenticated: %d", w.Code)
	}
	if w := numericPost(authenticatedServer(t, h), `{"action":"predict","x":4}`); w.Code != 403 {
		t.Fatalf("monitor: %d", w.Code)
	}
	for _, action := range []string{"initialize", "service"} {
		w := numericPost(a, `{"action":"`+action+`"}`)
		if w.Code >= 300 {
			t.Fatal(w.Body.String())
		}
	}
	if w := numericPost(a, `{"action":"predict","x":4}`); w.Code != 503 {
		t.Fatalf("no REST backend: %d %s", w.Code, w.Body.String())
	}
	n.Kube = nil
	if w := numericPost(a, `{"action":"predict","x":4}`); w.Code != 503 {
		t.Fatalf("no kube backend: %d", w.Code)
	}
}

func TestNumericInitializeMountsMetadataAndConflicts(t *testing.T) {
	n, h := numericTestServer(t)
	a := authenticatedServerAs(t, h, "admin", "shiyong")
	ctx := context.Background()
	w := numericPost(a, `{"action":"initialize"}`)
	if w.Code != 200 {
		t.Fatalf("initialize: %d %s", w.Code, w.Body.String())
	}
	var result NumericExampleResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Namespace != "ai-tenant-a" || !strings.Contains(result.ModelName, result.Namespace) || result.ModelName == result.DatasetName {
		t.Fatalf("response: %+v", result)
	}
	if again := numericPost(a, `{"action":"initialize"}`); again.Code != 200 {
		t.Fatalf("repeat: %s", again.Body.String())
	}
	cms, err := n.Kube.CoreV1().ConfigMaps(result.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	var files *corev1.ConfigMap
	for i := range cms.Items {
		if cms.Items[i].Data["runtime.py"] != "" {
			files = &cms.Items[i]
		}
	}
	if files == nil || files.Immutable == nil || !*files.Immutable || len(files.Data) != 3 {
		t.Fatalf("files: %+v", files)
	}
	reader := observe.NewClientWithKube(n.Dynamic, n.Kube)
	models, err := reader.ListArtifacts(ctx, result.Namespace)
	if err != nil {
		t.Fatal(err)
	}
	datasets, err := reader.ListDatasets(ctx, result.Namespace)
	if err != nil {
		t.Fatal(err)
	}
	uri := "configmap://" + result.Namespace + "/" + files.Name + "/model.json"
	if len(models) != 1 || models[0].ModelURI != uri || models[0].Owner != "admin" || models[0].Status != "ready" || models[0].Source != "numeric-example" || models[0].JobName != "" {
		t.Fatalf("models: %+v", models)
	}
	if len(datasets) != 1 || datasets[0].Source != "numeric-example" || datasets[0].Status != "ready" || datasets[0].Owner != "admin" {
		t.Fatalf("datasets: %+v", datasets)
	}
	if datasets[0].DisplayName != result.DatasetName {
		t.Fatalf("display name must identify tenant: %s", datasets[0].DisplayName)
	}
	for _, action := range []string{"job", "service"} {
		w = numericPost(a, `{"action":"`+action+`"}`)
		if w.Code != 201 {
			t.Fatalf("%s: %d %s", action, w.Code, w.Body.String())
		}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		app, err := n.Dynamic.Resource(domainapply.ApplicationGVR).Namespace(result.Namespace).Get(ctx, result.ApplicationName, metav1.GetOptions{})
		if err != nil {
			t.Fatal(err)
		}
		components := app.Object["spec"].(map[string]interface{})["components"].([]interface{})
		comp := components[0].(map[string]interface{})
		p := comp["properties"].(map[string]interface{})
		if comp["type"] != "ai-"+action || p["automountServiceAccountToken"] != false {
			t.Fatalf("component: %+v", comp)
		}
		volume := p["volumes"].([]interface{})[0].(map[string]interface{})
		mount := p["volumeMounts"].([]interface{})[0].(map[string]interface{})
		if volume["configMap"].(map[string]interface{})["name"] != files.Name || mount["mountPath"] != "/numeric" || mount["readOnly"] != true {
			t.Fatalf("mount: %v %v", volume, mount)
		}
		resources := p["resources"].(map[string]interface{})
		if resources["requests"].(map[string]interface{})["cpu"] != "100m" || resources["limits"].(map[string]interface{})["memory"] != "128Mi" {
			t.Fatalf("resources: %v", resources)
		}
		if again := numericPost(a, `{"action":"`+action+`"}`); again.Code != 409 {
			t.Fatalf("overwrite: %d", again.Code)
		}
		if action == "service" && p["endpoint"].(map[string]interface{})["type"] != "ClusterIP" {
			t.Fatal("service must explicitly use ClusterIP")
		}
		definition, err := os.ReadFile("../../../vela-templates/definitions/internal/component/ai-" + action + ".cue")
		if err != nil {
			t.Fatal(err)
		}
		params, err := json.Marshal(p)
		if err != nil {
			t.Fatal(err)
		}
		value := cuecontext.New().CompileString(string(definition) + "\ncontext: {appName: \"numeric\", name: \"runtime\"}\ntemplate: parameter: " + string(params))
		rendered, err := value.LookupPath(cue.ParsePath("template.output")).MarshalJSON()
		if err != nil {
			t.Fatalf("render %s: %v", action, err)
		}
		if !strings.Contains(string(rendered), `"mountPath":"/numeric"`) || !strings.Contains(string(rendered), `"automountServiceAccountToken":false`) {
			t.Fatalf("rendered workload: %s", rendered)
		}
	}
	b := authenticatedServerAs(t, h, "tenant-b", "tenant-b-123456")
	if numericPost(b, `{"action":"initialize"}`).Code != 200 {
		t.Fatal("tenant b initialize")
	}
	other, _ := reader.ListArtifacts(ctx, "ai-tenant-b")
	if len(other) != 1 || other[0].ModelURI == uri || other[0].Owner != "tenant-b" {
		t.Fatalf("isolation: %+v", other)
	}
	dir := t.TempDir()
	for name, data := range files.Data {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	// Execute the exact mounted payload, and change model weights to detect canned predictions.
	run := func(want string) {
		cmd := exec.Command("python3", filepath.Join(dir, "runtime.py"), "job")
		cmd.Env = append(os.Environ(), "NUMERIC_DIR="+dir, "MODEL_URI="+uri)
		out, err := cmd.CombinedOutput()
		if err != nil || !strings.Contains(string(out), "AI_RESULT_JSON=") || !strings.Contains(string(out), want) {
			t.Fatalf("runtime: %s %v", out, err)
		}
	}
	run(`"mae": 0.0`)
	ctxHTTP, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	httpScript := `import importlib.util, json, threading, urllib.request, urllib.error
spec = importlib.util.spec_from_file_location("numeric", "` + filepath.Join(dir, "runtime.py") + `")
m = importlib.util.module_from_spec(spec)
spec.loader.exec_module(m)
server = m.HTTPServer(("127.0.0.1", 0), m.Handler)
thread = threading.Thread(target=server.serve_forever, daemon=True)
thread.start()
base = "http://127.0.0.1:" + str(server.server_port)
try:
    for path in ("/", "/healthz"):
        assert json.load(urllib.request.urlopen(base + path))["status"] == "ok"
    assert json.load(urllib.request.urlopen(base + "/predict?x=4"))["y"] == 9
    m.model["slope"] = 3
    assert json.load(urllib.request.urlopen(base + "/predict?x=4"))["y"] == 13
    for path, status in (("/predict", 400), ("/predict?x=nan", 400), ("/predict?x=1&x=2", 400), ("/missing", 404)):
        try:
            urllib.request.urlopen(base + path)
            raise AssertionError(path)
        except urllib.error.HTTPError as error:
            assert error.code == status
finally:
    server.shutdown()
    server.server_close()
    thread.join()
`
	cmdHTTP := exec.CommandContext(ctxHTTP, "python3", "-c", httpScript)
	cmdHTTP.Env = append(os.Environ(), "NUMERIC_DIR="+dir, "MODEL_URI="+uri)
	if output, err := cmdHTTP.CombinedOutput(); err != nil {
		t.Fatalf("HTTP runtime: %s %v", output, err)
	}
	if err := os.WriteFile(filepath.Join(dir, "model.json"), []byte(`{"slope":3,"intercept":1}`), 0600); err != nil {
		t.Fatal(err)
	}
	run(`"mae": 2.0`)
}

func TestNumericRejectsScopeOriginAndMalformedBody(t *testing.T) {
	_, h := numericTestServer(t)
	a := authenticatedServerAs(t, h, "admin", "shiyong")
	if w := numericPost(authenticatedServer(t, h), `{"action":"initialize"}`); w.Code != 403 {
		t.Fatalf("monitor: %d", w.Code)
	}
	for _, body := range []string{`{"action":"initialize","namespace":"other"}`, `{"action":"initialize","owner":"other"}`, `{"action":"unknown"}`, `{"action":"initialize"} {}`, `null`} {
		if w := numericPost(a, body); w.Code != 400 && w.Code != 403 {
			t.Fatalf("body %s: %d", body, w.Code)
		}
	}
	for key, value := range map[string]string{"Origin": "https://evil.example", "Sec-Fetch-Site": "cross-site"} {
		r := httptest.NewRequest("POST", "/api/v1/ai/examples/numeric", strings.NewReader(`{"action":"initialize"}`))
		r.Header.Set(key, value)
		w := httptest.NewRecorder()
		a.ServeHTTP(w, r)
		if w.Code != 403 {
			t.Fatalf("%s: %d", key, w.Code)
		}
	}
	w := numericPost(a, strings.Repeat(" ", maxRequestBodyBytes+1))
	if w.Code != 413 {
		t.Fatalf("oversized: %d", w.Code)
	}
	r := httptest.NewRequest("POST", "/api/v1/ai/examples/numeric?namespace=other", strings.NewReader(`{"action":"initialize"}`))
	w = httptest.NewRecorder()
	a.ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatalf("query scope: %d", w.Code)
	}
}

func TestNumericLongNamespaceNames(t *testing.T) {
	n, _ := numericTestServer(t)
	ctx := context.Background()
	ns := strings.Repeat("tenant-", 8) + "abcdefg"
	if _, err := n.Kube.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}
	session := sessionRecord{Role: consoleRoleUser, Namespace: ns, Username: "long-user"}
	result, err := n.execute(ctx, session, "initialize")
	if err != nil {
		t.Fatal(err)
	}
	if problems := validation.IsValidLabelValue(result.ModelName); len(problems) != 0 {
		t.Fatalf("invalid model label: %s %v", result.ModelName, problems)
	}
	for _, action := range []string{"job", "service"} {
		result, err = n.execute(ctx, session, action)
		if err != nil {
			t.Fatal(err)
		}
		app, err := n.Dynamic.Resource(domainapply.ApplicationGVR).Namespace(ns).Get(ctx, result.ApplicationName, metav1.GetOptions{})
		if err != nil {
			t.Fatal(err)
		}
		c := app.Object["spec"].(map[string]interface{})["components"].([]interface{})[0].(map[string]interface{})
		workloadName := result.ApplicationName + "-" + c["name"].(string)
		if problems := validation.IsDNS1123Label(workloadName); len(problems) != 0 {
			t.Fatalf("invalid workload/service name %s: %v", workloadName, problems)
		}
		if c["properties"].(map[string]interface{})["annotations"].(map[string]interface{})["ai.oam.dev/purpose"] == "" {
			t.Fatal("missing purpose")
		}
	}
}

func TestNumericForeignRegistryNeverOverwritten(t *testing.T) {
	for _, key := range []string{"artifact.json", "dataset.json"} {
		t.Run(key, func(t *testing.T) {
			n, h := numericTestServer(t)
			a := authenticatedServerAs(t, h, "admin", "shiyong")
			ctx := context.Background()
			if w := numericPost(a, `{"action":"initialize"}`); w.Code != 200 {
				t.Fatal(w.Body.String())
			}
			cms, err := n.Kube.CoreV1().ConfigMaps("ai-tenant-a").List(ctx, metav1.ListOptions{})
			if err != nil {
				t.Fatal(err)
			}
			for _, cm := range cms.Items {
				if cm.Data[key] == "" {
					continue
				}
				cm.Data[key] = `{"owner":"foreign"}`
				if _, err = n.Kube.CoreV1().ConfigMaps(cm.Namespace).Update(ctx, &cm, metav1.UpdateOptions{}); err != nil {
					t.Fatal(err)
				}
				for _, action := range []string{"initialize", "job", "service"} {
					if w := numericPost(a, `{"action":"`+action+`"}`); w.Code != 409 {
						t.Fatalf("foreign registry %s: %d", action, w.Code)
					}
				}
				after, err := n.Kube.CoreV1().ConfigMaps(cm.Namespace).Get(ctx, cm.Name, metav1.GetOptions{})
				if err != nil || after.Data[key] != cm.Data[key] {
					t.Fatalf("overwrote foreign object: %v", err)
				}
			}
		})
	}
}

func TestNumericConcurrentInitializeAndForeignConfigMap(t *testing.T) {
	n, h := numericTestServer(t)
	a := authenticatedServerAs(t, h, "admin", "shiyong")
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := numericPost(a, `{"action":"initialize"}`)
			if w.Code != 200 {
				t.Errorf("concurrent: %d %s", w.Code, w.Body.String())
			}
		}()
	}
	wg.Wait()
	cms, _ := n.Kube.CoreV1().ConfigMaps("ai-tenant-a").List(context.Background(), metav1.ListOptions{})
	for _, cm := range cms.Items {
		if cm.Data["runtime.py"] == "" {
			continue
		}
		foreign := cm.DeepCopy()
		foreign.Annotations["ai.oam.dev/owner"] = "other"
		_, err := n.Kube.CoreV1().ConfigMaps(cm.Namespace).Update(context.Background(), foreign, metav1.UpdateOptions{})
		if err != nil {
			t.Fatal(err)
		}
	}
	if w := numericPost(a, `{"action":"initialize"}`); w.Code != 409 {
		t.Fatalf("foreign cm: %d %s", w.Code, w.Body.String())
	}
	if err := n.Kube.CoreV1().Namespaces().Delete(context.Background(), "ai-tenant-a", metav1.DeleteOptions{}); err != nil {
		t.Fatal(err)
	}
	if w := numericPost(a, `{"action":"initialize"}`); w.Code != 404 {
		t.Fatalf("missing namespace: %d", w.Code)
	}
}
