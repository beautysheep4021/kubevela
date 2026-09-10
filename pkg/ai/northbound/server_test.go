package northbound

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/oam-dev/kubevela/pkg/ai/domain"
	domainapply "github.com/oam-dev/kubevela/pkg/ai/domain/apply"
	"github.com/oam-dev/kubevela/pkg/ai/observe"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestServerHealthz(t *testing.T) {
	server := NewServer()
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if strings.TrimSpace(recorder.Body.String()) != "ok" {
		t.Fatalf("body = %q, want ok", recorder.Body.String())
	}
}

func TestServerRedirectsAnonymousUsersToLogin(t *testing.T) {
	server := NewServer()
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if recorder.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusFound)
	}
	if location := recorder.Header().Get("Location"); location != "/login" {
		t.Fatalf("location = %q, want /login", location)
	}
}

func TestServerServesLoginPage(t *testing.T) {
	server := NewServer()
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/login", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	contentType := recorder.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Fatalf("content-type = %q, want text/html", contentType)
	}
	body := recorder.Body.String()
	for _, expected := range []string{
		"登录",
		"角色选择",
		"使用方",
		"K8s 管理员",
		"admin / shiyong",
		"tenant-b / tenant-b-123456",
		"admin / jiankong",
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("login page missing %q", expected)
		}
	}
}

func TestServerLoginFlowRoutesByRole(t *testing.T) {
	server := NewServer()
	cases := []struct {
		name      string
		role      string
		password  string
		wantPath  string
		wantTitle string
	}{
		{name: "user", role: "user", password: "shiyong", wantPath: "/user", wantTitle: "使用方工作台"},
		{name: "monitor", role: "monitor", password: "jiankong", wantPath: "/monitor", wantTitle: "K8s 管理员工作台"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cookie := loginForRole(t, server, tc.role, "admin", tc.password)

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, tc.wantPath, nil)
			request.AddCookie(cookie)
			server.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
			}
			body := recorder.Body.String()
			for _, expected := range []string{
				tc.wantTitle,
				"data-role=\"" + tc.role + "\"",
				"退出登录",
			} {
				if !strings.Contains(body, expected) {
					t.Fatalf("console page missing %q", expected)
				}
			}

			otherPath := "/monitor"
			if tc.wantPath == "/monitor" {
				otherPath = "/user"
			}
			redirect := httptest.NewRecorder()
			otherReq := httptest.NewRequest(http.MethodGet, otherPath, nil)
			otherReq.AddCookie(cookie)
			server.ServeHTTP(redirect, otherReq)
			if redirect.Code != http.StatusFound {
				t.Fatalf("status = %d, want %d", redirect.Code, http.StatusFound)
			}
			if location := redirect.Header().Get("Location"); location != tc.wantPath {
				t.Fatalf("location = %q, want %q", location, tc.wantPath)
			}
		})
	}
}

func TestServerRendersScopedConsoleAccountContext(t *testing.T) {
	server := NewServer()
	cookie := loginForRole(t, server, "user", "tenant-b", "tenant-b-123456")

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/user", nil)
	request.AddCookie(cookie)
	server.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	body := recorder.Body.String()
	for _, expected := range []string{
		`data-account-tenant="tenant-b"`,
		`data-account-namespace="ai-tenant-b"`,
		"var accountNamespace = (document.body.getAttribute(\"data-account-namespace\") || \"\").trim();",
		"function applyAccountScopeDefaults()",
		"if (!accountNamespace)",
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("scoped console page missing %q", expected)
		}
	}
}

func TestConsoleDoesNotFallbackToDemoDataForScopedAccounts(t *testing.T) {
	for _, expected := range []string{
		"if (demoMode && demo !== null)",
		"if (demo.__demoError)",
		"function demoNamespaceError(",
		"requestedNamespace || accountNamespace",
	} {
		if !strings.Contains(consoleHTML, expected) {
			t.Fatalf("console demo isolation guard missing %q", expected)
		}
	}
	if strings.Contains(consoleHTML, "if (demo !== null) {\n          return demoClone(demo);") {
		t.Fatal("console must not fall back to demo data after a live API error")
	}
}

func TestConsoleDemoListValidationDoesNotRejectValidNamespace(t *testing.T) {
	const expected = `      if (pathNamespace) {
        var pathNamespaceError = demoNamespaceError(pathNamespace);
        if (pathNamespaceError) {
          return demoError(pathNamespaceError);
        }
      }`
	if !strings.Contains(consoleHTML, expected) {
		t.Fatalf("console demo path namespace validation must ignore empty list path namespace")
	}
}

func TestServerBlocksAnonymousApiRequests(t *testing.T) {
	server := NewServer()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/ai/validate", strings.NewReader("kind: AIJob"))
	server.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestServerPreservesDemoQueryOnLoginRedirect(t *testing.T) {
	server := NewServer()
	cases := []struct {
		name     string
		loginURL string
		want     string
	}{
		{name: "user", loginURL: "/login?demo=local", want: "/user?demo=local"},
		{name: "monitor", loginURL: "/login?demo=local", want: "/monitor?demo=local"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			form := url.Values{}
			if tc.name == "user" {
				form.Set("role", "user")
				form.Set("username", "admin")
				form.Set("password", "shiyong")
			} else {
				form.Set("role", "monitor")
				form.Set("username", "admin")
				form.Set("password", "jiankong")
			}
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, tc.loginURL, strings.NewReader(form.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			server.ServeHTTP(recorder, request)
			if recorder.Code != http.StatusFound {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusFound)
			}
			if location := recorder.Header().Get("Location"); location != tc.want {
				t.Fatalf("location = %q, want %q", location, tc.want)
			}
		})
	}
}

