package clients

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	jsoniter "github.com/json-iterator/go"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/resources_employee_access/internal/common"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/resources_employee_access/internal/models"
	fastclient "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
)

type AccessManagerApiClient struct {
	c *fastclient.HttpClient
}

func NewAccessManagerApiClient(ctx context.Context, config configs.Config) *AccessManagerApiClient {
	client := &AccessManagerApiClient{
		c: fastclient.NewHttpClient(),
	}

	client.c.InitHttpClient("access_manager_api_client")(ctx, config)

	return client
}

func (s *AccessManagerApiClient) GetPermittedAppActions(employeeID int64) ([]models.EmployeeAppAction, error) {
	const (
		apiKey      = "GetPermittedAppActions"
		base        = 10
		appClientID = 8
	)
	params := map[string]string{
		"employee_id": strconv.FormatInt(employeeID, base),
	}

	requestBody := struct {
		AppClientID int64 `json:"app_client_id"`
		EmployeeID  int64 `json:"employee_id"`
	}{appClientID, employeeID}

	response, err := s.c.HTTPRequest(nil, http.MethodPost, requestBody, apiKey, params)
	if err != nil {
		return nil, s.c.GenerateError(response, err, apiKey)
	}

	if response == nil {
		return nil, s.c.GenerateError(response, fmt.Errorf("invalid response: %w", common.ErrResponseNil), apiKey)
	}

	switch response.StatusCode() {
	case http.StatusOK:
		body := struct {
			Data []models.EmployeeAppAction `json:"data"`
		}{}

		if err = jsoniter.Unmarshal(response.Body(), &body); err != nil {
			return nil, fmt.Errorf("invalid response body: %w", err)
		}

		return body.Data, nil
	case http.StatusNoContent:
		return nil, nil
	default:
		return nil, s.c.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}
}
