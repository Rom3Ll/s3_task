package httpserver

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"

	"s3_task/internal/logger"
)

type objectEntry struct {
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"lastModified"`
}

func (s *Server) handleListObjects(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("prefix")

	ctx := r.Context()
	var entries []objectEntry
	var token *string

	for {
		out, err := s.s3.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:            aws.String(s.bucket),
			Prefix:            aws.String(prefix),
			ContinuationToken: token,
		})
		if err != nil {
			logger.Error("list objects", zap.Error(err))
			http.Error(w, "failed to list objects", http.StatusBadGateway)
			return
		}

		for _, obj := range out.Contents {
			if obj.Key == nil {
				continue
			}
			var lm time.Time
			if obj.LastModified != nil {
				lm = *obj.LastModified
			}
			entries = append(entries, objectEntry{
				Key:          *obj.Key,
				Size:         aws.ToInt64(obj.Size),
				LastModified: lm,
			})
		}

		if out.IsTruncated == nil || !*out.IsTruncated || out.NextContinuationToken == nil {
			break
		}
		token = out.NextContinuationToken
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(entries); err != nil {
		logger.Error("encode list response", zap.Error(err))
	}
}
