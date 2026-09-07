package handlers

import (
	"context"
	"fmt"

	"errors"

	"gitlab.wildberries.ru/wbwh/wh-core/gocore_auth.git"
	core_errors "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors/errors_keys"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/wh_storenv_tare_attachment_types_api_adapter/internal/clients"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/wh_storenv_tare_attachment_types_api_adapter/internal/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git/validators"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	utils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git"

	"github.com/gin-gonic/gin"
)

type ProxyService struct {
	client     *clients.WhStorenvTareAttachmentTypesApiClient
	errBuilder core_errors.ErrorBuilder
}

func New() *ProxyService {
	return &ProxyService{
		errBuilder: core_errors.NewErrorBuilder(errors_keys.NewErrorsRepository().GetErrors(), nil),
	}
}

func (s *ProxyService) Configure(ctx context.Context, config configs.Config) {
	s.client = clients.NewWhStorenvTareAttachmentTypesApiClient(ctx, config)
}

func (s *ProxyService) CheckTareStateByID(ctx *gin.Context, params map[string]interface{}) {
	employeeID := gocore_auth.MustGetCurrentEmployeeID(ctx)

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		StateID string `json:"value" binding:"required,len=3"`
	})
	if !ok {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	isValidTareStateID, msgErr, err := s.client.CheckTareStateByID(ctx, body.StateID, employeeID)
	if err != nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error check new wh abbreviation: %w", err))
		return
	}

	response := models.ResponseWithMsgError{
		IsValid: isValidTareStateID,
		ErrMsg:  msgErr,
	}

	utils.BindObjectToRestData(ctx, response)
}
