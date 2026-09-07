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
	createstorageplacetypemodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_storage_place_type/models"
	fastclient "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
)

type WbStorenvPlaceInfoApiClient struct {
	cli *fastclient.HttpClient
}

func NewWbStorenvPlaceInfoApiClient() *WbStorenvPlaceInfoApiClient {
	return &WbStorenvPlaceInfoApiClient{
		cli: fastclient.NewHttpClient(),
	}
}

func (t *WbStorenvPlaceInfoApiClient) Configure(ctx context.Context, config configs.Config) {
	t.cli.InitHttpClient("wbstorenv_place_info_api_client")(ctx, config)
}

func (t *WbStorenvPlaceInfoApiClient) CreateStoragePlaceType(ctx context.Context, body createstorageplacetypemodels.RequestForCreateStoragePlaceType) (result *createstorageplacetypemodels.StoragePlaceType, err error) {
	const apiKey = "StoragePlaceTypeUpd"

	var response *fasthttp.Response

	span := sentry.StartHTTPClientSpan(ctx, apiKey)
	defer func() {
		sentry.FinishHTTPClientSpan(span, response, err)
	}()

	response, err = t.cli.HttpRequestWithCtx(ctx, nil, http.MethodPost, body, apiKey, nil)
	if err != nil {
		return nil, t.cli.GenerateError(response, fmt.Errorf("can't http request: %w", err), apiKey)
	}
	if response == nil {
		return nil, t.cli.GenerateError(response, fmt.Errorf("invalid response: %w", internalerrors.ErrResponseNil), apiKey)
	}
	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusOK:
		var resp createstorageplacetypemodels.ResponseDataFromCreateStoragePlaceTypeApi
		if err = jsoniter.Unmarshal(response.Body(), &resp); err != nil {
			return nil, fmt.Errorf("[%s] Unmarshal data error: %w", apiKey, err)
		}
		return &resp.Data, nil
	case http.StatusNoContent:
		return nil, nil
	case http.StatusUnprocessableEntity:
		logrus.Debugf("can't process response, status code: %v err: %v", response.StatusCode(), t.cli.GenerateError(response, fmt.Errorf("unprocessable entity: %w", err), apiKey))

		errWithMsg, parseErr := internalerrors.ParseErrorMessageFromResponse(response)
		if parseErr != nil {
			logrus.Errorf("[%s] can't parse response error message from response: %v", apiKey, parseErr)
		}

		return nil, errWithMsg
	case http.StatusInternalServerError:
		return nil, t.cli.GenerateError(response, internalerrors.ErrInternalServerError, apiKey)
	default:
		return nil, t.cli.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}
}
