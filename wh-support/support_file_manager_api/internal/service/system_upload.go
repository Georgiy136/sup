package service

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/sirupsen/logrus"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/support_file_manager_api/internal/constant"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/support_file_manager_api/internal/models"
	file_utils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/support_file_manager_api/internal/utils"
	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"
	utils_time "gitlab.wildberries.ru/wbwh/wh-core/gocore-utils-time.git"
	utils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors/errors_keys"

	"github.com/gin-gonic/gin"
)

// InternalUploadFile загружает файл напрямую в S3
func (s *SupportFileManagerService) InternalUploadFile(ctx *gin.Context, _ map[string]interface{}) {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, fmt.Errorf("can't read file from multipart form: %w", err))
		return
	}
	entityType := ctx.PostForm("entity_type")
	if entityType == "" {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("entity_type is required"))
		return
	}

	if fileHeader.Size > constant.MaxUploadSize {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, fmt.Errorf("file size %d bytes exceeds maximum %d bytes", fileHeader.Size, constant.MaxUploadSize))
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, fmt.Errorf("can't open uploaded file: %w", err))
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			logrus.Errorf("can't close uploaded file: %v", err)
		}
	}()

	fileData, err := io.ReadAll(io.LimitReader(file, constant.MaxUploadSize))
	if err != nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, fmt.Errorf("can't read uploaded file content: %w", err))
		return
	}

	fileName := fileHeader.Filename
	if fileName == "" {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("file_name is required"))
		return
	}
	if len(fileName) > 100 {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("file_name must be less than 100 characters"))
		return
	}

	fileSize := fileHeader.Size
	mimeType := fileHeader.Header.Get("Content-Type")

	// Валидируем файл
	if err = s.fileValidator.ValidateFile(mimeType, fileSize, fileName); err != nil {
		s.errBuilder.BindError(ctx, support_err_keys.KeyErrorFileValidationFailed, err)
		return
	}

	// Резервируем новый upload ID
	uploadID, err := s.repository.GetNewUploadID()
	if err != nil {
		s.bindServiceError(ctx, err, "failed to get new upload ID")
		return
	}

	// Резервируем новый file ID
	fileID, err := s.repository.GetNewFileID()
	if err != nil {
		s.bindServiceError(ctx, err, "failed to get new file ID")
		return
	}

	// Создаем upload в сервисном статусе
	uploadParams := models.UploadAddParams{
		UploadID:          uploadID,
		FileID:            fileID,
		ChEmployeeID:      constant.SystemEmployeeID,
		EntityType:        entityType,
		OriginalName:      fileName,
		DeclaredMimeType:  mimeType,
		DeclaredSizeBytes: fileSize,
		Status:            constant.UploadStatusService,
		ExpiresAt:         time.Now().Add(5 * time.Minute).Format(utils_time.LayoutDataBasePG),
	}
	if err = s.repository.AddUpload(uploadParams); err != nil {
		s.bindServiceError(ctx, err, "failed to create service upload record")
		return
	}

	// Создаем объектный ключ для постоянного файла
	finalObjectKey := file_utils.GenerateFinalObjectKey(entityType, fileID)
	bucket := s.s3Client.GetBucketName()

	// Загружаем файл по постоянному ключу
	if err = s.s3Client.PutObject(ctx.Request.Context(), models.PutObjectRequest{
		Bucket:      bucket,
		ObjectKey:   finalObjectKey,
		ContentType: mimeType,
		FileData:    fileData,
	}); err != nil {
		s.bindServiceError(ctx, err, "failed to upload file to permanent storage")
		return
	}

	// Сохраняем file со статусом ATTACH_PENDING
	fileStatus := constant.FileStatusAttachPending
	expiresAt := time.Now().Add(constant.AttachPendingFileTTL).Format(utils_time.LayoutDataBasePG)
	fileExpiresAt := &expiresAt

	fileAddParams := []models.FileAddParams{
		{
			FileID:    fileID,
			UploadID:  uploadID,
			Bucket:    bucket,
			ObjectKey: finalObjectKey,
			Status:    fileStatus,
			ExpiresAt: fileExpiresAt,
		},
	}
	if err = s.repository.AddFiles(fileAddParams); err != nil {
		s.bindServiceError(ctx, err, "failed to create file record from service upload")
		return
	}

	// Обновляем статус аплоада на TRANSFERRED
	if err = s.repository.UpdateUploadStatus(uploadID, constant.SystemEmployeeID, constant.UploadStatusTransferred); err != nil {
		s.bindServiceError(ctx, err, "failed to update service upload status")
		return
	}

	utils.BindObjectToRestData(ctx, models.ServiceUploadResponse{
		FileID:    fileID,
		FileName:  fileName,
		FileSize:  fileSize,
		MimeType:  mimeType,
		UploadID:  uploadID,
		ObjectKey: finalObjectKey,
		Bucket:    bucket,
		Status:    fileStatus,
	})
}
