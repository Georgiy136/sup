package clients

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/models"
	fastclient "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	jsoniter "github.com/json-iterator/go"
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

func (s *EmployeeInfoApiClient) GetEmployeeSelfFullName(employeeID int64) (*models.EmployeeInfo, error) {
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

	if response == nil {
		return nil, nil
	}

	switch response.StatusCode() {
	case http.StatusOK:
		body := struct {
			Data models.EmployeeInfo `json:"data"`
		}{}

		if err = jsoniter.Unmarshal(response.Body(), &body); err != nil {
			return nil, fmt.Errorf("invalid response body: %w", err)
		}

		return &body.Data, nil
	case http.StatusNoContent:
		return nil, nil
	default:
		return nil, s.c.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}
}
