package clients

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/sprut_access_service/internal/clients"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/sprut_access_service/internal/common"
	http_client "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	gocore_constant "gitlab.wildberries.ru/wbwh/wh-core/gocore_service_const.git/constants"

	"net/http"
)

type SprutApiClient struct {
	httpClient *http_client.HttpClient
}

func NewSprutClient() *SprutApiClient {
	return &SprutApiClient{
		httpClient: http_client.NewHttpClient(),
	}
}

func (s *SprutApiClient) Configure(ctx context.Context, cfg configs.Config) {
	s.httpClient.InitHttpClient("sprut_api_client")(ctx, cfg)
}

func (s *SprutApiClient) CallSprutAPIByUrls(ctx *gin.Context, token string, urls []string, bytesToSend []byte) ([]byte, error) {
	resp, err := s.httpClient.HTTPRequestCustomUrl(ctx, http.MethodPost, bytesToSend, nil, urls, func(request *fasthttp.Request) error {
		request.Header.Set("Authorization", token)
		request.Header.Set(gocore_constant.HeaderWBWHLanguage, ctx.GetHeader(gocore_constant.HeaderWBWHLanguage))
		return nil
	})
	if err != nil {
		return nil, s.httpClient.GenerateErrorByUrls(resp, err, urls)
	}
	if resp == nil {
		return nil, s.httpClient.GenerateErrorByUrls(resp, fmt.Errorf("response is nil"), urls)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return resp.Body(), nil
	case http.StatusNoContent:
		return nil, nil
	case http.StatusUnprocessableEntity:
		logrus.Debugf("can't process response, status code: %v err: %v", resp.StatusCode(), s.httpClient.GenerateErrorByUrls(resp, fmt.Errorf("unprocessable entity: %w", err), urls))

		customErrWrapper, parseErr := common.ParseErrorMessageFromResponse(resp)
		if parseErr != nil {
			logrus.Errorf("[sprut_api_client] can't parse response error message from response: %v", parseErr)
		}

		return nil, customErrWrapper
	case http.StatusInternalServerError:
		return nil, s.httpClient.GenerateErrorByUrls(resp, clients.ErrInternalServerError, urls)
	default:
		return nil, s.httpClient.GenerateErrorByUrls(resp, fmt.Errorf("invalid response status %d", resp.StatusCode()), urls)
	}
}
