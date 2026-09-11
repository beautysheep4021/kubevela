package northbound

import (
	"context"
	"net/http"
	"strings"

	"github.com/oam-dev/kubevela/pkg/ai/observe"
)

type TenantResourceReader interface {
	ListTenantResources(ctx context.Context, namespace string) ([]observe.TenantResources, error)
}

type TenantResourcesResponse struct {
	Items []observe.TenantResources `json:"items"`
}

func tenantResources(reader TenantResourceReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := requestSession(r)
		if !ok || session.Role != consoleRoleMonitor {
			writeError(w, http.StatusForbidden, "monitor role is required")
			return
		}
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if reader == nil {
			writeError(w, http.StatusServiceUnavailable, "tenant resource reader is not configured")
			return
		}
		items, err := reader.ListTenantResources(r.Context(), strings.TrimSpace(r.URL.Query().Get("namespace")))
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		if items == nil {
			items = []observe.TenantResources{}
		}
		writeJSON(w, http.StatusOK, TenantResourcesResponse{Items: items})
	}
}
