package services

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/file_validation_cron/internal/models"
)

type FileDeletionService struct {
	storageClient         StorageClientInterface
	supportFileManagerApi SupportFileManagerApiClientInterface
}

func NewFileDeletionService(storageClient StorageClientInterface, supportFileManagerApi SupportFileManagerApiClientInterface) *FileDeletionService {
	return &FileDeletionService{
		storageClient:         storageClient,
		supportFileManagerApi: supportFileManagerApi,
	}
}

func (f *FileDeletionService) Process(ctx context.Context, cronTaskName string) error {
	files, err := f.supportFileManagerApi.GetFilesForDeletion()
	if err != nil {
		return fmt.Errorf("can't get files for deletion err: %w", err)
	}

	if len(files) == 0 {
		logrus.Debugf("no files for deletion found...")
		return nil
	}

	isSuccess := true
	for _, file := range files {
		if err := f.processFileDeletion(ctx, file); err != nil {
			logrus.Errorf("[%s] can't delete file %d with key %s err: %v", cronTaskName, file.FileID, file.ObjectKey, err)
			isSuccess = false
			continue
		}

		logrus.Debugf("file %d deleted", file.FileID)
	}

	if !isSuccess {
		return fmt.Errorf("can't process file deletion")
	}

	return nil
}

func (f *FileDeletionService) processFileDeletion(ctx context.Context, file models.File) error {
	_, err := f.storageClient.HeadObject(ctx, models.HeadObjectRequest{
		Bucket: file.Bucket,
		Key:    file.ObjectKey,
	})
	if err != nil {
		return fmt.Errorf("can't get object info from s3: %w", err)
	}

	err = f.storageClient.DeleteObject(ctx, models.DeleteObjectRequest{
		Bucket: file.Bucket,
		Key:    file.ObjectKey,
	})
	if err != nil {
		return fmt.Errorf("can't delete object from s3: %w", err)
	}

	err = f.supportFileManagerApi.MarkDeletedFile(models.MarkDeletedFileRequest{
		FileID: file.FileID,
	})
	if err != nil {
		return fmt.Errorf("can't update file status err: %w", err)
	}

	return nil
}
