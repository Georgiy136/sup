package clients

import (
	"context"
	"fmt"
	"net/http"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/wh_sprutenv_sprut_offices_api_adapter/internal/common"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/wh_sprutenv_sprut_offices_api_adapter/internal/models"
	client "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
)

type MapsApiClient struct {
	httpClient *client.HttpClient
}

func NewMapsApiClient(ctx context.Context, config configs.Config) *MapsApiClient {
	clientConfKey := "maps_api_client"

	apiClient := MapsApiClient{
		httpClient: client.NewHttpClient(),
	}

	apiClient.httpClient.InitHttpClient(clientConfKey)(ctx, config)
	return &apiClient
}

func (c *MapsApiClient) GetOfficesWithInactiveByOfficeIDs(ctx *gin.Context, officeIDs []int64) (*models.ResponseGetOfficeWithInactive, error) {
	const apiKey = "GetOfficesWithInactive"

	bodyRequest := struct {
		OfficeIDs []int64 `json:"office_ids"`
	}{officeIDs}

	response, err := c.httpClient.HTTPRequest(ctx, http.MethodPost, bodyRequest, apiKey, nil)
	if err != nil {
		return nil, c.httpClient.GenerateError(response, fmt.Errorf("can't make http request: %w", err), apiKey)
	}

	if response == nil {
		return nil, c.httpClient.GenerateError(response, fmt.Errorf("invalid response: %w", common.ErrResponseNil), apiKey)
	}

	switch response.StatusCode() {
	case http.StatusOK:
		var respData models.ResponseGetOfficeWithInactive
		if err := jsoniter.Unmarshal(response.Body(), &respData); err != nil {
			return nil, c.httpClient.GenerateError(response, fmt.Errorf("can't unmarshal response data: %w", err), apiKey)
		}
		return &respData, nil
	default:
		return nil, c.httpClient.GenerateError(
			response,
			fmt.Errorf("unexpected status code %d: %w", response.StatusCode(), common.ErrWrongStatusCode), apiKey)
	}
}
