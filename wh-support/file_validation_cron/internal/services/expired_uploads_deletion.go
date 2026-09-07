package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	serviceerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/file_validation_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/file_validation_cron/internal/models"
)

type ExpiredUploadsDeletionService struct {
	storageClient         StorageClientInterface
	supportFileManagerApi SupportFileManagerApiClientInterface
}

func NewExpiredUploadsDeletionService(storageClient StorageClientInterface, supportFileManagerApi SupportFileManagerApiClientInterface) *ExpiredUploadsDeletionService {
	return &ExpiredUploadsDeletionService{
		storageClient:         storageClient,
		supportFileManagerApi: supportFileManagerApi,
	}
}

func (e *ExpiredUploadsDeletionService) Process(ctx context.Context, cronTaskName string) error {
	uploads, err := e.supportFileManagerApi.GetExpiredUploads()
	if err != nil {
		return fmt.Errorf("can't get expired uploads for deleting err: %w", err)
	}

	if len(uploads) == 0 {
		logrus.Info("no expired uploads found...")
		return nil
	}

	isSuccess := true
	for _, upload := range uploads {
		if err := e.processExpiredUploadsDeletion(ctx, upload); err != nil {
			if errors.Is(err, serviceerrors.ErrObjectNotFound) {
				logrus.Debugf("[%s] file not found, file %d with key %s; upload_id: %d", cronTaskName, upload.FileID, upload.TmpObjectKey, upload.UploadID)
				continue
			}

			logrus.Errorf("[%s] can't delete expired uploads, file %d with key %s; upload_id: %d; err: %v", cronTaskName, upload.FileID, upload.TmpObjectKey, upload.UploadID, err)
			isSuccess = false
			continue
		}

		logrus.Debugf("file %d deleted; status upload_id %d updated", upload.FileID, upload.UploadID)
	}

	if !isSuccess {
		return fmt.Errorf("can't process expired uploads deletion")
	}

	return nil
}

func (e *ExpiredUploadsDeletionService) processExpiredUploadsDeletion(ctx context.Context, upload models.Upload) error {
	_, err := e.storageClient.HeadObject(ctx, models.HeadObjectRequest{
		Bucket: upload.TmpBucket,
		Key:    upload.TmpObjectKey,
	})
	if err != nil {
		return fmt.Errorf("can't get object info from s3: %w", err)
	}

	err = e.storageClient.DeleteObject(ctx, models.DeleteObjectRequest{
		Bucket: upload.TmpBucket,
		Key:    upload.TmpObjectKey,
	})
	if err != nil {
		return fmt.Errorf("can't delete object from s3: %w", err)
	}

	err = e.supportFileManagerApi.UpdateUploadStatus(models.ChangeStatusUploadRequest{
		UploadID: upload.UploadID,
		Status:   models.UploadStatusExpired,
	})
	if err != nil {
		return fmt.Errorf("can't update uploads status err: %w", err)
	}

	return nil
}
