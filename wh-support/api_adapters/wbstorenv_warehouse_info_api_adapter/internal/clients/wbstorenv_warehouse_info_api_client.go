package clients

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/valyala/fasthttp"
	gocore_constant "gitlab.wildberries.ru/wbwh/wh-core/gocore_service_const.git/constants"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/wbstorenv_warehouse_info_api_adapter/internal/models"
	client "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
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

func (c *WbstorenvWarehouseInfoApiClient) ListWarehousesOffices(ctx *gin.Context, employeeID int64) (*models.ListWarehousesOffices, error) {
	const (
		apiKey = "list_warehouses_offices"
	)

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
		var respData models.ListWarehousesOffices
		if err := jsoniter.Unmarshal(response.Body(), &respData); err != nil {
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

func (c *WbstorenvWarehouseInfoApiClient) GetWHByOffice(ctx *gin.Context, officeId string) (*models.WH, error) {
	const apiKey = "get_wh_by_office"

	response, err := c.httpClient.HTTPRequest(ctx, http.MethodGet, nil, apiKey, map[string]string{"office_id": officeId})
	if err != nil {
		return nil, c.httpClient.GenerateError(response, fmt.Errorf("can't make http request: %w", err), apiKey)
	}

	if response == nil {
		return nil, c.httpClient.GenerateError(response, fmt.Errorf("invalid response: %w", ErrResponseNil), apiKey)
	}

	switch response.StatusCode() {
	case http.StatusOK:
		var respData models.WH
		if err := jsoniter.Unmarshal(response.Body(), &respData); err != nil {
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

func (c *WbstorenvWarehouseInfoApiClient) GetWHByOfficeV2(ctx *gin.Context, officeId int64) ([]models.WHInfoV2, error) {
	const apiKey = "get_wh_by_office_v2"

	params := map[string]string{
		"office_id": strconv.FormatInt(officeId, 10),
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
		var respData struct {
			WhInfo []models.WHInfoV2 `json:"data"`
		}
		if err := jsoniter.Unmarshal(response.Body(), &respData); err != nil {
			return nil, c.httpClient.GenerateError(response, fmt.Errorf("can't unmarshal response data: %w", err), apiKey)
		}

		return respData.WhInfo, nil
	case http.StatusNoContent:
		return nil, nil
	default:
		return nil, c.httpClient.GenerateError(
			response,
			fmt.Errorf("unexpected status code %d: %w", response.StatusCode(), ErrWrongStatusCode), apiKey)
	}
}

func (c *WbstorenvWarehouseInfoApiClient) CheckNewWhAbbreviation(ctx *gin.Context, whAbbreviation string) (bool, string, error) {
	const apiKey = "check_new_wh_abbreviation"

	params := map[string]string{
		"place_name_prefix": whAbbreviation,
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

func (c *WbstorenvWarehouseInfoApiClient) CheckNewWhName(ctx *gin.Context, whName string) (bool, string, error) {
	const apiKey = "check_new_wh_name"

	params := map[string]string{
		"wh_name": whName,
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

func (c *WbstorenvWarehouseInfoApiClient) BusinessProcessesGetAll(ctx *gin.Context) (*models.BusinessProcesses, error) {
	const apiKey = "business_processes_get_all"

	response, err := c.httpClient.HTTPRequest(ctx, http.MethodGet, nil, apiKey, nil)
	if err != nil {
		return nil, c.httpClient.GenerateError(response, fmt.Errorf("can't make http request: %w", err), apiKey)
	}

	if response == nil {
		return nil, c.httpClient.GenerateError(response, fmt.Errorf("invalid response: %w", ErrResponseNil), apiKey)
	}

	switch response.StatusCode() {
	case http.StatusOK:
		var respData models.BusinessProcesses
		if err := jsoniter.Unmarshal(response.Body(), &respData); err != nil {
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

func (c *WbstorenvWarehouseInfoApiClient) ShkStateGetAll(ctx *gin.Context) (*models.GetShkStates, error) {
	const apiKey = "shk_state_get_all"

	response, err := c.httpClient.HTTPRequest(ctx, http.MethodGet, nil, apiKey, nil)
	if err != nil {
		return nil, c.httpClient.GenerateError(response, fmt.Errorf("can't make http request: %w", err), apiKey)
	}

	if response == nil {
		return nil, c.httpClient.GenerateError(response, fmt.Errorf("invalid response: %w", ErrResponseNil), apiKey)
	}

	switch response.StatusCode() {
	case http.StatusOK:
		var respData models.GetShkStates
		if err := jsoniter.Unmarshal(response.Body(), &respData); err != nil {
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

func (c *WbstorenvWarehouseInfoApiClient) ShkStateCheckByID(ctx *gin.Context, stateID string) (bool, string, error) {
	const apiKey = "shk_state_check_by_id"

	params := map[string]string{
		"state_id": stateID,
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
	case http.StatusOK:
		return true, "", nil

	case http.StatusNoContent:
		return true, "", nil

	case http.StatusUnprocessableEntity:
		var DataError DataError
		if err := jsoniter.Unmarshal(response.Body(), &DataError); err != nil {
			return false, "", c.httpClient.GenerateError(response, fmt.Errorf("can't unmarshal response err data: %w", err), apiKey)
		}
		return false, DataError.Errors[0].Message, nil

	default:
		return false, "", c.httpClient.GenerateError(
			response,
			fmt.Errorf("unexpected status code %d: %w", response.StatusCode(), ErrWrongStatusCode), apiKey)
	}
}

func (c *WbstorenvWarehouseInfoApiClient) GetBranchOfficeByOfficeIDs(ctx *gin.Context, officeIDs []int64, withDeleted bool) (*models.BranchOffices, error) {
	const apiKey = "branch_offices_list"

	params := map[string]string{
		"with_deleted": strconv.FormatBool(withDeleted),
	}

	response, err := c.httpClient.HTTPRequest(ctx, http.MethodPost,
		struct {
			OfficeIDs []int64 `json:"office_ids"`
		}{officeIDs},
		apiKey,
		params,
	)
	if err != nil {
		return nil, c.httpClient.GenerateError(response, fmt.Errorf("can't make http request: %w", err), apiKey)
	}

	if response == nil {
		return nil, c.httpClient.GenerateError(response, fmt.Errorf("invalid response: %w", ErrResponseNil), apiKey)
	}

	switch response.StatusCode() {
	case http.StatusOK:
		var respData models.BranchOffices
		if err := jsoniter.Unmarshal(response.Body(), &respData); err != nil {
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
