package service

import (
	"context"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/support_file_manager_api/internal/models"
)

type S3ClientInterface interface {
	GetPresignedUploadURL(ctx context.Context, input models.GetUploadURLRequest) (*models.GetUploadURLResponse, error)
	GetPresignedDownloadURL(ctx context.Context, input models.GetDownloadURLRequest) (string, error)
	GetTmpBucketName() string
	GetBucketName() string
	CopyObject(ctx context.Context, input models.CopyObjectRequest) error
	DeleteObject(ctx context.Context, bucket, objectKey string) error
	PutObject(ctx context.Context, input models.PutObjectRequest) error
}

type CentrifugoClientInterface interface {
	PublishUploadStatus(ctx context.Context, uploadID int64, isSuccess bool) error
}

type FileValidator interface {
	ValidateFile(mimeType string, fileSize int64, fileName string) error
}

type Repository interface {
	FileRepository
	UploadRepository
}

type FileRepository interface {
	GetNewFileID() (int64, error)
	GetFileInfoByID(fileID int64) (*models.File, error)
	UpdateFileStatus(fileID int64, status string) error
	AttachTicketIDToFiles(fileIDs []int64, ticketID int64) error
	GetFilesByStatus(status string) ([]models.File, error)
	AddFiles(params []models.FileAddParams) error
}

type UploadRepository interface {
	GetNewUploadID() (int64, error)
	AddUpload(params models.UploadAddParams) error
	UpdateUploadStatus(uploadID int64, employeeID int64, status string) error
	GetUploadByID(uploadID int64) (*models.Upload, error)
	GetUploadsByStatus(status string) ([]models.Upload, error)
	GetExpiredUploads() ([]models.Upload, error)
}
