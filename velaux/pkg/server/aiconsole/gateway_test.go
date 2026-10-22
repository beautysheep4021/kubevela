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

package aiconsole

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seen is what the console behind the gateway was asked.
type seen struct {
	path, query string
	header      http.Header
	cookies     []*http.Cookie
}

func backend(t *testing.T) (*httptest.Server, *seen) {
	t.Helper()
	got := &seen{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.path, got.query, got.header, got.cookies = r.URL.Path, r.URL.RawQuery, r.Header.Clone(), r.Cookies()
		_, _ = w.Write([]byte("console"))
	}))
	t.Cleanup(srv.Close)
	return srv, got
}

// clock is a gateway's time, moved by hand.
type clock struct{ t time.Time }

func (c *clock) now() time.Time { return c.t }

func gateway(t *testing.T, endpoint string) (*Gateway, *clock) {
	t.Helper()
	g, err := New(endpoint, "")
	require.NoError(t, err)
	c := &clock{t: time.Unix(1_800_000_000, 0)}
	g.now = c.now
	return g, c
}

// open redeems a ticket, as the iframe's first load does, and gives the
// session cookie it sets.
func open(t *testing.T, g *Gateway, link string) *http.Cookie {
	t.Helper()
	rec := httptest.NewRecorder()
	g.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, link, nil))
	require.Equal(t, http.StatusFound, rec.Code)
	loc, err := url.Parse(rec.Header().Get("Location"))
	require.NoError(t, err)
	assert.Equal(t, "", loc.Query().Get("ticket"), "the ticket leaves the URL")
	assert.Equal(t, "1", loc.Query().Get("embed"))
	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)
	c := cookies[0]
	assert.Equal(t, CookieName, c.Name)
	assert.Equal(t, RoutePrefix, c.Path)
	assert.True(t, c.HttpOnly)
	assert.Equal(t, http.SameSiteLaxMode, c.SameSite)
	return c
}

func TestTicketIsSpentOnce(t *testing.T) {
	srv, _ := backend(t)
	g, _ := gateway(t, srv.URL)
	_, link, err := g.IssueTicket(Identity{User: "alice", Project: "shop"})
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(link, "/ai-console/user?embed=1&ticket="))
	open(t, g, link)

	rec := httptest.NewRecorder()
	g.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, link, nil))
	assert.Equal(t, http.StatusUnauthorized, rec.Code, "a spent ticket opens nothing")
}

func TestTicketExpiresAfterAMinute(t *testing.T) {
	srv, _ := backend(t)
	g, c := gateway(t, srv.URL)
	_, link, err := g.IssueTicket(Identity{User: "alice", Project: "shop"})
	require.NoError(t, err)
	c.t = c.t.Add(61 * time.Second)
	rec := httptest.NewRecorder()
	g.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, link, nil))
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestMonitorTicketOpensTheMonitor(t *testing.T) {
	g, _ := gateway(t, "http://console.example")
	_, link, err := g.IssueTicket(Identity{User: "admin", Admin: true, Monitor: true})
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(link, "/ai-console/monitor?embed=1&ticket="))
}

