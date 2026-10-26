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
	"strings"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"

	"github.com/kubevela/velaux/pkg/server/aiconsole"
	"github.com/kubevela/velaux/pkg/server/domain/model"
	"github.com/kubevela/velaux/pkg/server/domain/service"
	apis "github.com/kubevela/velaux/pkg/server/interfaces/api/dto/v1"
	"github.com/kubevela/velaux/pkg/server/utils/bcode"
)

// aiConsole issues the tickets that open the AI console inside VelaUX.
type aiConsole struct {
	RbacService service.RBACService `inject:""`
	UserService service.UserService `inject:""`
	Gateway     *aiconsole.Gateway  `inject:"aiConsoleGateway"`
}

// NewAIConsole is the AI console's ticket API.
func NewAIConsole() Interface {
	return &aiConsole{}
}

// GetWebServiceRoute -
func (a *aiConsole) GetWebServiceRoute() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(versionPrefix + "/ai_console").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("api for opening the AI console")

	tags := []string{"ai_console"}

	// The user's own access is checked in the handler: a workspace needs the
	// project, and the monitor needs the platform.
	ws.Route(ws.POST("/tickets").To(a.createTicket).
		Doc("issue a single-use ticket that opens the AI console").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(apis.CreateAIConsoleTicketRequest{}).
		Returns(200, "OK", apis.AIConsoleTicketResponse{}).
		Returns(400, "Bad Request", bcode.Bcode{}).
		Returns(403, "Forbidden", bcode.Bcode{}).
		Returns(503, "Not connected", bcode.Bcode{}).
		Writes(apis.AIConsoleTicketResponse{}).Do(returns200, returns500))

	ws.Filter(authCheckFilter)
	return ws
}

func (a *aiConsole) createTicket(req *restful.Request, res *restful.Response) {
	if !a.Gateway.Enabled() {
		bcode.ReturnError(req, res, bcode.ErrAIConsoleNotConfigured)
		return
	}
	var body apis.CreateAIConsoleTicketRequest
	if err := req.ReadEntity(&body); err != nil {
		bcode.ReturnError(req, res, err)
		return
	}
	body.Project = strings.TrimSpace(body.Project)
	ctx := req.Request.Context()
	userName, ok := ctx.Value(&apis.CtxKeyUser).(string)
	if !ok {
		bcode.ReturnError(req, res, bcode.ErrUnauthorized)
		return
	}
	user, err := a.UserService.GetUser(ctx, userName)
	if err != nil {
		bcode.ReturnError(req, res, bcode.ErrUnauthorized)
		return
	}
	admin, err := a.allowed(ctx, user, "", "systemSetting:*", "update")
	if err != nil {
		bcode.ReturnError(req, res, err)
		return
	}
	id := aiconsole.Identity{User: userName, Project: body.Project, Admin: admin, Monitor: body.Monitor}
	switch {
	case body.Monitor:
		// The monitor spans every tenant, so it is the platform admins'.
		if !admin {
			bcode.ReturnError(req, res, bcode.ErrForbidden)
			return
		}
	case body.Project == "":
		bcode.ReturnError(req, res, bcode.ErrAIConsoleProjectRequired)
		return
	default:
		member, err := a.allowed(ctx, user, body.Project, "project:"+body.Project, "detail")
		if err != nil {
			bcode.ReturnError(req, res, err)
			return
		}
		if !member {
			bcode.ReturnError(req, res, bcode.ErrForbidden)
			return
		}
	}
	_, link, err := a.Gateway.IssueTicket(id)
	if err != nil {
		bcode.ReturnError(req, res, err)
		return
	}
	if err := res.WriteEntity(apis.AIConsoleTicketResponse{URL: link}); err != nil {
		bcode.ReturnError(req, res, err)
		return
	}
}

// allowed says whether the user's permissions, the project's and the
// platform's, grant the actions on the resource.
func (a *aiConsole) allowed(ctx context.Context, user *model.User, project, resource string, actions ...string) (bool, error) {
	perms, err := a.RbacService.GetUserPermissions(ctx, user, project, true)
	if err != nil {
		return false, err
	}
	ra := &service.RequestResourceAction{}
	ra.SetResourceWithName(resource, func(string) string { return "" })
	ra.SetActions(actions)
	return ra.Match(perms), nil
}
