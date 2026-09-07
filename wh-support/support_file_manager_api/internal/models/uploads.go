package models

import (
	"encoding/json"
	"time"
)

// Upload - модель записи в таблице files.uploads
type Upload struct {
	UploadID          int64           `json:"upload_id"`
	FileID            int64           `json:"file_id"`
	ChEmployeeID      int64           `json:"ch_employee_id"`
	EntityType        string          `json:"entity_type"`
	EntityData        json.RawMessage `json:"entity_data"`
	OriginalName      string          `json:"original_name"`
	DeclaredMimeType  string          `json:"declared_mime_type"`
	DeclaredSizeBytes int64           `json:"declared_size_bytes"`
	TmpBucket         string          `json:"tmp_bucket"`
	TmpObjectKey      string          `json:"tmp_object_key"`
	Status            string          `json:"status"`
	ExpiresAt         string          `json:"expires_at"`
	CreatedAt         string          `json:"created_at"`
	UpdatedAt         string          `json:"updated_at"`
	IsDel             bool            `json:"is_del"`
}

// UploadAddParams - параметры для создания записи аплоада
type UploadAddParams struct {
	UploadID          int64           `json:"upload_id"`
	FileID            int64           `json:"file_id"`
	ChEmployeeID      int64           `json:"ch_employee_id"`
	EntityType        string          `json:"entity_type"`
	EntityData        json.RawMessage `json:"entity_data,omitempty"`
	OriginalName      string          `json:"original_name"`
	DeclaredMimeType  string          `json:"declared_mime_type"`
	DeclaredSizeBytes int64           `json:"declared_size_bytes"`
	TmpBucket         string          `json:"tmp_bucket"`
	TmpObjectKey      string          `json:"tmp_object_key"`
	Status            string          `json:"status"`
	ExpiresAt         string          `json:"expires_at"`
}

// GetUploadURLRequest - запрос на получение ссылки для загрузки файла
type GetUploadURLRequest struct {
	ObjectKey   string
	ContentType string
	FileName    string
	Expiration  time.Duration
}

// GetUploadURLResponse - ответ на запрос ссылки для загрузки файла
type GetUploadURLResponse struct {
	UploadID           int64             `json:"upload_id"`
	FileID             int64             `json:"file_id"`
	UploadURL          string            `json:"upload_url"`
	Method             string            `json:"method"`
	ObjectKey          string            `json:"object_key"`
	ExpiresAt          string            `json:"expires_at"`
	RequiredFormFields map[string]string `json:"required_form_fields"`
}

// GetUploadStatusResponse - ответ со статусом аплоада
type GetUploadStatusResponse struct {
	UploadID int64  `json:"upload_id"`
	FileID   int64  `json:"file_id"`
	TicketID int64  `json:"ticket_id"`
	Status   string `json:"status"`
}

// CopyObjectRequest - запрос на копирование объекта между бакетами
type CopyObjectRequest struct {
	SourceBucket string
	SourceKey    string
	DestBucket   string
	DestKey      string
}

// PutObjectRequest - запрос на прямую загрузку объекта в S3
type PutObjectRequest struct {
	Bucket      string
	ObjectKey   string
	ContentType string
	FileData    []byte
}

// ServiceUploadResponse - ответ на сервисную загрузку файла
type ServiceUploadResponse struct {
	FileID    int64  `json:"file_id"`
	FileName  string `json:"file_name"`
	FileSize  int64  `json:"file_size"`
	MimeType  string `json:"mime_type"`
	UploadID  int64  `json:"upload_id"`
	ObjectKey string `json:"object_key"`
	Bucket    string `json:"bucket"`
	Status    string `json:"status"`
}
