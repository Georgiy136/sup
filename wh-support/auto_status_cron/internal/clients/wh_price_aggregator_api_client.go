package clients

import (
	"context"
	"fmt"
	"net/http"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/sentry"
	closewhmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/close_wh_new/models"
	fastclient "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
)

type WhPriceAggregatorApiClient struct {
	cli *fastclient.HttpClient
}

func NewWhPriceAggregatorApiClient() *WhPriceAggregatorApiClient {
	return &WhPriceAggregatorApiClient{
		cli: fastclient.NewHttpClient(),
	}
}

func (s *WhPriceAggregatorApiClient) Configure(ctx context.Context, config configs.Config) {
	s.cli.InitHttpClient("wh_price_aggregator_api_client")(ctx, config)
}

func (s *WhPriceAggregatorApiClient) GetAveragePriceNm(ctx context.Context, body closewhmodels.RequestPriceAggregation) (result *closewhmodels.ResponsePriceAggregation, err error) {
	const apiKey = "GetAveragePriceNm"

	var response *fasthttp.Response

	span := sentry.StartHTTPClientSpan(ctx, apiKey)
	defer func() {
		sentry.FinishHTTPClientSpan(span, response, err)
	}()

	response, err = s.cli.HttpRequestWithCtx(ctx, nil, http.MethodPost, body, apiKey, nil)
	if err != nil {
		return nil, s.cli.GenerateError(response, fmt.Errorf("can't http request: %w", err), apiKey)
	}
	if response == nil {
		return nil, s.cli.GenerateError(response, fmt.Errorf("invalid response: %w", internalerrors.ErrResponseNil), apiKey)
	}
	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusOK:
		var result closewhmodels.ResponsePriceAggregation
		if err = jsoniter.Unmarshal(response.Body(), &result); err != nil {
			return nil, s.cli.GenerateError(response, fmt.Errorf("can't unmarshal response: %w", err), apiKey)
		}
		return &result, nil

	case http.StatusNoContent:
		return nil, nil
	case http.StatusUnprocessableEntity:
		logrus.Debugf("can't process response, status code: %v", response.StatusCode())

		errWithMsg, parseErr := internalerrors.ParseErrorMessageFromResponse(response)
		if parseErr != nil {
			logrus.Errorf("[%s] can't parse response error message: %v", apiKey, parseErr)
		}

		return nil, errWithMsg
	case http.StatusInternalServerError:
		return nil, s.cli.GenerateError(response, internalerrors.ErrInternalServerError, apiKey)
	default:
		return nil, s.cli.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}
}
