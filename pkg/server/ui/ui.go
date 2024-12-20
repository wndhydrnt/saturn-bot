package ui

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api"
	"github.com/wndhydrnt/saturn-bot/pkg/server/api/openapi"
)

type Ui struct {
	API openapi.StrictServerInterface
}

func RegisterUiRoutes(router chi.Router, apiServer *api.APIServer) {
	app := &Ui{API: apiServer}
	router.Get("/ui", app.HandleIndex)
	router.Get("/ui/runs", app.ListRuns)
	router.Get("/ui/runs/{runId}", app.GetRun)
	router.Group(func(r chi.Router) {
		// Strip the prefix "/ui" from request path
		// because http.FileServerFS() doesn't expect it.
		r.Use(middleware.PathRewrite("/ui", ""))
		r.Handle("/ui/assets/*", http.FileServerFS(assetsFS))
	})
}
