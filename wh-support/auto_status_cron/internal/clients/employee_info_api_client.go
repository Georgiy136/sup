package clients

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/clients/models"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/sentry"
	fastclient "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
)

type EmployeeInfoApiClient struct {
	c *fastclient.HttpClient
}

func NewEmployeeInfoApiClient() *EmployeeInfoApiClient {
	return &EmployeeInfoApiClient{
		c: fastclient.NewHttpClient(),
	}
}

func (s *EmployeeInfoApiClient) Configure(ctx context.Context, config configs.Config) {
	s.c.InitHttpClient("employee_info_api_client")(ctx, config)
}

func (s *EmployeeInfoApiClient) GetEmployeeName(ctx context.Context, employeeID int64) (name string, err error) {
	const (
		apiKey = "GetEmployeeSelfFullName"
		base   = 10
	)

	var response *fasthttp.Response

	span := sentry.StartHTTPClientSpan(ctx, apiKey)
	defer func() {
		sentry.FinishHTTPClientSpan(span, response, err)
	}()

	params := map[string]string{
		"employee_id": strconv.FormatInt(employeeID, base),
	}

	response, err = s.c.HttpRequestWithCtx(ctx, nil, http.MethodGet, nil, apiKey, params)
	if err != nil {
		return "", s.c.GenerateError(response, err, apiKey)
	}
	if response == nil {
		return "", nil
	}
	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusOK:
		var result models.DataWrapper[struct {
			models.EmployeeInfo
		}]
		if err = jsoniter.Unmarshal(response.Body(), &result); err != nil {
			return "", fmt.Errorf("invalid response body: %w", err)
		}

		return result.Data.Name, nil
	case http.StatusNoContent:
		return "", nil
	default:
		return "", s.c.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}
}
