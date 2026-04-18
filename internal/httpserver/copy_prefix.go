package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"

	"s3_task/internal/logger"
)

type copyPrefixRequest struct {
	From         string `json:"from"`
	To           string `json:"to"`
	DeleteSource bool   `json:"delete_source"`
}

type copyPrefixResult struct {
	Copied  int      `json:"copied"`
	Deleted int      `json:"deleted"`
	Errors  []string `json:"errors,omitempty"`
}

func prefixDir(p string) string {
	p = strings.ReplaceAll(p, `\`, `/`)
	if p != "" && !strings.HasSuffix(p, "/") {
		return p + "/"
	}
	return p
}

func copySuffix(key, from string) string {
	if !strings.HasPrefix(key, from) {
		return ""
	}
	rest := key[len(from):]
	return strings.TrimPrefix(rest, "/")
}

func buildCopySource(bucket, key string) string {
	k := strings.ReplaceAll(key, "%", "%25")
	return bucket + "/" + url.PathEscape(k)
}

func (s *Server) handleCopyPrefix(w http.ResponseWriter, r *http.Request) {
	var req copyPrefixRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	from := strings.TrimSpace(strings.ReplaceAll(req.From, `\`, `/`))
	to := strings.TrimSpace(strings.ReplaceAll(req.To, `\`, `/`))
	if from == "" || to == "" {
		http.Error(w, `"from" and "to" must be non-empty`, http.StatusBadRequest)
		return
	}
	if strings.TrimSuffix(from, "/") == strings.TrimSuffix(to, "/") {
		http.Error(w, `"from" and "to" must differ`, http.StatusBadRequest)
		return
	}

	fromDir := prefixDir(from)
	toDir := prefixDir(to)
	destInsideSource := strings.HasPrefix(toDir, fromDir) || toDir == from

	ctx := r.Context()
	var keys []string
	var token *string
	for {
		out, err := s.s3.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:            aws.String(s.bucket),
			Prefix:            aws.String(from),
			ContinuationToken: token,
		})
		if err != nil {
			logger.Error("copy-prefix list", zap.Error(err))
			http.Error(w, "failed to list objects", http.StatusBadGateway)
			return
		}
		for _, obj := range out.Contents {
			if obj.Key == nil {
				continue
			}
			key := *obj.Key
			if !strings.HasPrefix(key, from) {
				continue
			}
			if destInsideSource && strings.HasPrefix(key, toDir) {
				continue
			}
			suf := copySuffix(key, from)
			if suf == "" {
				continue
			}
			keys = append(keys, key)
		}
		if out.IsTruncated == nil || !*out.IsTruncated || out.NextContinuationToken == nil {
			break
		}
		token = out.NextContinuationToken
	}

	var res copyPrefixResult
	for _, srcKey := range keys {
		suf := copySuffix(srcKey, from)
		if suf == "" {
			continue
		}
		destKey := strings.TrimSuffix(toDir, "/") + "/" + suf

		copySrc := buildCopySource(s.bucket, srcKey)
		_, err := s.s3.CopyObject(ctx, &s3.CopyObjectInput{
			Bucket:     aws.String(s.bucket),
			CopySource: aws.String(copySrc),
			Key:        aws.String(destKey),
		})
		if err != nil {
			msg := fmt.Sprintf("%s: %v", srcKey, err)
			logger.Error("copy-prefix copy", zap.String("key", srcKey), zap.Error(err))
			res.Errors = append(res.Errors, msg)
			continue
		}
		res.Copied++

		if req.DeleteSource {
			_, delErr := s.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
				Bucket: aws.String(s.bucket),
				Key:    aws.String(srcKey),
			})
			if delErr != nil {
				msg := fmt.Sprintf("delete %s: %v", srcKey, delErr)
				logger.Error("copy-prefix delete", zap.String("key", srcKey), zap.Error(delErr))
				res.Errors = append(res.Errors, msg)
				continue
			}
			res.Deleted++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		logger.Error("encode copy-prefix response", zap.Error(err))
	}
}
