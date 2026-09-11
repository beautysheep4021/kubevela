package northbound

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type consoleRole string

const (
	consoleRoleUser           consoleRole = "user"
	consoleRoleMonitor        consoleRole = "monitor"
	sessionCookieName                     = "kubevela_ai_session"
	sessionTTL                            = 12 * time.Hour
	defaultRemoteUserHeader               = "X-Remote-User"
	defaultRemoteGroupsHeader             = "X-Remote-Groups"
)

type consoleRolePage struct {
	Label      string
	Title      string
	Lede       string
	RoleLabel  string
	RedirectTo string
	Username   string
	Password   string
}

type consoleAccount struct {
	Role      consoleRole
	Username  string
	Password  string
	Tenant    string
	Namespace string
}

type authConfig struct {
	RemoteAuthEnabled  bool
	RemoteUserHeader   string
	RemoteGroupsHeader string
	AdminGroups        []string
}

var consoleAccounts = []consoleAccount{
	{
		Role:      consoleRoleUser,
		Username:  "admin",
		Password:  "shiyong",
		Tenant:    "tenant-a",
		Namespace: "ai-tenant-a",
	},
	{
		Role:      consoleRoleUser,
		Username:  "tenant-b",
		Password:  "tenant-b-123456",
		Tenant:    "tenant-b",
		Namespace: "ai-tenant-b",
	},
	{
		Role:     consoleRoleMonitor,
		Username: "admin",
		Password: "jiankong",
	},
}

var consoleRolePages = map[consoleRole]consoleRolePage{
	consoleRoleUser: {
		Label:      "智算纳管 · 使用方工作台 PoC",
		Title:      "使用方工作台",
		Lede:       "面向业务用户提交模型服务或批任务，默认只展示使用方页面。",
		RoleLabel:  "使用方",
		RedirectTo: "/user",
		Username:   "admin",
		Password:   "shiyong",
	},
	consoleRoleMonitor: {
		Label:      "智算纳管 · K8s 管理员工作台 PoC",
		Title:      "K8s 管理员工作台",
		Lede:       "面向 K8s 管理员查看任务概览、资源总览和异常任务，默认只展示管理页面。",
		RoleLabel:  "K8s 管理员",
		RedirectTo: "/monitor",
		Username:   "admin",
		Password:   "jiankong",
	},
}

type sessionRecord struct {
	Username  string
	Role      consoleRole
	Tenant    string
	Namespace string
	Expires   time.Time
}

type authSessionContextKey struct{}

type consoleAuth struct {
	mu       sync.RWMutex
	sessions map[string]sessionRecord
	now      func() time.Time
	config   authConfig
	admins   map[string]struct{}
}

func newConsoleAuth() *consoleAuth {
	return newConsoleAuthWithConfig(nil)
}

func newConsoleAuthWithConfig(cfg *authConfig) *consoleAuth {
	config := defaultAuthConfig()
	if cfg != nil {
		config = config.withOverrides(*cfg)
	}
	return &consoleAuth{
		sessions: map[string]sessionRecord{},
		now:      time.Now,
		config:   config,
		admins:   makeGroupSet(config.AdminGroups),
	}
}

func (a *consoleAuth) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	suffix := demoQuerySuffix(r)
	if session, ok := a.authenticatedSession(r); ok {
		http.Redirect(w, r, consoleRolePages[session.Role].RedirectTo+suffix, http.StatusFound)
		return
	}
	http.Redirect(w, r, "/login"+suffix, http.StatusFound)
}

