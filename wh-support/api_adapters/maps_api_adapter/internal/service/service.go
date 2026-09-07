package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"sync"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/maps_api_adapter/internal/clients"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/maps_api_adapter/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git/validators"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	utils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git"
	core_errors "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors/errors_keys"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

type ProxyService struct {
	client                *clients.MapsApiClient
	allowOfficeTypePoints map[int64]struct{}
	errBuilder            core_errors.ErrorBuilder
}

func New() *ProxyService {
	return &ProxyService{
		allowOfficeTypePoints: make(map[int64]struct{}),
		errBuilder:            core_errors.NewErrorBuilder(errors_keys.NewErrorsRepository().GetErrors(), nil),
	}
}

func (s *ProxyService) Configure(ctx context.Context, config configs.Config) {
	s.client = clients.NewMapsApiClient(ctx, config)

	var conf models.ParamsKeys
	if err := jsoniter.Unmarshal(config.GetByServiceKeyRequired("service_param_keys"), &conf); err != nil {
		logrus.Panicf("error parsing service_param_keys configs - %v", err)
	}
	for _, val := range conf.AllowOfficeTypePoints {
		s.allowOfficeTypePoints[val] = struct{}{}
	}

	logrus.Infof("allowed office type points: %v", conf.AllowOfficeTypePoints)
}

func (s *ProxyService) GetOffices(ctx *gin.Context, params map[string]interface{}) {
	body, ok := params[validators.RawBODY].([]byte)
	if !ok {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if len(body) == 0 {
		s.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	resp, err := s.client.GetOffices(ctx, body)
	if err != nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get all info about offices: %w", err))
		return
	}
	if resp == nil || len(resp.Offices) == 0 {
		utils.BindNoContent(ctx)
		return
	}
	ctx.JSON(http.StatusOK, resp)
	utils.BindNoContent(ctx)
}

func (s *ProxyService) GetSpruts(ctx *gin.Context, _ map[string]interface{}) {
	resp, err := s.client.GetSpruts(ctx)
	if err != nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get all info about spruts: %w", err))
		return
	}
	if resp == nil || len(resp.Spruts) == 0 {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), errors.New("total len spruts is 0"))
		return
	}
	ctx.JSON(http.StatusOK, resp)
	utils.BindNoContent(ctx)
}

func (s *ProxyService) GetFilteredOffices(ctx *gin.Context, _ map[string]interface{}) {
	resp, err := s.client.GetSpruts(ctx)
	if err != nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get all info about spruts: %w", err))
		return
	}
	if resp == nil || len(resp.Spruts) == 0 {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), errors.New("total len spruts is 0"))
		return
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, runtime.GOMAXPROCS(0))

	var mu sync.Mutex
	result := make([]models.Office, 0)

	wg.Add(len(resp.Spruts))

	for i := range resp.Spruts {
		offices := resp.Spruts[i].Offices

		sem <- struct{}{}

		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			allowedOffices := s.getAllowedOffices(offices)

			mu.Lock()
			result = append(result, allowedOffices...)
			mu.Unlock()
		}()
	}

	wg.Wait()
	utils.BindObjectToRestData(ctx, result)
}

func (s *ProxyService) getAllowedOffices(offices []models.Office) []models.Office {
	var result []models.Office
	for i := range offices {
		if _, ok := s.allowOfficeTypePoints[offices[i].TypePoint]; ok {
			result = append(result, offices[i])
		}
	}
	return result
}
