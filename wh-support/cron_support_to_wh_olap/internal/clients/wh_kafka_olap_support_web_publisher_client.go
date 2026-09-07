package clients

import (
	"context"
	"fmt"
	"net/http"

	client "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
)

type WhKafkaOlapWebPublisherClient struct {
	client *client.HttpClient
}

func NewWhKafkaOlapWebPublisherClient(ctx context.Context, config configs.Config) *WhKafkaOlapWebPublisherClient {
	httpClient := client.NewHttpClient()
	httpClient.InitHttpClient("wh_kafka_olap_support_web_publisher_client")(ctx, config)

	return &WhKafkaOlapWebPublisherClient{
		client: httpClient,
	}
}

func (w *WhKafkaOlapWebPublisherClient) SendData(apiKey string, body any) error {
	resp, err := w.client.HTTPRequest(nil, http.MethodPost, body, apiKey, nil)
	if err != nil {
		return fmt.Errorf("failed to send data: %w", err)
	}

	if resp.StatusCode() == http.StatusNoContent {
		return nil
	}

	return w.client.GenerateError(resp, fmt.Errorf("unexpected response status code: %d, err: %v", resp.StatusCode(), err), apiKey)
}
