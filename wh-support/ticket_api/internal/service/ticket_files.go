package service

import (
	"errors"
	"fmt"

	custom_errors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/common"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/models"
	"gitlab.wildberries.ru/wbwh/support/utils.git/access_actions"
	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"

	"gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git/validators"
	utils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git"
	core_errors "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors/errors_keys"

	"github.com/gin-gonic/gin"
)

type TicketFilesService struct {
	errBuilder                   core_errors.ErrorBuilder
	redisClient                  RedisClientInterface
	resourceEmployeeAccessClient ResourceEmployeeAccessClientInterface
	fileManager                  FileManagerClientInterface
	ticketAccess                 *TicketAccessChecker
}

func NewTicketFilesService(
	redisClient RedisClientInterface,
	resourceEmployeeAccessClient ResourceEmployeeAccessClientInterface,
	fileManager FileManagerClientInterface,
) *TicketFilesService {
	errorsRepo := errors_keys.NewErrorsRepository()
	errorsRepo.SetErrors(support_err_keys.ErrorsRepository)

	return &TicketFilesService{
		errBuilder:                   core_errors.NewErrorBuilder(errorsRepo.GetErrors(), nil),
		redisClient:                  redisClient,
		resourceEmployeeAccessClient: resourceEmployeeAccessClient,
		fileManager:                  fileManager,
		ticketAccess:                 NewTicketAccessChecker(redisClient, resourceEmployeeAccessClient),
	}
}

func (t *TicketFilesService) GetUploadFileLink(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		MimeType string `json:"mime_type" binding:"required,min=1,max=150"`
		FileSize int64  `json:"file_size" binding:"required,gt=0"`
		FileName string `json:"file_name" binding:"required,min=1,max=100"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	req := models.FileManagerInitUploadRequest{
		FileName:   body.FileName,
		FileSize:   body.FileSize,
		MimeType:   body.MimeType,
		EntityType: models.EntityTypeTicketAttachment,
	}

	data, err := t.fileManager.InitUpload(ctx, employeeID, req)
	if err != nil {
		t.handleErrorResponse(ctx, err)
		return
	}

	utils.BindObjectToRestData(ctx, data)
}

func (t *TicketFilesService) CompleteFileUpload(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		UploadID int64 `json:"upload_id" binding:"required,gt=0"`
		FileID   int64 `json:"file_id" binding:"required,gt=0"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	err := t.fileManager.ConfirmUpload(ctx, employeeID, models.FileManagerConfirmUploadRequest{
		UploadID: body.UploadID,
		FileID:   body.FileID,
	})
	if err != nil {
		t.handleErrorResponse(ctx, err)
		return
	}
	utils.BindNoContent(ctx)
}

func (t *TicketFilesService) GetFileDownloadURL(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		FileID   int64 `json:"file_id" binding:"required,gt=0"`
		TicketID int64 `json:"ticket_id" binding:"required,gt=0"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	hasAccess, err := t.ticketAccess.HasEmployeeTicketAccess(ctx, employeeID, body.TicketID,
		access_actions.TypeActionViewTicketCategory,
		access_actions.TypeActionStatusApproveCategory,
		access_actions.TypeActionStatusPerformCategory,
	)
	if err != nil {
		if errors.Is(err, custom_errors.ErrTicketNotFound) {
			t.errBuilder.BindError(ctx, support_err_keys.KeyErrorTicketNotFound, err)
			return
		}
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, fmt.Errorf("check ticket access error: %w", err))
		return
	}
	if !hasAccess {
		t.errBuilder.BindError(ctx, errors_keys.AuthErrCritTokenNotAllowed, errors.New("no access to download file for ticket"))
		return
	}

	data, err := t.fileManager.GetDownloadURL(ctx, employeeID, body.FileID)
	if err != nil {
		t.handleErrorResponse(ctx, err)
		return
	}

	utils.BindObjectToRestData(ctx, data)
}

func (t *TicketFilesService) CancelFileUpload(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		UploadID int64 `json:"upload_id" binding:"required,gt=0"`
		FileID   int64 `json:"file_id" binding:"required,gt=0"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	err := t.fileManager.CancelUpload(ctx, employeeID, models.FileManagerCancelUploadRequest{
		UploadID: body.UploadID,
		FileID:   body.FileID,
	})
	if err != nil {
		t.handleErrorResponse(ctx, err)
		return
	}
	utils.BindNoContent(ctx)
}

func (s *TicketFilesService) handleErrorResponse(ctx *gin.Context, err error) {
	if rd := custom_errors.ToRestData(err); rd != nil {
		utils.BindRestData(ctx, *rd)
		return
	}
	s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), err)
}
