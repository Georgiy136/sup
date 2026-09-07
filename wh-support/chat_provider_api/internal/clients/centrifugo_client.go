package clients

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/valyala/fasthttp"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_provider_api/internal/models"
	fastclient "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

type CentrifugoClient struct {
	client *fastclient.HttpClient
}

func NewCentrifugoClient() *CentrifugoClient {
	return &CentrifugoClient{
		client: fastclient.NewHttpClient(),
	}
}

func (c *CentrifugoClient) Configure(ctx context.Context, config configs.Config) {
	c.client.InitHttpClient("centrifugo_client")(ctx, config)
}

func (c *CentrifugoClient) Publish(ctx context.Context, channel string, data any) error {
	const apiKey = "Publish"

	payload, err := jsoniter.Marshal(models.CentrifugoPublishRequest{
		Channel: channel,
		Data:    data,
	})
	if err != nil {
		return fmt.Errorf("marshal centrifugo publish payload: %w", err)
	}

	response, err := c.client.HttpRequestWithCtx(ctx, nil, http.MethodPost, payload, apiKey, nil)
	if err != nil {
		return fmt.Errorf("centrifugo publish request failed: %w", c.client.GenerateError(response, err, apiKey))
	}
	if response == nil {
		return fmt.Errorf("centrifugo publish empty response: %w", c.client.GenerateError(response, errors.New("empty response"), apiKey))
	}
	defer fasthttp.ReleaseResponse(response)

	if response.StatusCode() != http.StatusOK {
		return fmt.Errorf("centrifugo publish unexpected status %d: %w", response.StatusCode(), c.client.GenerateError(response, fmt.Errorf("unexpected status code %d", response.StatusCode()), apiKey))
	}

	err = c.ensureNoCentrifugoError(response.Body())
	if err != nil {
		return fmt.Errorf("can't ensure centrifugo publish: %w", err)
	}

	logrus.Infof("success centrifugo publish info in %s", channel)

	return nil
}

func (c *CentrifugoClient) ensureNoCentrifugoError(rawBody []byte) error {
	var resp models.CentrifugoResponse
	if err := jsoniter.Unmarshal(rawBody, &resp); err != nil {
		return fmt.Errorf("unmarshal centrifugo response: %w", err)
	}

	logrus.Infof("centrifugo resp: %v", resp)

	if resp.Error != nil {
		return fmt.Errorf("centrifugo error code=%d message=%s", resp.Error.Code, resp.Error.Message)
	}
	if len(resp.Result) == 0 || string(resp.Result) == "null" {
		return errors.New("centrifugo response has empty result")
	}

	return nil
}