func TestConsoleShowsAndEnforcesIsolationPolicy(t *testing.T) {
	for _, expected := range []string{
		"隔离策略预览",
		"id=\"isolationPolicyPreview\"",
		"Namespace：",
		"ResourceQuota：",
		"LimitRange：",
		"配额校验：",
		"function renderIsolationPolicyPreview",
		"function validateIsolationPolicy",
		"资源隔离校验未通过",
		"validateIsolationPolicy()",
		"parseComputeQuantity",
	} {
		if !strings.Contains(consoleHTML, expected) {
			t.Fatalf("console isolation policy enforcement missing %q", expected)
		}
	}
}

func TestConsoleShowsMonitorResourceOverviewAndTenantView(t *testing.T) {
	for _, forbidden := range []string{
		"基础模型 URI",
		"训练数据 URI",
		"待发布模型 URI",
		"评测数据 URI",
		"数据集内部 URI",
	} {
		if strings.Contains(consoleHTML, forbidden) {
			t.Fatalf("console page still exposes user-facing URI label %q", forbidden)
		}
	}
}

func TestConsoleUsesSidebarAndSinglePublishServiceEntry(t *testing.T) {
	for _, expected := range []string{
		"class=\"sidebar\"",
		"data-section=\"training\"",
		"data-section=\"platform\"",
		"data-section=\"overview\"",
		"data-section=\"tasks\"",
		"data-section=\"audits\"",
		"id=\"nav-training\"",
		"id=\"nav-platform\"",
		"id=\"nav-monitor-overview\"",
		"id=\"nav-monitor-tasks\"",
		"id=\"nav-monitor-audits\"",
		"id=\"publish-delivery-service\"",
		"id=\"audit-list\"",
	} {
		if !strings.Contains(consoleHTML, expected) {
			t.Fatalf("console page missing %q", expected)
		}
	}
	for _, forbidden := range []string{
		"id=\"template-service\"",
		"发布服务</strong>",
	} {
		if strings.Contains(consoleHTML, forbidden) {
			t.Fatalf("console page still contains duplicate publish-service affordance %q", forbidden)
		}
	}
}

func TestConsoleUsesLightThemeAndFoldableSections(t *testing.T) {
	for _, expected := range []string{
		"--paper: #f2f3f3;",
		"--panel: #ffffff;",
		"--line: #d5d9d9;",
		"font-family: -apple-system",
		"class=\"fold-section\"",
		"summary>资源与调度（以 K8s 调度为准）</summary",
		"summary>租户资源隔离</summary",
		"summary>评测配置</summary",
		"summary>应用交付</summary",
		"summary>筛选条件</summary",
		"任务类型（AIService / AIJob）",
		"载入服务画像",
		"服务画像基座",
		"服务画像会把当前模型引用整理为 AIService 配置，不新增运行时类型。",
		"Kubernetes 调度：source of truth",
		"集群能力边界：CPU Manager static / CPU pinning、GPU device plugin、Topology Manager",
	} {
		if !strings.Contains(consoleHTML, expected) {
			t.Fatalf("console page missing %q", expected)
		}
	}
}

func TestConsoleMakesAIServiceAndAIJobExplicit(t *testing.T) {
	for _, expected := range []string{
		"<option value=\"AIService\">AIService</option>",
		"<option value=\"AIJob\">AIJob</option>",
		"高级配置：查看和调整底层 AIService / AIJob 字段",
		"当前调度映射（以 K8s 调度为准）",
	} {
		if !strings.Contains(consoleHTML, expected) {
			t.Fatalf("console page missing %q", expected)
		}
	}
}

func TestConsoleSeparatesTrainingAndEvaluationModes(t *testing.T) {
	for _, expected := range []string{
		"summary>训练配置</summary",
		"id=\"evaluation-band\"",
		"id=\"workflow-model-hint\"",
		"function syncWorkflowMode(",
		"function currentJobDatasetURI(",
		"evaluation_threshold=",
		"summary: template === \"evaluation\" ? \"evaluation-passed\"",
	} {
		if !strings.Contains(consoleHTML, expected) {
			t.Fatalf("console page missing %q", expected)
		}
	}
}

func TestConsoleMonitorDashboardAddsChartsAndExtraKPIs(t *testing.T) {
	for _, expected := range []string{
		"id=\"monitor-kpi-grid\"",
		"id=\"monitor-trend-chart\"",
		"id=\"monitor-status-chart\"",
		"id=\"monitor-tenant-chart\"",
		"最近刷新趋势",
		"任务健康分布",
		"租户任务 Top5",
		"健康率",
		"活跃租户",
		"function buildMonitorSummary(",
		"function renderMonitorTrendChart(",
		"function renderMonitorStatusChart(",
		"function renderMonitorTenantChart(",
		"var monitorHistory = []",
	} {
		if !strings.Contains(consoleHTML, expected) {
			t.Fatalf("console monitor dashboard missing %q", expected)
		}
	}
}

func TestConsoleContainsDemoModeSupport(t *testing.T) {
	for _, expected := range []string{
		"var demoMode =",
		"var demoState =",
		"function demoResponse(",
		"demo=local",
	} {
		if !strings.Contains(consoleHTML, expected) {
			t.Fatalf("console page missing %q", expected)
		}
	}
}

func loginForRole(t *testing.T, server http.Handler, role, username, password string) *http.Cookie {
	t.Helper()
	form := url.Values{}
	form.Set("role", role)
	form.Set("username", username)
	form.Set("password", password)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	server.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusFound)
	}
	if recorder.Header().Get("Location") == "" {
		t.Fatal("login response missing redirect location")
	}
	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name != "" {
			return cookie
		}
	}
	t.Fatal("login response missing session cookie")
	return nil
}

type authenticatedHandler struct {
	http.Handler
	cookie *http.Cookie
}

func (h authenticatedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/login" && r.URL.Path != "/healthz" && r.URL.Path != "/" {
		if _, err := r.Cookie(sessionCookieName); err != nil {
			r.AddCookie(h.cookie)
		}
	}
	h.Handler.ServeHTTP(w, r)
}

func authenticatedServer(t *testing.T, handler http.Handler) http.Handler {
	t.Helper()
	return authenticatedServerAsRole(t, handler, "monitor", "admin", "jiankong")
}

