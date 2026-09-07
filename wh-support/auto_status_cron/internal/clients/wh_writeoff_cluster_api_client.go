package clients

import (
	"context"
	"fmt"
	"net/http"

	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/sentry"
	excludetaremodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/exclude_tare/models"
	fastclient "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

type WhWriteoffClusterApiClient struct {
	cli *fastclient.HttpClient
}

func NewWhWriteoffClusterApiClient() *WhWriteoffClusterApiClient {
	return &WhWriteoffClusterApiClient{
		cli: fastclient.NewHttpClient(),
	}
}

func (t *WhWriteoffClusterApiClient) Configure(ctx context.Context, config configs.Config) {
	t.cli.InitHttpClient("wh_writeoff_cluster_api_client")(ctx, config)
}

func (t *WhWriteoffClusterApiClient) ExcludeTareOnWriteoff(ctx context.Context, body []excludetaremodels.RequestBodyForExcludeTareOnWriteoffApi) (err error) {
	const apiKey = "TareToPostponedWriteoffUpd"

	var response *fasthttp.Response

	span := sentry.StartHTTPClientSpan(ctx, apiKey)
	defer func() {
		sentry.FinishHTTPClientSpan(span, response, err)
	}()

	response, err = t.cli.HttpRequestWithCtx(ctx, nil, http.MethodPost, body, apiKey, nil)
	if err != nil {
		return t.cli.GenerateError(response, fmt.Errorf("can't http request: %w", err), apiKey)
	}
	if response == nil {
		return t.cli.GenerateError(response, fmt.Errorf("invalid response: %w", internalerrors.ErrResponseNil), apiKey)
	}
	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusNoContent:
		return nil
	case http.StatusUnprocessableEntity:
		logrus.Debugf("can't process response, status code: %v err: %v", response.StatusCode(), t.cli.GenerateError(response, fmt.Errorf("unprocessable entity: %w", err), apiKey))

		errWithMsg, parseErr := internalerrors.ParseErrorMessageFromResponse(response)
		if parseErr != nil {
			logrus.Errorf("[%s] can't parse response error message from response: %v", apiKey, parseErr)
		}

		return errWithMsg
	case http.StatusInternalServerError:
		return t.cli.GenerateError(response, internalerrors.ErrInternalServerError, apiKey)
	default:
		return t.cli.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}
}
