package clients

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/wh_storenv_offices_api_adapter/internal/models"
	client "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
)

type WhStorenvOfficesApiClient struct {
	httpClient *client.HttpClient
}

const (
	ClientConfKey = "wh_storenv_offices_api_client"
)

func NewWhStorenvOfficesApiClient(ctx context.Context, config configs.Config) *WhStorenvOfficesApiClient {
	apiClient := WhStorenvOfficesApiClient{
		httpClient: client.NewHttpClient(),
	}

	apiClient.httpClient.InitHttpClient(ClientConfKey)(ctx, config)
	return &apiClient
}

func (c *WhStorenvOfficesApiClient) GetAllTypePoint(ctx *gin.Context, employeeID int64) ([]models.TypePoint, error) {
	const apiKey = "get_all_type_point"

	params := map[string]string{
		"employee_id": strconv.FormatInt(employeeID, 10),
	}

	response, err := c.httpClient.HTTPRequest(ctx, http.MethodGet, nil, apiKey, params)
	if err != nil {
		return nil, c.httpClient.GenerateError(response, fmt.Errorf("can't make http request: %w", err), apiKey)
	}

	if response == nil {
		return nil, c.httpClient.GenerateError(response, fmt.Errorf("invalid response: %w", ErrResponseNil), apiKey)
	}

	switch response.StatusCode() {
	case http.StatusOK:
		var respData models.ResponseGetAllTypePoint
		if err = jsoniter.Unmarshal(response.Body(), &respData); err != nil {
			return nil, c.httpClient.GenerateError(response, fmt.Errorf("can't unmarshal response data: %w", err), apiKey)
		}

		return respData.Data, nil
	case http.StatusNoContent:
		return nil, nil
	default:
		return nil, c.httpClient.GenerateError(
			response,
			fmt.Errorf("unexpected status code %d: %w", response.StatusCode(), ErrWrongStatusCode), apiKey)
	}
}
