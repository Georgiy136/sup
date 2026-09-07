package sprut

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/sentry"
	client "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_core.git/rest_data"

	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
)

type WhBackendApiClient struct {
	httpClient *client.HttpClient
}

func NewWhBackendApiClient() *WhBackendApiClient {
	return &WhBackendApiClient{
		httpClient: client.NewHttpClient(),
	}
}

func (c *WhBackendApiClient) Configure(ctx context.Context, config configs.Config) {
	c.httpClient.InitHttpClient("sprut_backend_api_client")(ctx, config)
}

func (c *WhBackendApiClient) GetCallBacks(ctx context.Context, apiKey string, bytesToSend []byte) (result *PositiveResponseV001, err error) {
	var response *fasthttp.Response

	span := sentry.StartHTTPClientSpan(ctx, apiKey)
	defer func() {
		sentry.FinishHTTPClientSpan(span, response, err)
	}()

	response, err = c.httpClient.HttpRequestWithCtx(ctx, nil, http.MethodPost, bytesToSend, apiKey, nil)
	if err != nil {
		return nil, fmt.Errorf("HTTPRequest error: %w", err)
	}
	if response == nil {
		return nil, c.httpClient.GenerateError(response, fmt.Errorf("invalid response: %w", internalerrors.ErrResponseNil), apiKey)
	}
	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusOK:
		var rd rest_data.RestData
		if err = jsoniter.Unmarshal(response.Body(), &rd); err != nil {
			return nil, c.httpClient.GenerateError(response, err, apiKey)
		}
		var dataResp PositiveResponseV001
		if err = jsoniter.Unmarshal(rd.GetData(), &dataResp); err != nil {
			return nil, c.httpClient.GenerateError(response, err, apiKey)
		}
		if len(dataResp.ApiCallBacks.Apis) == 0 {
			return nil, c.httpClient.GenerateError(response, errors.New("apis empty"), apiKey)
		}
		return &dataResp, nil
	case http.StatusNoContent:
		return nil, nil
	default:
		return nil, c.httpClient.GenerateError(response, errors.New("unexpected status code"), apiKey)
	}
}
