package clients

import (
	"context"
	"fmt"
	"net/http"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/support_file_manager_api/internal/models"
	fastclient "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
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

func (c *CentrifugoClient) PublishUploadStatus(ctx context.Context, uploadID int64, isSuccess bool) error {
	channelName := fmt.Sprintf("upload:%d", uploadID)
	req := models.BodyByUploadIDRequest{
		IsSuccess: isSuccess,
		Event:     "upload.status.updated",
		Meta: models.Meta{
			PublishedAt: time.Now(),
		},
	}

	return c.publish(ctx, channelName, req)
}

func (c *CentrifugoClient) publish(ctx context.Context, channelName string, data any) error {
	const apiKey = "Publish"

	payload, err := jsoniter.Marshal(models.CentrifugoPublishRequest{
		Channel: channelName,
		Data:    data,
	})
	if err != nil {
		return fmt.Errorf("error marshalling centrifugo publish payload: %w", err)
	}

	response, err := c.client.HttpRequestWithCtx(ctx, nil, http.MethodPost, payload, apiKey, nil)
	if err != nil {
		return c.client.GenerateError(response, err, apiKey)
	}

	if response == nil {
		return c.client.GenerateError(response, fmt.Errorf("empty response"), apiKey)
	}

	defer fasthttp.ReleaseResponse(response)

	if response.StatusCode() != http.StatusOK {
		return c.client.GenerateError(response, fmt.Errorf("unexpected status code %d", response.StatusCode()), apiKey)
	}

	return c.ensureNoCentrifugoError(response.Body())
}

func (c *CentrifugoClient) ensureNoCentrifugoError(rawBody []byte) error {
	var resp models.CentrifugoResponse
	if err := jsoniter.Unmarshal(rawBody, &resp); err != nil {
		return fmt.Errorf("error unmarshal centrifugo response: %w", err)
	}

	if resp.Error != nil {
		return fmt.Errorf("centrifugo error code=%d message=%s", resp.Error.Code, resp.Error.Message)
	}
	if len(resp.Result) == 0 || string(resp.Result) == "null" {
		return fmt.Errorf("centrifugo response has empty result")
	}

	return nil
}
