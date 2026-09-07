package clients

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/valyala/fasthttp"
	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/sentry"
	additionaltransportmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/additional_transport/models"
	fastclient "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

const (
	CargoAggregatorLogisticClientName = "CargoAggregatorLogisticClient"
)

type responseCreateLogisticTicket struct {
	Cargo struct {
		CargoId string `json:"cargo_id"`
	} `json:"cargo"`
}

type CargoAggregatorLogisticClient struct {
	cli *fastclient.HttpClient
}

func NewCargoAggregatorLogisticClient() *CargoAggregatorLogisticClient {
	return &CargoAggregatorLogisticClient{
		cli: fastclient.NewHttpClient(),
	}
}

func (f *CargoAggregatorLogisticClient) Configure(ctx context.Context, config configs.Config) {
	f.cli.InitHttpClient("cargo_aggregator_logistic_client")(ctx, config)
}

func (f *CargoAggregatorLogisticClient) CreateLogisticTicket(ctx context.Context, req additionaltransportmodels.RequestForCreateLogisticCargo) (cargoID string, err error) {
	const apiKey = "CargoCreateAPI"

	var response *fasthttp.Response

	span := sentry.StartHTTPClientSpan(ctx, apiKey)
	defer func() {
		sentry.FinishHTTPClientSpan(span, response, err)
	}()

	response, err = f.cli.HttpRequestWithCtx(ctx, nil, http.MethodPost, req, apiKey, nil)
	if err != nil {
		return "", f.cli.GenerateError(response, err, apiKey)
	}
	if response == nil {
		return "", f.cli.GenerateError(response, fmt.Errorf("invalid response: %w", internalerrors.ErrResponseNil), apiKey)
	}
	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusOK:
		var cargo responseCreateLogisticTicket

		if err = jsoniter.Unmarshal(response.Body(), &cargo); err != nil {
			return "", f.cli.GenerateError(response, fmt.Errorf("can't unmarshal response cargo: %w", err), apiKey)
		}

		return cargo.Cargo.CargoId, nil
	case http.StatusUnprocessableEntity:
		logrus.Errorf("can't process response, status code: %v err: %v", response.StatusCode(), f.cli.GenerateError(response, fmt.Errorf("unprocessable entity: %w", err), apiKey))

		errWithMsg, parseErr := internalerrors.ParseExternalErrorMessageFromResponse(response)
		if parseErr != nil {
			logrus.Errorf("[%s] can't parse response error message: %v", CargoAggregatorLogisticClientName, parseErr)
		}

		return "", errWithMsg
	case http.StatusInternalServerError:
		return "", fmt.Errorf("[%s] can't process request: %w", CargoAggregatorLogisticClientName, internalerrors.ErrInternalServerError)
	default:
		return "", f.cli.GenerateError(response, fmt.Errorf("[%s] invalid response status %d", CargoAggregatorLogisticClientName, response.StatusCode()), apiKey)
	}
}

func (f *CargoAggregatorLogisticClient) GetStatusLogisticTicketsBatched(ctx context.Context, req additionaltransportmodels.RequestForGetStatusLogisticTickets) (additionaltransportmodels.ResponseGetStatusLogisticTickets, error) {
	const batchSize = 100
	const requestDelay = 2 * time.Second

	var result additionaltransportmodels.ResponseGetStatusLogisticTickets

	for from := 0; from < len(req.CargoIds); from += batchSize {
		to := min(from+batchSize, len(req.CargoIds))

		batchReq := additionaltransportmodels.RequestForGetStatusLogisticTickets{
			CargoIds: req.CargoIds[from:to],
		}

		batchResp, err := f.getStatusLogisticTickets(ctx, batchReq)
		if err != nil {
			return result, fmt.Errorf("can't get status logistic tickets: %w", err)
		}

		result.Cargos = append(result.Cargos, batchResp.Cargos...)

		if to < len(req.CargoIds) {
			time.Sleep(requestDelay)
		}
	}

	return result, nil
}

