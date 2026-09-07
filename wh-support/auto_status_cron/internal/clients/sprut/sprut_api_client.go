package sprut

import (
	"context"
	"fmt"
	"net/http"

	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/sentry"
	client "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"

	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

type SprutApiClient struct {
	cli *client.HttpClient
}

func NewSprutClient() *SprutApiClient {
	return &SprutApiClient{
		cli: client.NewHttpClient(),
	}
}

func (s *SprutApiClient) Configure(ctx context.Context, config configs.Config) {
	s.cli.InitHttpClient("sprut_api_client")(ctx, config)
}

func (s *SprutApiClient) CallSprutAPIByUrls(ctx context.Context, token string, urls []string, bytesToSend []byte) (response *fasthttp.Response, err error) {
	var sprutApiClientName = "SprutApiClient"

	span := sentry.StartHTTPClientSpan(ctx, sprutApiClientName)
	defer func() {
		sentry.FinishHTTPClientSpan(span, response, err)
	}()

	response, err = s.cli.HTTPRequestCustomUrl(nil, http.MethodPost, bytesToSend, nil, urls, func(request *fasthttp.Request) error {
		request.Header.Set("Authorization", token)
		return nil
	})
	if err != nil {
		return nil, s.cli.GenerateErrorByUrls(response, err, urls)
	}
	if response == nil {
		return nil, s.cli.GenerateErrorByUrls(response, fmt.Errorf("invalid response: %w", internalerrors.ErrResponseNil), urls)
	}
	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusOK:
		return response, nil
	case http.StatusNoContent:
		return nil, nil
	case http.StatusUnprocessableEntity:
		logrus.Debugf("can't process response, status code: %v err: %v", response.StatusCode(), s.cli.GenerateErrorByUrls(response, fmt.Errorf("unprocessable entity: %w", err), urls))

		return response, fmt.Errorf("[%s] can't process request: %w", sprutApiClientName, internalerrors.ErrUnprocessableEntity)
	case http.StatusInternalServerError:
		return response, fmt.Errorf("[%s] can't process request: %w", sprutApiClientName, internalerrors.ErrInternalServerError)
	default:
		return nil, s.cli.GenerateErrorByUrls(response, fmt.Errorf("[%s] invalid response status %d", sprutApiClientName, response.StatusCode()), urls)
	}
}
