package clients

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/external_mobile_control_pa_adapter/internal/models"
	client "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
)

const (
	ClientConfKey = "external_mobile_control_proxy_adapter_client"
)

type ExternalMobileControlProxyAdapterClient struct {
	httpClient *client.HttpClient
}

func NewExternalMobileControlProxyAdapterClient(ctx context.Context, config configs.Config) *ExternalMobileControlProxyAdapterClient {
	apiClient := ExternalMobileControlProxyAdapterClient{
		httpClient: client.NewHttpClient(),
	}
	apiClient.httpClient.InitHttpClient(ClientConfKey)(ctx, config)
	return &apiClient
}

func (m *ExternalMobileControlProxyAdapterClient) CheckDevice(ctx *gin.Context, deviceSerialNumber string) (*models.DeviceInfo, string, error) {
	const apiKey = "CheckDevice"

	params := map[string]string{
		"device_serial_number": deviceSerialNumber,
	}

	response, err := m.httpClient.HttpRequestWithCtx(ctx, nil, http.MethodGet, nil, apiKey, params)
	if err != nil {
		return nil, "", m.httpClient.GenerateError(response, fmt.Errorf("can't make http request: %w", err), apiKey)
	}

	if response == nil {
		return nil, "", m.httpClient.GenerateError(response, errors.New("response is nil"), apiKey)
	}

	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusOK:
		respData := struct {
			DeviceInfo models.DeviceInfo `json:"data"`
		}{}

		if err = jsoniter.Unmarshal(response.Body(), &respData); err != nil {
			return nil, "", m.httpClient.GenerateError(response, fmt.Errorf("can't unmarshal response data: %w", err), apiKey)
		}
		return &respData.DeviceInfo, "", nil
	case http.StatusUnprocessableEntity:
		var DataError DataError

		if err = jsoniter.Unmarshal(response.Body(), &DataError); err != nil {
			return nil, "", m.httpClient.GenerateError(response, fmt.Errorf("can't unmarshal response err data: %w", err), apiKey)
		}

		return nil, DataError.Errors[0].Message, nil
	default:
		return nil, "", m.httpClient.GenerateError(response, fmt.Errorf("unexpected status code %d", response.StatusCode()), apiKey)
	}
}
