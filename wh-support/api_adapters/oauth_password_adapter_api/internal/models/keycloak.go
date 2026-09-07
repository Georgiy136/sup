package models

import (
	"fmt"
	"path"

	"github.com/AlekSi/pointer"
	"github.com/go-playground/validator"
	jsoniter "github.com/json-iterator/go"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/oauth_password_adapter_api/internal/constants"
	client "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
)

//nolint:gosec
type KeycloakConfig struct {
	ClientID        string `json:"client_id" validate:"required"`
	ClientSecret    string `json:"client_secret" validate:"required"`
	Realm           string `json:"realm" validate:"required"`
	KeycloakHost    string `json:"keycloak_host" validate:"required,url"`
	ResponseTimeout string `json:"response_timeout" validate:"required"`
	RetryCount      int    `json:"retry_count" validate:"required"`
}

func (k *KeycloakConfig) Parse(config configs.Config) error {
	ok, keycloakConfigRaw := config.GetByServiceKey(constants.ConfigKeyKeycloak)
	if !ok {
		return fmt.Errorf("config not found for key %s", constants.ConfigKeyKeycloak)
	}

	if err := jsoniter.Unmarshal(keycloakConfigRaw, k); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	if err := validator.New().Struct(k); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	return nil
}

func (k *KeycloakConfig) GetRawKeycloakClientConfig() ([]byte, error) {
	clientConfig := client.HttpClientRawConfig{
		Urls: []string{k.KeycloakHost},
		Apis: map[string]string{
			constants.ApiKeyKeycloakPassword: path.Join(
				constants.ApiPathKeycloakRealms,
				k.Realm,
				constants.ApiPathKeycloakToken,
			),
		},
		Auth: &client.Auth{
			Type: pointer.To(client.AuthTypeWithout),
			Headers: map[string]string{
				"Content-Type": "application/x-www-form-urlencoded",
			},
		},
		ResponseTimeout:             k.ResponseTimeout,
		RetryCount:                  k.RetryCount,
		PassResponseForUnauthorized: true,
	}

	rawClientConfig, err := jsoniter.Marshal(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}

	return rawClientConfig, nil
}

//nolint:gosec
type Tokens struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	NotBeforePolicy  int64  `json:"not-before-policy"`
	SessionState     string `json:"session_state"`
	Scope            string `json:"scope"`
	IDToken          string `json:"id_token"`
}
