package clients

import (
	"context"
	"errors"
	"fmt"

	"github.com/valyala/fasthttp"
	gocore_constant "gitlab.wildberries.ru/wbwh/wh-core/gocore_service_const.git/constants"

	"net/http"
	"strconv"

	client "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
)

type WhStorenvTareAttachmentTypesApiClient struct {
	httpClient *client.HttpClient
}

const (
	ClientConfKey = "wh_storenv_tare_attachment_types_api_client"
)

func NewWhStorenvTareAttachmentTypesApiClient(ctx context.Context, config configs.Config) *WhStorenvTareAttachmentTypesApiClient {
	apiClient := WhStorenvTareAttachmentTypesApiClient{
		httpClient: client.NewHttpClient(),
	}

	apiClient.httpClient.InitHttpClient(ClientConfKey)(ctx, config)
	return &apiClient
}

func (c *WhStorenvTareAttachmentTypesApiClient) CheckTareStateByID(ctx *gin.Context, stateID string, employeeID int64) (bool, string, error) {
	const apiKey = "check_tare_state_by_id"

	body := struct {
		StateID string `json:"state_id"`
	}{StateID: stateID}

	params := map[string]string{
		"employee_id": strconv.FormatInt(employeeID, 10),
	}

	response, err := c.httpClient.HTTPRequestWithOpts(ctx, http.MethodPost, body, apiKey, params, func(request *fasthttp.Request) error {
		request.Header.Set(gocore_constant.HeaderWBWHLanguage, ctx.GetHeader(gocore_constant.HeaderWBWHLanguage))
		return nil
	})
	if err != nil {
		return false, "", c.httpClient.GenerateError(response, fmt.Errorf("can't make http request: %w", err), apiKey)
	}

	if response == nil {
		return false, "", c.httpClient.GenerateError(response, fmt.Errorf("invalid response: %w", ErrResponseNil), apiKey)
	}

	switch response.StatusCode() {
	case http.StatusNoContent:
		return true, "", nil
	case http.StatusUnprocessableEntity:
		var respErr DataError

		err := jsoniter.Unmarshal(response.Body(), &respErr)
		if err != nil {
			return false, "", c.httpClient.GenerateError(response, fmt.Errorf("can't unmarshal response error: %w", err), apiKey)
		}

		if len(respErr.Errors) == 0 {
			return false, "", c.httpClient.GenerateError(response, errors.New("empty response error"), apiKey)
		}

		return false, respErr.Errors[0].Message, nil
	default:
		return false, "", c.httpClient.GenerateError(response, fmt.Errorf("unexpected status code %d: %w", response.StatusCode(), ErrWrongStatusCode), apiKey)
	}
}