func authenticatedServerAs(t *testing.T, handler http.Handler, username, password string) http.Handler {
	return authenticatedServerAsRole(t, handler, "user", username, password)
}

func authenticatedServerAsRole(t *testing.T, handler http.Handler, role, username, password string) http.Handler {
	t.Helper()
	return authenticatedHandler{
		Handler: handler,
		cookie:  loginForRole(t, handler, role, username, password),
	}
}

func TestServerListsApplications(t *testing.T) {
	reader := &recordingApplicationReader{
		items: []observe.ApplicationListItem{
			{
				Name:          "ai-service-northbound-demo",
				Namespace:     "sock-shop",
				Phase:         "running",
				Healthy:       true,
				Message:       "Ready:1/1",
				WorkloadTypes: []string{"service"},
				AIMetadata: map[string]string{
					"ai.oam.dev/tenant": "demo-tenant",
				},
			},
		},
	}
	server := authenticatedServer(t, NewServerWithOptions(Options{Reader: reader}))
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/ai/applications?namespace=sock-shop", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var payload ApplicationsResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if reader.listNamespace != "sock-shop" {
		t.Fatalf("namespace = %q, want sock-shop", reader.listNamespace)
	}
	if len(payload.Items) != 1 || payload.Items[0].Name != "ai-service-northbound-demo" || payload.Items[0].WorkloadTypes[0] != "service" {
		t.Fatalf("unexpected application list: %#v", payload)
	}
}

func TestServerScopesTenantBApplicationListToItsNamespace(t *testing.T) {
	reader := &recordingApplicationReader{}
	server := authenticatedServerAs(t, NewServerWithOptions(Options{Reader: reader}), "tenant-b", "tenant-b-123456")

	for _, tc := range []struct {
		name          string
		url           string
		wantStatus    int
		wantNamespace string
	}{
		{
			name:          "defaults to account namespace",
			url:           "/api/v1/ai/applications",
			wantStatus:    http.StatusOK,
			wantNamespace: "ai-tenant-b",
		},
		{
			name:       "rejects another namespace",
			url:        "/api/v1/ai/applications?namespace=ai-tenant-a",
			wantStatus: http.StatusForbidden,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reader.listNamespace = ""
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.url, nil))

			if recorder.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d: %s", recorder.Code, tc.wantStatus, recorder.Body.String())
			}
			if reader.listNamespace != tc.wantNamespace {
				t.Fatalf("reader namespace = %q, want %q", reader.listNamespace, tc.wantNamespace)
			}
		})
	}
}

func TestServerScopesPrimaryUserApplicationListToItsNamespace(t *testing.T) {
	reader := &recordingApplicationReader{}
	server := authenticatedServerAs(t, NewServerWithOptions(Options{Reader: reader}), "admin", "shiyong")

	for _, tc := range []struct {
		name          string
		url           string
		wantStatus    int
		wantNamespace string
	}{
		{
			name:          "defaults to account namespace",
			url:           "/api/v1/ai/applications",
			wantStatus:    http.StatusOK,
			wantNamespace: "ai-tenant-a",
		},
		{
			name:       "rejects another namespace",
			url:        "/api/v1/ai/applications?namespace=ai-tenant-b",
			wantStatus: http.StatusForbidden,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reader.listNamespace = ""
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.url, nil))

			if recorder.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d: %s", recorder.Code, tc.wantStatus, recorder.Body.String())
			}
			if tc.wantNamespace != "" && reader.listNamespace != tc.wantNamespace {
				t.Fatalf("namespace = %q, want %q", reader.listNamespace, tc.wantNamespace)
			}
		})
	}
}

func TestConsoleIncludesIsolatedDemoTaskPairs(t *testing.T) {
	for _, expected := range []string{
		`namespace: "ai-tenant-a"`,
		`namespace: "ai-tenant-b"`,
		`name: "tenant-a-training-job"`,
		`name: "tenant-a-inference-service"`,
		`name: "tenant-b-evaluation-job"`,
		`name: "tenant-b-chat-service"`,
	} {
		if !strings.Contains(consoleHTML, expected) {
			t.Fatalf("console demo data missing %q", expected)
		}
	}
}

func TestServerRejectsTenantBDeploymentOutsideItsNamespace(t *testing.T) {
	applier := &recordingApplicationApplier{}
	server := authenticatedServerAs(t, NewServerWithOptions(Options{Applier: applier}), "tenant-b", "tenant-b-123456")
	content := strings.Replace(validAIServiceYAML(), "namespace: ai-demo", "namespace: ai-platform", 1)

	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/ai/applications", strings.NewReader(content)))

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusForbidden, recorder.Body.String())
	}
	if applier.called {
		t.Fatal("applier should not be called for a cross-tenant deployment")
	}
}

func TestServerReturnsApplicationStatus(t *testing.T) {
	reader := &recordingApplicationReader{
		summaries: map[string]*observe.Summary{
			"sock-shop/ai-service-northbound-demo": {
				Name:      "ai-service-northbound-demo",
				Namespace: "sock-shop",
				Phase:     "running",
				Healthy:   true,
				Message:   "Ready:1/1",
			},
		},
	}
	server := authenticatedServer(t, NewServerWithOptions(Options{Reader: reader}))
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/ai/applications/sock-shop/ai-service-northbound-demo/status", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var summary observe.Summary
	if err := json.Unmarshal(recorder.Body.Bytes(), &summary); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if summary.Name != "ai-service-northbound-demo" || summary.Namespace != "sock-shop" || !summary.Healthy {
		t.Fatalf("unexpected summary: %#v", summary)
	}
}