func (a *consoleAuth) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/login" {
		http.NotFound(w, r)
		return
	}
	suffix := demoQuerySuffix(r)
	switch r.Method {
	case http.MethodGet:
		if session, ok := a.authenticatedSession(r); ok {
			http.Redirect(w, r, consoleRolePages[session.Role].RedirectTo+suffix, http.StatusFound)
			return
		}
		a.writeLoginPage(w, consoleRoleUser, "", suffix)
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			a.writeLoginPage(w, consoleRoleUser, "表单解析失败", suffix)
			return
		}
		role := consoleRole(strings.TrimSpace(r.Form.Get("role")))
		if _, ok := consoleRolePages[role]; !ok {
			role = consoleRoleUser
		}
		username := strings.TrimSpace(r.Form.Get("username"))
		password := r.Form.Get("password")
		account, ok := findConsoleAccount(role, username, password)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			a.writeLoginPage(w, role, "账号或密码错误", suffix)
			return
		}
		token, err := a.createSessionForAccount(account)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     sessionCookieName,
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Expires:  a.now().Add(sessionTTL),
			MaxAge:   int(sessionTTL / time.Second),
		})
		http.Redirect(w, r, consoleRolePages[role].RedirectTo+suffix, http.StatusFound)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *consoleAuth) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/logout" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	suffix := demoQuerySuffix(r)
	if cookie, err := r.Cookie(sessionCookieName); err == nil && cookie.Value != "" {
		a.deleteSession(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
	http.Redirect(w, r, "/login"+suffix, http.StatusFound)
}

func (a *consoleAuth) handleConsolePage(role consoleRole) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != consoleRolePages[role].RedirectTo {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		session, ok := a.authenticatedSession(r)
		if !ok {
			http.Redirect(w, r, "/login"+demoQuerySuffix(r), http.StatusFound)
			return
		}
		if session.Role != role {
			http.Redirect(w, r, consoleRolePages[session.Role].RedirectTo+demoQuerySuffix(r), http.StatusFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = io.WriteString(w, renderConsoleHTMLForSession(role, session))
	}
}

func (a *consoleAuth) requireSession(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := a.authenticatedSession(r)
		if !ok {
			writeError(w, http.StatusUnauthorized, "login required")
			return
		}
		ctx := context.WithValue(r.Context(), authSessionContextKey{}, session)
		next(w, r.WithContext(ctx))
	}
}

func (a *consoleAuth) writeLoginPage(w http.ResponseWriter, role consoleRole, message, querySuffix string) {
	selected := consoleRoleUser
	if _, ok := consoleRolePages[role]; ok {
		selected = role
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.WriteString(w, renderLoginHTML(selected, message, querySuffix))
}

func (a *consoleAuth) authenticatedSession(r *http.Request) (sessionRecord, bool) {
	if session, ok := a.sessionFromRequest(r); ok {
		return session, true
	}
	if session, ok := a.requestIdentityFromHeaders(r); ok {
		return session, true
	}
	return sessionRecord{}, false
}

func (a *consoleAuth) sessionFromRequest(r *http.Request) (sessionRecord, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return sessionRecord{}, false
	}
	a.mu.RLock()
	session, ok := a.sessions[cookie.Value]
	a.mu.RUnlock()
	if !ok {
		return sessionRecord{}, false
	}
	if a.now().After(session.Expires) {
		a.deleteSession(cookie.Value)
		return sessionRecord{}, false
	}
	return session, true
}

func (a *consoleAuth) requestIdentityFromHeaders(r *http.Request) (sessionRecord, bool) {
	if a == nil || !a.config.RemoteAuthEnabled {
		return sessionRecord{}, false
	}
	username := requestIdentityUser(r, a.config)
	if username == "" {
		return sessionRecord{}, false
	}
	role := consoleRoleUser
	if a.roleFromGroups(requestIdentityGroups(r, a.config)) == consoleRoleMonitor {
		role = consoleRoleMonitor
	}
	return sessionRecord{
		Username: username,
		Role:     role,
	}, true
}

func findConsoleAccount(role consoleRole, username, password string) (consoleAccount, bool) {
	for _, account := range consoleAccounts {
		if account.Role == role && account.Username == username && account.Password == password {
			return account, true
		}
	}
	return consoleAccount{}, false
}

func findConsoleAccountByIdentity(role consoleRole, username string) (consoleAccount, bool) {
	for _, account := range consoleAccounts {
		if account.Role == role && account.Username == username {
			return account, true
		}
	}
	return consoleAccount{}, false
}

func (a *consoleAuth) roleFromGroups(groups []string) consoleRole {
	for _, group := range groups {
		if _, ok := a.admins[strings.ToLower(strings.TrimSpace(group))]; ok {
			return consoleRoleMonitor
		}
	}
	return consoleRoleUser
}

func (a *consoleAuth) createSession(role consoleRole, username string) (string, error) {
	account, ok := findConsoleAccountByIdentity(role, username)
	if !ok {
		account = consoleAccount{
			Role:     role,
			Username: username,
		}
	}
	return a.createSessionForAccount(account)
}

func (a *consoleAuth) createSessionForAccount(account consoleAccount) (string, error) {
	token, err := randomToken()
	if err != nil {
		return "", err
	}
	a.mu.Lock()
	a.sessions[token] = sessionRecord{
		Username:  account.Username,
		Role:      account.Role,
		Tenant:    account.Tenant,
		Namespace: account.Namespace,
		Expires:   a.now().Add(sessionTTL),
	}
	a.mu.Unlock()
	return token, nil
}

func requestSession(r *http.Request) (sessionRecord, bool) {
	if r == nil {
		return sessionRecord{}, false
	}
	session, ok := r.Context().Value(authSessionContextKey{}).(sessionRecord)
	return session, ok
}

func namespaceForListRequest(r *http.Request, requested string) (string, error) {
	requested = strings.TrimSpace(requested)
	scope := namespaceScopeForRequest(r)
	if scope == "" {
		return requested, nil
	}
	if requested == "" {
		return scope, nil
	}
	if requested != scope {
		return "", fmt.Errorf("namespace %q is outside account scope %q", requested, scope)
	}
	return requested, nil
}

func authorizeNamespaceRequest(r *http.Request, requested string) error {
	requested = strings.TrimSpace(requested)
	scope := namespaceScopeForRequest(r)
	if scope == "" {
		return nil
	}
	if requested == "" {
		return fmt.Errorf("namespace is required for account scope %q", scope)
	}
	if requested != scope {
		return fmt.Errorf("namespace %q is outside account scope %q", requested, scope)
	}
	return nil
}

func namespaceScopeForRequest(r *http.Request) string {
	session, ok := requestSession(r)
	if !ok || session.Role == consoleRoleMonitor {
		return ""
	}
	return strings.TrimSpace(session.Namespace)
}

func (a *consoleAuth) deleteSession(token string) {
	a.mu.Lock()
	delete(a.sessions, token)
	a.mu.Unlock()
}

func randomToken() (string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw[:]), nil
}

