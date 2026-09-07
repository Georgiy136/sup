package client

import (
	"context"
	"fmt"
	"github.com/valyala/fasthttp"
	app_errors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/oauth_password_adapter_api/internal/errors"
	client "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"net/http"
	"net/url"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/oauth_password_adapter_api/internal/constants"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/oauth_password_adapter_api/internal/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
)

const (
	KeycloakGrantTypePassword string = "password"
)

type Keycloak struct {
	client *client.HttpClient

	clientID     string
	clientSecret string
}

func NewKeycloak() *Keycloak {
	return &Keycloak{
		client: client.NewHttpClient(),
	}
}

func (k *Keycloak) Configure(ctx context.Context, config configs.Config) {
	keycloakConfig := models.KeycloakConfig{}

	if err := keycloakConfig.Parse(config); err != nil {
		logrus.Panicf("[Keycloak Client] failed to parse config: %s", err)
		return
	}

	rawClientConfig, err := keycloakConfig.GetRawKeycloakClientConfig()
	if err != nil {
		logrus.Panicf("[Keycloak Client] failed to get raw client config: %s", err)
		return
	}

	config.SetConfigs(map[string][]byte{constants.ConfigKeyKeycloakClient: rawClientConfig})
	k.client.InitHttpClient(constants.ConfigKeyKeycloakClient)(ctx, config)

	k.clientID = keycloakConfig.ClientID
	k.clientSecret = keycloakConfig.ClientSecret
}

func (k *Keycloak) GetTokenByPassword(ctx context.Context, username, password string) (*models.Tokens, error) {
	values := url.Values{
		"client_id":     []string{k.clientID},
		"client_secret": []string{k.clientSecret},
		"grant_type":    []string{KeycloakGrantTypePassword},
		"username":      []string{username},
		"password":      []string{password},
	}

	resp, err := k.client.HttpRequestWithCtx(ctx, nil, http.MethodPost, []byte(values.Encode()), constants.ApiKeyKeycloakPassword, nil)
	if err != nil && resp == nil {
		return nil, fmt.Errorf("make http request: %w", err)
	}
	defer func() {
		fasthttp.ReleaseResponse(resp)
	}()

	switch resp.StatusCode() {
	case http.StatusOK:
		jwtTokens := models.Tokens{}
		if err = jsoniter.Unmarshal(resp.Body(), &jwtTokens); err != nil {
			return nil, fmt.Errorf("parse response json body: %w", err)
		}
		return &jwtTokens, nil
	case http.StatusUnauthorized:
		return nil, app_errors.NewCustomError(http.StatusUnauthorized, string(resp.Body()))
	case http.StatusBadRequest:
		return nil, app_errors.NewCustomError(http.StatusBadRequest, string(resp.Body()))
	case http.StatusForbidden:
		return nil, app_errors.NewCustomError(http.StatusForbidden, string(resp.Body()))
	default:
		return nil, k.client.GenerateError(resp, err, constants.ApiKeyKeycloakPassword)
	}
}
