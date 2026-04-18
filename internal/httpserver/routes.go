package httpserver

import (
	"net/http"
	"path/filepath"
)

func (s *Server) Register(mux *http.ServeMux, webDir string) {
	webDir = filepath.Clean(webDir)
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /api/objects", s.handleListObjects)
	mux.HandleFunc("POST /api/upload-local", s.handleUploadLocal)
	mux.HandleFunc("POST /api/copy-prefix", s.handleCopyPrefix)
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(webDir, "index.html"))
	})
	mux.HandleFunc("GET /app.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(webDir, "app.js"))
	})
	mux.HandleFunc("GET /styles.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(webDir, "styles.css"))
	})
}