func TestServerReturnsApplicationLogs(t *testing.T) {
	reader := &recordingApplicationReader{
		logs: map[string]*observe.Logs{
			"sock-shop/ai-job-demo": {
				Namespace:   "sock-shop",
				Application: "ai-job-demo",
				Pod:         "ai-job-demo-pod",
				Container:   "main",
				TailLines:   80,
				Logs:        "epoch=1 loss=0.42\n",
				Pods: []observe.LogPod{
					{
						Name:       "ai-job-demo-pod",
						Phase:      "Succeeded",
						Containers: []string{"main"},
					},
				},
			},
		},
	}
	server := authenticatedServer(t, NewServerWithOptions(Options{Reader: reader}))
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ai/applications/sock-shop/ai-job-demo/logs?pod=ai-job-demo-pod&container=main&tailLines=80", nil)
	server.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var payload observe.Logs
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if payload.Pod != "ai-job-demo-pod" || payload.Container != "main" || payload.TailLines != 80 || !strings.Contains(payload.Logs, "loss=0.42") {
		t.Fatalf("unexpected logs payload: %#v", payload)
	}
	if reader.logRequest.Namespace != "sock-shop" || reader.logRequest.Name != "ai-job-demo" || reader.logRequest.Pod != "ai-job-demo-pod" || reader.logRequest.Container != "main" || reader.logRequest.TailLines != 80 {
		t.Fatalf("unexpected log request: %#v", reader.logRequest)
	}
}

func TestServerProbesApplicationService(t *testing.T) {
	prober := &recordingApplicationProber{
		result: &observe.ProbeResult{
			Namespace:   "sock-shop",
			Application: "delivery-service",
			Path:        "/healthz",
			StatusCode:  200,
			Healthy:     true,
			Body:        `{"status":"ok"}`,
		},
	}
	server := authenticatedServer(t, NewServerWithOptions(Options{Prober: prober}))
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/applications/sock-shop/delivery-service/probe", strings.NewReader(`{"path":"/healthz","timeoutSeconds":3}`))
	req.Header.Set("Content-Type", "application/json")
	server.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var payload observe.ProbeResult
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if !payload.Healthy || payload.StatusCode != 200 || payload.Body != `{"status":"ok"}` {
		t.Fatalf("unexpected probe payload: %#v", payload)
	}
	if prober.request.Namespace != "sock-shop" || prober.request.Name != "delivery-service" || prober.request.Path != "/healthz" || prober.request.TimeoutSeconds != 3 {
		t.Fatalf("unexpected probe request: %#v", prober.request)
	}
}

func TestServerDeletesApplicationAndRecordsAudit(t *testing.T) {
	manager := &recordingApplicationManager{}
	audits := &recordingAuditStore{}
	server := authenticatedServer(t, NewServerWithOptions(Options{Manager: manager, Audits: audits}))
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/ai/applications/sock-shop/ai-job-demo", nil)
	req.Header.Set("X-AI-User", "tester")
	server.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var payload observe.LifecycleResult
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if payload.Action != "delete" || payload.Namespace != "sock-shop" || payload.Name != "ai-job-demo" {
		t.Fatalf("unexpected lifecycle result: %#v", payload)
	}
	if manager.deleted != "sock-shop/ai-job-demo" {
		t.Fatalf("deleted = %q, want sock-shop/ai-job-demo", manager.deleted)
	}
	if len(audits.events) != 1 || audits.events[0].Action != "delete" || audits.events[0].Actor != "tester" || audits.events[0].Success != true {
		t.Fatalf("unexpected audit events: %#v", audits.events)
	}
}

func TestServerRestartsApplicationAndRecordsAudit(t *testing.T) {
	manager := &recordingApplicationManager{}
	audits := &recordingAuditStore{}
	server := authenticatedServer(t, NewServerWithOptions(Options{Manager: manager, Audits: audits}))
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/applications/sock-shop/ai-service-demo/restart", nil)
	server.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if manager.restarted != "sock-shop/ai-service-demo" {
		t.Fatalf("restarted = %q, want sock-shop/ai-service-demo", manager.restarted)
	}
	if len(audits.events) != 1 || audits.events[0].Action != "restart" || audits.events[0].Actor != "anonymous" {
		t.Fatalf("unexpected audit events: %#v", audits.events)
	}
}

func TestServerRerunsApplicationAndRecordsAudit(t *testing.T) {
	manager := &recordingApplicationManager{}
	audits := &recordingAuditStore{}
	server := authenticatedServer(t, NewServerWithOptions(Options{Manager: manager, Audits: audits}))
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/applications/sock-shop/ai-job-demo/rerun", nil)
	server.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if manager.rerun != "sock-shop/ai-job-demo" {
		t.Fatalf("rerun = %q, want sock-shop/ai-job-demo", manager.rerun)
	}
	if len(audits.events) != 1 || audits.events[0].Action != "rerun" {
		t.Fatalf("unexpected audit events: %#v", audits.events)
	}
}

func TestServerListsAuditEvents(t *testing.T) {
	audits := &recordingAuditStore{
		events: []observe.AuditEvent{
			{
				Namespace: "sock-shop",
				Name:      "ai-job-demo",
				Action:    "delete",
				Actor:     "tester",
				Success:   true,
			},
		},
	}
	server := authenticatedServer(t, NewServerWithOptions(Options{Audits: audits}))
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/ai/audits?namespace=sock-shop", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var payload observe.AuditList
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if len(payload.Items) != 1 || payload.Items[0].Namespace != "sock-shop" || payload.Items[0].Action != "delete" {
		t.Fatalf("unexpected audits: %#v", payload)
	}
	if audits.listNamespace != "sock-shop" {
		t.Fatalf("list namespace = %q, want sock-shop", audits.listNamespace)
	}
}

func TestServerExtractsDeliveryResultFromAIJobLogs(t *testing.T) {
	reader := &recordingApplicationReader{
		logs: map[string]*observe.Logs{
			"sock-shop/train-demo": {
				Namespace:   "sock-shop",
				Application: "train-demo",
				Pod:         "train-demo-pod",
				Container:   "trainer",
				Logs: strings.Join([]string{
					"epoch=1 loss=0.3",
					`AI_RESULT_JSON={"modelURI":"inline://models/train-demo/v1","metrics":{"loss":0.12,"accuracy":0.98},"summary":"trained"}`,
					"done",
				}, "\n"),
			},
		},
	}
	server := authenticatedServer(t, NewServerWithOptions(Options{Reader: reader}))
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/ai/deliveries/sock-shop/train-demo/result", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var payload DeliveryResultResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if payload.Namespace != "sock-shop" || payload.JobName != "train-demo" || payload.Result.ModelURI != "inline://models/train-demo/v1" {
		t.Fatalf("unexpected delivery result: %#v", payload)
	}
	if payload.Result.Metrics["accuracy"] != 0.98 || payload.Result.Summary != "trained" {
		t.Fatalf("unexpected metrics: %#v", payload.Result)
	}
}

