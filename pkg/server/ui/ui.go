package ui

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
)

// Ui defines all HTTP handlers that render views of the UI.
type Ui struct {
	API openapi.StrictServerInterface
}

// RegisterUiRoutes initializes [Ui] using apiServer and registers its handlers with the router.
func RegisterUiRoutes(router chi.Router, apiServer *api.APIServer) {
	app := &Ui{API: apiServer}
	router.Get("/ui", app.GetHome)
	router.Get("/ui/runs", app.ListRuns)
	router.Get("/ui/runs/{runId}", app.GetRun)
	router.Group(func(r chi.Router) {
		r.Use(
			// Strip the prefix "/ui" from request path
			// because http.FileServerFS() doesn't expect it.
			middleware.PathRewrite("/ui", ""),
			// Ensure that the static assets are cached.
			addCacheHeaders(),
		)
		r.Handle("/ui/assets/*", http.FileServerFS(assetsFS))
	})
}
