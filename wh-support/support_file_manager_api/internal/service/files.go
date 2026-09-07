package service

import (
	"errors"
	"fmt"
	"time"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/support_file_manager_api/internal/constant"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/support_file_manager_api/internal/models"
	file_utils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/support_file_manager_api/internal/utils"
	time_utils "gitlab.wildberries.ru/wbwh/wh-core/gocore-utils-time.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git/validators"
	utils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git"
	core_errors "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors/errors_keys"

	"github.com/gin-gonic/gin"
)

func (s *SupportFileManagerService) GetDownloadURL(ctx *gin.Context, params map[string]interface{}) {
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		FileID int64 `json:"file_id" binding:"required,gt=0"`
	})
	if !ok {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	// Получаем инфо о файле по file_id
	fileInfo, err := s.repository.GetFileInfoByID(body.FileID)
	if err != nil {
		s.bindServiceError(ctx, err, "failed to get file")
		return
	}

	if fileInfo.Status != constant.FileStatusActive {
		s.errBuilder.BindError(ctx, errors_keys.ErrBizDbResponseEmpty, fmt.Errorf("file is not active: %d", body.FileID))
		return
	}

	// Определяем время истечения ссылки
	expiration := time.Duration(s.urlExpiration.DownloadMinutes) * time.Minute

	// Генерируем ссылку на скачивание
	downloadURL, err := s.s3Client.GetPresignedDownloadURL(ctx.Request.Context(), models.GetDownloadURLRequest{
		Bucket:           fileInfo.Bucket,
		ObjectKey:        fileInfo.ObjectKey,
		OriginalFileName: file_utils.EncodeFilenameForMetadata(fileInfo.OriginalName),
		ContentType:      fileInfo.MimeType,
		Expiration:       expiration,
	})
	if err != nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("failed to generate download URL: %w", err))
		return
	}

	utils.BindObjectToRestData(ctx, models.GetDownloadURLResponse{
		FileID:      fileInfo.FileID,
		DownloadURL: downloadURL,
		TicketID:    fileInfo.TicketID,
		FileName:    fileInfo.OriginalName,
		ExpiresAt:   time.Now().Add(expiration),
	})
}

func (s *SupportFileManagerService) DeleteFile(ctx *gin.Context, params map[string]interface{}) {
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		FileID int64 `json:"file_id" binding:"required,gt=0"`
	})
	if !ok {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	// Получаем инфо о файле по file_id
	file, err := s.repository.GetFileInfoByID(body.FileID)
	if err != nil {
		s.bindServiceError(ctx, err, "failed to get file")
		return
	}

	newStatus, err := s.statusManager.CanDeleteFile(file.Status)
	if err != nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrBizDbResponseEmpty, fmt.Errorf("file %d: %w", body.FileID, err))
		return
	}

	// Помечаем файл как готовый к удалению
	if err = s.repository.UpdateFileStatus(body.FileID, newStatus); err != nil {
		s.bindServiceError(ctx, err, "failed to mark file as ready for deletion")
		return
	}

	utils.BindNoContent(ctx)
}

func (s *SupportFileManagerService) GetFilesForDeletion(ctx *gin.Context, params map[string]interface{}) {
	files, err := s.repository.GetFilesByStatus(constant.FileStatusDeletePending)
	if err != nil {
		s.bindServiceError(ctx, err, "failed to get files for deletion")
		return
	}

	if len(files) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	utils.BindObjectToRestData(ctx, files)
}

