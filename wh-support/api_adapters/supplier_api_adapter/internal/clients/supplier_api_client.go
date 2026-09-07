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
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/supplier_api_adapter/internal/models"
	client "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
)

const (
	ClientConfKey = "supplier_api_client"
)

type SupplierApiClient struct {
	httpClient *client.HttpClient
}

func NewSupplierApiClient(ctx context.Context, config configs.Config) *SupplierApiClient {
	apiClient := SupplierApiClient{
		httpClient: client.NewHttpClient(),
	}

	apiClient.httpClient.InitHttpClient(ClientConfKey)(ctx, config)
	return &apiClient
}

func (s *SupplierApiClient) GetSupplierInfoByNmIDs(ctx *gin.Context, employeeID int64, nmIDs []int64) ([]models.SupplierInfo, error) {
	const apiKey = "get_supplier_info_by_nm_ids"

	params := map[string]string{
		"employee_id": strconv.FormatInt(employeeID, 10),
	}

	bodyRequest := struct {
		NmIDs []int64 `json:"nm_ids"`
	}{nmIDs}

	response, err := s.httpClient.HTTPRequest(ctx, http.MethodPost, bodyRequest, apiKey, params)
	if err != nil {
		return nil, s.httpClient.GenerateError(response, fmt.Errorf("can't make http request: %w", err), apiKey)
	}

	if response == nil {
		return nil, s.httpClient.GenerateError(response, errors.New("response is nil"), apiKey)
	}

	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusOK:
		body := struct {
			SupplierInfo []models.SupplierInfo `json:"data"`
		}{}
		if err := jsoniter.Unmarshal(response.Body(), &body); err != nil {
			return nil, s.httpClient.GenerateError(response, fmt.Errorf("can't unmarshal response data: %w", err), apiKey)
		}

		return body.SupplierInfo, nil
	case http.StatusNoContent:
		return nil, nil
	default:
		return nil, s.httpClient.GenerateError(
			response,
			fmt.Errorf("unexpected status code %d", response.StatusCode()), apiKey)
	}
}
