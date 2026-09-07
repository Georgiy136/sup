package clients

import (
	"context"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/employee_group_api/internal/models"
	fastclient "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	"net/http"
	"strconv"
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

func (s *EmployeeInfoApiClient) GetEmployeesFullName(chEmployeeID int64, employeeIDs []int64) ([]models.EmployeeInfo, error) {
	const (
		apiKey = "GetEmployeeFullName"
		base   = 10
	)
	params := map[string]string{
		"employee_id": strconv.FormatInt(chEmployeeID, base),
	}

	reqBody := struct {
		EmployeeIDs []int64 `json:"employee_ids"`
	}{employeeIDs}

	response, err := s.c.HTTPRequest(nil, http.MethodPost, reqBody, apiKey, params)
	if err != nil {
		return nil, s.c.GenerateError(response, err, apiKey)
	}

	switch response.StatusCode() {
	case http.StatusNoContent:
		return nil, nil
	case http.StatusOK:
		respData := struct {
			Data []models.EmployeeInfo `json:"data"`
		}{}

		if err = jsoniter.Unmarshal(response.Body(), &respData); err != nil {
			return nil, fmt.Errorf("invalid response body: %w", err)
		}

		return respData.Data, nil
	default:
		return nil, s.c.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}
}
