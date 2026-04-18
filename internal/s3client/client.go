package s3client

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	appcfg "s3_task/internal/config"
)

func New(ctx context.Context, c appcfg.App) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(c.S3Region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(c.S3AccessKey, c.S3SecretKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("aws config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(c.S3Endpoint)
		o.UsePathStyle = c.S3UsePathStyle
	})

	return client, nil
}
