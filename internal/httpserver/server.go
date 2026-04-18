package httpserver

import (
	appcfg "s3_task/internal/config"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Server struct {
	cfg    appcfg.App
	s3     *s3.Client
	bucket string
}

func New(cfg appcfg.App, client *s3.Client) *Server {
	return &Server{cfg: cfg, s3: client, bucket: cfg.S3Bucket}
}
