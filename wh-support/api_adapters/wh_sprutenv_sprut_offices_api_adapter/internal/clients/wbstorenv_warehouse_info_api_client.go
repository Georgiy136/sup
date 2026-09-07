package clients

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/wh_sprutenv_sprut_offices_api_adapter/internal/common"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/wh_sprutenv_sprut_offices_api_adapter/internal/models"
	client "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
)

type WbstorenvWarehouseInfoApiClient struct {
	httpClient *client.HttpClient
}

const (
	ClientConfKey = "wbstorenv_warehouse_info_api_client"
)

func NewWbstorenvWarehouseInfoApiClient(ctx context.Context, config configs.Config) *WbstorenvWarehouseInfoApiClient {
	apiClient := WbstorenvWarehouseInfoApiClient{
		httpClient: client.NewHttpClient(),
	}

	apiClient.httpClient.InitHttpClient(ClientConfKey)(ctx, config)
	return &apiClient
}

func (c *WbstorenvWarehouseInfoApiClient) GetBranchOfficeInfoByID(ctx *gin.Context, employeeID, officeID int64) (*models.ResponseGetBranchOfficeInfo, error) {
	const (
		apiKey = "GetBranchOfficeInfoByID"
	)

	params := map[string]string{
		"employee_id": strconv.FormatInt(employeeID, 10),
	}

	body := struct {
		OfficeIDs []int64 `json:"office_ids"`
	}{[]int64{officeID}}

	response, err := c.httpClient.HTTPRequest(ctx, http.MethodPost,
		body,
		apiKey,
		params,
	)
	if err != nil {
		return nil, c.httpClient.GenerateError(response, fmt.Errorf("can't make http request: %w", err), apiKey)
	}

	if response == nil {
		return nil, c.httpClient.GenerateError(response, fmt.Errorf("invalid response: %w", common.ErrResponseNil), apiKey)
	}
	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusOK:
		var respData models.ResponseGetBranchOfficeInfo
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
