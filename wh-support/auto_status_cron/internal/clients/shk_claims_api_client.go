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
	removeremainswrrmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/remove_remains_wrr/models"
	fastclient "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

type ShkClaimsApiClient struct {
	cli *fastclient.HttpClient
}

func NewShkClaimsApiClient() *ShkClaimsApiClient {
	return &ShkClaimsApiClient{
		cli: fastclient.NewHttpClient(),
	}
}

func (s *ShkClaimsApiClient) Configure(ctx context.Context, config configs.Config) {
	s.cli.InitHttpClient("shk_claims_api_client")(ctx, config)
}

func (s *ShkClaimsApiClient) ShksRelease(ctx context.Context, body removeremainswrrmodels.RequestDataForShksRelease, employeeID int64) (err error) {
	const apiKey = "ShkClaimsShksRelease"

	var response *fasthttp.Response

	span := sentry.StartHTTPClientSpan(ctx, apiKey)
	defer func() {
		sentry.FinishHTTPClientSpan(span, response, err)
	}()

	params := map[string]string{
		"employee_id": strconv.FormatInt(employeeID, 10),
	}

	response, err = s.cli.HttpRequestWithCtx(ctx, nil, http.MethodPost, body, apiKey, params)
	if err != nil {
		return s.cli.GenerateError(response, fmt.Errorf("can't http request: %w", err), apiKey)
	}

	if response == nil {
		return s.cli.GenerateError(response, fmt.Errorf("invalid response: %w", internalerrors.ErrResponseNil), apiKey)
	}
	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusOK:
		return nil
	case http.StatusNoContent:
		return nil
	case http.StatusUnprocessableEntity:
		logrus.Debugf("can't process response, status code: %v err: %v", response.StatusCode(), s.cli.GenerateError(response, fmt.Errorf("unprocessable entity"), apiKey))

		errWithMsg, parseErr := internalerrors.ParseErrorMessageFromResponse(response)
		if parseErr != nil {
			logrus.Errorf("[%s] can't parse response error message from response: %v", apiKey, parseErr)
		}

		return errWithMsg
	case http.StatusInternalServerError:
		return s.cli.GenerateError(response, internalerrors.ErrInternalServerError, apiKey)
	default:
		return s.cli.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}
}

func (s *ShkClaimsApiClient) GoodsValidation(ctx context.Context, body internalresortmodels.RequestDataForGoodsValidation, employeeID int64) (result []internalresortmodels.ResponseDataForGoodsValidation, err error) {
	const apiKey = "GoodsValidation"

	var response *fasthttp.Response

	span := sentry.StartHTTPClientSpan(ctx, apiKey)
	defer func() {
		sentry.FinishHTTPClientSpan(span, response, err)
	}()

	params := map[string]string{
		"employee_id": strconv.FormatInt(employeeID, 10),
	}

	response, err = s.cli.HttpRequestWithCtx(ctx, nil, http.MethodPost, body, apiKey, params)
	if err != nil {
		return nil, s.cli.GenerateError(response, fmt.Errorf("can't http request: %w", err), apiKey)
	}

	if response == nil {
		return nil, s.cli.GenerateError(response, fmt.Errorf("invalid response: %w", internalerrors.ErrResponseNil), apiKey)
	}
	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusNoContent:
		return nil, nil
	case http.StatusOK:
		var wrapper clientsmodels.DataWrapper[[]internalresortmodels.ResponseDataForGoodsValidation]
		if err = jsoniter.Unmarshal(response.Body(), &wrapper); err != nil {
			return nil, fmt.Errorf("can't parse response body: %w", err)
		}

		return wrapper.Data, nil
	case http.StatusInternalServerError:
		return nil, s.cli.GenerateError(response, internalerrors.ErrInternalServerError, apiKey)
	default:
		return nil, s.cli.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}
}
