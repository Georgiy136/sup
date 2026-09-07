package clients

import (
	"context"
	"fmt"
	"github.com/valyala/fasthttp"
	gocore_constant "gitlab.wildberries.ru/wbwh/wh-core/gocore_service_const.git/constants"

	"net/http"
	"strconv"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/wbstorenv_place_info_api_adapter/internal/models"
	client "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
)

type WbstorenvPlaceInfoApiClient struct {
	httpClient *client.HttpClient
}

const (
	ClientConfKey = "wbstorenv_place_info_api_client"
)

func NewWbstorenvPlaceInfoApiClient(ctx context.Context, config configs.Config) *WbstorenvPlaceInfoApiClient {
	apiClient := WbstorenvPlaceInfoApiClient{
		httpClient: client.NewHttpClient(),
	}

	apiClient.httpClient.InitHttpClient(ClientConfKey)(ctx, config)
	return &apiClient
}

func (c *WbstorenvPlaceInfoApiClient) GetStoragePlaceStickerPrefixes(ctx *gin.Context) (*models.StickerPrefixes, error) {
	const apiKey = "get_sticker_prefixes"

	response, err := c.httpClient.HTTPRequest(ctx, http.MethodGet, nil, apiKey, nil)
	if err != nil {
		return nil, c.httpClient.GenerateError(response, fmt.Errorf("can't make http request: %w", err), apiKey)
	}

	if response == nil {
		return nil, c.httpClient.GenerateError(response, fmt.Errorf("invalid response: %w", ErrResponseNil), apiKey)
	}

	switch response.StatusCode() {
	case http.StatusOK:
		var respData models.StickerPrefixes
		if err = jsoniter.Unmarshal(response.Body(), &respData); err != nil {
			return nil, c.httpClient.GenerateError(response, fmt.Errorf("can't unmarshal response data: %w", err), apiKey)
		}

		return &respData, nil
	case http.StatusNoContent:
		return nil, nil
	default:
		return nil, c.httpClient.GenerateError(
			response,
			fmt.Errorf("unexpected status code %d: %w", response.StatusCode(), ErrWrongStatusCode), apiKey)
	}
}

func (c *WbstorenvPlaceInfoApiClient) CheckStoragePlaceType(ctx *gin.Context, placeTypeName string) (bool, string, error) {
	const apiKey = "check_storage_place_type"

	params := map[string]string{
		"place_type_name": placeTypeName,
	}
	response, err := c.httpClient.HTTPRequestWithOpts(ctx, http.MethodGet, nil, apiKey, params, func(request *fasthttp.Request) error {
		request.Header.Set(gocore_constant.HeaderWBWHLanguage, ctx.GetHeader(gocore_constant.HeaderWBWHLanguage))
		return nil
	})
	if err != nil {
		return false, "", c.httpClient.GenerateError(response, fmt.Errorf("can't make http request: %w", err), apiKey)
	}

	if response == nil {
		return false, "", c.httpClient.GenerateError(response, fmt.Errorf("invalid response: %w", ErrResponseNil), apiKey)
	}

	switch response.StatusCode() {
	case http.StatusNoContent:
		return true, "", nil
	case http.StatusUnprocessableEntity:
		var DataError DataError

		if err = jsoniter.Unmarshal(response.Body(), &DataError); err != nil {
			return false, "", c.httpClient.GenerateError(response, fmt.Errorf("can't unmarshal response err data: %w", err), apiKey)
		}

		return false, DataError.Errors[0].Message, nil
	default:
		return false, "", c.httpClient.GenerateError(
			response,
			fmt.Errorf("unexpected status code %d: %w", response.StatusCode(), ErrWrongStatusCode), apiKey)
	}
}

func (c *WbstorenvPlaceInfoApiClient) GetNamesOfStoragePlaceTypes(ctx *gin.Context, employeeID int64) (*models.StoragePlaceTypes, error) {
	const apiKey = "get_names_of_storage_place_types"

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
		var respData models.StoragePlaceTypes
		if err = jsoniter.Unmarshal(response.Body(), &respData); err != nil {
			return nil, c.httpClient.GenerateError(response, fmt.Errorf("can't unmarshal response data: %w", err), apiKey)
		}

		return &respData, nil
	case http.StatusNoContent:
		return nil, nil
	default:
		return nil, c.httpClient.GenerateError(
			response,
			fmt.Errorf("unexpected status code %d: %w", response.StatusCode(), ErrWrongStatusCode), apiKey)
	}
}
