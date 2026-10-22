package northbound

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
)

func TestSanitizeBasePath(t *testing.T) {
	for _, tc := range []struct {
		raw  string
		want string
	}{
		{raw: "", want: ""},
		{raw: "/", want: ""},
		{raw: "/ai-console", want: "/ai-console"},
		{raw: " /ai-console/ ", want: "/ai-console"},
		{raw: "/a/b_c/d-1", want: "/a/b_c/d-1"},
		{raw: "ai-console", want: ""},
		{raw: "//evil.example", want: ""},
		{raw: "/ai console", want: ""},
		{raw: "/ai\"><script>", want: ""},
		{raw: "/a/../b", want: ""},
		{raw: "/a?x=1", want: ""},
	} {
		if got := sanitizeBasePath(tc.raw); got != tc.want {
			t.Errorf("sanitizeBasePath(%q) = %q, want %q", tc.raw, got, tc.want)
		}
	}
}

func TestBasePathPrecedence(t *testing.T) {
	for _, tc := range []struct {
		name      string
		config    authConfig
		forwarded string
		want      string
	}{
		{name: "none", want: ""},
		{name: "configured", config: authConfig{BasePath: "/ai"}, want: "/ai"},
		{name: "forwarded ignored without remote auth", config: authConfig{BasePath: "/ai"}, forwarded: "/ai-console", want: "/ai"},
		{name: "forwarded wins with remote auth", config: authConfig{RemoteAuthEnabled: true, BasePath: "/ai"}, forwarded: "/ai-console", want: "/ai-console"},
		{name: "bad forwarded falls back", config: authConfig{RemoteAuthEnabled: true, BasePath: "/ai"}, forwarded: "https://evil", want: "/ai"},
		{name: "bad configured ignored", config: authConfig{BasePath: "no-slash"}, want: ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			auth := &consoleAuth{config: tc.config}
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.forwarded != "" {
				req.Header.Set(forwardedPrefixHeader, tc.forwarded)
			}
			if got := auth.basePath(req); got != tc.want {
				t.Fatalf("basePath = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestBasePathPrefixesRedirectsAndPages(t *testing.T) {
	server := NewServerWithOptions(Options{Auth: &authConfig{BasePath: "/ai-console"}})

	for _, tc := range []struct {
		name     string
		method   string
		path     string
		form     url.Values
		wantCode int
		wantLoc  string
	}{
		{name: "root to login", method: http.MethodGet, path: "/", wantCode: http.StatusFound, wantLoc: "/ai-console/login"},
		{name: "page to login", method: http.MethodGet, path: "/user", wantCode: http.StatusFound, wantLoc: "/ai-console/login"},
		{name: "logout to login", method: http.MethodPost, path: "/logout", wantCode: http.StatusFound, wantLoc: "/ai-console/login"},
		{
			name: "login to role page", method: http.MethodPost, path: "/login",
			form:     url.Values{"role": {"user"}, "username": {"admin"}, "password": {"shiyong"}},
			wantCode: http.StatusFound, wantLoc: "/ai-console/user",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request
			if tc.form != nil {
				req = httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.form.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			} else {
				req = httptest.NewRequest(tc.method, tc.path, nil)
			}
			rec := httptest.NewRecorder()
			server.ServeHTTP(rec, req)
			if rec.Code != tc.wantCode {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantCode)
			}
			if loc := rec.Header().Get("Location"); loc != tc.wantLoc {
				t.Fatalf("location = %q, want %q", loc, tc.wantLoc)
			}
			for _, cookie := range rec.Result().Cookies() {
				if cookie.Name == sessionCookieName && cookie.Path != "/ai-console/" {
					t.Fatalf("cookie path = %q, want /ai-console/", cookie.Path)
				}
			}
		})
	}

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/login", nil))
	body := rec.Body.String()
	if strings.Contains(body, basePathPlaceholder) {
		t.Fatal("login page still holds the base path placeholder")
	}
	for _, want := range []string{`action="/ai-console/login"`, `formaction="/ai-console/login?demo=local"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("login page missing %s", want)
		}
	}
}

func TestRenderConsoleHTMLReplacesBasePath(t *testing.T) {
	for _, role := range []consoleRole{consoleRoleUser, consoleRoleMonitor} {
		out := renderConsoleHTMLForSession(role, sessionRecord{Role: role}, "/ai-console")
		if strings.Contains(out, basePathPlaceholder) {
			t.Fatalf("%s console still holds the base path placeholder", role)
		}
		source := userConsoleHTML()
		if role == consoleRoleMonitor {
			source = monitorConsoleHTML()
		}
		if strings.Contains(source, basePathPlaceholder+"/console/") && !strings.Contains(out, "/ai-console/console/") {
			t.Fatalf("%s console assets are not under the base path", role)
		}
		if bare := renderConsoleHTMLForSession(role, sessionRecord{Role: role}, ""); strings.Contains(bare, basePathPlaceholder) {
			t.Fatalf("%s console without a base path still holds the placeholder", role)
		}
	}
}

func TestRemoteHeadersScopeTenantNamespace(t *testing.T) {
	reader := &recordingApplicationReader{}
	server := NewServerWithOptions(Options{Reader: reader, Auth: &authConfig{
		RemoteAuthEnabled: true,
		AdminGroups:       []string{"platform-admin"},
	}})

	for _, tc := range []struct {
		name          string
		user          string
		groups        string
		tenant        string
		namespace     string
		url           string
		wantStatus    int
		wantNamespace string
	}{
		{name: "defaults to header namespace", user: "alice", tenant: "team-a", namespace: "ai-team-a", url: "/api/v1/ai/applications", wantStatus: http.StatusOK, wantNamespace: "ai-team-a"},
		{name: "in scope", user: "alice", namespace: "ai-team-a", url: "/api/v1/ai/applications?namespace=ai-team-a", wantStatus: http.StatusOK, wantNamespace: "ai-team-a"},
		{name: "out of scope", user: "alice", namespace: "ai-team-a", url: "/api/v1/ai/applications?namespace=ai-team-b", wantStatus: http.StatusForbidden},
		{name: "detail out of scope", user: "alice", namespace: "ai-team-a", url: "/api/v1/ai/applications/ai-team-b/app", wantStatus: http.StatusForbidden},
		{name: "no namespace", user: "alice", url: "/api/v1/ai/applications", wantStatus: http.StatusForbidden},
		{name: "invalid namespace", user: "alice", namespace: "Bad_NS", url: "/api/v1/ai/applications", wantStatus: http.StatusForbidden},
		{name: "falls back to static account", user: "tenant-b", url: "/api/v1/ai/applications", wantStatus: http.StatusOK, wantNamespace: "ai-tenant-b"},
		{name: "monitor unscoped", user: "root", groups: "platform-admin", url: "/api/v1/ai/applications?namespace=anything", wantStatus: http.StatusOK, wantNamespace: "anything"},
		{name: "monitor ignores namespace header", user: "root", groups: "platform-admin", namespace: "ai-team-a", url: "/api/v1/ai/applications", wantStatus: http.StatusOK, wantNamespace: ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reader.listNamespace = "unset"
			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			req.Header.Set("X-Remote-User", tc.user)
			if tc.groups != "" {
				req.Header.Set("X-Remote-Groups", tc.groups)
			}
			if tc.tenant != "" {
				req.Header.Set("X-Remote-Tenant", tc.tenant)
			}
			if tc.namespace != "" {
				req.Header.Set("X-Remote-Namespace", tc.namespace)
			}
			rec := httptest.NewRecorder()
			server.ServeHTTP(rec, req)
			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d: %s", rec.Code, tc.wantStatus, rec.Body.String())
			}
			if tc.wantStatus == http.StatusOK && reader.listNamespace != tc.wantNamespace {
				t.Fatalf("reader namespace = %q, want %q", reader.listNamespace, tc.wantNamespace)
			}
		})
	}
}

func TestRemoteHeadersUseConfiguredTenantHeaders(t *testing.T) {
	auth := newConsoleAuthWithConfig(&authConfig{
		RemoteAuthEnabled:     true,
		RemoteTenantHeader:    "X-Test-Tenant",
		RemoteNamespaceHeader: "X-Test-Namespace",
	})
	req := httptest.NewRequest(http.MethodGet, "/user", nil)
	req.Header.Set("X-Remote-User", "alice")
	req.Header.Set("X-Test-Tenant", "team-a")
	req.Header.Set("X-Test-Namespace", "ai-team-a")
	req.Header.Set("X-Remote-Namespace", "ai-other")
	session, ok := auth.authenticatedSession(req)
	if !ok {
		t.Fatal("expected a session from the headers")
	}
	if session.Tenant != "team-a" || session.Namespace != "ai-team-a" || !session.Remote {
		t.Fatalf("session = %+v, want tenant team-a in ai-team-a from the proxy", session)
	}
}

func TestSessionEndpointReportsScopeAndBasePath(t *testing.T) {
	server := NewServerWithOptions(Options{Auth: &authConfig{RemoteAuthEnabled: true}})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ai/session", nil)
	req.Header.Set("X-Remote-User", "alice")
	req.Header.Set("X-Remote-Tenant", "team-a")
	req.Header.Set("X-Remote-Namespace", "ai-team-a")
	req.Header.Set(forwardedPrefixHeader, "/ai-console")
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}
	var payload map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	want := map[string]string{"username": "alice", "role": "user", "tenant": "team-a", "namespace": "ai-team-a", "basePath": "/ai-console"}
	for key, value := range want {
		if payload[key] != value {
			t.Fatalf("%s = %q, want %q (%v)", key, payload[key], value, payload)
		}
	}
}

func TestForwardedPrefixIgnoredWithoutRemoteAuth(t *testing.T) {
	t.Setenv("KUBEVELA_AI_REMOTE_AUTH_ENABLED", "false")
	t.Setenv("KUBEVELA_AI_BASE_PATH", "")
	server := NewServerWithOptions(Options{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(forwardedPrefixHeader, "/elsewhere")
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)
	if loc := rec.Header().Get("Location"); loc != "/login" {
		t.Fatalf("location = %q, want /login", loc)
	}
}

func TestConsolePagesOnlyReferenceServedAssets(t *testing.T) {
	server := authenticatedServer(t, NewServerWithOptions(Options{}))
	asset := regexp.MustCompile(`/console/([A-Za-z0-9._-]+)`)
	for name, page := range map[string]string{"index.html": userConsoleHTML(), "monitor.html": monitorConsoleHTML()} {
		for _, match := range asset.FindAllStringSubmatch(page, -1) {
			rec := httptest.NewRecorder()
			server.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/console/"+match[1], nil))
			if rec.Code != http.StatusOK {
				t.Errorf("%s references /console/%s, served with status %d", name, match[1], rec.Code)
			}
		}
	}
}
