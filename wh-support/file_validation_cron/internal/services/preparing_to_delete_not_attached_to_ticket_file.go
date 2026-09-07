package services

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/file_validation_cron/internal/models"
)

type PreparingToDeleteNotAttachedToTicketFile struct {
	storageClient         StorageClientInterface
	supportFileManagerApi SupportFileManagerApiClientInterface
}

func NewPreparingToDeleteNotAttachedToTicketFile(storageClient StorageClientInterface, supportFileManagerApi SupportFileManagerApiClientInterface) *PreparingToDeleteNotAttachedToTicketFile {
	return &PreparingToDeleteNotAttachedToTicketFile{
		storageClient:         storageClient,
		supportFileManagerApi: supportFileManagerApi,
	}
}

func (f *PreparingToDeleteNotAttachedToTicketFile) Process(ctx context.Context, cronTaskName string) error {
	files, err := f.supportFileManagerApi.GetFilesNotAttachedToTicketForDeleting()
	if err != nil {
		return fmt.Errorf("can't get not attached to tickets files for preparing to delete err: %w", err)
	}

	if len(files) == 0 {
		logrus.Info("no files found that are not attached to tickets")
		return nil
	}

	isSuccess := true
	countFailedDeletionFiles := 0
	for _, file := range files {
		request := models.DeleteFileRequest{
			FileID: file.FileID,
		}

		if err := f.supportFileManagerApi.DeleteFile(request); err != nil {
			logrus.Errorf("[%s] can't delete not attached to ticket file %d with key %s err: %v", cronTaskName, file.FileID, file.ObjectKey, err)
			isSuccess = false
			countFailedDeletionFiles++
			continue
		}

		logrus.Debugf("not attached to ticket file %d delete", file.FileID)
	}

	if !isSuccess {
		return fmt.Errorf("can't process preparing to delete not attached to ticket file; count of failed deletion files: %d", countFailedDeletionFiles)
	}

	return nil
}
