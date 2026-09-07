package clients

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/finance_tariff_cluster_api_adapter/common"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/finance_tariff_cluster_api_adapter/internal/models"
	client "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	"net/http"
)

type FinanceTariffClusterClient struct {
	httpClient *client.HttpClient
}

func NewFinanceTariffClusterClient(ctx context.Context, config configs.Config) *FinanceTariffClusterClient {
	apiClient := FinanceTariffClusterClient{
		httpClient: client.NewHttpClient(),
	}

	apiClient.httpClient.InitHttpClient("finance_tariff_cluster_client")(ctx, config)
	return &apiClient
}

func (c *FinanceTariffClusterClient) GetAllProdTypes(ctx *gin.Context) (*models.ProdTypes, error) {
	const apiKey = "prodtypes_get_all_with_parts"

	response, err := c.httpClient.HTTPRequest(ctx, http.MethodGet, nil, apiKey, nil)
	if err != nil {
		return nil, c.httpClient.GenerateError(response, fmt.Errorf("can't make http request: %w", err), apiKey)
	}

	if response == nil {
		return nil, c.httpClient.GenerateError(response, fmt.Errorf("invalid response: %w", common.ErrResponseNil), apiKey)
	}

	switch response.StatusCode() {
	case http.StatusOK:
		var respData models.ProdTypes
		if err := jsoniter.Unmarshal(response.Body(), &respData); err != nil {
			return nil, c.httpClient.GenerateError(response, fmt.Errorf("can't unmarshal response data: %w", err), apiKey)
		}

		return &respData, nil
	case http.StatusNoContent:
		return nil, nil
	default:
		return nil, c.httpClient.GenerateError(
			response,
			fmt.Errorf("unexpected status code %d: %w", response.StatusCode(), common.ErrWrongStatusCode), apiKey)
	}
}

func (c *FinanceTariffClusterClient) GetAllProdTypeParts(ctx *gin.Context) (*models.ProdTypeParts, error) {
	const apiKey = "prodtype_part_get_all"

	response, err := c.httpClient.HTTPRequest(ctx, http.MethodGet, nil, apiKey, nil)
	if err != nil {
		return nil, c.httpClient.GenerateError(response, fmt.Errorf("can't make http request: %w", err), apiKey)
	}

	if response == nil {
		return nil, c.httpClient.GenerateError(response, fmt.Errorf("invalid response: %w", common.ErrResponseNil), apiKey)
	}

	switch response.StatusCode() {
	case http.StatusOK:
		var respData models.ProdTypeParts
		if err := jsoniter.Unmarshal(response.Body(), &respData); err != nil {
			return nil, c.httpClient.GenerateError(response, fmt.Errorf("can't unmarshal response data: %w", err), apiKey)
		}

		return &respData, nil
	case http.StatusNoContent:
		return nil, nil
	default:
		return nil, c.httpClient.GenerateError(
			response,
			fmt.Errorf("unexpected status code %d: %w", response.StatusCode(), common.ErrWrongStatusCode), apiKey)
	}
}
