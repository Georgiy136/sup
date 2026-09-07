package clients

import (
	"context"
	"fmt"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/categories_api/internal/consts"
	"net/http"
	"strconv"

	jsoniter "github.com/json-iterator/go"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/categories_api/internal/common"
	fastclient "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
)

type ResourceEmployeeAccessClient struct {
	c *fastclient.HttpClient
}

func NewResourceEmployeeAccessClient() *ResourceEmployeeAccessClient {
	return &ResourceEmployeeAccessClient{
		c: fastclient.NewHttpClient(),
	}
}

func (a *ResourceEmployeeAccessClient) Configure(ctx context.Context, config configs.Config) {
	a.c.InitHttpClient("resources_employee_access_client")(ctx, config)
}

func (r *ResourceEmployeeAccessClient) GetAccessActionsByEmployeeID(employeeID int64) ([]string, error) {
	const apiKey = "GetAccessActionsByEmployeeID"

	params := map[string]string{
		"employee_id": strconv.FormatInt(employeeID, consts.Base10),
	}

	response, err := r.c.HTTPRequest(nil, http.MethodGet, nil, apiKey, params)
	if err != nil {
		return nil, r.c.GenerateError(response, err, apiKey)
	}

	if response == nil {
		return nil, r.c.GenerateError(response, fmt.Errorf("invalid response: %w", common.ErrResponseNil), apiKey)
	}

	switch response.StatusCode() {
	case http.StatusOK:
		body := struct {
			Data struct {
				ActionIds []string `json:"action_ids"`
			} `json:"data"`
		}{}
		if err = jsoniter.Unmarshal(response.Body(), &body); err != nil {
			return nil, fmt.Errorf("can't parse response body: %w", err)
		}

		return body.Data.ActionIds, nil
	case http.StatusNoContent:
		return nil, nil
	default:
		return nil, r.c.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}
}
