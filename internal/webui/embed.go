package webui

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

// Assets contains the production frontend bundle.
//
//go:embed dist/*
var Assets embed.FS

func Handler() http.Handler {
	dist, _ := fs.Sub(Assets, "dist")
	files := http.FileServer(http.FS(dist))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path != "" {
			if _, err := fs.Stat(dist, path); err == nil {
				files.ServeHTTP(w, r)
				return
			}
		}
		request := r.Clone(r.Context())
		request.URL.Path = "/"
		files.ServeHTTP(w, request)
	})
}
