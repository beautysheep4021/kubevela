package northbound

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oam-dev/kubevela/pkg/ai/datasetstorage"
)

func datasetUploadRequest(t *testing.T, namespace, name, content string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("namespace", namespace); err != nil {
		t.Fatal(err)
	}
	if err := writer.WriteField("name", name); err != nil {
		t.Fatal(err)
	}
	file, err := writer.CreateFormFile("file", "data.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = file.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err = writer.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", "/api/v1/ai/dataset-uploads", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func TestDatasetUploadsPersistAndIsolate(t *testing.T) {
	repo, err := datasetstorage.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	server := NewServerWithOptions(Options{DatasetUploads: repo})
	a := authenticatedServerAs(t, server, "admin", "shiyong")
	b := authenticatedServerAs(t, server, "tenant-b", "tenant-b-123456")
	post := func(handler http.Handler, namespace, content string) *httptest.ResponseRecorder {
		out := httptest.NewRecorder()
		handler.ServeHTTP(out, datasetUploadRequest(t, namespace, "customer-data", content))
		return out
	}
	out := post(a, "ai-tenant-a", "{\"text\":\"private A\"}\n")
	if out.Code != 201 {
		t.Fatalf("upload: %d %s", out.Code, out.Body.String())
	}
	var asset datasetstorage.Asset
	if err := json.Unmarshal(out.Body.Bytes(), &asset); err != nil {
		t.Fatal(err)
	}
	if asset.Owner != "admin" || asset.RuntimeReady || asset.Rows != 1 {
		t.Fatalf("asset: %+v", asset)
	}
	if post(a, "ai-tenant-a", "{\"text\":\"private A\"}\n").Code != 409 {
		t.Fatal("duplicate must not overwrite")
	}
	if post(b, "ai-tenant-a", "{}\n").Code != 403 {
		t.Fatal("foreign upload allowed")
	}
	if post(a, "ai-tenant-a", "bad json").Code != 400 {
		t.Fatal("invalid JSON accepted")
	}
	path := "/api/v1/ai/dataset-uploads/ai-tenant-a/customer-data/" + asset.Version
	for _, tc := range []struct {
		handler http.Handler
		path    string
		want    int
	}{
		{a, path, 200}, {b, path, 403}, {b, "/api/v1/ai/dataset-uploads?namespace=ai-tenant-a", 403},
		{b, "/api/v1/ai/dataset-uploads", 200}, {server, path, 401},
	} {
		r := httptest.NewRecorder()
		tc.handler.ServeHTTP(r, httptest.NewRequest("GET", tc.path, nil))
		if r.Code != tc.want {
			t.Fatalf("%s: %d %s", tc.path, r.Code, r.Body.String())
		}
		if tc.handler == b && strings.Contains(r.Body.String(), "private A") {
			t.Fatal("leaked preview")
		}
	}
}

func TestDatasetUploadsDisabledAndCrossOrigin(t *testing.T) {
	server := NewServerWithOptions(Options{})
	a := authenticatedServerAs(t, server, "admin", "shiyong")
	out := httptest.NewRecorder()
	a.ServeHTTP(out, httptest.NewRequest("GET", "/api/v1/ai/dataset-uploads", nil))
	if out.Code != 503 {
		t.Fatalf("disabled: %d", out.Code)
	}
	repo, err := datasetstorage.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	server = NewServerWithOptions(Options{DatasetUploads: repo})
	a = authenticatedServerAs(t, server, "admin", "shiyong")
	req := datasetUploadRequest(t, "ai-tenant-a", "data", "{}\n")
	req.Header.Set("Origin", "https://evil.example")
	out = httptest.NewRecorder()
	a.ServeHTTP(out, req)
	if out.Code != 403 {
		t.Fatalf("cross-origin: %d", out.Code)
	}
	monitor := authenticatedServer(t, server)
	out = httptest.NewRecorder()
	monitor.ServeHTTP(out, datasetUploadRequest(t, "ai-tenant-a", "data", "{}\n"))
	if out.Code != 403 {
		t.Fatalf("monitor writes: %d", out.Code)
	}
}

func TestRegisteredModelDoesNotImplyTraining(t *testing.T) {
	store := &recordingArtifactStore{}
	server := authenticatedServerAs(t, NewServerWithOptions(Options{Artifacts: store}), "admin", "shiyong")
	out := httptest.NewRecorder()
	server.ServeHTTP(out, httptest.NewRequest("POST", "/api/v1/ai/models", strings.NewReader(`{"namespace":"ai-tenant-a","name":"base-model","modelURI":"hf://org/model"}`)))
	if out.Code != 201 || len(store.saved) != 1 || store.saved[0].Status != "registered" {
		t.Fatalf("registration: %d %s", out.Code, out.Body.String())
	}
}

func TestDatasetUploadRejectsMalformedAndOversizedBodies(t *testing.T) {
	repo, err := datasetstorage.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	server := authenticatedServerAs(t, NewServerWithOptions(Options{DatasetUploads: repo}), "admin", "shiyong")
	for _, tc := range []struct {
		name, content string
		want          int
	}{
		{"empty", "", 400}, {"array", "[]\n", 400}, {"scalar", "3\n", 400},
		{"oversized", strings.Repeat("x", (10<<20)+(64<<10)), 413},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := httptest.NewRecorder()
			server.ServeHTTP(out, datasetUploadRequest(t, "ai-tenant-a", "invalid-data", tc.content))
			if out.Code != tc.want {
				t.Fatalf("status %d: %s", out.Code, out.Body.String())
			}
		})
	}
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	for key, value := range map[string]string{"namespace": "ai-tenant-a", "name": "data", "owner": "another-user"} {
		if err := w.WriteField(key, value); err != nil {
			t.Fatal(err)
		}
	}
	f, err := w.CreateFormFile("file", "../../escape")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.Write([]byte("{}\n")); err != nil {
		t.Fatal(err)
	}
	if err = w.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", "/api/v1/ai/dataset-uploads", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	out := httptest.NewRecorder()
	server.ServeHTTP(out, req)
	if out.Code != 400 {
		t.Fatalf("forged owner: %d", out.Code)
	}
}
