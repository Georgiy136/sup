package clients

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/maps_api_adapter/models"
	client "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	"net/http"
)

type MapsApiClient struct {
	httpClient *client.HttpClient
}

const (
	ClientConfKey = "maps_api_client"
)

func NewMapsApiClient(ctx context.Context, config configs.Config) *MapsApiClient {
	apiClient := MapsApiClient{
		httpClient: client.NewHttpClient(),
	}
	apiClient.httpClient.InitHttpClient(ClientConfKey)(ctx, config)
	return &apiClient
}

func (c *MapsApiClient) GetOffices(ctx *gin.Context, body []byte) (*models.GetOfficesResponse, error) {
	const apiKey = "get_offices"

	response, err := c.httpClient.HTTPRequest(ctx, http.MethodPost, body, apiKey, nil)
	if err != nil {
		return nil, c.httpClient.GenerateError(response, fmt.Errorf("can't make http request: %w", err), apiKey)
	}

	if response == nil {
		return nil, c.httpClient.GenerateError(response, errors.New("response is nil"), apiKey)
	}

	switch response.StatusCode() {
	case http.StatusOK:
		var respData models.GetOfficesResponse
		if err = jsoniter.Unmarshal(response.Body(), &respData); err != nil {
			return nil, c.httpClient.GenerateError(response, fmt.Errorf("can't unmarshal response data: %w", err), apiKey)
		}
		return &respData, nil
	case http.StatusNoContent:
		return nil, nil
	default:
		return nil, c.httpClient.GenerateError(
			response,
			fmt.Errorf("unexpected status code %d", response.StatusCode()), apiKey)
	}
}

func (c *MapsApiClient) GetSpruts(ctx *gin.Context) (*models.GetSprutsResponse, error) {
	const apiKey = "get_spruts"

	response, err := c.httpClient.HTTPRequest(ctx, http.MethodGet, nil, apiKey, nil)
	if err != nil {
		return nil, c.httpClient.GenerateError(response, fmt.Errorf("can't make http request: %w", err), apiKey)
	}

	if response == nil {
		return nil, c.httpClient.GenerateError(response, errors.New("response is nil"), apiKey)
	}

	switch response.StatusCode() {
	case http.StatusOK:
		var respData models.GetSprutsResponse
		if err = jsoniter.Unmarshal(response.Body(), &respData); err != nil {
			return nil, c.httpClient.GenerateError(response, fmt.Errorf("can't unmarshal response data: %w", err), apiKey)
		}
		return &respData, nil
	case http.StatusNoContent:
		return nil, nil
	default:
		return nil, c.httpClient.GenerateError(
			response,
			fmt.Errorf("unexpected status code %d", response.StatusCode()), apiKey)
	}
}
