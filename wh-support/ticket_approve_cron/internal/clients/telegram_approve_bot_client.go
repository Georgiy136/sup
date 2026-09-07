package clients

import (
	"context"
	"fmt"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_approve_cron/internal/models"
	fastclient "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	"net/http"
)

type TelegramApproveBotClient struct {
	c *fastclient.HttpClient
}

func NewTelegramApproveBotClient() *TelegramApproveBotClient {
	return &TelegramApproveBotClient{
		c: fastclient.NewHttpClient(),
	}
}

func (s *TelegramApproveBotClient) Configure(ctx context.Context, config configs.Config) {
	s.c.InitHttpClient("telegram_approve_bot_client")(ctx, config)
}

func (s *TelegramApproveBotClient) SendNotification(in models.NotificationRequest) error {
	const (
		apiKey = "send_notification"
	)

	response, err := s.c.HTTPRequest(nil, http.MethodPost, in, apiKey, nil)
	if err != nil {
		return s.c.GenerateError(response, err, apiKey)
	}

	switch response.StatusCode() {
	case http.StatusNoContent:
		return nil
	default:
		return s.c.GenerateError(response, fmt.Errorf("invalid response status %d", response.StatusCode()), apiKey)
	}
}
