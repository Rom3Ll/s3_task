package localupload

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"

	"s3_task/internal/logger"
)

const jpegQuality = 75

type Result struct {
	Source string `json:"source"`
	Key    string `json:"key,omitempty"`
	OK     bool   `json:"ok"`
	Error  string `json:"error,omitempty"`
}

var imageExts = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".png":  {},
	".gif":  {},
}

func UploadJPEGFromDir(ctx context.Context, client *s3.Client, bucket, dir, keyPrefix string) []Result {
	keyPrefix = strings.TrimSuffix(keyPrefix, "/")
	if keyPrefix != "" {
		keyPrefix += "/"
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return []Result{{
			Source: dir,
			OK:     false,
			Error:  err.Error(),
		}}
	}

	var out []Result
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if _, ok := imageExts[ext]; !ok {
			continue
		}

		full := filepath.Join(dir, name)
		res := Result{Source: name}
		key, buf, encErr := compressToJPEG(full, name, keyPrefix)
		if encErr != nil {
			res.Error = encErr.Error()
			out = append(out, res)
			continue
		}

		_, err := client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:      aws.String(bucket),
			Key:         aws.String(key),
			Body:        bytes.NewReader(buf),
			ContentType: aws.String("image/jpeg"),
		})
		if err != nil {
			logger.Error("put object", zap.String("key", key), zap.Error(err))
			res.Error = err.Error()
			out = append(out, res)
			continue
		}

		res.OK = true
		res.Key = key
		out = append(out, res)
	}

	return out
}

func compressToJPEG(fullPath, baseName, keyPrefix string) (key string, jpegBytes []byte, err error) {
	f, err := os.Open(fullPath)
	if err != nil {
		return "", nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return "", nil, fmt.Errorf("decode: %w", err)
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: jpegQuality}); err != nil {
		return "", nil, fmt.Errorf("encode jpeg: %w", err)
	}

	stem := strings.TrimSuffix(baseName, filepath.Ext(baseName))
	key = keyPrefix + stem + ".jpg"
	return key, buf.Bytes(), nil
}