func TestServerPublishesDeliveryResultAsAIService(t *testing.T) {
	reader := &recordingApplicationReader{
		logs: map[string]*observe.Logs{
			"sock-shop/train-demo": {
				Logs: `AI_RESULT_JSON={"modelURI":"inline://models/train-demo/v1","metrics":{"loss":0.12}}`,
			},
		},
	}
	applier := &recordingApplicationApplier{}
	audits := &recordingAuditStore{}
	server := authenticatedServer(t, NewServerWithOptions(Options{Reader: reader, Applier: applier, Audits: audits}))
	body := strings.NewReader(`{"serviceName":"train-demo-service","image":"python:3.11-slim","port":8080,"servicePort":80}`)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/deliveries/sock-shop/train-demo/publish-service", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AI-User", "tester")
	server.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var payload DeliveryPublishResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if payload.ServiceName != "train-demo-service" || payload.ModelURI != "inline://models/train-demo/v1" {
		t.Fatalf("unexpected publish response: %#v", payload)
	}
	if !applier.called || applier.namespace != "sock-shop" || applier.name != "train-demo-service" {
		t.Fatalf("unexpected applier call: %#v", applier)
	}
	content := string(applier.content)
	for _, expected := range []string{"kind: Application", "type: ai-service", "modelURI: inline://models/train-demo/v1", "name: train-demo-service", "cmd:", "socketserver.TCPServer", "serve model"} {
		if !strings.Contains(content, expected) {
			t.Fatalf("expected translated AIService Application to contain %q, got:\n%s", expected, content)
		}
	}
	if len(audits.events) != 1 || audits.events[0].Action != "publish-service" || audits.events[0].Actor != "tester" {
		t.Fatalf("unexpected audit events: %#v", audits.events)
	}
}

func TestServerPublishesDeliveryResultAndRecordsArtifact(t *testing.T) {
	reader := &recordingApplicationReader{
		logs: map[string]*observe.Logs{
			"sock-shop/train-demo": {
				Pod:  "train-demo-pod",
				Logs: `AI_RESULT_JSON={"modelURI":"inline://models/train-demo/v1","metrics":{"loss":0.12},"summary":"trained"}`,
			},
		},
	}
	artifacts := &recordingArtifactStore{}
	server := authenticatedServer(t, NewServerWithOptions(Options{Reader: reader, Applier: &recordingApplicationApplier{}, Artifacts: artifacts}))
	body := strings.NewReader(`{"serviceName":"train-demo-service"}`)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/ai/deliveries/sock-shop/train-demo/publish-service", body))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if len(artifacts.saved) != 1 {
		t.Fatalf("saved artifacts = %d, want 1", len(artifacts.saved))
	}
	artifact := artifacts.saved[0]
	if artifact.Namespace != "sock-shop" || artifact.JobName != "train-demo" || artifact.ModelURI != "inline://models/train-demo/v1" {
		t.Fatalf("unexpected artifact: %#v", artifact)
	}
	if len(artifact.PublishedServices) != 1 || artifact.PublishedServices[0] != "train-demo-service" {
		t.Fatalf("unexpected published services: %#v", artifact.PublishedServices)
	}
	if artifact.Status != "published" || artifact.Visibility != "private" || artifact.EvaluationStatus != "passed" {
		t.Fatalf("unexpected artifact lifecycle fields: %#v", artifact)
	}
}

func TestServerListsArtifacts(t *testing.T) {
	artifacts := &recordingArtifactStore{
		list: []observe.ModelArtifact{
			{
				Namespace: "sock-shop",
				Name:      "train-demo",
				JobName:   "train-demo",
				ModelURI:  "inline://models/train-demo/v1",
			},
		},
	}
	server := authenticatedServer(t, NewServerWithOptions(Options{Artifacts: artifacts}))
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/ai/artifacts?namespace=sock-shop", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var payload observe.ArtifactList
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if len(payload.Items) != 1 || payload.Items[0].ModelURI != "inline://models/train-demo/v1" {
		t.Fatalf("unexpected artifacts: %#v", payload)
	}
	if artifacts.listNamespace != "sock-shop" {
		t.Fatalf("namespace = %q, want sock-shop", artifacts.listNamespace)
	}
}

func TestServerListsModelAssets(t *testing.T) {
	artifacts := &recordingArtifactStore{
		list: []observe.ModelArtifact{
			{
				Namespace:        "sock-shop",
				Name:             "customer-sft-demo",
				ModelURI:         "inline://models/customer-sft-demo/v1",
				Status:           "evaluated",
				Visibility:       "private",
				EvaluationStatus: "passed",
			},
		},
	}
	server := authenticatedServer(t, NewServerWithOptions(Options{Artifacts: artifacts}))
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/ai/models?namespace=sock-shop", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var payload observe.ArtifactList
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if len(payload.Items) != 1 || payload.Items[0].Name != "customer-sft-demo" || payload.Items[0].Visibility != "private" {
		t.Fatalf("unexpected model assets: %#v", payload)
	}
	if artifacts.listNamespace != "sock-shop" {
		t.Fatalf("namespace = %q, want sock-shop", artifacts.listNamespace)
	}
}

