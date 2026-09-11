package northbound

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUserConsoleAssetsAndSession(t *testing.T) {
	server := NewServer()
	cookie := loginForRole(t, server, "user", "tenant-b", "tenant-b-123456")
	for _, path := range []string{"/user", "/console/styles.css", "/console/app.js", "/api/v1/ai/session"} {
		t.Run(path, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, path, nil)
			r.AddCookie(cookie)
			w := httptest.NewRecorder()
			server.ServeHTTP(w, r)
			if w.Code != http.StatusOK {
				t.Fatalf("GET %s: %d %s", path, w.Code, w.Body.String())
			}
			if path == "/user" && (!strings.Contains(w.Body.String(), "/console/app.js") || strings.Contains(w.Body.String(), "平台视图预览")) {
				t.Fatal("user console must serve dedicated workspace")
			}
			if path == "/api/v1/ai/session" {
				var identity map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &identity); err != nil {
					t.Fatal(err)
				}
				if identity["namespace"] != "ai-tenant-b" || identity["username"] != "tenant-b" {
					t.Fatalf("identity: %v", identity)
				}
				if _, ok := identity["password"]; ok {
					t.Fatal("identity exposes credentials")
				}
			}
		})
	}
	for _, path := range []string{"/api/v1/ai/session", "/console/app.js", "/console/monitor.js", "/console/monitor-api.js"} {
		w := httptest.NewRecorder()
		server.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("anonymous %s: %d", path, w.Code)
		}
	}
	for _, path := range []string{"/console/api.test.cjs", "/console/monitor-api.test.cjs", "/console/index.html", "/console/monitor.html", "/console/../auth.go"} {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, path, nil)
		r.AddCookie(cookie)
		server.ServeHTTP(w, r)
		if w.Code == http.StatusOK {
			t.Fatalf("unexpected exposed asset %s", path)
		}
	}
}

func TestMonitorUsesSharedConsoleAssets(t *testing.T) {
	server := NewServer()
	cookie := loginForRole(t, server, "monitor", "admin", "jiankong")
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/monitor", nil)
	r.AddCookie(cookie)
	server.ServeHTTP(w, r)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "/console/monitor.js") || !strings.Contains(w.Body.String(), "/console/styles.css") {
		t.Fatal("monitor must use the shared console design and dedicated controller")
	}
}
