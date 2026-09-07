package services

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/file_validation_cron/internal/models"
)

type FileTransferService struct {
	storageClient         StorageClientInterface
	supportFileManagerApi SupportFileManagerApiClientInterface
}

func NewFileTransferService(storageClient StorageClientInterface, supportFileManagerApi SupportFileManagerApiClientInterface) *FileTransferService {
	return &FileTransferService{
		storageClient:         storageClient,
		supportFileManagerApi: supportFileManagerApi,
	}
}

func (f *FileTransferService) Process(ctx context.Context, cronTaskName string) error {
	uploads, err := f.supportFileManagerApi.GetUploadsForTransfer()
	if err != nil {
		return fmt.Errorf("can't get uploads for transferring err: %w", err)
	}

	if len(uploads) == 0 {
		logrus.Info("no uploads for transfer found...")
		return nil
	}

	isSuccess := true
	for _, upload := range uploads {
		if err := f.processFileTransfer(ctx, upload); err != nil {
			logrus.Errorf("[%s] can't transfer file %d with key %s; upload_id: %d; err: %v", cronTaskName, upload.FileID, upload.TmpObjectKey, upload.UploadID, err)
			isSuccess = false
			continue
		}

		logrus.Debugf("file %d transferred; status upload_id %d updated", upload.FileID, upload.UploadID)
	}

	if !isSuccess {
		return fmt.Errorf("can't process file transfer")
	}

	return nil
}

func (f *FileTransferService) processFileTransfer(ctx context.Context, upload models.Upload) error {
	_, err := f.storageClient.HeadObject(ctx, models.HeadObjectRequest{
		Bucket: upload.TmpBucket,
		Key:    upload.TmpObjectKey,
	})
	if err != nil {
		return fmt.Errorf("can't get object info from s3: %w", err)
	}

	err = f.supportFileManagerApi.TransferFile(models.TransferFileRequest{UploadID: upload.UploadID})
	if err != nil {
		return fmt.Errorf("can't transfer file err: %w", err)
	}

	return nil
}
