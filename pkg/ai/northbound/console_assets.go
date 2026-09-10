package northbound

import (
	"embed"
	"net/http"
	"strings"
)

//go:embed console/*.html console/*.js console/*.css console/*.svg
var userConsoleFiles embed.FS

func userConsoleHTML() string {
	content, _ := userConsoleFiles.ReadFile("console/index.html")
	return string(content)
}

func serveConsoleAsset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	name := strings.TrimPrefix(r.URL.Path, "/console/")
	contentType, ok := map[string]string{
		"app.js": "text/javascript", "api.js": "text/javascript", "create.js": "text/javascript",
		"state.js": "text/javascript", "lucide.js": "text/javascript",
		"styles.css": "text/css", "brand.svg": "image/svg+xml",
	}[name]
	if !ok {
		http.NotFound(w, r)
		return
	}
	content, err := userConsoleFiles.ReadFile("console/" + name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", contentType+"; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(content)
}

func currentConsoleSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	session, _ := r.Context().Value(authSessionContextKey{}).(sessionRecord)
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, map[string]string{
		"username": session.Username, "role": string(session.Role),
		"tenant": session.Tenant, "namespace": session.Namespace,
	})
}