func defaultAuthConfig() authConfig {
	return authConfig{
		RemoteAuthEnabled:  parseBoolEnv("KUBEVELA_AI_REMOTE_AUTH_ENABLED", false),
		RemoteUserHeader:   envOrDefault("KUBEVELA_AI_REMOTE_USER_HEADER", defaultRemoteUserHeader),
		RemoteGroupsHeader: envOrDefault("KUBEVELA_AI_REMOTE_GROUPS_HEADER", defaultRemoteGroupsHeader),
		AdminGroups:        parseListEnv("KUBEVELA_AI_ADMIN_GROUPS", []string{"system:masters", "cluster-admin"}),
	}
}

func (c authConfig) withOverrides(override authConfig) authConfig {
	if override.RemoteAuthEnabled {
		c.RemoteAuthEnabled = true
	}
	if strings.TrimSpace(override.RemoteUserHeader) != "" {
		c.RemoteUserHeader = strings.TrimSpace(override.RemoteUserHeader)
	}
	if strings.TrimSpace(override.RemoteGroupsHeader) != "" {
		c.RemoteGroupsHeader = strings.TrimSpace(override.RemoteGroupsHeader)
	}
	if len(override.AdminGroups) > 0 {
		c.AdminGroups = append([]string(nil), override.AdminGroups...)
	}
	return c
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func parseBoolEnv(key string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return value
}

func parseListEnv(key string, fallback []string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return append([]string(nil), fallback...)
	}
	items := splitAndTrim(raw)
	if len(items) == 0 {
		return append([]string(nil), fallback...)
	}
	return items
}

func splitAndTrim(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';'
	})
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if value := strings.TrimSpace(part); value != "" {
			out = append(out, value)
		}
	}
	return out
}

func makeGroupSet(groups []string) map[string]struct{} {
	set := make(map[string]struct{}, len(groups))
	for _, group := range groups {
		if value := strings.ToLower(strings.TrimSpace(group)); value != "" {
			set[value] = struct{}{}
		}
	}
	return set
}

