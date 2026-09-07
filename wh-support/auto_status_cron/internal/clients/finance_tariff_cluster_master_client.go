package clients

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/sentry"
	createoperationmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_operations/models"
	fastclient "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

const (
	FinanceTariffClusterMasterClientName = "FinanceTariffClusterMasterClient"
)

type FinanceTariffClusterMasterClient struct {
	cli *fastclient.HttpClient
}

func NewFinanceTariffClusterMasterClient() *FinanceTariffClusterMasterClient {
	return &FinanceTariffClusterMasterClient{
		cli: fastclient.NewHttpClient(),
	}
}

func (f *FinanceTariffClusterMasterClient) Configure(ctx context.Context, config configs.Config) {
	f.cli.InitHttpClient("finance_tariff_cluster_master_client")(ctx, config)
}

func (f *FinanceTariffClusterMasterClient) ProdtypesAddNew(ctx context.Context, body createoperationmodels.RequestDataForProdtypesAddNew, employeeID int64) (err error) {
	const (
		apiKey = "ProdtypesAddNew"
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

	response, err = f.cli.HttpRequestWithCtx(ctx, nil, http.MethodPost, body, apiKey, params)
	if err != nil {
		return f.cli.GenerateError(response, err, apiKey)
	}
	if response == nil {
		return f.cli.GenerateError(response, fmt.Errorf("invalid response: %w", internalerrors.ErrResponseNil), apiKey)
	}
	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusOK:
		return nil
	case http.StatusNoContent:
		return nil
	case http.StatusUnprocessableEntity:
		logrus.Debugf("can't process response, status code: %v err: %v", response.StatusCode(), f.cli.GenerateError(response, fmt.Errorf("unprocessable entity: %w", err), apiKey))

		errWithMsg, parseErr := internalerrors.ParseErrorMessageFromResponse(response)
		if parseErr != nil {
			logrus.Errorf("[%s] can't parse response error message: %v", FinanceTariffClusterMasterClientName, parseErr)
		}

		return errWithMsg
	case http.StatusInternalServerError:
		return fmt.Errorf("[%s] can't process request: %w", FinanceTariffClusterMasterClientName, internalerrors.ErrInternalServerError)
	default:
		return f.cli.GenerateError(response, fmt.Errorf("[%s] invalid response status %d", FinanceTariffClusterMasterClientName, response.StatusCode()), apiKey)
	}
}