func (s *SupportFileManagerService) GetFilesNotAttachedToTicketForDeleting(ctx *gin.Context, _ map[string]any) {
	files, err := s.repository.GetFilesByStatus(constant.FileStatusAttachPending)
	if err != nil {
		s.bindServiceError(ctx, err, "failed to get files not attached to tickets")
		return
	}

	filesForDeleting := make([]models.File, 0, len(files))
	timeNow := time.Now()

	for i := range files {
		if files[i].ExpiresAt == nil {
			s.errBuilder.BindError(ctx, errors_keys.ErrVldDatetimeWrong, fmt.Errorf("expires time can't be nil in file %d", files[i].FileID))
			return
		}

		timeExpires, err := time.Parse(time_utils.LayoutDataBasePG, *files[i].ExpiresAt)
		if err != nil {
			s.bindServiceError(ctx, err, "parse RFC3339Nano time err")
			return
		}

		if timeNow.After(timeExpires) {
			filesForDeleting = append(filesForDeleting, files[i])
		}
	}

	if len(filesForDeleting) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	utils.BindObjectToRestData(ctx, filesForDeleting)
}

func (s *SupportFileManagerService) AttachFilesToTicket(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		FileIDs  []int64 `json:"file_ids" binding:"required,min=1,dive,gt=0"`
		TicketID int64   `json:"ticket_id" binding:"required,gt=0"`
	})
	if !ok {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	for _, fileID := range body.FileIDs {
		file, err := s.repository.GetFileInfoByID(fileID)
		if err != nil {
			s.bindServiceError(ctx, err, "failed to get file")
			return
		}

		if file.CreateEmployeeID != employeeID {
			s.errBuilder.BindError(ctx, errors_keys.ErrBizDbResponseEmpty, fmt.Errorf("employee does not have access to file: %d", fileID))
			return
		}

		if file.TicketID != nil {
			s.errBuilder.BindError(ctx, errors_keys.ErrBizDbResponseEmpty, fmt.Errorf("file %d already attached to ticket", fileID))
			return
		}

		if _, err = s.statusManager.CanAttachFile(file.Status); err != nil {
			s.errBuilder.BindError(ctx, errors_keys.ErrBizDbResponseEmpty, fmt.Errorf("file %d: %w", fileID, err))
			return
		}
	}

	if err := s.repository.AttachTicketIDToFiles(body.FileIDs, body.TicketID); err != nil {
		s.bindServiceError(ctx, err, "failed to update files ticket ID")
		return
	}

	for _, fileID := range body.FileIDs {
		if err := s.repository.UpdateFileStatus(fileID, constant.FileStatusActive); err != nil {
			s.bindServiceError(ctx, err, fmt.Sprintf("failed to mark file %d as active", fileID))
			return
		}
	}

	utils.BindNoContent(ctx)
}

func (s *SupportFileManagerService) MarkFileDeleted(ctx *gin.Context, params map[string]interface{}) {
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		FileID int64 `json:"file_id" binding:"required,gt=0"`
	})
	if !ok {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	// Получаем инфо о файле по file_id
	file, err := s.repository.GetFileInfoByID(body.FileID)
	if err != nil {
		s.bindServiceError(ctx, err, "failed to get file")
		return
	}

	newStatus, err := s.statusManager.CanMarkFileDeleted(file.Status)
	if err != nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrBizDbResponseEmpty, err)
		return
	}

	// Удаляем файл из бакета
	if err = s.s3Client.DeleteObject(ctx.Request.Context(), file.Bucket, file.ObjectKey); err != nil {
		s.bindServiceError(ctx, err, "failed to delete object from bucket")
		return
	}

	// Меняем статус на DELETED
	if err = s.repository.UpdateFileStatus(body.FileID, newStatus); err != nil {
		s.bindServiceError(ctx, err, "failed to mark file as deleted")
		return
	}

	utils.BindNoContent(ctx)
}

func (s *SupportFileManagerService) GetFileInfo(ctx *gin.Context, params map[string]interface{}) {
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		FileID int64 `json:"file_id" binding:"required,gt=0"`
	})
	if !ok {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	fileInfo, err := s.repository.GetFileInfoByID(body.FileID)
	if err != nil {
		s.bindServiceError(ctx, err, fmt.Sprintf("failed to get file info by id: %d", body.FileID))
		return
	}

	utils.BindObjectToRestData(ctx, fileInfo)
}