func TestServerRegistersModelAsset(t *testing.T) {
	artifacts := &recordingArtifactStore{}
	server := authenticatedServer(t, NewServerWithOptions(Options{Artifacts: artifacts}))
	body := strings.NewReader(`{
		"namespace": "sock-shop",
		"name": "customer-sft-demo",
		"jobName": "customer-sft-demo",
		"modelURI": "inline://models/customer-sft-demo/v1",
		"baseModelURI": "modelscope://qwen/Qwen2.5-0.5B",
		"datasetURI": "inline://datasets/customer-sft-demo",
		"status": "trained",
		"visibility": "private",
		"evaluationStatus": "pending"
	}`)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/ai/models", body))

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}
	if len(artifacts.saved) != 1 {
		t.Fatalf("saved artifacts = %d, want 1", len(artifacts.saved))
	}
	asset := artifacts.saved[0]
	if asset.Namespace != "sock-shop" || asset.Name != "customer-sft-demo" || asset.ModelURI != "inline://models/customer-sft-demo/v1" {
		t.Fatalf("unexpected saved model asset: %#v", asset)
	}
	if asset.BaseModelURI != "modelscope://qwen/Qwen2.5-0.5B" || asset.DatasetURI != "inline://datasets/customer-sft-demo" {
		t.Fatalf("unexpected model lineage: %#v", asset)
	}
	if asset.Visibility != "private" || asset.Status != "trained" || asset.EvaluationStatus != "pending" {
		t.Fatalf("unexpected model lifecycle fields: %#v", asset)
	}
}

func TestServerRegistersModelAssetDerivesNameFromVersionedURI(t *testing.T) {
	artifacts := &recordingArtifactStore{}
	server := authenticatedServer(t, NewServerWithOptions(Options{Artifacts: artifacts}))
	body := strings.NewReader(`{
		"namespace": "sock-shop",
		"modelURI": "inline://models/customer-sft-demo/v1"
	}`)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/ai/models", body))

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}
	if len(artifacts.saved) != 1 || artifacts.saved[0].Name != "customer-sft-demo" {
		t.Fatalf("unexpected derived model asset: %#v", artifacts.saved)
	}
}

func TestServerListsDatasetAssets(t *testing.T) {
	datasets := &recordingDatasetStore{
		list: []observe.DatasetArtifact{
			{
				Namespace:  "sock-shop",
				Name:       "customer-sft",
				DatasetURI: "dataset://sock-shop/customer-sft/v1",
				Format:     "sharegpt-jsonl",
				Purpose:    "sft",
				Status:     "validated",
			},
		},
	}
	server := authenticatedServer(t, NewServerWithOptions(Options{Datasets: datasets}))
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/ai/datasets?namespace=sock-shop", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var payload observe.DatasetList
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if len(payload.Items) != 1 || payload.Items[0].DatasetURI != "dataset://sock-shop/customer-sft/v1" || payload.Items[0].Status != "validated" {
		t.Fatalf("unexpected dataset assets: %#v", payload)
	}
	if datasets.listNamespace != "sock-shop" {
		t.Fatalf("namespace = %q, want sock-shop", datasets.listNamespace)
	}
}

func TestServerRegistersDatasetAsset(t *testing.T) {
	datasets := &recordingDatasetStore{}
	server := authenticatedServer(t, NewServerWithOptions(Options{Datasets: datasets}))
	body := strings.NewReader(`{
		"namespace": "sock-shop",
		"name": "customer-sft",
		"displayName": "客服问答 SFT 数据集",
		"datasetURI": "oss://datasets/customer-sft/v1/train.jsonl",
		"format": "sharegpt-jsonl",
		"purpose": "sft",
		"status": "validated"
	}`)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/datasets", body)
	req.Header.Set("X-AI-User", "tester")
	server.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}
	if len(datasets.saved) != 1 {
		t.Fatalf("saved datasets = %d, want 1", len(datasets.saved))
	}
	asset := datasets.saved[0]
	if asset.Namespace != "sock-shop" || asset.Name != "customer-sft" || asset.DatasetURI != "oss://datasets/customer-sft/v1/train.jsonl" {
		t.Fatalf("unexpected saved dataset asset: %#v", asset)
	}
	if asset.DisplayName != "客服问答 SFT 数据集" || asset.Format != "sharegpt-jsonl" || asset.Purpose != "sft" {
		t.Fatalf("unexpected dataset metadata: %#v", asset)
	}
	if asset.Status != "validated" || asset.Owner != "tester" {
		t.Fatalf("unexpected dataset lifecycle fields: %#v", asset)
	}
}

func TestServerStartsModelEvaluationJob(t *testing.T) {
	applier := &recordingApplicationApplier{}
	artifacts := &recordingArtifactStore{
		list: []observe.ModelArtifact{
			{
				Namespace: "sock-shop",
				Name:      "customer-sft-demo",
				ModelURI:  "inline://models/customer-sft-demo/v1",
			},
		},
	}
	server := authenticatedServer(t, NewServerWithOptions(Options{Applier: applier, Artifacts: artifacts}))
	body := strings.NewReader(`{
		"evaluationDatasetURI": "inline://datasets/customer-eval",
		"evaluationType": "accuracy",
		"passThreshold": 0.8
	}`)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/ai/models/sock-shop/customer-sft-demo/evaluate", body))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if !applier.called {
		t.Fatalf("expected evaluation Application to be applied")
	}
	content := string(applier.content)
	for _, expected := range []string{
		"name: customer-sft-demo-eval",
		"type: ai-job",
		"jobKind: evaluation",
		"model_uri=inline://models/customer-sft-demo/v1",
		"evaluation_dataset=inline://datasets/customer-eval",
		"pass_threshold=0.8",
		"AI_RESULT_JSON=",
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("expected evaluation Application to contain %q, got:\n%s", expected, content)
		}
	}
}