func requestIdentityUser(r *http.Request, config authConfig) string {
	if r == nil {
		return ""
	}
	for _, header := range configuredUserHeaders(config) {
		if value := strings.TrimSpace(r.Header.Get(header)); value != "" {
			return value
		}
	}
	return ""
}

func requestIdentityGroups(r *http.Request, config authConfig) []string {
	if r == nil {
		return nil
	}
	header := strings.TrimSpace(config.RemoteGroupsHeader)
	if header == "" {
		header = defaultRemoteGroupsHeader
	}
	return splitAndTrim(r.Header.Get(header))
}

func configuredUserHeaders(config authConfig) []string {
	header := strings.TrimSpace(config.RemoteUserHeader)
	if header == "" {
		header = defaultRemoteUserHeader
	}
	return []string{
		header,
		"X-AI-User",
		"X-User",
		"X-Forwarded-User",
	}
}

func renderConsoleHTML(role consoleRole) string {
	return renderConsoleHTMLForSession(role, sessionRecord{Role: role})
}

func renderConsoleHTMLForSession(role consoleRole, session sessionRecord) string {
	page, ok := consoleRolePages[role]
	if !ok {
		page = consoleRolePages[consoleRoleUser]
	}
	replacer := strings.NewReplacer(
		"__PAGE_ROLE__", string(role),
		"__ACCOUNT_TENANT__", html.EscapeString(session.Tenant),
		"__ACCOUNT_NAMESPACE__", html.EscapeString(session.Namespace),
		"__ACCOUNT_USERNAME__", html.EscapeString(session.Username),
		"__PAGE_LABEL__", page.Label,
		"__PAGE_TITLE__", page.Title,
		"__PAGE_LEDE__", page.Lede,
		"__PAGE_ROLE_LABEL__", page.RoleLabel,
	)
	if role == consoleRoleUser {
		return replacer.Replace(userConsoleHTML())
	}
	if role == consoleRoleMonitor {
		return replacer.Replace(monitorConsoleHTML())
	}
	return replacer.Replace(consoleHTML)
}

func renderLoginHTML(role consoleRole, message, querySuffix string) string {
	errorClass := ""
	if strings.TrimSpace(message) != "" {
		errorClass = "show"
	}
	action := "/login"
	if querySuffix != "" {
		action += querySuffix
	}
	monitorRoleLabel := consoleRolePages[consoleRoleMonitor].RoleLabel
	replacer := strings.NewReplacer(
		"__ERROR_CLASS__", errorClass,
		"__ERROR_MESSAGE__", html.EscapeString(message),
		"__USER_SELECTED__", selectedAttr(role == consoleRoleUser),
		"__MONITOR_SELECTED__", selectedAttr(role == consoleRoleMonitor),
		"__MONITOR_ROLE_LABEL__", html.EscapeString(monitorRoleLabel),
		"__DEFAULT_ACCOUNT_HINT__", html.EscapeString(defaultAccountHint(monitorRoleLabel)),
		"__LOGIN_TITLE__", "智算纳管登录",
		"__LOGIN_ACTION__", html.EscapeString(action),
	)
	return replacer.Replace(loginHTML)
}

func defaultAccountHint(monitorRoleLabel string) string {
	user := consoleAccounts[0]
	scoped := consoleAccounts[1]
	monitor := consoleAccounts[2]
	return fmt.Sprintf(
		"使用方：%s / %s（Namespace：%s）；隔离测试使用方：%s / %s（Namespace：%s）；%s：%s / %s",
		user.Username,
		user.Password,
		user.Namespace,
		scoped.Username,
		scoped.Password,
		scoped.Namespace,
		monitorRoleLabel,
		monitor.Username,
		monitor.Password,
	)
}

func demoQuerySuffix(r *http.Request) string {
	if r == nil {
		return ""
	}
	if strings.EqualFold(r.URL.Query().Get("demo"), "local") {
		return "?demo=local"
	}
	return ""
}

func selectedAttr(selected bool) string {
	if selected {
		return "selected"
	}
	return ""
}

