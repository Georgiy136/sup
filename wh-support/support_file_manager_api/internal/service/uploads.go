package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/support_file_manager_api/internal/constant"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/support_file_manager_api/internal/models"
	file_utils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/support_file_manager_api/internal/utils"
	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"
	utils_time "gitlab.wildberries.ru/wbwh/wh-core/gocore-utils-time.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git/validators"
	utils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git"
	core_errors "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors/errors_keys"

	"github.com/gin-gonic/gin"
)

func (s *SupportFileManagerService) InitUpload(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		FileName   string          `json:"file_name" binding:"required,min=1,max=100"`
		FileSize   int64           `json:"file_size" binding:"required,gt=0"`
		MimeType   string          `json:"mime_type" binding:"required,min=1,max=150"`
		EntityType string          `json:"entity_type" binding:"required,min=1,max=25"`
		EntityData json.RawMessage `json:"entity_data" binding:"omitempty"`
	})
	if !ok {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	// Валидация файла по MIME type, размеру и расширению
	if err := s.fileValidator.ValidateFile(body.MimeType, body.FileSize, body.FileName); err != nil {
		s.errBuilder.BindError(ctx, support_err_keys.KeyErrorFileValidationFailed, err)
		return
	}

	// Резервируем file_id
	fileID, err := s.repository.GetNewFileID()
	if err != nil {
		s.bindServiceError(ctx, err, "failed to get new file ID")
		return
	}

	// Резервируем upload_id
	uploadID, err := s.repository.GetNewUploadID()
	if err != nil {
		s.bindServiceError(ctx, err, "failed to get new upload ID")
		return
	}

	// Генерируем временный объектный ключ
	// Формат: tmp/{shardId}/{shardId2}/{uploadId}
	objectKey := file_utils.GenerateTmpObjectKey(uploadID)

	// Генерируем presigned URL для загрузки во временный бакет
	fileNameEncoded := file_utils.EncodeFilenameForMetadata(body.FileName)
	expiration := time.Duration(s.urlExpiration.UploadMinutes) * time.Minute

	response, err := s.s3Client.GetPresignedUploadURL(ctx.Request.Context(), models.GetUploadURLRequest{
		ObjectKey:   objectKey,
		ContentType: body.MimeType,
		FileName:    fileNameEncoded,
		Expiration:  expiration,
	})
	if err != nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("failed to generate upload URL: %w", err))
		return
	}

	// Создаем запись аплоада в БД
	uploadParams := models.UploadAddParams{
		UploadID:          uploadID,
		FileID:            fileID,
		ChEmployeeID:      employeeID,
		EntityType:        body.EntityType,
		EntityData:        body.EntityData,
		OriginalName:      body.FileName,
		DeclaredMimeType:  body.MimeType,
		DeclaredSizeBytes: body.FileSize,
		TmpBucket:         s.s3Client.GetTmpBucketName(),
		TmpObjectKey:      objectKey,
		Status:            constant.UploadStatusInitiated,
		ExpiresAt:         time.Now().Add(expiration).Format(utils_time.LayoutDataBasePG),
	}
	if err = s.repository.AddUpload(uploadParams); err != nil {
		s.bindServiceError(ctx, err, "failed to create upload record")
		return
	}

	utils.BindObjectToRestData(ctx, models.GetUploadURLResponse{
		UploadID:           uploadID,
		FileID:             fileID,
		UploadURL:          response.UploadURL,
		Method:             response.Method,
		ObjectKey:          objectKey,
		ExpiresAt:          uploadParams.ExpiresAt,
		RequiredFormFields: response.RequiredFormFields,
	})
}

