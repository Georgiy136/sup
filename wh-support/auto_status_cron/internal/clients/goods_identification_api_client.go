package clients

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	clientsmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/clients/models"
	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/sentry"
	internalresortmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/internal_resort/models"
	fastclient "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

type GoodsIdentificationApiClient struct {
	cli *fastclient.HttpClient
}

func NewGoodsIdentificationApiClient() *GoodsIdentificationApiClient {
	return &GoodsIdentificationApiClient{
		cli: fastclient.NewHttpClient(),
	}
}

func (g *GoodsIdentificationApiClient) Configure(ctx context.Context, config configs.Config) {
	g.cli.InitHttpClient("goods_identification_api_client")(ctx, config)
}

func (g *GoodsIdentificationApiClient) CreateInternalOrder(ctx context.Context, body internalresortmodels.RequestCreateInternalOrder, employeeID int64) (result *internalresortmodels.ResponseCreateInternalOrder, err error) {
	const apiKey = "CreateInternalOrder"

	var response *fasthttp.Response

	span := sentry.StartHTTPClientSpan(ctx, apiKey)
	defer func() {
		sentry.FinishHTTPClientSpan(span, response, err)
	}()

	params := map[string]string{
		"employee_id": strconv.FormatInt(employeeID, 10),
	}

	response, err = g.cli.HttpRequestWithCtx(ctx, nil, http.MethodPost, body, apiKey, params)
	if err != nil {
		return nil, g.cli.GenerateError(response, fmt.Errorf("can't http request: %w", err), apiKey)
	}

	if response == nil {
		return nil, g.cli.GenerateError(response, fmt.Errorf("invalid response: %w", internalerrors.ErrResponseNil), apiKey)
	}
	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusNoContent:
		return nil, nil
	case http.StatusOK:
		var wrapper clientsmodels.DataWrapper[internalresortmodels.ResponseCreateInternalOrder]
		if err = jsoniter.Unmarshal(response.Body(), &wrapper); err != nil {
			return nil, fmt.Errorf("can't parse response body: %w", err)
		}

		return &wrapper.Data, nil
	case http.StatusUnprocessableEntity:
		logrus.Errorf("can't process response, status code: %v err: %v", response.StatusCode(), g.cli.GenerateError(response, fmt.Errorf("unprocessable entity: %w", err), apiKey))

		errWithMsg, parseErr := internalerrors.ParseErrorMessageFromResponse(response)
		if parseErr != nil {
			logrus.Errorf("[%s] can't parse response error message: %v", apiKey, parseErr)
		}

		return nil, errWithMsg
	case http.StatusInternalServerError:
		return nil, g.cli.GenerateError(response, internalerrors.ErrInternalServerError, apiKey)
	default:
		return nil, g.cli.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}
}
