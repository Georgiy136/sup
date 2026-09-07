package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_auth_service/internal/models"
	utils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git"
)

// возвращает ошибку в формате Centrifugo https://centrifugal.dev/docs/server/proxy#error
func BindCentrifugoError(ctx *gin.Context, httpCode int, message string) {
	utils.BindNoContent(ctx)

	response := models.CentrifugoErrorResponse{
		Error: models.CentrifugoError{
			Code:    httpCode,
			Message: message,
		},
	}
	// Centrifugo proxy всегда ожидает HTTP 200
	ctx.JSON(http.StatusOK, response)
}

// возвращает успешный ответ для Centrifugo proxy https://centrifugal.dev/docs/server/proxy
func BindCentrifugoSuccess(ctx *gin.Context, result interface{}) {
	utils.BindNoContent(ctx)

	ctx.JSON(http.StatusOK, gin.H{
		"result": result,
	})
}
