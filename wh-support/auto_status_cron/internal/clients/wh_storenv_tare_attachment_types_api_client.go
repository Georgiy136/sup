package clients

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	createtarestatemodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_tare_state/models"

	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/sentry"
	fastclient "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

type WhStorenvTareAttachmentTypesApiClient struct {
	cli *fastclient.HttpClient
}

func NewWhStorenvTareAttachmentTypesApiClient() *WhStorenvTareAttachmentTypesApiClient {
	return &WhStorenvTareAttachmentTypesApiClient{
		cli: fastclient.NewHttpClient(),
	}
}

func (w *WhStorenvTareAttachmentTypesApiClient) Configure(ctx context.Context, config configs.Config) {
	w.cli.InitHttpClient("wh_storenv_tare_attachment_types_api_client")(ctx, config)
}

func (w *WhStorenvTareAttachmentTypesApiClient) CreateTareState(ctx context.Context, body createtarestatemodels.RequestForTareStateCreate, employeeID int64) (err error) {
	const (
		apiKey = "TareStateCreate"
		base   = 10
	)

	var response *fasthttp.Response

	span := sentry.StartHTTPClientSpan(ctx, apiKey)
	defer func() {
		sentry.FinishHTTPClientSpan(span, response, err)
	}()

	params := map[string]string{
		"employee_id": strconv.FormatInt(employeeID, base),
	}

	response, err = w.cli.HttpRequestWithCtx(ctx, nil, http.MethodPost, body, apiKey, params)
	if err != nil {
		return w.cli.GenerateError(response, fmt.Errorf("can't http request: %w", err), apiKey)
	}
	if response == nil {
		return w.cli.GenerateError(response, fmt.Errorf("invalid response: %w", internalerrors.ErrResponseNil), apiKey)
	}
	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusOK:
		return nil
	case http.StatusNoContent:
		return nil
	case http.StatusUnprocessableEntity:
		logrus.Debugf("can't process response, status code: %v err: %v", response.StatusCode(), w.cli.GenerateError(response, fmt.Errorf("unprocessable entity: %w", err), apiKey))

		errWithMsg, parseErr := internalerrors.ParseErrorMessageFromResponse(response)
		if parseErr != nil {
			logrus.Errorf("[%s] can't parse response error message from response: %v", apiKey, parseErr)
		}

		return errWithMsg
	case http.StatusInternalServerError:
		return w.cli.GenerateError(response, internalerrors.ErrInternalServerError, apiKey)
	default:
		return w.cli.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}
}
