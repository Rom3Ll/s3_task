package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type App struct {
	HTTPAddr string

	LocalSourceDir    string
	LocalUploadPrefix string

	S3Endpoint     string
	S3Region       string
	S3AccessKey    string
	S3SecretKey    string
	S3Bucket       string
	S3UsePathStyle bool
}

func Load() (App, error) {
	cfg := App{
		HTTPAddr:          getenv("HTTP_ADDR", ":8080"),
		LocalSourceDir:    getenv("LOCAL_SOURCE_DIR", "assets/source"),
		LocalUploadPrefix: getenv("LOCAL_UPLOAD_PREFIX", "uploads"),
		S3Endpoint:        strings.TrimRight(getenv("S3_ENDPOINT", "http://localhost:9000"), "/"),
		S3Region:       getenv("S3_REGION", "us-east-1"),
		S3AccessKey:    getenv("S3_ACCESS_KEY", "minioadmin"),
		S3SecretKey:    getenv("S3_SECRET_KEY", "minioadmin"),
		S3Bucket:       getenv("S3_BUCKET", "images"),
		S3UsePathStyle: parseBool(getenv("S3_USE_PATH_STYLE", "true")),
	}

	if cfg.S3Endpoint == "" {
		return App{}, fmt.Errorf("S3_ENDPOINT must not be empty")
	}
	if cfg.S3Bucket == "" {
		return App{}, fmt.Errorf("S3_BUCKET must not be empty")
	}

	return cfg, nil
}

func getenv(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func parseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "1" || s == "true" || s == "yes" || s == "on" {
		return true
	}
	if s == "0" || s == "false" || s == "no" || s == "off" {
		return false
	}
	b, err := strconv.ParseBool(s)
	if err != nil {
		return false
	}
	return b
}
