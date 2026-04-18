package httpserver

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"

	"s3_task/internal/localupload"
	"s3_task/internal/logger"
)

func (s *Server) handleUploadLocal(w http.ResponseWriter, r *http.Request) {
	dir, err := filepath.Abs(s.cfg.LocalSourceDir)
	if err != nil {
		logger.Error("resolve local source dir", zap.Error(err))
		http.Error(w, "invalid local source directory", http.StatusInternalServerError)
		return
	}

	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "local source directory does not exist", http.StatusBadRequest)
			return
		}
		logger.Error("stat local source dir", zap.Error(err))
		http.Error(w, "cannot read local source directory", http.StatusInternalServerError)
		return
	}
	if !info.IsDir() {
		http.Error(w, "local source path is not a directory", http.StatusBadRequest)
		return
	}

	keyPrefix := strings.Trim(strings.ReplaceAll(s.cfg.LocalUploadPrefix, `\`, `/`), "/")
	results := localupload.UploadJPEGFromDir(r.Context(), s.s3, s.bucket, dir, keyPrefix)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string][]localupload.Result{"results": results}); err != nil {
		logger.Error("encode upload response", zap.Error(err))
	}
}
