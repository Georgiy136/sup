package clients

import (
	"context"
	"fmt"
	"net/http"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	clientsmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/clients/models"
	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/sentry"
	createstorageplaceinventtaskmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_storage_place_invent_task/models"
	fastclient "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
)

type StoragePlaceAdministrationApiClient struct {
	cli *fastclient.HttpClient
}

func NewStoragePlaceAdministrationApiClient() *StoragePlaceAdministrationApiClient {
	return &StoragePlaceAdministrationApiClient{
		cli: fastclient.NewHttpClient(),
	}
}

func (s *StoragePlaceAdministrationApiClient) Configure(ctx context.Context, config configs.Config) {
	s.cli.InitHttpClient("storage_place_administration_api")(ctx, config)
}

func (s *StoragePlaceAdministrationApiClient) CreateStoragePlaceInventTask(ctx context.Context, body createstorageplaceinventtaskmodels.RequestDataForCreateStoragePlaceInventTask) (createdQty int64, err error) {
	const apiKey = "StoragePlaceInventTaskCreate"

	var response *fasthttp.Response

	span := sentry.StartHTTPClientSpan(ctx, apiKey)
	defer func() {
		sentry.FinishHTTPClientSpan(span, response, err)
	}()

	response, err = s.cli.HttpRequestWithCtx(ctx, nil, http.MethodPost, body, apiKey, nil)
	if err != nil {
		return 0, s.cli.GenerateError(response, fmt.Errorf("can't http request: %w", err), apiKey)
	}
	if response == nil {
		return 0, s.cli.GenerateError(response, fmt.Errorf("invalid response: %w", internalerrors.ErrResponseNil), apiKey)
	}
	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusOK:
		var resp clientsmodels.DataWrapper[clientsmodels.CreateStoragePlaceInventTaskData]
		if err = jsoniter.Unmarshal(response.Body(), &resp); err != nil {
			return 0, s.cli.GenerateError(response, fmt.Errorf("can't unmarshal response: %w", err), apiKey)
		}

		return resp.Data.CreatedQty, nil
	case http.StatusUnprocessableEntity:
		logrus.Debugf("can't process response, status code: %v err: %v", response.StatusCode(), s.cli.GenerateError(response, fmt.Errorf("api error"), apiKey))

		errWithMsg, parseErr := internalerrors.ParseErrorMessageFromResponse(response)
		if parseErr != nil {
			logrus.Errorf("[%s] can't parse response error message from response: %v", apiKey, parseErr)
		}

		return 0, errWithMsg
	case http.StatusInternalServerError:
		return 0, s.cli.GenerateError(response, internalerrors.ErrInternalServerError, apiKey)
	default:
		return 0, s.cli.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}
}
