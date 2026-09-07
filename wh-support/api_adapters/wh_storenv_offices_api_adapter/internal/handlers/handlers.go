package handlers

import (
	"context"
	"fmt"

	core_errors "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors/errors_keys"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/wh_storenv_offices_api_adapter/internal/clients"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/wh_storenv_offices_api_adapter/internal/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	utils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

type ProxyService struct {
	client            *clients.WhStorenvOfficesApiClient
	allowedTypePoints map[int64]struct{}
	errBuilder        core_errors.ErrorBuilder
}

func New() *ProxyService {
	return &ProxyService{
		allowedTypePoints: make(map[int64]struct{}),
		errBuilder:        core_errors.NewErrorBuilder(errors_keys.NewErrorsRepository().GetErrors(), nil),
	}
}

func (s *ProxyService) Configure(ctx context.Context, config configs.Config) {
	s.client = clients.NewWhStorenvOfficesApiClient(ctx, config)

	var conf models.ParamsKeys
	if err := jsoniter.Unmarshal(config.GetByServiceKeyRequired("service_param_keys"), &conf); err != nil {
		logrus.Panicf("error parsing service_param_keys configs - %v", err)
	}

	for _, val := range conf.AllowedTypePoints {
		s.allowedTypePoints[val] = struct{}{}
	}

	logrus.Infof("allowed office type points: %v", conf.AllowedTypePoints)
}

func (s *ProxyService) GetFilteredTypePoint(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	resp, err := s.client.GetAllTypePoint(ctx, employeeID)
	if err != nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get all type point: %w", err))
		return
	}

	if len(resp) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	selectorTypePoints := make([]models.TypePoint, 0, len(resp))
	for _, v := range resp {
		if _, ok := s.allowedTypePoints[v.TypePoint]; ok {
			selectorTypePoints = append(selectorTypePoints, v)
		}
	}

	if len(selectorTypePoints) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	utils.BindObjectToRestData(ctx, selectorTypePoints)
}
