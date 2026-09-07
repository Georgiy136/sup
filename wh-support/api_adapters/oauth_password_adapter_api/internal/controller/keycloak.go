package controller

import (
	"context"
	"errors"
	"fmt"
	app_errors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/oauth_password_adapter_api/internal/errors"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors/errors_keys"
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/api_adapters/oauth_password_adapter_api/internal/models"
	utils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git"
	core_errors "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors"
)

type keycloakService interface {
	AuthByPassword(ctx context.Context, username, password string, employeeID int64) (*models.Tokens, error)
}

type Keycloak struct {
	keycloakService keycloakService
	errBuilder      core_errors.ErrorBuilder
}

func NewKeycloak(keycloakService keycloakService) *Keycloak {
	return &Keycloak{
		keycloakService: keycloakService,
		errBuilder:      core_errors.NewErrorBuilder(errors_keys.NewErrorsRepository().GetErrors(), nil),
	}
}

func (k *Keycloak) AuthKeycloakByPassword(ctx *gin.Context, params map[string]any) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		k.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	login, password, ok := ctx.Request.BasicAuth()
	if !ok {
		k.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("basic auth not provided"))
		return
	}

	tokens, err := k.keycloakService.AuthByPassword(ctx.Request.Context(), login, password, employeeID)
	if err != nil {
		var keycloakErr *app_errors.CustomError
		if errors.As(err, &keycloakErr) {
			switch keycloakErr.StatusCode {
			case http.StatusUnauthorized:
				k.errBuilder.BindError(ctx, errors_keys.AuthErrCritTokenFailure, err)
				return
			case http.StatusBadRequest:
				k.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, err)
				return
			case http.StatusForbidden:
				k.errBuilder.BindError(ctx, errors_keys.ErrBizProcessForbidden, err)
				return
			}
		}
		k.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyServiceError), fmt.Errorf("failed to auth by password: %w", err))
		return
	}

	utils.BindObjectToRestData(ctx, models.TokensResponse{
		AccessToken:      tokens.AccessToken,
		ExpiresIn:        tokens.ExpiresIn,
		RefreshExpiresIn: tokens.RefreshExpiresIn,
		RefreshToken:     tokens.RefreshToken,
		IDToken:          tokens.IDToken,
	})
}