const loginHTML = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>__LOGIN_TITLE__</title>
  <style>
    :root {
      --ink: #16191f;
      --muted: #5f6b7a;
      --paper: #f2f3f3;
      --panel: #ffffff;
      --line: #d5d9d9;
      --blue: #1677ff;
      --shadow: 0 8px 28px rgba(22, 25, 31, .14);
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      min-height: 100vh;
      display: grid;
      place-items: center;
      padding: 20px;
      color: var(--ink);
      font-family: -apple-system, "Segoe UI", system-ui, "PingFang SC", "Microsoft YaHei", Roboto, sans-serif;
      background: var(--paper);
    }
    main {
      width: min(520px, 100%);
    }
    .panel {
      border: 1px solid var(--line);
      border-radius: 12px;
      background: var(--panel);
      box-shadow: var(--shadow);
      padding: 24px;
    }
    .eyebrow {
      color: #0b5cad;
      font-size: 12px;
      font-weight: 900;
      letter-spacing: .12em;
    }
    h1 {
      margin: 10px 0 8px;
      font-size: 28px;
      line-height: 1.1;
    }
    p {
      margin: 0 0 18px;
      color: var(--muted);
      line-height: 1.55;
      font-size: 13px;
    }
    .error {
      display: none;
      margin-bottom: 14px;
      padding: 10px 12px;
      border: 1px solid #f5c6cb;
      border-radius: 8px;
      color: #b91c1c;
      background: #fff1f2;
      font-size: 13px;
    }
    .error.show { display: block; }
    .form-grid {
      display: grid;
      gap: 14px;
    }
    label {
      display: block;
      margin-bottom: 7px;
      color: #5f6b7a;
      font-size: 12px;
      font-weight: 900;
    }
    input, select {
      width: 100%;
      min-height: 40px;
      border: 1px solid var(--line);
      border-radius: 8px;
      padding: 8px 10px;
      color: var(--ink);
      background: #f8fafc;
      font: inherit;
    }
    .default-accounts {
      margin-top: 14px;
      padding: 10px 12px;
      border: 1px solid var(--line);
      border-radius: 8px;
      color: var(--ink);
      background: #f8fafc;
      font-size: 12px;
      line-height: 1.6;
      white-space: pre-line;
    }
    .actions {
      display: flex;
      justify-content: space-between;
      gap: 12px;
      align-items: center;
      margin-top: 14px;
    }
    .hint {
      color: var(--muted);
      font-size: 12px;
      line-height: 1.5;
    }
    button {
      border: 0;
      border-radius: 8px;
      padding: 10px 14px;
      color: #fff;
      background: var(--blue);
      font: inherit;
      font-weight: 900;
      cursor: pointer;
    }
    button.ghost {
      color: var(--ink);
      background: #f7f8f8;
      border: 1px solid var(--line);
    }
  </style>
</head>
<body>
  <main>
    <section class="panel">
      <div class="eyebrow">智算纳管北向验证台</div>
      <h1>登录</h1>
      <p>先选择使用方或 __MONITOR_ROLE_LABEL__，再输入对应的默认账号。</p>
      <div class="error __ERROR_CLASS__">__ERROR_MESSAGE__</div>
        <form method="post" action="__LOGIN_ACTION__">
        <div class="form-grid">
          <div>
            <label for="role">角色选择</label>
            <select id="role" name="role">
              <option value="user" __USER_SELECTED__>使用方</option>
              <option value="monitor" __MONITOR_SELECTED__>__MONITOR_ROLE_LABEL__</option>
            </select>
          </div>
          <div>
            <label for="username">账号</label>
            <input id="username" name="username" value="admin" autocomplete="username">
          </div>
          <div>
            <label for="password">密码</label>
            <input id="password" name="password" type="password" autocomplete="current-password">
          </div>
        </div>
        <div class="default-accounts">__DEFAULT_ACCOUNT_HINT__</div>
        <div class="actions">
          <div class="hint">登录后会自动跳转到对应页面。</div>
          <button type="submit">登录</button>
          <button class="ghost" type="submit" formaction="/login?demo=local">本地演示</button>
        </div>
      </form>
    </section>
  </main>
</body>
</html>`