func (f *CargoAggregatorLogisticClient) getStatusLogisticTickets(ctx context.Context, req additionaltransportmodels.RequestForGetStatusLogisticTickets) (result additionaltransportmodels.ResponseGetStatusLogisticTickets, err error) {
	const apiKey = "CargoGetStatusAPI"

	var response *fasthttp.Response

	span := sentry.StartHTTPClientSpan(ctx, apiKey)
	defer func() {
		sentry.FinishHTTPClientSpan(span, response, err)
	}()
	var cargosInfo additionaltransportmodels.ResponseGetStatusLogisticTickets

	response, err = f.cli.HttpRequestWithCtx(ctx, nil, http.MethodPost, req, apiKey, nil)
	if err != nil {
		return cargosInfo, f.cli.GenerateError(response, err, apiKey)
	}
	if response == nil {
		return cargosInfo, f.cli.GenerateError(response, fmt.Errorf("invalid response: %w", internalerrors.ErrResponseNil), apiKey)
	}
	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusOK:
		if err = jsoniter.Unmarshal(response.Body(), &cargosInfo); err != nil {
			return cargosInfo, f.cli.GenerateError(response, fmt.Errorf("can't unmarshal response cargo: %w", err), apiKey)
		}

		return cargosInfo, nil
	case http.StatusUnprocessableEntity:
		logrus.Errorf("can't process response, status code: %v err: %v", response.StatusCode(), f.cli.GenerateError(response, fmt.Errorf("unprocessable entity: %w", err), apiKey))

		errWithMsg, parseErr := internalerrors.ParseExternalErrorMessageFromResponse(response)
		if parseErr != nil {
			logrus.Errorf("[%s] can't parse response error message: %v", CargoAggregatorLogisticClientName, parseErr)
		}

		return cargosInfo, errWithMsg
	case http.StatusInternalServerError:
		return cargosInfo, fmt.Errorf("[%s] can't process request: %w", CargoAggregatorLogisticClientName, internalerrors.ErrInternalServerError)
	default:
		return cargosInfo, f.cli.GenerateError(response, fmt.Errorf("[%s] invalid response status %d", CargoAggregatorLogisticClientName, response.StatusCode()), apiKey)
	}
}

func (f *CargoAggregatorLogisticClient) RejectLogisticTicket(ctx context.Context, req additionaltransportmodels.RequestRejectLogisticTicket) (err error) {
	const (
		apiBase   = "/api/cargo/v1/wh-support/v1/cargos/"
		apiAction = "/cancel"
	)

	var response *fasthttp.Response

	span := sentry.StartHTTPClientSpan(ctx, apiBase)
	defer func() {
		sentry.FinishHTTPClientSpan(span, response, err)
	}()

	api, err := url.JoinPath(apiBase, req.CargoId, apiAction)
	if err != nil {
		return fmt.Errorf("can't create request reject logistic ticket: %w", err)
	}

	response, err = f.cli.HttpRequestWithCtx(ctx, nil, http.MethodPost, req, api, nil)
	if err != nil {
		return f.cli.GenerateError(response, err, api)
	}
	if response == nil {
		return f.cli.GenerateError(response, fmt.Errorf("invalid response: %w", internalerrors.ErrResponseNil), api)
	}
	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusOK, http.StatusNoContent:
		return nil
	case http.StatusUnprocessableEntity:
		logrus.Errorf("can't process response, status code: %v err: %v", response.StatusCode(), f.cli.GenerateError(response, fmt.Errorf("unprocessable entity: %w", err), api))

		errWithMsg, parseErr := internalerrors.ParseExternalErrorMessageFromResponse(response)
		if parseErr != nil {
			logrus.Errorf("[%s] can't parse response error message: %v", CargoAggregatorLogisticClientName, parseErr)
		}

		return errWithMsg
	case http.StatusInternalServerError:
		return fmt.Errorf("[%s] can't process request: %w", CargoAggregatorLogisticClientName, internalerrors.ErrInternalServerError)
	default:
		return f.cli.GenerateError(response, fmt.Errorf("[%s] invalid response status %d", CargoAggregatorLogisticClientName, response.StatusCode()), api)
	}
}
