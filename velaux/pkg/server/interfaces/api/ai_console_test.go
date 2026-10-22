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

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/emicklei/go-restful/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubevela/velaux/pkg/server/aiconsole"
	"github.com/kubevela/velaux/pkg/server/domain/model"
	"github.com/kubevela/velaux/pkg/server/domain/service"
)

// grants answers each user's permissions: the platform's, and a project's
// when asked for one.
type grants struct {
	service.RBACService
	platform map[string][]*model.Permission
	project  map[string]map[string][]*model.Permission
}

func (g grants) GetUserPermissions(_ context.Context, user *model.User, project string, withPlatform bool) ([]*model.Permission, error) {
	var perms []*model.Permission
	if withPlatform {
		perms = append(perms, g.platform[user.Name]...)
	}
	if project != "" {
		perms = append(perms, g.project[project][user.Name]...)
	}
	return perms, nil
}

type users struct{ service.UserService }

func (users) GetUser(_ context.Context, name string) (*model.User, error) {
	return &model.User{Name: name}, nil
}

func allow(resource string, actions ...string) *model.Permission {
	return &model.Permission{Resources: []string{resource}, Actions: actions, Effect: "Allow"}
}

func ticketAPI(t *testing.T, endpoint string) *restful.Container {
	t.Helper()
	gw, err := aiconsole.New(endpoint, "")
	require.NoError(t, err)
	api := &aiConsole{
		RbacService: grants{
			platform: map[string][]*model.Permission{"admin": {allow("*", "*")}},
			project:  map[string]map[string][]*model.Permission{"shop": {"alice": {allow("project:shop/*", "detail")}}},
		},
		UserService: users{},
		Gateway:     gw,
	}
	container := restful.NewContainer()
	container.Add(api.GetWebServiceRoute())
	return container
}

// askTicket posts a ticket request as the user, and gives the status and the
// URL it answers with.
func askTicket(t *testing.T, c *restful.Container, user, body string) (int, string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai_console/tickets", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken(t, user, time.Now().Add(time.Hour)))
	rec := httptest.NewRecorder()
	c.ServeHTTP(rec, req)
	var out struct {
		URL string `json:"url"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec.Code, out.URL
}

func TestAIConsoleTicketForAProjectMember(t *testing.T) {
	c := ticketAPI(t, "http://console.example")
	code, url := askTicket(t, c, "alice", `{"project":"shop"}`)
	assert.Equal(t, http.StatusOK, code)
	assert.True(t, strings.HasPrefix(url, "/ai-console/user?embed=1&ticket="), url)

	code, _ = askTicket(t, c, "alice", `{"project":"bank"}`)
	assert.Equal(t, http.StatusForbidden, code, "not a member of bank")

	code, _ = askTicket(t, c, "alice", `{}`)
	assert.Equal(t, http.StatusBadRequest, code, "a workspace needs a project")
}

func TestAIConsoleMonitorIsTheAdmins(t *testing.T) {
	c := ticketAPI(t, "http://console.example")
	code, _ := askTicket(t, c, "alice", `{"project":"shop","monitor":true}`)
	assert.Equal(t, http.StatusForbidden, code)

	code, url := askTicket(t, c, "admin", `{"monitor":true}`)
	assert.Equal(t, http.StatusOK, code)
	assert.True(t, strings.HasPrefix(url, "/ai-console/monitor?embed=1&ticket="), url)

	code, _ = askTicket(t, c, "admin", `{"project":"bank"}`)
	assert.Equal(t, http.StatusOK, code, "a platform admin opens any project's workspace")
}

func TestAIConsoleTicketWhenNotConnected(t *testing.T) {
	c := ticketAPI(t, "")
	code, _ := askTicket(t, c, "admin", `{"project":"shop"}`)
	assert.Equal(t, http.StatusServiceUnavailable, code)
}
