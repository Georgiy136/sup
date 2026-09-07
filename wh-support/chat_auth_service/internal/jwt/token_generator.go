package jwt

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type CentrifugoTokenClaims struct {
	jwt.RegisteredClaims
}

type JwtTokenGenerator struct {
	secretKey string
}

func NewJwtTokenGenerator() *JwtTokenGenerator {
	return &JwtTokenGenerator{}
}

func (g *JwtTokenGenerator) Configure(ctx context.Context, config configs.Config) {
	if g.secretKey = string(config.GetByServiceKeyRequired("centrifugo_jwt_secret_key")); g.secretKey == "" {
		logrus.Panic("centrifugo_jwt_secret_key is empty")
	}
}

func (g *JwtTokenGenerator) GenerateToken(employeeID int64, ttl time.Duration) (string, error) {
	now := time.Now()
	employeeIDStr := strconv.FormatInt(employeeID, 10)

	claims := CentrifugoTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   employeeIDStr,
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(g.secretKey))
}

func (g *JwtTokenGenerator) ParseToken(tokenString string) (*CentrifugoTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CentrifugoTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(g.secretKey), nil
	})

	if err != nil {
		if !errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("failed to parse token: %w", err)
		}
	}

	claims, ok := token.Claims.(*CentrifugoTokenClaims)
	if !ok {
		return nil, errors.New("failed to extract claims")
	}

	return claims, nil
}

func (c *CentrifugoTokenClaims) GetEmployeeID() (int64, error) {
	if c.Subject == "" {
		return 0, errors.New("employee_id not found in token")
	}

	employeeID, err := strconv.ParseInt(c.Subject, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse employee_id: %w", err)
	}

	return employeeID, nil
}

func (c *CentrifugoTokenClaims) IsExpired() bool {
	if c.ExpiresAt == nil {
		return true
	}
	return time.Now().After(c.ExpiresAt.Time)
}
