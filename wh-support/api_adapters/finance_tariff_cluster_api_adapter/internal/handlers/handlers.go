package handlers

import (
	"context"
	"fmt"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/finance_tariff_cluster_api_adapter/internal/clients"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/finance_tariff_cluster_api_adapter/internal/models"
	core_errors "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors/errors_keys"

	"github.com/gin-gonic/gin"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	utils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git"
)

type ProxyService struct {
	client     *clients.FinanceTariffClusterClient
	errBuilder core_errors.ErrorBuilder
}

func New() *ProxyService {
	return &ProxyService{
		errBuilder: core_errors.NewErrorBuilder(errors_keys.NewErrorsRepository().GetErrors(), nil),
	}
}

func (s *ProxyService) Configure(ctx context.Context, config configs.Config) {
	s.client = clients.NewFinanceTariffClusterClient(ctx, config)
}

func (s *ProxyService) GetAllProdTypesHandler(ctx *gin.Context, _ map[string]interface{}) {
	resp, err := s.client.GetAllProdTypes(ctx)
	if err != nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get all info about prodtypes: %w", err))
		return
	}

	if resp == nil || len(resp.Data) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	utils.BindObjectToRestData(ctx, resp)
}

func (s *ProxyService) GetSelectorProdTypePartsHandler(ctx *gin.Context, _ map[string]interface{}) {
	resp, err := s.client.GetAllProdTypeParts(ctx)
	if err != nil {
		s.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("error get all info about prodtype parts: %w", err))
		return
	}

	if resp == nil || len(resp.Data) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	selectorProdTypes := make([]models.SelectorProdType, 0, len(resp.Data))
	for i := range resp.Data {
		selectorProdTypes = append(selectorProdTypes, models.SelectorProdType{
			ProdTypeID:   resp.Data[i].ProdTypePartID,
			ProdTypeName: resp.Data[i].ProdTypePartName,
		})
	}

	utils.BindObjectToRestData(ctx, selectorProdTypes)
}
