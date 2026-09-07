package clients

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_api/internal/models"
	utils "gitlab.wildberries.ru/wbwh/support/utils.git/common"
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

func (e *EmployeeInfoApiClient) Configure(ctx context.Context, config configs.Config) {
	e.c.InitHttpClient("employee_info_api_client")(ctx, config)
}

func (e *EmployeeInfoApiClient) GetEmployeesFullName(chEmployeeID int64, employeeIDs []int64) ([]models.EmployeeInfo, error) {
	const (
		apiKey = "GetEmployeesFullName"
		base   = 10
	)

	params := map[string]string{
		"employee_id": strconv.FormatInt(chEmployeeID, base),
	}

	requestBody := struct {
		EmployeeIDs []int64 `json:"employee_ids"`
	}{employeeIDs}

	response, err := e.c.HTTPRequest(nil, http.MethodPost, requestBody, apiKey, params)
	if err != nil {
		return nil, e.c.GenerateError(response, err, apiKey)
	}
	if response == nil {
		return nil, e.c.GenerateError(response, errors.New("invalid response: response is nil"), apiKey)
	}
	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusOK:
		var body utils.DataWrapper[[]models.EmployeeInfo]
		if err = jsoniter.Unmarshal(response.Body(), &body); err != nil {
			return nil, fmt.Errorf("invalid response body: %w", err)
		}

		return body.Data, nil
	case http.StatusNoContent:
		return nil, nil
	default:
		return nil, e.c.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}
}
