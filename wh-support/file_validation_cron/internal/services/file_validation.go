package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gabriel-vasile/mimetype"

	"github.com/sirupsen/logrus"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/file_validation_cron/internal/models"
)

const (
	minBytesSizeFileForDetectType = 3072
)

type FileValidationService struct {
	storageClient         StorageClientInterface
	supportFileManagerApi SupportFileManagerApiClientInterface
}

func NewFileValidationService(storageClient StorageClientInterface, supportFileManagerApi SupportFileManagerApiClientInterface) *FileValidationService {
	return &FileValidationService{
		storageClient:         storageClient,
		supportFileManagerApi: supportFileManagerApi,
	}
}

func (f *FileValidationService) Process(ctx context.Context, cronTaskName string) error {
	uploads, err := f.supportFileManagerApi.GetUploadsForValidation()
	if err != nil {
		return fmt.Errorf("can't get uploads for validating err: %w", err)
	}

	if len(uploads) == 0 {
		logrus.Info("no uploads for validation found...")
		return nil
	}

	isSuccess := true
	for _, upload := range uploads {
		if err := f.processFileValidation(ctx, upload); err != nil {
			logrus.Errorf("[%s] can't validate file %d with key %s; upload_id: %d; err: %v", cronTaskName, upload.FileID, upload.TmpObjectKey, upload.UploadID, err)
			isSuccess = false
			continue
		}

		logrus.Debugf("file %d validated; status upload_id %d updated", upload.FileID, upload.UploadID)
	}

	if !isSuccess {
		return fmt.Errorf("can't process file validation")
	}

	return nil
}

func (f *FileValidationService) processFileValidation(ctx context.Context, upload models.Upload) error {
	rangeHeader := fmt.Sprintf("bytes=0-%d", minBytesSizeFileForDetectType-1)
	object, err := f.storageClient.GetObject(ctx, models.GetObjectRequest{
		Bucket: upload.TmpBucket,
		Key:    upload.TmpObjectKey,
		Range:  rangeHeader,
	})
	if err != nil {
		if _, ok := errors.AsType[*types.NoSuchKey](err); ok {
			logrus.Debugf("object with key %s doesn't exist in bucket %s", upload.TmpObjectKey, upload.TmpBucket)
			return nil
		}
		return fmt.Errorf("can't get object from s3: %w", err)
	}

	sizeObject := parseContentRangeSize(object.ContentRange)

	fileType, err := detectFileType(object.Body, sizeObject)
	if err != nil {
		return fmt.Errorf("can't detect file type: %w", err)
	}

	status := getUploadStatus(fileType, sizeObject, upload)
	err = f.supportFileManagerApi.UpdateUploadStatus(models.ChangeStatusUploadRequest{
		UploadID: upload.UploadID,
		Status:   status,
	})
	if err != nil {
		return fmt.Errorf("can't update upload status err: %w", err)
	}

	return nil
}

func detectFileType(bodyObject io.ReadCloser, sizeObject int64) (string, error) {
	defer func() {
		if err := bodyObject.Close(); err != nil {
			logrus.Errorf("can't close body object err: %v", err)
		}
	}()

	body, err := io.ReadAll(bodyObject)
	if err != nil {
		return "", fmt.Errorf("failed to read object body: %w", err)
	}

	if len(body) == 0 {
		return "", fmt.Errorf("file can't be empty")
	}

	if sizeObject == 0 {
		return "", fmt.Errorf("invalid file size from S3")
	}

	return mimetype.Detect(body).String(), nil
}

func parseContentRangeSize(contentRange *string) int64 {
	if contentRange == nil || *contentRange == "" {
		return 0
	}

	idx := strings.LastIndex(*contentRange, "/")
	if idx == -1 {
		return 0
	}

	sizeStr := (*contentRange)[idx+1:]
	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		logrus.Errorf("can't parse content range err: %v", err)
		return 0
	}

	return size
}

func getUploadStatus(fileType string, sizeObject int64, upload models.Upload) string {
	const (
		prefixTextMimeType = "text/plain"
		csvMimeType        = "text/csv"

		videoMp4MimeType       = "video/mp4"
		videoQuicktimeMimeType = "video/quicktime"
	)

	status := models.UploadStatusReady
	sizeValid := sizeObject == upload.DeclaredSizeBytes
	mimeValid := mimetype.EqualsAny(fileType, upload.DeclaredMimeType)

	if !mimeValid {
		if strings.HasPrefix(fileType, prefixTextMimeType) && upload.DeclaredMimeType == csvMimeType {
			mimeValid = true
		}

		if upload.DeclaredMimeType == videoQuicktimeMimeType && fileType == videoMp4MimeType {
			mimeValid = true
		}
	}

	if !mimeValid || !sizeValid {
		status = models.UploadStatusInvalid
		logrus.Debugf("invalid file; param from s3: %s, %d; param from db: %s, %d", fileType, sizeObject, upload.DeclaredMimeType, upload.DeclaredSizeBytes)
	}

	return status
}