func (s *SupportFileManagerService) ConfirmUpload(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		UploadID int64 `json:"upload_id" binding:"required,gt=0"`
		FileID   int64 `json:"file_id" binding:"required,gt=0"`
	})
	if !ok {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	// Проверяем существование аплоада и его текущий статус
	uploadStatus, err := s.repository.GetUploadByID(body.UploadID)
	if err != nil {
		s.bindServiceError(ctx, err, "failed to get upload")
		return
	}

	// Проверяем, что сотрудник является создателем аплоада
	if uploadStatus.ChEmployeeID != employeeID {
		s.errBuilder.BindError(ctx, errors_keys.ErrBizDbResponseEmpty, fmt.Errorf("employee does not have access to upload: %d", body.UploadID))
		return
	}

	// Проверяем, что file_id совпадает
	if uploadStatus.FileID != body.FileID {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, fmt.Errorf("file_id not match"))
		return
	}

	newStatus, err := s.statusManager.CanConfirmUpload(uploadStatus.Status)
	if err != nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrBizDbResponseEmpty, err)
		return
	}

	if err = s.repository.UpdateUploadStatus(body.UploadID, employeeID, newStatus); err != nil {
		s.bindServiceError(ctx, err, "failed to update upload status")
		return
	}

	utils.BindNoContent(ctx)
}

func (s *SupportFileManagerService) UpdateUploadStatus(ctx *gin.Context, params map[string]interface{}) {
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		UploadID int64  `json:"upload_id" binding:"required,gt=0"`
		Status   string `json:"status" binding:"required,oneof=READY INVALID EXPIRED"`
	})
	if !ok {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	// Проверяем существование аплоада
	upload, err := s.repository.GetUploadByID(body.UploadID)
	if err != nil {
		s.bindServiceError(ctx, err, "failed to get upload")
		return
	}

	// Проверяем допустимость перехода статуса
	if err = s.statusManager.ValidateUploadStatusTransition(upload.Status, body.Status); err != nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrBizDbResponseEmpty, err)
		return
	}

	// Обновляем статус
	if err = s.repository.UpdateUploadStatus(body.UploadID, upload.ChEmployeeID, body.Status); err != nil {
		s.bindServiceError(ctx, err, "failed to update upload status")
		return
	}

	// Отправляем в канал аплоада данные по невалидному файлу
	if body.Status == constant.UploadStatusInvalid {
		if err := s.centrifugoClient.PublishUploadStatus(ctx, body.UploadID, false); err != nil {
			s.bindServiceError(ctx, err, "failed to publish centrifugo invalid upload status")
			return
		}
	}

	utils.BindNoContent(ctx)
}

func (s *SupportFileManagerService) CancelUpload(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		UploadID int64 `json:"upload_id" binding:"required,gt=0"`
		FileID   int64 `json:"file_id" binding:"required,gt=0"`
	})
	if !ok {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	// Получаем информацию об аплоаде
	uploadStatus, err := s.repository.GetUploadByID(body.UploadID)
	if err != nil {
		s.bindServiceError(ctx, err, "failed to get upload")
		return
	}

	if uploadStatus.ChEmployeeID != employeeID {
		s.errBuilder.BindError(ctx, errors_keys.ErrBizDbResponseEmpty, fmt.Errorf("employee does not have access to upload: %d", body.UploadID))
		return
	}

	// Проверяем, что file_id совпадает
	if uploadStatus.FileID != body.FileID {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, fmt.Errorf("file_id not match"))
		return
	}

	newStatus, err := s.statusManager.CanCancelUpload(uploadStatus.Status)
	if err != nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrBizDbResponseEmpty, err)
		return
	}

	if err = s.repository.UpdateUploadStatus(body.UploadID, employeeID, newStatus); err != nil {
		s.bindServiceError(ctx, err, "failed to update upload status")
		return
	}

	utils.BindNoContent(ctx)
}

func (s *SupportFileManagerService) GetUploadStatus(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		UploadID int64 `json:"upload_id" binding:"required,gt=0"`
	})
	if !ok {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	uploadStatus, err := s.repository.GetUploadByID(body.UploadID)
	if err != nil {
		s.bindServiceError(ctx, err, "failed to get upload")
		return
	}

	if uploadStatus.ChEmployeeID != employeeID {
		s.errBuilder.BindError(ctx, errors_keys.ErrBizDbResponseEmpty, fmt.Errorf("employee does not have access to upload: %d", body.UploadID))
		return
	}

	response := models.GetUploadStatusResponse{
		UploadID: uploadStatus.UploadID,
		FileID:   uploadStatus.FileID,
		Status:   uploadStatus.Status,
	}

	utils.BindObjectToRestData(ctx, response)
}