func TestServerSyncsModelEvaluationResult(t *testing.T) {
	reader := &recordingApplicationReader{
		logs: map[string]*observe.Logs{
			"sock-shop/customer-sft-demo-eval": {
				Pod:  "customer-sft-demo-eval-pod",
				Logs: `AI_RESULT_JSON={"modelURI":"inline://models/customer-sft-demo/v1","metrics":{"accuracy":0.91},"summary":"evaluation-passed"}`,
			},
		},
	}
	artifacts := &recordingArtifactStore{
		list: []observe.ModelArtifact{
			{
				Namespace:        "sock-shop",
				Name:             "customer-sft-demo",
				ModelURI:         "inline://models/customer-sft-demo/v1",
				EvaluationStatus: "pending",
			},
		},
	}
	server := authenticatedServer(t, NewServerWithOptions(Options{Reader: reader, Artifacts: artifacts}))
	body := strings.NewReader(`{"evaluationJobName":"customer-sft-demo-eval"}`)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/ai/models/sock-shop/customer-sft-demo/sync-evaluation", body))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if len(artifacts.saved) != 1 {
		t.Fatalf("saved artifacts = %d, want 1", len(artifacts.saved))
	}
	asset := artifacts.saved[0]
	if asset.EvaluationStatus != "passed" || asset.Status != "evaluated" || asset.Metrics["accuracy"] != 0.91 {
		t.Fatalf("unexpected evaluated model asset: %#v", asset)
	}
	if asset.Summary != "evaluation-passed" || asset.SourceApplication != "customer-sft-demo-eval" {
		t.Fatalf("unexpected evaluation lineage: %#v", asset)
	}
}

func TestServerSyncsFailedModelEvaluationResult(t *testing.T) {
	reader := &recordingApplicationReader{
		logs: map[string]*observe.Logs{
			"sock-shop/customer-sft-demo-eval": {
				Pod:  "customer-sft-demo-eval-pod",
				Logs: `AI_RESULT_JSON={"modelURI":"inline://models/customer-sft-demo/v1","metrics":{"accuracy":0.42},"summary":"evaluation-failed"}`,
			},
		},
	}
	artifacts := &recordingArtifactStore{
		list: []observe.ModelArtifact{
			{
				Namespace:        "sock-shop",
				Name:             "customer-sft-demo",
				ModelURI:         "inline://models/customer-sft-demo/v1",
				EvaluationStatus: "pending",
			},
		},
	}
	server := authenticatedServer(t, NewServerWithOptions(Options{Reader: reader, Artifacts: artifacts}))
	body := strings.NewReader(`{"evaluationJobName":"customer-sft-demo-eval"}`)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/ai/models/sock-shop/customer-sft-demo/sync-evaluation", body))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if len(artifacts.saved) != 1 {
		t.Fatalf("saved artifacts = %d, want 1", len(artifacts.saved))
	}
	asset := artifacts.saved[0]
	if asset.EvaluationStatus != "failed" || asset.Status != "evaluated" || asset.Metrics["accuracy"] != 0.42 {
		t.Fatalf("unexpected failed evaluation asset: %#v", asset)
	}
	if asset.Summary != "evaluation-failed" || asset.SourceApplication != "customer-sft-demo-eval" {
		t.Fatalf("unexpected failed evaluation lineage: %#v", asset)
	}
}

func TestServerListReturnsUnavailableWhenReaderIsNotConfigured(t *testing.T) {
	server := authenticatedServer(t, NewServer())
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/ai/applications", nil))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusServiceUnavailable, recorder.Body.String())
	}
}

func TestServerDeploysDomainYAMLWithDryRun(t *testing.T) {
	applier := &recordingApplicationApplier{}
	server := authenticatedServer(t, NewServerWithOptions(Options{Applier: applier}))
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/ai/applications?dryRun=true", bytes.NewReader([]byte(validAIServiceYAML()))))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var result DeployResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if result.Application.Name != "sentiment-demo" || result.Application.Namespace != "ai-demo" || !result.Application.DryRun {
		t.Fatalf("unexpected application result: %#v", result.Application)
	}
	if result.Normalized.Kind != "AIService" || result.Normalized.GovernanceIntent.Tenant != "demo-tenant" {
		t.Fatalf("unexpected normalized payload: %#v", result.Normalized)
	}
	if !applier.called || applier.namespace != "ai-demo" || applier.name != "sentiment-demo" {
		t.Fatalf("unexpected applier call: %#v", applier)
	}
	if len(applier.options.DryRun) != 1 || applier.options.DryRun[0] != metav1.DryRunAll {
		t.Fatalf("expected server dry-run option, got %#v", applier.options.DryRun)
	}
	if !strings.Contains(string(applier.content), "kind: Application") || !strings.Contains(string(applier.content), "type: ai-service") {
		t.Fatalf("expected translated Application YAML, got:\n%s", string(applier.content))
	}
}

func TestServerDeployReturnsUnavailableWhenApplyIsNotConfigured(t *testing.T) {
	server := authenticatedServer(t, NewServer())
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/ai/applications", bytes.NewReader([]byte(validAIServiceYAML()))))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusServiceUnavailable, recorder.Body.String())
	}
	var payload errorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if !strings.Contains(payload.Error, "deployment is not configured") {
		t.Fatalf("unexpected error: %#v", payload)
	}
}

func TestServerValidatesDomainYAML(t *testing.T) {
	server := authenticatedServer(t, NewServer())
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/ai/validate", bytes.NewReader([]byte(validAIJobYAML()))))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var result domain.ValidationResult
	if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if len(result.Errors) != 0 {
		t.Fatalf("expected no validation errors, got %#v", result.Errors)
	}
}

