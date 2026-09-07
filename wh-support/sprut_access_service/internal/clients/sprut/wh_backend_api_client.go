package clients

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	jsoniter "github.com/json-iterator/go"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/sprut_access_service/internal/models"
	client "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_core.git/rest_data"
)

type WhBackendApiClient struct {
	httpClient *client.HttpClient
}

func NewWhBackendApiClient() *WhBackendApiClient {
	return &WhBackendApiClient{
		httpClient: client.NewHttpClient(),
	}
}

func (b *WhBackendApiClient) Configure(ctx context.Context, cfg configs.Config) {
	b.httpClient.InitHttpClient("wh_backend_api_client")(ctx, cfg)
}

func (b *WhBackendApiClient) GetCallBacks(apiKey string, bytesToSend []byte) (*models.ResponseWithCallBacks, error) {
	res, err := b.httpClient.HTTPRequest(nil, http.MethodPost, bytesToSend, apiKey, nil)
	if err != nil {
		return nil, fmt.Errorf("[%s] HTTPRequest error: %w", apiKey, err)
	}

	switch res.StatusCode() {
	case http.StatusOK:
		var rd rest_data.RestData
		if err = jsoniter.Unmarshal(res.Body(), &rd); err != nil {
			return nil, b.httpClient.GenerateError(res, err, apiKey)
		}
		var dataResp models.ResponseWithCallBacks
		if err = jsoniter.Unmarshal(rd.GetData(), &dataResp); err != nil {
			return nil, b.httpClient.GenerateError(res, err, apiKey)
		}
		if len(dataResp.ApiCallBacks.Apis) == 0 {
			return nil, b.httpClient.GenerateError(res, errors.New("apis empty"), apiKey)
		}
		return &dataResp, nil
	case http.StatusNoContent:
		return nil, nil
	default:
		return nil, b.httpClient.GenerateError(res, errors.New("unexpected status code"), apiKey)
	}
}