func TestSessionProxiesWithItsIdentity(t *testing.T) {
	srv, got := backend(t)
	g, _ := gateway(t, srv.URL)
	_, link, err := g.IssueTicket(Identity{User: "alice", Project: "shop"})
	require.NoError(t, err)
	session := open(t, g, link)

	req := httptest.NewRequest(http.MethodGet, "/ai-console/api/v1/ai/applications?namespace=ai-shop", nil)
	req.AddCookie(session)
	req.AddCookie(&http.Cookie{Name: "other", Value: "kept"})
	// A browser may not speak for itself.
	req.Header.Set("X-Remote-User", "mallory")
	req.Header.Set("X-Remote-Groups", "cluster-admin")
	req.Header.Set("X-Remote-Namespace", "kube-system")
	req.Header.Set("X-Remote-Extra", "x")
	req.Header.Set("X-Forwarded-Prefix", "/elsewhere")
	req.Header.Set("Authorization", "Bearer velaux-token")
	rec := httptest.NewRecorder()
	g.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "/api/v1/ai/applications", got.path, "the prefix is stripped")
	assert.Equal(t, "namespace=ai-shop", got.query)
	assert.Equal(t, "alice", got.header.Get("X-Remote-User"))
	assert.Equal(t, "", got.header.Get("X-Remote-Groups"))
	assert.Equal(t, "shop", got.header.Get("X-Remote-Tenant"))
	assert.Equal(t, "ai-shop", got.header.Get("X-Remote-Namespace"))
	assert.Equal(t, "", got.header.Get("X-Remote-Extra"))
	assert.Equal(t, "/ai-console", got.header.Get("X-Forwarded-Prefix"))
	assert.Equal(t, "", got.header.Get("Authorization"), "VelaUX's token stays in VelaUX")
	names := []string{}
	for _, c := range got.cookies {
		names = append(names, c.Name)
	}
	assert.Equal(t, []string{"other"}, names, "the session cookie stays in VelaUX")
}

func TestAdminSessionCarriesTheAdminGroup(t *testing.T) {
	srv, got := backend(t)
	g, _ := gateway(t, srv.URL)
	g.namespaceTemplate = "tenant-{project}-ai"
	_, link, err := g.IssueTicket(Identity{User: "admin", Project: "ml", Admin: true, Monitor: true})
	require.NoError(t, err)
	session := open(t, g, link)
	req := httptest.NewRequest(http.MethodGet, "/ai-console/monitor", nil)
	req.AddCookie(session)
	g.ServeHTTP(httptest.NewRecorder(), req)
	assert.Equal(t, "/monitor", got.path)
	assert.Equal(t, "cluster-admin", got.header.Get("X-Remote-Groups"))
	assert.Equal(t, "tenant-ml-ai", got.header.Get("X-Remote-Namespace"))
}

func TestAdminWorkspaceGoesWithoutTheAdminGroup(t *testing.T) {
	srv, got := backend(t)
	g, _ := gateway(t, srv.URL)
	_, link, err := g.IssueTicket(Identity{User: "admin", Project: "ml", Admin: true})
	require.NoError(t, err)
	session := open(t, g, link)
	req := httptest.NewRequest(http.MethodGet, "/ai-console/api/v1/ai/applications", nil)
	req.AddCookie(session)
	g.ServeHTTP(httptest.NewRecorder(), req)
	assert.Equal(t, "/api/v1/ai/applications", got.path)
	assert.Equal(t, "", got.header.Get("X-Remote-Groups"), "an admin in a workspace is a tenant user there")
	assert.Equal(t, "ai-ml", got.header.Get("X-Remote-Namespace"))
}

func TestNoSessionIsRefused(t *testing.T) {
	srv, got := backend(t)
	g, c := gateway(t, srv.URL)
	rec := httptest.NewRecorder()
	g.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ai-console/user", nil))
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	_, link, err := g.IssueTicket(Identity{User: "alice", Project: "shop"})
	require.NoError(t, err)
	session := open(t, g, link)
	c.t = c.t.Add(13 * time.Hour)
	req := httptest.NewRequest(http.MethodGet, "/ai-console/user", nil)
	req.AddCookie(session)
	rec = httptest.NewRecorder()
	g.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code, "a session lasts 12 hours")
	assert.Equal(t, "", got.path, "nothing reached the console")
}

func TestDisabledGateway(t *testing.T) {
	g, err := New("  ", "")
	require.NoError(t, err)
	assert.False(t, g.Enabled())
	rec := httptest.NewRecorder()
	g.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ai-console/user", nil))
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	_, err = New("console:8080", "")
	assert.Error(t, err, "an endpoint without a scheme is refused")
}
