package clients

import (
	"context"
	"fmt"
	"net/http"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/wbstorenv_api_adapter/internal/common"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/wbstorenv_api_adapter/internal/models"
	client "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
)

type LocalizationApiClient struct {
	httpClient *client.HttpClient
}

const (
	localizationConfKey = "fasthttp_localization_client"
)

func NewLocalizationApiClient(ctx context.Context, config configs.Config) *LocalizationApiClient {
	apiClient := LocalizationApiClient{
		httpClient: client.NewHttpClient(),
	}
	apiClient.httpClient.InitHttpClient(localizationConfKey)(ctx, config)
	return &apiClient
}

func (l *LocalizationApiClient) GetTranslation(ctx *gin.Context, lang string, keysWithValues models.KeysWithValues) (string, error) {
	const apiKey = "api/translation_with_fmt_get_by_key"

	requestBody := models.TranslationGetByKeyRequest{
		Lang:           lang,
		KeysWithValues: []models.KeysWithValues{keysWithValues},
	}

	response, err := l.httpClient.HTTPRequest(ctx, http.MethodPost, &requestBody, apiKey, nil)
	if err != nil {
		return "", l.httpClient.GenerateError(response, fmt.Errorf("can't make http request: %w", err), apiKey)
	}
	if response == nil {
		return "", l.httpClient.GenerateError(response, fmt.Errorf("invalid response: %w", common.ErrResponseNil), apiKey)
	}

	defer fasthttp.ReleaseResponse(response)

	switch response.StatusCode() {
	case http.StatusOK:
		var respData models.TranslationGetByKeyResponse
		if err = jsoniter.Unmarshal(response.Body(), &respData); err != nil {
			return "", l.httpClient.GenerateError(response, fmt.Errorf("can't unmarshal response data: %w", err), apiKey)
		}

		if len(respData.Data) == 0 {
			return "", common.ErrTranslationNotFound
		}

		return respData.Data[0].LocalizationMessage, nil
	default:
		return "", l.httpClient.GenerateError(response, fmt.Errorf("unexpected status code %d: %w", response.StatusCode(), common.ErrWrongStatusCode), apiKey)
	}
}
