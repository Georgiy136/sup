package clients

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	clientsmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/clients/models"
	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/sentry"
	assemblysheetterminatemodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/assembly_sheet_terminate/models"
	closewhmodelsnew "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/close_wh_new/models"
	commonhandlersmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/common_handlers/models"
	createinventtaskmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_invent_task/models"
	createofficemodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_office/models"
	createstreetsmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_streets/models"
	createwhmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_wh/models"
	transferstreetsbetweenwhmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/transfer_streets_between_wh/models"
	fastclient "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

type WhSprutenvAdministrationApiClient struct {
	cli *fastclient.HttpClient
}

func NewWhSprutenvAdministrationApiClient() *WhSprutenvAdministrationApiClient {
	return &WhSprutenvAdministrationApiClient{
		cli: fastclient.NewHttpClient(),
	}
}

func (s *WhSprutenvAdministrationApiClient) Configure(ctx context.Context, config configs.Config) {
	s.cli.InitHttpClient("wh_sprutenv_administration_api_client")(ctx, config)
}

func (s *WhSprutenvAdministrationApiClient) AddNewWh(ctx context.Context, body createwhmodels.RequestDataForAddNewWh) (whId int64, err error) {
	const apiKey = "WhSprutenvAdministrationWhAdd"

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
		var whInfo clientsmodels.DataWrapper[struct {
			WhID int64 `json:"wh_id"`
		}]
		err := jsoniter.Unmarshal(response.Body(), &whInfo)
		if err != nil {
			return 0, s.cli.GenerateError(response, fmt.Errorf("can't unmarshal response wh add: %w", err), apiKey)
		}

		return whInfo.Data.WhID, nil
	case http.StatusUnprocessableEntity:
		logrus.Debugf("can't process response, status code: %v err: %v", response.StatusCode(), s.cli.GenerateError(response, fmt.Errorf("unprocessable entity"), apiKey))

		errWithMsg, parseErr := internalerrors.ParseErrorMessageFromResponse(response)
		if parseErr != nil {
			logrus.Errorf("[%s] can't parse response error message from response: %v", apiKey, parseErr)
		}

		return 0, errWithMsg
	case http.StatusInternalServerError:
		return 0, fmt.Errorf("can't process request: %w", internalerrors.ErrInternalServerError)
	default:
		return 0, s.cli.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}
}

