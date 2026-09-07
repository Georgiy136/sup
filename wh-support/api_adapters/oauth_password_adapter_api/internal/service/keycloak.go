package service

import (
	"context"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	app_errors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/oauth_password_adapter_api/internal/errors"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	"net/http"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/oauth_password_adapter_api/internal/models"
)

type keycloakClient interface {
	GetTokenByPassword(ctx context.Context, username, password string) (*models.Tokens, error)
}

type Keycloak struct {
	keycloakClient keycloakClient
	entries        map[string]int64
}

func NewKeycloak(keycloakClient keycloakClient) *Keycloak {
	return &Keycloak{
		keycloakClient: keycloakClient,
	}
}

func (k *Keycloak) Configure(ctx context.Context, config configs.Config) {
	if err := jsoniter.Unmarshal(config.GetByServiceKeyRequired("entries"), &k.entries); err != nil {
		logrus.Panicf("error parsing entries configs - %v", err)
	}
}

func (k *Keycloak) AuthByPassword(ctx context.Context, username, password string, employeeID int64) (*models.Tokens, error) {
	if foundEmployeeID, ok := k.entries[username]; !ok || employeeID != foundEmployeeID {
		return nil, app_errors.NewCustomError(http.StatusUnauthorized, "employee not found")
	}

	tokens, err := k.keycloakClient.GetTokenByPassword(ctx, username, password)
	if err != nil {
		return nil, fmt.Errorf("get token by password error: %w", err)
	}

	return tokens, nil
}
