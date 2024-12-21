package ui

import (
	"embed"
	"net/http"
)

//go:embed assets/*
var assetsFS embed.FS

func addCacheHeaders() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "max-age=3600")
			next.ServeHTTP(w, r)
		})
	}
}
