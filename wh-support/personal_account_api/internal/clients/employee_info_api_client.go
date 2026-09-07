package clients

import (
	"context"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/personal_account_api/internal/models"
	fastclient "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	"net/http"
	"strconv"
)

type EmployeeInfoApiClient struct {
	c *fastclient.HttpClient
}

func NewEmployeeInfoApiClient(ctx context.Context, config configs.Config) *EmployeeInfoApiClient {
	client := &EmployeeInfoApiClient{
		c: fastclient.NewHttpClient(),
	}

	client.c.InitHttpClient("employee_info_api_client")(ctx, config)

	return client
}

func (s *EmployeeInfoApiClient) GetEmployeeSelfFullName(employeeID int64) (*models.EmployeeNameInfo, error) {
	const (
		apiKey = "GetEmployeeSelfFullName"
		base   = 10
	)
	params := map[string]string{
		"employee_id": strconv.FormatInt(employeeID, base),
	}

	response, err := s.c.HTTPRequest(nil, http.MethodGet, nil, apiKey, params)
	if err != nil {
		return nil, s.c.GenerateError(response, err, apiKey)
	}

	body := struct {
		Data models.EmployeeNameInfo `json:"data"`
	}{}

	switch response.StatusCode() {
	case http.StatusNoContent:
		return nil, nil
	case http.StatusOK:
		if err = jsoniter.Unmarshal(response.Body(), &body); err != nil {
			return nil, fmt.Errorf("invalid response body: %w", err)
		}
	default:
		return nil, s.c.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}

	return &body.Data, nil
}