func TestServerReturnsBadRequestForInvalidDomainYAML(t *testing.T) {
	server := authenticatedServer(t, NewServer())
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/ai/validate", bytes.NewReader([]byte(`
apiVersion: ai.oam.dev/v1alpha1
kind: AIJob
metadata:
  name: invalid
spec:
  properties:
    image: busybox:1.36
`))))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
	var payload errorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if !strings.Contains(payload.Error, "spec.properties.jobKind is required") {
		t.Fatalf("expected jobKind error, got %#v", payload)
	}
}

func TestServerNormalizesDomainYAML(t *testing.T) {
	server := authenticatedServer(t, NewServer())
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/ai/normalize", bytes.NewReader([]byte(validAIServiceYAML()))))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var normalized domain.NormalizedObject
	if err := json.Unmarshal(recorder.Body.Bytes(), &normalized); err != nil {
		t.Fatalf("decode response: %v\n%s", err, recorder.Body.String())
	}
	if normalized.Kind != "AIService" || normalized.WorkloadType != "service" {
		t.Fatalf("unexpected normalized identity: %#v", normalized)
	}
	if normalized.GovernanceIntent.Tenant != "demo-tenant" || normalized.WorkloadIntent.Service.Model.Name != "sentiment" {
		t.Fatalf("unexpected normalized intent: %#v", normalized)
	}
}

func TestServerRejectsUnsupportedMethod(t *testing.T) {
	server := authenticatedServer(t, NewServer())
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/ai/normalize", nil))

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusMethodNotAllowed)
	}
}

func validAIServiceYAML() string {
	return `
apiVersion: ai.oam.dev/v1alpha1
kind: AIService
metadata:
  name: sentiment-demo
  namespace: ai-demo
spec:
  componentName: sentiment-api
  properties:
    image: hashicorp/http-echo:0.2.3
    model:
      name: sentiment
  runtime:
    runtime: http
    tenant: demo-tenant
`
}

func validAIJobYAML() string {
	return `
apiVersion: ai.oam.dev/v1alpha1
kind: AIJob
metadata:
  name: evaluator-demo
spec:
  componentName: batch-evaluator
  properties:
    image: busybox:1.36
    jobKind: evaluation
  runtime:
    runtime: batch
`
}

type recordingApplicationApplier struct {
	called    bool
	namespace string
	name      string
	content   []byte
	options   metav1.PatchOptions
}

func (a *recordingApplicationApplier) ApplyApplication(_ context.Context, namespace, name string, content []byte, opts metav1.PatchOptions) (*unstructured.Unstructured, error) {
	a.called = true
	a.namespace = namespace
	a.name = name
	a.content = append([]byte(nil), content...)
	a.options = opts
	return &unstructured.Unstructured{}, nil
}

var _ domainapply.ApplicationApplier = (*recordingApplicationApplier)(nil)

type recordingApplicationReader struct {
	items         []observe.ApplicationListItem
	summaries     map[string]*observe.Summary
	logs          map[string]*observe.Logs
	listNamespace string
	logRequest    observe.LogOptions
}

func (r *recordingApplicationReader) ListApplications(_ context.Context, namespace string) ([]observe.ApplicationListItem, error) {
	r.listNamespace = namespace
	return r.items, nil
}

func (r *recordingApplicationReader) SummarizeApplication(_ context.Context, namespace, name string) (*observe.Summary, error) {
	key := namespace + "/" + name
	summary := r.summaries[key]
	if summary == nil {
		return nil, fmt.Errorf("not found")
	}
	return summary, nil
}

func (r *recordingApplicationReader) GetApplicationLogs(_ context.Context, options observe.LogOptions) (*observe.Logs, error) {
	r.logRequest = options
	key := options.Namespace + "/" + options.Name
	logs := r.logs[key]
	if logs == nil {
		return nil, fmt.Errorf("not found")
	}
	return logs, nil
}

type recordingApplicationProber struct {
	request observe.ProbeOptions
	result  *observe.ProbeResult
	err     error
}

func (p *recordingApplicationProber) ProbeApplication(_ context.Context, options observe.ProbeOptions) (*observe.ProbeResult, error) {
	p.request = options
	if p.err != nil {
		return nil, p.err
	}
	return p.result, nil
}

type recordingApplicationManager struct {
	deleted   string
	restarted string
	rerun     string
}

func (m *recordingApplicationManager) DeleteApplication(_ context.Context, namespace, name string) (*observe.LifecycleResult, error) {
	m.deleted = namespace + "/" + name
	return &observe.LifecycleResult{Action: "delete", Namespace: namespace, Name: name, Message: "delete requested"}, nil
}

func (m *recordingApplicationManager) RestartApplication(_ context.Context, namespace, name string) (*observe.LifecycleResult, error) {
	m.restarted = namespace + "/" + name
	return &observe.LifecycleResult{Action: "restart", Namespace: namespace, Name: name, Message: "restart requested"}, nil
}

func (m *recordingApplicationManager) RerunApplication(_ context.Context, namespace, name string) (*observe.LifecycleResult, error) {
	m.rerun = namespace + "/" + name
	return &observe.LifecycleResult{Action: "rerun", Namespace: namespace, Name: name, Message: "rerun requested"}, nil
}

type recordingAuditStore struct {
	events        []observe.AuditEvent
	listNamespace string
}

type recordingArtifactStore struct {
	saved         []observe.ModelArtifact
	list          []observe.ModelArtifact
	listNamespace string
}

type recordingDatasetStore struct {
	saved         []observe.DatasetArtifact
	list          []observe.DatasetArtifact
	listNamespace string
}

func (s *recordingArtifactStore) SaveArtifact(_ context.Context, artifact observe.ModelArtifact) error {
	s.saved = append(s.saved, artifact)
	return nil
}

func (s *recordingArtifactStore) ListArtifacts(_ context.Context, namespace string) ([]observe.ModelArtifact, error) {
	s.listNamespace = namespace
	return s.list, nil
}

func (s *recordingDatasetStore) SaveDataset(_ context.Context, dataset observe.DatasetArtifact) error {
	s.saved = append(s.saved, dataset)
	return nil
}

func (s *recordingDatasetStore) ListDatasets(_ context.Context, namespace string) ([]observe.DatasetArtifact, error) {
	s.listNamespace = namespace
	return s.list, nil
}

func (s *recordingAuditStore) RecordAudit(_ context.Context, event observe.AuditEvent) error {
	s.events = append(s.events, event)
	return nil
}

func (s *recordingAuditStore) ListAudits(_ context.Context, namespace string) ([]observe.AuditEvent, error) {
	s.listNamespace = namespace
	return s.events, nil
}
