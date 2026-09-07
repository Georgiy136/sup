package clients

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
	fastclient "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
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

func (s *EmployeeInfoApiClient) GetEmployeeSelfFullName(ctx context.Context, employeeID int64) (string, error) {
	const (
		apiKey = "GetEmployeeSelfFullName"
		base   = 10
	)
	params := map[string]string{
		"employee_id": strconv.FormatInt(employeeID, base),
	}

	response, err := s.c.HttpRequestWithCtx(ctx, nil, http.MethodGet, nil, apiKey, params)
	if err != nil {
		return "", fmt.Errorf("get employee self full name: %w", s.c.GenerateError(response, err, apiKey))
	}
	if response == nil {
		return "", errors.New("response is nil")
	}
	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusNoContent:
		return "", nil
	case http.StatusOK:
		body := struct {
			Data struct {
				EmployeeName string `json:"employee_name"`
			} `json:"data"`
		}{}
		if err = jsoniter.Unmarshal(response.Body(), &body); err != nil {
			return "", fmt.Errorf("invalid response body: %w", err)
		}
		return body.Data.EmployeeName, nil
	default:
		return "", fmt.Errorf("get employee self full name: %w", s.c.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey))
	}
}
