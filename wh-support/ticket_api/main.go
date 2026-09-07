package main

import (
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/clients"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/service"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/service/chat_mapper"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository/connection_control"
	rest_auto_api "gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs/env_config"

	"gitlab.wildberries.ru/wbwh/support/utils.git/body_size_limiter"
	"gitlab.wildberries.ru/wbwh/support/utils.git/middleware"
)

func main() {
	conf := env_config.GetCommonEnvConfigs()
	runner := rest_auto_api.NewService()

	loggerToClickhouse := middleware.NewClickhouseLoggerMiddleware()
	reqSizeLimiterOption := body_size_limiter.NewDefaultReqSizeLimiterMiddleware()

	runner.RegisterConfigurableEntity(
		connection_control.InitConnections(conf),
	)

	runner.DisableJWTTokenAuth()
	runner.DisableHmacAuth()
	runner.EnableAuthActionsCheck()

	redisClient := clients.NewRedisClient()
	resourceEmployeeAccessClient := clients.NewResourceEmployeeAccessClient()
	fileManagerClient := clients.NewSupportFileManagerApiClient()

	runner.RegisterConfigurableEntity(redisClient.Configure)
	runner.RegisterConfigurableEntity(resourceEmployeeAccessClient.Configure)
	runner.RegisterConfigurableEntity(fileManagerClient.Configure)

	chatMapper := chat_mapper.New(redisClient)
	chatMapperStreaming := chat_mapper.NewChatMapperStreaming(redisClient)

	srvCategories := service.NewCategoryInfo()
	srvTickets := service.NewTicketsService(redisClient, resourceEmployeeAccessClient, fileManagerClient, chatMapper)
	srvTicketFiles := service.NewTicketFilesService(redisClient, resourceEmployeeAccessClient, fileManagerClient)
	srvAdmin := service.NewAdminService(redisClient, resourceEmployeeAccessClient, chatMapperStreaming)

	// categories
	runner.RegisterCustomHandler("AddCategory", srvCategories.AddCategory)
	runner.RegisterCustomHandler("GetCategoryInfoByIDV2", srvCategories.GetCategoryInfoByIDV2)
	runner.RegisterCustomHandler("UpdateCategoryStatus", srvCategories.UpdateCategoryStatus)
	runner.RegisterCustomHandler("UpdateCategory", srvCategories.UpdateCategory)
	runner.RegisterCustomHandler("GetCategoriesTree", srvCategories.GetCategoriesTree)
	runner.RegisterCustomHandler("GetCategoryStatusesInfo", srvCategories.GetCategoryStatusesInfo)
	runner.RegisterCustomHandler("CopyCategoryTicket", srvCategories.CopyCategoryTicket)
	runner.RegisterCustomHandler("CopyCategorySettings", srvCategories.CopyCategorySettings)
	runner.RegisterCustomHandler("MoveCategory", srvCategories.MoveCategory)

	// tickets
	runner.RegisterCustomHandler("CreateTicket", srvTickets.CreateTicket)
	runner.RegisterCustomHandler("CreateTicketV2", srvTickets.CreateTicketV2)
	runner.RegisterCustomHandler("CreateTicketV3", srvTickets.CreateTicketV3)
	runner.RegisterCustomHandler("PerformTicket", srvTickets.PerformTicket)
	runner.RegisterCustomHandler("PerformTicketV2", srvTickets.PerformTicketV2)
	runner.RegisterCustomHandler("PerformTicketV3", srvTickets.PerformTicketV3)
	runner.RegisterCustomHandler("ApproveTicket", srvTickets.ApproveTicket)
	runner.RegisterCustomHandler("ApproveTicketV2", srvTickets.ApproveTicketV2)
	runner.RegisterCustomHandler("ApproveTicketV3", srvTickets.ApproveTicketV3)
	runner.RegisterCustomHandler("BookTicket", srvTickets.BookTicket)
	runner.RegisterCustomHandler("BookTicketV2", srvTickets.BookTicketV2)
	runner.RegisterCustomHandler("UnbookTicket", srvTickets.UnbookTicket)
	runner.RegisterCustomHandler("UnbookTicketV2", srvTickets.UnbookTicketV2)
	runner.RegisterCustomHandler("RejectTicket", srvTickets.RejectTicket)
	runner.RegisterCustomHandler("RejectTicketV2", srvTickets.RejectTicketV2)
	runner.RegisterCustomHandler("RejectMyTicketOrTicketAtPerform", srvTickets.RejectMyTicketOrTicketAtPerform)
	runner.RegisterCustomHandler("RejectTicketAtApproveOrBook", srvTickets.RejectTicketAtApproveOrBook)
	runner.RegisterCustomHandler("GetTicketsCreatedByEmployee", srvTickets.GetTicketsCreatedByEmployee)
	runner.RegisterCustomHandler("GetTicket", srvTickets.GetTicket)
	runner.RegisterCustomHandler("GetTicketsForWorkV2", srvTickets.GetTicketsForWorkV2)
	runner.RegisterCustomHandler("GetWorkTickets", srvTickets.GetWorkTickets)
	runner.RegisterCustomHandler("UpdateFavouriteTickets", srvTickets.UpdateFavouriteTickets)
	runner.RegisterCustomHandler("TicketsFavouriteGetByEmployeeV2", srvTickets.TicketsFavouriteGetByEmployeeV2)
	runner.RegisterCustomHandler("TicketsFavouriteGetByEmployeeForWorkV2", srvTickets.TicketsFavouriteGetByEmployeeForWorkV2)
	runner.RegisterCustomHandler("GetTicketInfoByTypeAction", srvTickets.GetTicketInfoByTypeAction)
	runner.RegisterCustomHandler("GetTicketInfoForCenter", srvTickets.GetTicketInfoForCenter)
	runner.RegisterCustomHandler("GetTicketInfoForMyTickets", srvTickets.GetTicketInfoForMyTickets)
	runner.RegisterCustomHandler("ShareMyTicket", srvTickets.ShareMyTicket)
	runner.RegisterCustomHandler("ShareWorkTicket", srvTickets.ShareWorkTicket)
	runner.RegisterCustomHandler("GetTicketForLink", srvTickets.GetTicketForLink)
	runner.RegisterCustomHandler("GetMyTickets", srvTickets.GetMyTickets)
	runner.RegisterCustomHandler("CheckTicketAccessForSubTab", srvTickets.CheckTicketAccessForSubTab)
	runner.RegisterCustomHandler("GetCategoryStatusesForMyTickets", srvTickets.GetCategoryStatusesForMyTickets)
	runner.RegisterCustomHandler("GetCategoryStatusesForWorkTickets", srvTickets.GetCategoryStatusesForWorkTickets)
	runner.RegisterCustomHandler("GetTicketsByIDs", srvTickets.GetTicketsByIDs)
	runner.RegisterCustomHandler("DeleteTicketFromObservables", srvTickets.DeleteTicketFromObservables)
	runner.RegisterCustomHandler("RequestToRejectTicket", srvTickets.RequestToRejectTicket)
	runner.RegisterCustomHandler("UpdateTicketExt", srvTickets.UpdateTicketExt)
	runner.RegisterCustomHandler("GetTicketFieldInfoByCategoryStatus", srvTickets.GetTicketFieldInfoByCategoryStatus)
	runner.RegisterCustomHandler("ReturnTicketStatus", srvTickets.ReturnTicketStatus)

	// tickets files
	runner.RegisterCustomHandler("GetUploadFileLink", srvTicketFiles.GetUploadFileLink)
	runner.RegisterCustomHandler("CompleteFileUpload", srvTicketFiles.CompleteFileUpload)
	runner.RegisterCustomHandler("GetFileDownloadURL", srvTicketFiles.GetFileDownloadURL)
	runner.RegisterCustomHandler("CancelFileUpload", srvTicketFiles.CancelFileUpload)

	// admin
	runner.RegisterCustomHandler("GetTicketsByCategoryIDs", srvAdmin.GetTicketsByCategoryIDs)
	runner.RegisterCustomHandler("GetCenterTickets", srvAdmin.GetCenterTickets)
	runner.RegisterCustomHandler("GetCenterTicketIDs", srvAdmin.GetCenterTicketIDs)
	runner.RegisterCustomHandler("GetCenterTicketByTicketIDs", srvAdmin.GetCenterTicketByTicketIDs)

	runOptions := append([]rest_auto_api.CustomRunOption{reqSizeLimiterOption}, loggerToClickhouse...)

	runner.Run(runOptions...)
}
