package jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJWT(t *testing.T) {
	secretKey := "test-secret-key"
	employeeID := int64(14545)

	t.Run("генерация и парсинг токена", func(t *testing.T) {
		gen := NewJwtTokenGenerator()
		gen.secretKey = secretKey

		token, err := gen.GenerateToken(employeeID, time.Hour)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		claims, err := gen.ParseToken(token)
		assert.NoError(t, err)
		assert.NotNil(t, claims)

		parsedEmployeeID, err := claims.GetEmployeeID()
		assert.NoError(t, err)
		assert.Equal(t, employeeID, parsedEmployeeID)
		assert.False(t, claims.IsExpired())
	})

	t.Run("парсинг истёкшего токена", func(t *testing.T) {
		gen := NewJwtTokenGenerator()
		gen.secretKey = secretKey

		token, err := gen.GenerateToken(employeeID, -time.Second)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		claims, err := gen.ParseToken(token)
		assert.NoError(t, err)
		assert.NotNil(t, claims)

		parsedEmployeeID, err := claims.GetEmployeeID()
		assert.NoError(t, err)
		assert.Equal(t, employeeID, parsedEmployeeID)
		assert.True(t, claims.IsExpired())
	})
}
