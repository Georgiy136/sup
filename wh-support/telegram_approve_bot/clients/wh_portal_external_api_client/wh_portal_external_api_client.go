package clients

import (
	"context"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/models"
	fastclient "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	"net/http"
	"strconv"
)

type WhPortalExternalApiClient struct {
	cli *fastclient.HttpClient
}

func NewWhPortalExternalApiClient(ctx context.Context, config configs.Config) *WhPortalExternalApiClient {
	client := &WhPortalExternalApiClient{
		cli: fastclient.NewHttpClient(),
	}

	client.cli.InitHttpClient("wh_portal_external_api_client")(ctx, config)

	return client
}

func (apiClient *WhPortalExternalApiClient) CheckTelegramOnWhPortal(telegramID int64) (*models.CheckTelegramInfo, error) {
	const (
		apiKey = "check_telegram"
		base   = 10
	)
	params := map[string]string{
		"tg_id": strconv.FormatInt(telegramID, base),
	}

	response, err := apiClient.cli.HTTPRequest(nil, http.MethodGet, nil, apiKey, params)
	if err != nil {
		return nil, fmt.Errorf("[CheckTelegramOnWhPortal] error check telegram on wh portal, error: %w", err)
	}

	if response == nil {
		return nil, nil
	}

	body := struct {
		Data models.CheckTelegramInfo `json:"data"`
	}{}

	switch response.StatusCode() {
	case http.StatusNoContent:
		return nil, nil
	case http.StatusOK:
		if err = jsoniter.Unmarshal(response.Body(), &body); err != nil {
			return nil, fmt.Errorf("invalid response body: %w", err)
		}
	default:
		return nil, fmt.Errorf("[CheckTelegramOnWhPortal] unexpected status code, error: %w", err)
	}

	return &body.Data, nil
}
