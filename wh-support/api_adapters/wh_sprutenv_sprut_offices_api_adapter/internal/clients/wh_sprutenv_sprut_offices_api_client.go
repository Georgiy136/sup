package clients

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/wh_sprutenv_sprut_offices_api_adapter/internal/common"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/wh_sprutenv_sprut_offices_api_adapter/internal/models"
	client "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
)

type WhSprutenvSprutOfficesApiClient struct {
	httpClient *client.HttpClient
}

func NewWhSprutenvSprutOfficesApiClient(ctx context.Context, config configs.Config) *WhSprutenvSprutOfficesApiClient {
	clientConfKey := "wh_sprutenv_sprut_offices_api"

	apiClient := WhSprutenvSprutOfficesApiClient{
		httpClient: client.NewHttpClient(),
	}

	apiClient.httpClient.InitHttpClient(clientConfKey)(ctx, config)
	return &apiClient
}

func (c *WhSprutenvSprutOfficesApiClient) GetSprutNames(ctx *gin.Context) (*models.Spruts, error) {
	const apiKey = "GetSprutNames"

	response, err := c.httpClient.HTTPRequest(ctx, http.MethodGet, nil, apiKey, nil)
	if err != nil {
		return nil, c.httpClient.GenerateError(response, fmt.Errorf("can't make http request: %w", err), apiKey)
	}

	if response == nil {
		return nil, c.httpClient.GenerateError(response, fmt.Errorf("invalid response: %w", common.ErrResponseNil), apiKey)
	}

	switch response.StatusCode() {
	case http.StatusOK:
		var respData models.Spruts
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

func (c *WhSprutenvSprutOfficesApiClient) CheckExistSprutOffice(ctx *gin.Context, officeID int64) (*models.OfficeOnSprut, error) {
	const apiKey = "CheckExistSprutOffice"

	body := models.RequestChesExistSprutOffice{
		OfficeID: officeID,
	}

	response, err := c.httpClient.HTTPRequest(ctx, http.MethodPost, body, apiKey, nil)
	if err != nil {
		return nil, c.httpClient.GenerateError(response, fmt.Errorf("can't make http request: %w", err), apiKey)
	}

	if response == nil {
		return nil, c.httpClient.GenerateError(response, fmt.Errorf("invalid response: %w", common.ErrResponseNil), apiKey)
	}

	switch response.StatusCode() {
	case http.StatusOK:
		var respData models.DataOfficeOnSprut
		if err := jsoniter.Unmarshal(response.Body(), &respData); err != nil {
			return nil, c.httpClient.GenerateError(response, fmt.Errorf("can't unmarshal response data: %w", err), apiKey)
		}
		return &respData.Data, nil
	default:
		return nil, c.httpClient.GenerateError(
			response,
			fmt.Errorf("unexpected status code %d: %w", response.StatusCode(), common.ErrWrongStatusCode), apiKey)
	}
}
