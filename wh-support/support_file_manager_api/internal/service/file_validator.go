package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
)

type fileValidator struct {
	config *FileValidationConfig
}

type FileValidationConfig struct {
	AllowedMimeTypes map[string]MimeConfig `json:"allowed_mime_types"`
}

type MimeConfig struct {
	Extensions   []string `json:"extensions"`
	MaxSizeBytes int64    `json:"max_size_bytes"`
}

func NewFileValidator() *fileValidator {
	return &fileValidator{}
}

func (v *fileValidator) Configure(ctx context.Context, config configs.Config) {
	const cfgKey = "file_validation"

	if err := jsoniter.Unmarshal(config.GetByServiceKeyRequired(cfgKey), &v.config); err != nil {
		logrus.Panicf("error unmarshaling file validation config: %v", err)
	}
}

func (v *fileValidator) ValidateFile(mimeType string, fileSize int64, fileName string) error {
	mime := strings.ToLower(mimeType)
	if mime == "" {
		return fmt.Errorf("mime type is required")
	}

	mimeConfig, ok := v.config.AllowedMimeTypes[mime]
	if !ok {
		return fmt.Errorf("mime type '%s' is not allowed", mime)
	}

	if fileSize > mimeConfig.MaxSizeBytes {
		return fmt.Errorf("file size %d bytes exceeds maximum %d bytes", fileSize, mimeConfig.MaxSizeBytes)
	}

	// Проверка расширения файла
	ext := strings.ToLower(filepath.Ext(fileName))
	if ext == "" {
		return fmt.Errorf("file must have an extension")
	}

	extAllowed := false
	for _, allowedExt := range mimeConfig.Extensions {
		if ext == strings.ToLower(allowedExt) {
			extAllowed = true
			break
		}
	}
	if !extAllowed {
		return fmt.Errorf("extension '%s' is not allowed for mime type '%s'", ext, mime)
	}

	return nil
}
