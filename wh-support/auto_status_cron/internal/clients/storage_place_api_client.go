package clients

import (
	"context"
	"fmt"
	"net/http"

	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/clients/models"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/sentry"
	closewhmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/close_wh_new/models"
	createstorageplacemodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_storage_place/models"
	deletestorageplacemodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/delete_storage_place/models"
	fastclient "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
)

type StoragePlaceApiClient struct {
	cli *fastclient.HttpClient
}

func NewStoragePlaceApiClient() *StoragePlaceApiClient {
	return &StoragePlaceApiClient{
		cli: fastclient.NewHttpClient(),
	}
}

func (s *StoragePlaceApiClient) Configure(ctx context.Context, config configs.Config) {
	s.cli.InitHttpClient("storage_place_api_client")(ctx, config)
}

func (s *StoragePlaceApiClient) DeleteStoragePlace(ctx context.Context, body deletestorageplacemodels.RequestDataForDeleteStoragePlace) (err error) {
	const apiKey = "StoragePlaceDelete"

	var response *fasthttp.Response

	span := sentry.StartHTTPClientSpan(ctx, apiKey)
	defer func() {
		sentry.FinishHTTPClientSpan(span, response, err)
	}()

	response, err = s.cli.HttpRequestWithCtx(ctx, nil, http.MethodPost, body, apiKey, nil)
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

func (s *StoragePlaceApiClient) CreateStoragePlace(ctx context.Context, body createstorageplacemodels.RequestDataForCreateStoragePlace) (storagePlaces []createstorageplacemodels.StoragePlace, err error) {
	const apiKey = "StoragePlaceCreate"

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
		body := createstorageplacemodels.ResponseDataFromCreateStoragePlaceApi{}
		if err = jsoniter.Unmarshal(response.Body(), &body); err != nil {
			return nil, fmt.Errorf("can't parse response body: %w", err)
		}

		return body.Data, nil
	case http.StatusUnprocessableEntity:
		logrus.Debugf("can't process response, status code: %v err: %v", response.StatusCode(), s.cli.GenerateError(response, fmt.Errorf("unprocessable entity"), apiKey))

		errWithMsg, parseErr := internalerrors.ParseErrorMessageFromResponse(response)
		if parseErr != nil {
			logrus.Errorf("[%s] can't parse response error message from response: %v", apiKey, parseErr)
		}

		return nil, errWithMsg
	case http.StatusInternalServerError:
		return nil, s.cli.GenerateError(response, internalerrors.ErrInternalServerError, apiKey)
	default:
		return nil, s.cli.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}
}

func (s *StoragePlaceApiClient) DeleteStoragePlacesFromInactiveWh(ctx context.Context, body closewhmodels.RequestDeleteStoragePlace) (result *closewhmodels.DeleteStoragePlaceData, err error) {
	const apiKey = "StoragePlaceDeleteFromInactiveWh"

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
		var result models.DataWrapper[closewhmodels.DeleteStoragePlaceData]
		if err = jsoniter.Unmarshal(response.Body(), &result); err != nil {
			return nil, s.cli.GenerateError(response, fmt.Errorf("can't unmarshal response: %w", err), apiKey)
		}
		return &result.Data, nil
	case http.StatusUnprocessableEntity:
		logrus.Debugf("can't process response, status code: %v", response.StatusCode())

		errWithMsg, parseErr := internalerrors.ParseErrorMessageFromResponse(response)
		if parseErr != nil {
			logrus.Errorf("[%s] can't parse response error message from response: %v", apiKey, parseErr)
		}

		return nil, errWithMsg
	case http.StatusInternalServerError:
		return nil, s.cli.GenerateError(response, internalerrors.ErrInternalServerError, apiKey)
	default:
		return nil, s.cli.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}
}

func (s *StoragePlaceApiClient) StoragePlaceGoodsExportTaskCreate(ctx context.Context, body closewhmodels.RequestCreateGoodsTask) (result *closewhmodels.ResponseGoodsTask, err error) {
	const apiKey = "StoragePlaceGoodsExportTaskCreate"

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
		var result closewhmodels.ResponseGoodsTask
		if err = jsoniter.Unmarshal(response.Body(), &result); err != nil {
			return nil, s.cli.GenerateError(response, fmt.Errorf("can't unmarshal response: %w", err), apiKey)
		}
		return &result, nil
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

func (s *StoragePlaceApiClient) StoragePlaceGetGoodsListByWh(ctx context.Context, body closewhmodels.RequestGetGoodsPage) (result *closewhmodels.ResponseGoodsPage, err error) {
	const apiKey = "StoragePlaceGetGoodsListByWh"

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
		var result closewhmodels.ResponseGoodsPage
		if err = jsoniter.Unmarshal(response.Body(), &result); err != nil {
			return nil, s.cli.GenerateError(response, fmt.Errorf("can't unmarshal response: %w", err), apiKey)
		}
		return &result, nil
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

func (s *StoragePlaceApiClient) StoragePlaceDeleteGoodsByWh(ctx context.Context, body closewhmodels.RequestDeleteGoodsData) (err error) {
	const apiKey = "StoragePlaceDeleteGoodsByWh"

	var response *fasthttp.Response

	span := sentry.StartHTTPClientSpan(ctx, apiKey)
	defer func() {
		sentry.FinishHTTPClientSpan(span, response, err)
	}()

	response, err = s.cli.HttpRequestWithCtx(ctx, nil, http.MethodPost, body, apiKey, nil)
	if err != nil {
		return s.cli.GenerateError(response, fmt.Errorf("can't http request: %w", err), apiKey)
	}
	if response == nil {
		return s.cli.GenerateError(response, fmt.Errorf("invalid response: %w", internalerrors.ErrResponseNil), apiKey)
	}
	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusOK, http.StatusNoContent:
		return nil
	case http.StatusUnprocessableEntity:
		logrus.Debugf("can't process response, status code: %v", response.StatusCode())

		errWithMsg, parseErr := internalerrors.ParseErrorMessageFromResponse(response)
		if parseErr != nil {
			logrus.Errorf("[%s] can't parse response error message: %v", apiKey, parseErr)
		}

		return errWithMsg
	case http.StatusInternalServerError:
		return s.cli.GenerateError(response, internalerrors.ErrInternalServerError, apiKey)
	default:
		return s.cli.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}
}
