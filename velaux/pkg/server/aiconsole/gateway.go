/*
Copyright 2026 The KubeVela Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package aiconsole opens the AI console, served by the northbound server,
// inside VelaUX. A signed-in user asks for a ticket for a project; the ticket,
// redeemed once at /ai-console/, becomes a cookie session there; and every
// request in that session is proxied to the console with the user, tenant
// and namespace it is for set as trusted headers.
package aiconsole

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	// RoutePrefix is where VelaUX serves the console.
	RoutePrefix = "/ai-console"
	// CookieName holds a browser's console session.
	CookieName = "velaux_ai_console"
	// AdminGroup is the group the console takes for a monitor.
	AdminGroup = "cluster-admin"

	ticketTTL  = 60 * time.Second
	sessionTTL = 12 * time.Hour
)

// The headers the console trusts from its proxy; a browser's own are dropped.
const (
	headerUser      = "X-Remote-User"
	headerGroups    = "X-Remote-Groups"
	headerTenant    = "X-Remote-Tenant"
	headerNamespace = "X-Remote-Namespace"
	headerPrefix    = "X-Forwarded-Prefix"
)

// Identity is who a ticket or session is for: the user, the project the
// console is scoped to, and whether they may monitor the whole platform.
type Identity struct {
	User    string
	Project string
	Admin   bool
	// Monitor opens the monitor console rather than the user's workspace.
	Monitor bool
}

type grant struct {
	Identity
	expires time.Time
}

// Gateway issues tickets, keeps the sessions they become, and proxies the
// sessions' requests to the console.
type Gateway struct {
	target            *url.URL
	namespaceTemplate string
	proxy             *httputil.ReverseProxy
	now               func() time.Time

	mu       sync.Mutex
	tickets  map[string]grant
	sessions map[string]grant
}

// New makes a gateway to the console at endpoint, scoping each project to the
// namespace the template names, with {project} standing for the project. An
// empty endpoint makes a gateway that is not enabled.
func New(endpoint, namespaceTemplate string) (*Gateway, error) {
	g := &Gateway{
		namespaceTemplate: namespaceTemplate,
		now:               time.Now,
		tickets:           map[string]grant{},
		sessions:          map[string]grant{},
	}
	if g.namespaceTemplate == "" {
		g.namespaceTemplate = "ai-{project}"
	}
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return g, nil
	}
	target, err := url.Parse(endpoint)
	if err != nil || target.Scheme == "" || target.Host == "" {
		return nil, fmt.Errorf("the AI console endpoint %q is not an absolute URL", endpoint)
	}
	g.target = target
	g.proxy = &httputil.ReverseProxy{Director: g.direct}
	return g, nil
}

// Enabled says whether a console endpoint is set.
func (g *Gateway) Enabled() bool {
	return g != nil && g.target != nil
}

// Namespace is the namespace a project's workspace is scoped to.
func (g *Gateway) Namespace(project string) string {
	return strings.ReplaceAll(g.namespaceTemplate, "{project}", project)
}

// IssueTicket gives a single-use ticket for the identity, good for a minute,
// and the URL that redeems it.
func (g *Gateway) IssueTicket(id Identity) (string, string, error) {
	ticket, err := randomToken()
	if err != nil {
		return "", "", err
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.sweep()
	g.tickets[ticket] = grant{Identity: id, expires: g.now().Add(ticketTTL)}
	page := "/user"
	if id.Monitor {
		page = "/monitor"
	}
	return ticket, RoutePrefix + page + "?embed=1&ticket=" + url.QueryEscape(ticket), nil
}

// redeem spends a ticket, once, while it is fresh.
func (g *Gateway) redeem(ticket string) (Identity, bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	t, ok := g.tickets[ticket]
	delete(g.tickets, ticket)
	if !ok || g.now().After(t.expires) {
		return Identity{}, false
	}
	return t.Identity, true
}

func (g *Gateway) openSession(id Identity) (string, error) {
	token, err := randomToken()
	if err != nil {
		return "", err
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.sessions[token] = grant{Identity: id, expires: g.now().Add(sessionTTL)}
	return token, nil
}

func (g *Gateway) session(req *http.Request) (Identity, bool) {
	cookie, err := req.Cookie(CookieName)
	if err != nil || cookie.Value == "" {
		return Identity{}, false
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	s, ok := g.sessions[cookie.Value]
	if !ok {
		return Identity{}, false
	}
	if g.now().After(s.expires) {
		delete(g.sessions, cookie.Value)
		return Identity{}, false
	}
	return s.Identity, true
}

// sweep drops spent tickets and sessions; the caller holds the lock.
func (g *Gateway) sweep() {
	now := g.now()
	for k, t := range g.tickets {
		if now.After(t.expires) {
			delete(g.tickets, k)
		}
	}
	for k, s := range g.sessions {
		if now.After(s.expires) {
			delete(g.sessions, k)
		}
	}
}

// ServeHTTP serves /ai-console/: a request bearing a ticket redeems it for a
// session cookie and comes back without it; one in a session is proxied.
func (g *Gateway) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if !g.Enabled() {
		http.Error(res, "The AI console is not connected to VelaUX.", http.StatusServiceUnavailable)
		return
	}
	if ticket := req.URL.Query().Get("ticket"); ticket != "" {
		g.redeemAndRedirect(res, req, ticket)
		return
	}
	id, ok := g.session(req)
	if !ok {
		res.Header().Set("Content-Type", "text/html; charset=utf-8")
		res.WriteHeader(http.StatusUnauthorized)
		_, _ = res.Write([]byte(`<!doctype html><meta charset="utf-8"><title>AI console</title>` +
			`<p style="font:14px sans-serif;color:#4e5969;padding:24px">The AI console session has expired. Reopen it from VelaUX.</p>`))
		return
	}
	g.proxy.ServeHTTP(res, req.WithContext(withIdentity(req.Context(), id)))
}

func (g *Gateway) redeemAndRedirect(res http.ResponseWriter, req *http.Request, ticket string) {
	id, ok := g.redeem(ticket)
	if !ok {
		http.Error(res, "The AI console link has expired. Reopen it from VelaUX.", http.StatusUnauthorized)
		return
	}
	token, err := g.openSession(id)
	if err != nil {
		http.Error(res, "Could not open the AI console session.", http.StatusInternalServerError)
		return
	}
	http.SetCookie(res, &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     RoutePrefix,
		MaxAge:   int(sessionTTL.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   req.TLS != nil,
	})
	next := *req.URL
	query := next.Query()
	query.Del("ticket")
	next.RawQuery = query.Encode()
	http.Redirect(res, req, next.RequestURI(), http.StatusFound)
}

// direct points a proxied request at the console: the prefix stripped, the
// browser's identity headers and session cookie dropped, and the session's
// identity set in their place.
func (g *Gateway) direct(req *http.Request) {
	id, _ := identityFrom(req.Context())
	req.URL.Scheme = g.target.Scheme
	req.URL.Host = g.target.Host
	path := strings.TrimPrefix(req.URL.Path, RoutePrefix)
	if path == "" {
		path = "/"
	}
	req.URL.Path = strings.TrimSuffix(g.target.Path, "/") + path
	req.URL.RawPath = ""
	req.Host = g.target.Host

	for name := range req.Header {
		if strings.HasPrefix(strings.ToLower(name), "x-remote-") {
			req.Header.Del(name)
		}
	}
	req.Header.Del(headerPrefix)
	req.Header.Del("Authorization")
	dropCookie(req, CookieName)

	req.Header.Set(headerPrefix, RoutePrefix)
	req.Header.Set(headerUser, id.User)
	// The console takes the admin group for its monitor role, which it then
	// holds a session to, so an admin opening their own workspace goes
	// without it, and stays a tenant user there.
	if id.Admin && id.Monitor {
		req.Header.Set(headerGroups, AdminGroup)
	}
	if id.Project != "" {
		req.Header.Set(headerTenant, id.Project)
		req.Header.Set(headerNamespace, g.Namespace(id.Project))
	}
}

// dropCookie removes one cookie from a request, keeping the others.
func dropCookie(req *http.Request, name string) {
	cookies := req.Cookies()
	req.Header.Del("Cookie")
	for _, c := range cookies {
		if c.Name != name {
			req.AddCookie(c)
		}
	}
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
