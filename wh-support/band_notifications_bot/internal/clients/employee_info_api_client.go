package clients

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/valyala/fasthttp"
	customerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/models"
	"gitlab.wildberries.ru/wbwh/support/utils.git/common"
	fastclient "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	jsoniter "github.com/json-iterator/go"
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

func (s *EmployeeInfoApiClient) GetEmployeeName(employeeID int64) (string, error) {
	const (
		apiKey = "GetEmployeeSelfFullName"
		base   = 10
	)
	params := map[string]string{
		"employee_id": strconv.FormatInt(employeeID, base),
	}

	response, err := s.c.HTTPRequest(nil, http.MethodGet, nil, apiKey, params)
	if err != nil {
		return "", s.c.GenerateError(response, err, apiKey)
	}

	if response == nil {
		return "", customerrors.ErrNilResponse
	}
	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusOK:
		var body common.DataWrapper[models.EmployeeInfo]
		if err = jsoniter.Unmarshal(response.Body(), &body); err != nil {
			return "", fmt.Errorf("invalid response body: %w", err)
		}
		return body.Data.Name, nil
	case http.StatusNoContent:
		return "", nil
	default:
		return "", s.c.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}
}
