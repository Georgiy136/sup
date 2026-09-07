package service

import (
	"context"
	"time"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_auth_service/internal/jwt"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_auth_service/internal/models"
)

type redisClientInterface interface {
	GetAccessPolicies(ctx context.Context, employeeID int64, typeActions ...string) (*models.AccessPoliciesResult, error)
}

type resourceEmployeeAccessClientInterface interface {
	GetAccessActionsByEmployeeID(employeeID int64) ([]string, error)
}

type jwtTokenGeneratorInterface interface {
	GenerateToken(employeeID int64, ttl time.Duration) (string, error)
	ParseToken(tokenString string) (*jwt.CentrifugoTokenClaims, error)
}
