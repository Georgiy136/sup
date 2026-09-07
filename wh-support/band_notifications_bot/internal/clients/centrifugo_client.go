package clients

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/valyala/fasthttp"
	customerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/models"
	fastclient "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	jsoniter "github.com/json-iterator/go"
)

const defaultHistoryBatchLimit int64 = 100

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

func (c *CentrifugoClient) History(ctx context.Context, channel string, since *models.StreamPosition) (*models.HistoryResult, error) {
	result, err := c.history(ctx, models.HistoryRequest{
		Channel: channel,
		Limit:   defaultHistoryBatchLimit,
		Reverse: false,
		Since:   since,
	})
	if err != nil {
		if streamErr, ok := errors.AsType[*customerrors.StreamError](err); ok && streamErr.Code == customerrors.StreamErrorCodeHistoryExpired {
			return nil, fmt.Errorf("%w: %w", customerrors.ErrHistoryExpired, err)
		}
		return nil, err
	}
	return result, nil
}

func (c *CentrifugoClient) Presence(ctx context.Context, req models.PresenceRequest) (*models.PresenceResult, error) {
	const apiKey = "Presence"

	body, err := jsoniter.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal presence body: %w", err)
	}

	resp, err := c.client.HttpRequestWithCtx(ctx, nil, http.MethodPost, body, apiKey, nil)
	if err != nil {
		return nil, c.client.GenerateError(resp, err, apiKey)
	}
	if resp == nil {
		return nil, c.client.GenerateError(resp, fmt.Errorf("empty response"), apiKey)
	}
	defer fasthttp.ReleaseResponse(resp)

	if resp.StatusCode() != http.StatusOK {
		return nil, c.client.GenerateError(resp, fmt.Errorf("unexpected status code %d", resp.StatusCode()), apiKey)
	}

	result, err := parseStreamResponse[models.PresenceResult](resp.Body())
	if err != nil {
		return nil, fmt.Errorf("can't parse presence response: %w", err)
	}

	return result, nil
}

func (c *CentrifugoClient) history(ctx context.Context, req models.HistoryRequest) (*models.HistoryResult, error) {
	const apiKey = "History"

	body, err := jsoniter.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal history body: %w", err)
	}

	resp, err := c.client.HttpRequestWithCtx(ctx, nil, http.MethodPost, body, apiKey, nil)
	if err != nil {
		return nil, c.client.GenerateError(resp, err, apiKey)
	}
	if resp == nil {
		return nil, c.client.GenerateError(resp, fmt.Errorf("empty response"), apiKey)
	}
	defer fasthttp.ReleaseResponse(resp)

	if resp.StatusCode() != http.StatusOK {
		return nil, c.client.GenerateError(resp, fmt.Errorf("unexpected status code %d", resp.StatusCode()), apiKey)
	}

	result, err := parseStreamResponse[models.HistoryResult](resp.Body())
	if err != nil {
		return nil, fmt.Errorf("can't parse history response: %w", err)
	}

	return result, nil
}

func parseStreamResponse[T any](rawBody []byte) (*T, error) {
	var resp models.StreamResponse[T]
	if err := jsoniter.Unmarshal(rawBody, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal stream response: %w", err)
	}
	if resp.Error != nil {
		return nil, customerrors.NewStreamError(resp.Error.Code, resp.Error.Message)
	}
	if resp.Result == nil {
		return nil, fmt.Errorf("stream response has empty result")
	}

	return resp.Result, nil
}
