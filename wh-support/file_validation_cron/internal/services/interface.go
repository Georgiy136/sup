package services

import (
	"context"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/file_validation_cron/internal/models"
)

type StorageClientInterface interface {
	HeadObject(ctx context.Context, input models.HeadObjectRequest) (*models.HeadObjectResponse, error)
	GetObject(ctx context.Context, input models.GetObjectRequest) (*models.GetObjectResponse, error)
	DeleteObject(ctx context.Context, input models.DeleteObjectRequest) error
}

type SupportFileManagerApiClientInterface interface {
	GetUploadsForValidation() ([]models.Upload, error)
	GetExpiredUploads() ([]models.Upload, error)
	GetUploadsForTransfer() ([]models.Upload, error)
	UpdateUploadStatus(body models.ChangeStatusUploadRequest) error
	GetFilesForDeletion() ([]models.File, error)
	GetFilesNotAttachedToTicketForDeleting() ([]models.File, error)
	MarkDeletedFile(body models.MarkDeletedFileRequest) error
	TransferFile(body models.TransferFileRequest) error
	DeleteFile(body models.DeleteFileRequest) error
}
