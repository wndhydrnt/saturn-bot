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
	router.Handle("/", http.RedirectHandler("/ui", http.StatusMovedPermanently))
	router.Get("/ui", app.Home)
	router.Get("/ui/runs", app.RunsIndex)
	router.Get("/ui/runs/{runId}", app.RunsShow)
	router.Get("/ui/runs/{runId}/{repositoryName}/error", app.RunsRepositoryErrorShow)
	router.Get("/ui/tasks", app.TasksIndex)
	router.Get("/ui/tasks/{name}/file", app.TasksFileShow)
	router.Get("/ui/tasks/{name}/results", app.ResultsIndex)
	router.Get("/ui/status", app.StatusIndex)
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
