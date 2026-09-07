package main

import (
	"context"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/file_validation_cron/internal/clients"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/file_validation_cron/internal/cron"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/file_validation_cron/internal/services"

	"gitlab.wildberries.ru/wbwh/wh-core/gocore_cron_core.git/cron_core"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository/connection_control"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs/env_config"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_core.git/app"
)

func main() {
	conf := env_config.GetCommonEnvConfigs()
	serviceWorker := app.NewServerScript()
	serviceWorker.InitConfigManagerDefault()

	storageClient := clients.NewStorageClient()
	supportFileManagerApiClient := clients.NewSupportFileManagerApiClient()

	serviceWorker.Configuration(
		connection_control.InitConnections(conf),
		storageClient.Configure,
		supportFileManagerApiClient.Configure,
	)

	fileValidationService := services.NewFileValidationService(storageClient, supportFileManagerApiClient)
	expiredUploadsDeletionService := services.NewExpiredUploadsDeletionService(storageClient, supportFileManagerApiClient)
	fileDeletionService := services.NewFileDeletionService(storageClient, supportFileManagerApiClient)
	fileTransferService := services.NewFileTransferService(storageClient, supportFileManagerApiClient)
	preparingToDeleteNotAttachedToTicketFile := services.NewPreparingToDeleteNotAttachedToTicketFile(storageClient, supportFileManagerApiClient)

	serviceWorker.Tasks(func(ctx context.Context, conf configs.Config) {
		worker := cron_core.HandlerRegistration{Workers: map[string]cron_core.Worker{
			"file_validation_cron":                            cron.NewCron(fileValidationService),
			"expired_uploads_deletion_cron":                   cron.NewCron(expiredUploadsDeletionService),
			"file_deletion_cron":                              cron.NewCron(fileDeletionService),
			"file_transfer_cron":                              cron.NewCron(fileTransferService),
			"preparing_to_delete_not_attached_to_ticket_file": cron.NewCron(preparingToDeleteNotAttachedToTicketFile),
		}}
		worker.InitCrons(ctx, conf)
	})

	app.StartServer(serviceWorker)
}
