package service

import (
	"context"
	"errors"
	"fmt"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/external_mobile_control_pa_adapter/internal/clients"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/external_mobile_control_pa_adapter/internal/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git/validators"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	utils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git"
	core_errors "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors/errors_keys"

	"github.com/gin-gonic/gin"
)

type ProxyService struct {
	client     *clients.ExternalMobileControlProxyAdapterClient
	errBuilder core_errors.ErrorBuilder
}

func New() *ProxyService {
	return &ProxyService{
		errBuilder: core_errors.NewErrorBuilder(errors_keys.NewErrorsRepository().GetErrors(), nil),
	}
}

func (s *ProxyService) Configure(ctx context.Context, config configs.Config) {
	s.client = clients.NewExternalMobileControlProxyAdapterClient(ctx, config)
}

func (s *ProxyService) CheckDevice(ctx *gin.Context, params map[string]interface{}) {
	body, ok := params[validators.BodyValidatorBODY].(*struct {
		DeviceSerialNumber string `json:"value" binding:"required,max=36"`
	})
	if !ok {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	resp, msgErr, err := s.client.CheckDevice(ctx, body.DeviceSerialNumber)
	if err != nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get info about device: %w", err))
		return
	}

	if msgErr != "" {
		utils.BindObjectToRestData(ctx, models.CheckDeviceResponse{
			IsValid: false,
			ErrMsg:  msgErr,
		})
		return
	}

	utils.BindObjectToRestData(ctx, models.CheckDeviceResponse{
		IsValid:     true,
		DevicesInfo: []models.DeviceInfo{*resp},
	})
}