func (s *WhSprutenvAdministrationApiClient) CreateOfficeOnSprut(ctx context.Context, body createofficemodels.RequestForCreateOffice) (err error) {
	const apiKey = "WhSprutenvAdministrationSprutOfficeAdd"

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

func (s *WhSprutenvAdministrationApiClient) CloseWh(ctx context.Context, body closewhmodelsnew.RequestForCloseWh) (err error) {
	const apiKey = "WhSprutenvAdministrationWhDelete"

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

func (s *WhSprutenvAdministrationApiClient) GetStoragePlacesByWh(ctx context.Context, body closewhmodelsnew.RequestGetStoragePlacesByWh) (storagePlaceIDs []int64, err error) {
	const apiKey = "WhSprutenvAdministrationStoragePlacesGetByWh"

	var response *fasthttp.Response

	span := sentry.StartHTTPClientSpan(ctx, apiKey)
	defer func() {
		sentry.FinishHTTPClientSpan(span, response, err)
	}()

	response, err = s.cli.HttpRequestWithCtx(ctx, nil, http.MethodPost, body, apiKey, nil)
	if err != nil {
		logrus.Errorf("[GetStoragePlacesByWh] body for get storage places by Wh - %+v", body)
		return nil, s.cli.GenerateError(response, fmt.Errorf("can't http request: %w", err), apiKey)
	}
	if response == nil {
		return nil, s.cli.GenerateError(response, fmt.Errorf("invalid response: %w", internalerrors.ErrResponseNil), apiKey)
	}
	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusOK:
		storagePlace := struct {
			Data []int64 `json:"data"`
		}{}

		if err = jsoniter.Unmarshal(response.Body(), &storagePlace); err != nil {
			return nil, s.cli.GenerateError(response, fmt.Errorf("can't unmarshal response: %w", err), apiKey)
		}

		return storagePlace.Data, nil
	case http.StatusNoContent:
		return nil, nil
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

func (s *WhSprutenvAdministrationApiClient) DeactivateWh(ctx context.Context, body closewhmodelsnew.RequestForDeactivateWh) (err error) {
	const apiKey = "WhSprutenvAdministrationWhDeactivateV001"

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
	case http.StatusUnprocessableEntity:
		logrus.Debugf("can't process response, status code: %v", response.StatusCode())

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

func (s *WhSprutenvAdministrationApiClient) CreateStage(ctx context.Context, body commonhandlersmodels.BodyRequestForCreateStage) (err error) {
	const apiKey = "WhSprutenvAdministrationStageAdd"

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

func (s *WhSprutenvAdministrationApiClient) CreatePart(ctx context.Context, body commonhandlersmodels.BodyRequestForCreatePart) (err error) {
	const apiKey = "WhSprutenvAdministrationPartAdd"

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

func (s *WhSprutenvAdministrationApiClient) AssemblySheetTerminate(ctx context.Context, body assemblysheetterminatemodels.RequestDataForAssemblySheetTerminate) (err error) {
	const apiKey = "WhSprutenvAssemblyAdministrationAssemblySheetTerminate"

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
	case http.StatusUnprocessableEntity:
		logrus.Debugf("can't process response, status code: %v err: %v", response.StatusCode(), s.cli.GenerateError(response, fmt.Errorf("unprocessable entity"), apiKey))

		errWithMsg, parseErr := internalerrors.ParseErrorMessageFromResponse(response)
		if parseErr != nil {
			logrus.Errorf("[%s] can't parse response error message from response: %v", apiKey, parseErr)
		}

		return errWithMsg
	default:
		return s.cli.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}
}

func (s *WhSprutenvAdministrationApiClient) AssemblyStreetExclusionUpdate(ctx context.Context, body commonhandlersmodels.BodyRequestForExclusionUpdateStreet) (err error) {
	const apiKey = "WhSprutenvAssemblyAdministrationStreetExclusionUpdate"

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

func (s *WhSprutenvAdministrationApiClient) StreetExclusionFromSaleUpdateV002(ctx context.Context, body commonhandlersmodels.BodyRequestForStreetExclusionFromSaleUpdateV002) (err error) {
	const apiKey = "WhSprutenvAssemblyAdministrationStreetExclusionFromSaleUpdateV002"

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

func (s *WhSprutenvAdministrationApiClient) StageOrWhExclusionFromSaleUpdateV001(ctx context.Context, body commonhandlersmodels.RequestForStageOrWhExclusionFromSaleUpdateV001) (err error) {
	const apiKey = "WhSprutenvAssemblyAdministrationStageExclusionFromSaleUpdateV001"

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
	case http.StatusUnprocessableEntity:
		logrus.Debugf("can't process response, status code: %v err: %v", response.StatusCode(), s.cli.GenerateError(response, errors.New("unprocessable entity"), apiKey))

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

func (s *WhSprutenvAdministrationApiClient) StageOrWhExclusionFromAssemblyUpdateV001(ctx context.Context, body commonhandlersmodels.RequestForStageOrWhExclusionFromAssemblyUpdateV001) (err error) {
	const apiKey = "WhSprutenvAssemblyAdministrationStageExclusionFromAssemblyUpdateV001"

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
	case http.StatusUnprocessableEntity:
		logrus.Debugf("can't process response, status code: %v err: %v", response.StatusCode(), s.cli.GenerateError(response, errors.New("unprocessable entity"), apiKey))

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

func (s *WhSprutenvAdministrationApiClient) AddGoodsOnStockInventTask(ctx context.Context, body commonhandlersmodels.BodyRequestForAddGoodsOnStockTask) (goodsNotUpdated []int64, err error) {
	const apiKey = "InventAdministrationGoodsOnStockTaskAdd"

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
		var resp clientsmodels.DataWrapper[clientsmodels.AddGoodsOnStockTaskDataResponse]
		if err = jsoniter.Unmarshal(response.Body(), &resp); err != nil {
			return nil, s.cli.GenerateError(response, fmt.Errorf("can't unmarshal response: %w", err), apiKey)
		}

		return resp.Data.GoodsNotUpdated, nil
	case http.StatusBadRequest:
		return nil, s.cli.GenerateError(response, fmt.Errorf("invalid response"), apiKey)
	case http.StatusUnprocessableEntity:
		logrus.Debugf("can't process response, status code: %v err: %v", response.StatusCode(), s.cli.GenerateError(response, fmt.Errorf("unprocessable entity"), apiKey))

		errWithMsg, parseErr := internalerrors.ParseErrorMessageFromResponse(response)
		if parseErr != nil {
			logrus.Errorf("[%s] can't parse response error message from response: %v", apiKey, parseErr)
		}

		return nil, errWithMsg
	case http.StatusNoContent:
		return nil, nil
	case http.StatusInternalServerError:
		return nil, s.cli.GenerateError(response, internalerrors.ErrInternalServerError, apiKey)
	default:
		return nil, s.cli.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}
}

func (s *WhSprutenvAdministrationApiClient) CreateInventTaskNoWithdrawal(ctx context.Context, body createinventtaskmodels.RequestDataForCreateInventTask) (err error) {
	const apiKey = "SupportInventAdministrationInventTaskCreateInventTaskNoWithdrawal"

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

func (s *WhSprutenvAdministrationApiClient) CreateInventTaskWithWithdrawal(ctx context.Context, body createinventtaskmodels.RequestDataForCreateInventTask) (err error) {
	const apiKey = "SupportInventAdministrationInventTaskCreateInventTaskWithWithdrawal"

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

func (s *WhSprutenvAdministrationApiClient) CreateStreet(ctx context.Context, body createstreetsmodels.RequestForCreateStreet) (err error) {
	const apiKey = "WhSprutenvStageStreetUpsertInnerV001"

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

func (s *WhSprutenvAdministrationApiClient) TransferWhForStock(ctx context.Context, body transferstreetsbetweenwhmodels.RequestForTransferWhForStock) (err error) {
	const apiKey = "WhSprutenvStagePartSettingsTransferWhForStockV001"

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
