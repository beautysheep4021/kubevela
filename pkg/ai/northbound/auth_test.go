package northbound

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestConsoleAuthIgnoresRemoteHeadersByDefault(t *testing.T) {
	t.Setenv("KUBEVELA_AI_REMOTE_AUTH_ENABLED", "false")
	auth := newConsoleAuth()
	req := httptest.NewRequest(http.MethodGet, "/monitor", nil)
	req.Header.Set("X-Remote-User", "alice")
	req.Header.Set("X-Remote-Groups", "system:masters")

	if _, ok := auth.authenticatedSession(req); ok {
		t.Fatal("expected remote headers to be ignored by default")
	}
}

func TestConsoleAuthResolvesRemoteHeadersToMonitorRole(t *testing.T) {
	auth := newConsoleAuthWithConfig(&authConfig{
		RemoteAuthEnabled:  true,
		RemoteUserHeader:   "X-Test-User",
		RemoteGroupsHeader: "X-Test-Groups",
		AdminGroups:        []string{"platform-admin", "system:masters"},
	})
	req := httptest.NewRequest(http.MethodGet, "/monitor", nil)
	req.Header.Set("X-Test-User", "alice")
	req.Header.Set("X-Test-Groups", "dev,platform-admin")

	session, ok := auth.authenticatedSession(req)
	if !ok {
		t.Fatal("expected remote headers to resolve an authenticated session")
	}
	if session.Username != "alice" {
		t.Fatalf("username = %q, want alice", session.Username)
	}
	if session.Role != consoleRoleMonitor {
		t.Fatalf("role = %q, want monitor", session.Role)
	}
}

func TestServerUsesConfiguredRemoteHeadersForK8sAdminRole(t *testing.T) {
	server := NewServerWithOptions(Options{Auth: &authConfig{
		RemoteAuthEnabled:  true,
		RemoteUserHeader:   "X-Test-User",
		RemoteGroupsHeader: "X-Test-Groups",
		AdminGroups:        []string{"platform-admin", "system:masters"},
	}})

	rootReq := httptest.NewRequest(http.MethodGet, "/", nil)
	rootReq.Header.Set("X-Test-User", "alice")
	rootReq.Header.Set("X-Test-Groups", "dev,platform-admin")
	rootRec := httptest.NewRecorder()
	server.ServeHTTP(rootRec, rootReq)

	if rootRec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rootRec.Code, http.StatusFound)
	}
	if location := rootRec.Header().Get("Location"); location != "/monitor" {
		t.Fatalf("location = %q, want /monitor", location)
	}

	pageReq := httptest.NewRequest(http.MethodGet, "/monitor", nil)
	pageReq.Header.Set("X-Test-User", "alice")
	pageReq.Header.Set("X-Test-Groups", "dev,platform-admin")
	pageRec := httptest.NewRecorder()
	server.ServeHTTP(pageRec, pageReq)

	if pageRec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", pageRec.Code, http.StatusOK)
	}
	body := pageRec.Body.String()
	for _, expected := range []string{"K8s 管理员工作台", "K8s 管理员", "data-role=\"monitor\""} {
		if !strings.Contains(body, expected) {
			t.Fatalf("monitor page missing %q", expected)
		}
	}
}

func TestConsoleAuthPrefersSessionCookieOverRemoteHeaders(t *testing.T) {
	auth := newConsoleAuthWithConfig(&authConfig{
		RemoteAuthEnabled:  true,
		RemoteUserHeader:   "X-Remote-User",
		RemoteGroupsHeader: "X-Remote-Groups",
		AdminGroups:        []string{"system:masters"},
	})
	token, err := auth.createSession(consoleRoleUser, "bob")
	if err != nil {
		t.Fatalf("createSession: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/user", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	req.Header.Set("X-Remote-User", "alice")
	req.Header.Set("X-Remote-Groups", "system:masters")

	session, ok := auth.authenticatedSession(req)
	if !ok {
		t.Fatal("expected cookie session to resolve")
	}
	if session.Username != "bob" {
		t.Fatalf("username = %q, want bob", session.Username)
	}
	if session.Role != consoleRoleUser {
		t.Fatalf("role = %q, want user", session.Role)
	}
}

func TestConsoleAuthAllowsTenantBScopedAccount(t *testing.T) {
	auth := newConsoleAuth()
	form := url.Values{}
	form.Set("role", string(consoleRoleUser))
	form.Set("username", "tenant-b")
	form.Set("password", "tenant-b-123456")

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	auth.handleLogin(recorder, request)

	if recorder.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusFound)
	}

	var cookie *http.Cookie
	for _, candidate := range recorder.Result().Cookies() {
		if candidate.Name == sessionCookieName {
			cookie = candidate
			break
		}
	}
	if cookie == nil {
		t.Fatal("login response missing session cookie")
	}

	sessionRequest := httptest.NewRequest(http.MethodGet, "/user", nil)
	sessionRequest.AddCookie(cookie)
	session, ok := auth.authenticatedSession(sessionRequest)
	if !ok {
		t.Fatal("expected tenant-b account to authenticate")
	}
	if session.Username != "tenant-b" || session.Role != consoleRoleUser {
		t.Fatalf("unexpected session identity: %#v", session)
	}
	if session.Tenant != "tenant-b" || session.Namespace != "ai-tenant-b" {
		t.Fatalf("unexpected tenant scope: %#v", session)
	}
}

func TestConsoleAuthScopesPrimaryUserAccount(t *testing.T) {
	auth := newConsoleAuth()
	form := url.Values{}
	form.Set("role", string(consoleRoleUser))
	form.Set("username", "admin")
	form.Set("password", "shiyong")

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	auth.handleLogin(recorder, request)

	if recorder.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusFound)
	}

	var cookie *http.Cookie
	for _, candidate := range recorder.Result().Cookies() {
		if candidate.Name == sessionCookieName {
			cookie = candidate
			break
		}
	}
	if cookie == nil {
		t.Fatal("login response missing session cookie")
	}

	sessionRequest := httptest.NewRequest(http.MethodGet, "/user", nil)
	sessionRequest.AddCookie(cookie)
	session, ok := auth.authenticatedSession(sessionRequest)
	if !ok {
		t.Fatal("expected primary user account to authenticate")
	}
	if session.Username != "admin" || session.Role != consoleRoleUser {
		t.Fatalf("unexpected session identity: %#v", session)
	}
	if session.Tenant != "tenant-a" || session.Namespace != "ai-tenant-a" {
		t.Fatalf("unexpected tenant scope: %#v", session)
	}
}
