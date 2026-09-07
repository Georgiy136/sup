package clients

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/wbstorenv_api_adapter/internal/models"
	client "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
)

const (
	ClientConfKey = "wbstorenv_api_client"
)

type WbstorenvClient struct {
	httpClient *client.HttpClient
}

func NewWbstorenvClient(ctx context.Context, config configs.Config) *WbstorenvClient {
	apiClient := WbstorenvClient{
		httpClient: client.NewHttpClient(),
	}

	apiClient.httpClient.InitHttpClient(ClientConfKey)(ctx, config)
	return &apiClient
}

func (w *WbstorenvClient) GetWarehouseInfoByWhID(ctx *gin.Context, whID int64) (*models.WarehouseInfo, error) {
	const apiKey = "get_warehouse_info_by_wh_id"

	params := map[string]string{
		"wh_id": strconv.FormatInt(whID, 10),
	}

	response, err := w.httpClient.HTTPRequest(ctx, http.MethodGet, nil, apiKey, params)
	if err != nil {
		return nil, w.httpClient.GenerateError(response, fmt.Errorf("can't make http request: %w", err), apiKey)
	}

	if response == nil {
		return nil, w.httpClient.GenerateError(response, errors.New("response is nil"), apiKey)
	}
	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusOK:
		body := struct {
			WarehouseInfo models.WarehouseInfo `json:"data"`
		}{}
		if err := jsoniter.Unmarshal(response.Body(), &body); err != nil {
			return nil, w.httpClient.GenerateError(response, fmt.Errorf("can't unmarshal response data: %w", err), apiKey)
		}

		return &body.WarehouseInfo, nil
	case http.StatusNoContent:
		return nil, nil
	default:
		return nil, w.httpClient.GenerateError(
			response,
			fmt.Errorf("unexpected status code %d", response.StatusCode()), apiKey)
	}
}
