package main

import (
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/handlers"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/services"
	rest_auto_api "gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git"

	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository/connection_control"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs/env_config"

	"github.com/go-playground/validator/v10"
)

func main() {
	bot := services.NewBot()
	runner := rest_auto_api.NewService()
	runner.DisableJWTTokenAuth()
	runner.DisableHmacAuth()
	runner.RegisterConfigurableEntity(connection_control.InitConnections(env_config.GetCommonEnvConfigs()))
	runner.RegisterConfigurableEntity(bot.Init)

	valid := validator.New()
	handlers := handlers.New(bot, valid)
	runner.RegisterConfigurableEntity(handlers.Configure)
	runner.RegisterCustomHandler("SendNotification", handlers.SendNotificationHandler)

	runner.Run()
}