func (s *SupportFileManagerService) GetUploadsForValidation(ctx *gin.Context, params map[string]interface{}) {
	uploads, err := s.repository.GetUploadsByStatus(constant.UploadStatusValidating)
	if err != nil {
		s.bindServiceError(ctx, err, "failed to get uploads for validation")
		return
	}

	if len(uploads) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	utils.BindObjectToRestData(ctx, uploads)
}

func (s *SupportFileManagerService) GetExpiredUploads(ctx *gin.Context, params map[string]interface{}) {
	uploads, err := s.repository.GetExpiredUploads()
	if err != nil {
		s.bindServiceError(ctx, err, "failed to get expired uploads")
		return
	}

	if len(uploads) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	utils.BindObjectToRestData(ctx, uploads)
}

// GetUploadsForTransfer возвращает аплоады в статусе READY, готовые к переносу в постоянное хранилище
func (s *SupportFileManagerService) GetUploadsForTransfer(ctx *gin.Context, params map[string]interface{}) {
	uploads, err := s.repository.GetUploadsByStatus(constant.UploadStatusReady)
	if err != nil {
		s.bindServiceError(ctx, err, "failed to get uploads for transfer")
		return
	}

	if len(uploads) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	utils.BindObjectToRestData(ctx, uploads)
}

// TransferFile переносит файл из временного бакета в постоянный
func (s *SupportFileManagerService) TransferFile(ctx *gin.Context, params map[string]interface{}) {
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		UploadID int64 `json:"upload_id" binding:"required,gt=0"`
	})
	if !ok {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	// Получаем информацию об аплоаде
	upload, err := s.repository.GetUploadByID(body.UploadID)
	if err != nil {
		s.bindServiceError(ctx, err, "failed to get upload")
		return
	}

	newStatus, err := s.statusManager.CanTransferFile(upload.Status)
	if err != nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrBizDbResponseEmpty, err)
		return
	}

	// Генерируем ключ для постоянного хранилища
	// Формат: final/{entityType}/{shard1}/{shard2}/{fileId}
	finalObjectKey := file_utils.GenerateFinalObjectKey(upload.EntityType, upload.FileID)
	bucket := s.s3Client.GetBucketName()

	// Копируем файл из временного бакета в постоянный
	err = s.s3Client.CopyObject(ctx.Request.Context(), models.CopyObjectRequest{
		SourceBucket: upload.TmpBucket,
		SourceKey:    upload.TmpObjectKey,
		DestBucket:   bucket,
		DestKey:      finalObjectKey,
	})
	if err != nil {
		s.bindServiceError(ctx, err, "failed to copy file to permanent storage")
		return
	}

	fileStatus := constant.FileStatusAttachPending
	expiresAt := time.Now().Add(constant.AttachPendingFileTTL).Format(utils_time.LayoutDataBasePG)
	fileExpiresAt := &expiresAt

	// Добавляем запись о файле в БД
	fileAddParams := []models.FileAddParams{
		{
			FileID:    upload.FileID,
			UploadID:  upload.UploadID,
			Bucket:    bucket,
			ObjectKey: finalObjectKey,
			Status:    fileStatus,
			ExpiresAt: fileExpiresAt,
		},
	}
	if err = s.repository.AddFiles(fileAddParams); err != nil {
		s.bindServiceError(ctx, err, "failed to add file record")
		return
	}

	if err = s.repository.UpdateUploadStatus(body.UploadID, upload.ChEmployeeID, newStatus); err != nil {
		s.bindServiceError(ctx, err, "failed to update upload status")
		return
	}

	// Отправляем в канал аплоада данные по перенесенному файлу
	if err := s.centrifugoClient.PublishUploadStatus(ctx, body.UploadID, true); err != nil {
		s.bindServiceError(ctx, err, "failed to publish centrifugo transferred upload status")
		return
	}

	// Удаляем файл из временного бакета
	if err = s.s3Client.DeleteObject(ctx.Request.Context(), upload.TmpBucket, upload.TmpObjectKey); err != nil {
		s.bindServiceError(ctx, err, "failed to delete object from temporary bucket")
		return
	}

	utils.BindNoContent(ctx)
}
