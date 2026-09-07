package main

import (
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/support_file_manager_api/internal/clients"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/support_file_manager_api/internal/service"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/support_file_manager_api/internal/storage/postgres"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository/connection_control"
	rest_auto_api "gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs/env_config"

	"gitlab.wildberries.ru/wbwh/support/utils.git/middleware"
)

func main() {
	conf := env_config.GetCommonEnvConfigs()
	runner := rest_auto_api.NewService()

	loggerToClickhouse := middleware.NewClickhouseLoggerMiddleware()

	runner.RegisterConfigurableEntity(
		connection_control.InitConnections(conf),
	)

	runner.DisableJWTTokenAuth()
	runner.DisableHmacAuth()

	s3Client := clients.NewS3Client()
	centrifugoClient := clients.NewCentrifugoClient()
	fileValidator := service.NewFileValidator()
	repository := postgres.NewRepository()
	srvFileManager := service.NewSupportFileManagerService(s3Client, centrifugoClient, fileValidator, repository)

	runner.RegisterConfigurableEntity(s3Client.Configure)
	runner.RegisterConfigurableEntity(centrifugoClient.Configure)
	runner.RegisterConfigurableEntity(fileValidator.Configure)
	runner.RegisterConfigurableEntity(srvFileManager.Configure)

	runner.RegisterCustomHandler("InitUpload", srvFileManager.InitUpload)
	runner.RegisterCustomHandler("ConfirmUpload", srvFileManager.ConfirmUpload)
	runner.RegisterCustomHandler("UpdateUploadStatus", srvFileManager.UpdateUploadStatus)
	runner.RegisterCustomHandler("CancelUpload", srvFileManager.CancelUpload)
	runner.RegisterCustomHandler("GetUploadStatus", srvFileManager.GetUploadStatus)
	runner.RegisterCustomHandler("GetUploadsForValidation", srvFileManager.GetUploadsForValidation)
	runner.RegisterCustomHandler("GetExpiredUploads", srvFileManager.GetExpiredUploads)
	runner.RegisterCustomHandler("DeleteFile", srvFileManager.DeleteFile)
	runner.RegisterCustomHandler("AttachFilesToTicket", srvFileManager.AttachFilesToTicket)
	runner.RegisterCustomHandler("GetDownloadURL", srvFileManager.GetDownloadURL)
	runner.RegisterCustomHandler("GetFilesForDeletion", srvFileManager.GetFilesForDeletion)
	runner.RegisterCustomHandler("GetFilesNotAttachedToTicketForDeleting", srvFileManager.GetFilesNotAttachedToTicketForDeleting)
	runner.RegisterCustomHandler("MarkFileDeleted", srvFileManager.MarkFileDeleted)
	runner.RegisterCustomHandler("GetUploadsForTransfer", srvFileManager.GetUploadsForTransfer)
	runner.RegisterCustomHandler("TransferFile", srvFileManager.TransferFile)
	runner.RegisterCustomHandler("GetFileInfo", srvFileManager.GetFileInfo)
	runner.RegisterCustomHandler("InternalUploadFile", srvFileManager.InternalUploadFile)

	runner.Run(loggerToClickhouse...)
}
