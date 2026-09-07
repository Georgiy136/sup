package main

import (
	"context"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/docgen"
	assemblysheetterminate "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/assembly_sheet_terminate"
	closewh "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/close_wh"
	createstorageplace "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_storage_place"
	createtarestatecreate "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_tare_state"
	deletestorageplace "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/delete_storage_place"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/clients"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/clients/redis"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/clients/sprut"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/cron"
	repo "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/repository"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services"
	additionaltransport "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/additional_transport"
	close_wh_new "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/close_wh_new"
	commonhandlers "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/common_handlers"
	create_goods_invent_task "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_goods_invent_task"
	create_goods_search_invent_task "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_goods_search_invent_task"
	createinventtask "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_invent_task"
	createoffice "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_office"
	createoperations "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_operations"
	createshkstates "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_shk_states"
	createstages "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_stages"
	createstorageplaceinventtask "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_storage_place_invent_task"
	createmxtype "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_storage_place_type"
	createstreets "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_streets"
	createwh "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_wh"
	excludestreetfromsale "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/exclude_street_from_sale"
	excludetare "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/exclude_tare"
	excludewhorstagesfromassembly "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/exclude_wh_or_stages_from_assembly"
	excludewhorstagesforsale "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/exclude_wh_or_stages_from_sale"
	inclusionstreetinsale "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/inclusion_street_in_sale"
	inclusionwhorstagesinassembly "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/inclusion_wh_or_stages_in_assembly"
	inclusionwhorstagesinsale "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/inclusion_wh_or_stages_in_sale"
	internalresort "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/internal_resort"
	removeremainswrr "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/remove_remains_wrr"
	transferstreetsbetweenwh "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/transfer_streets_between_wh"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_cron_core.git/cron_core"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository/connection_control"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_sentry.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs/env_config"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_core.git/app"
)

const (
	categoryIdOldCreateOffice = 100
	categoryIdCreateOffice    = 206

	categoryIdCreateStages = 160
	categoryIdExcludeTare  = 5

	categoryIdExcludeStreetFromSaleOld = 184
	categoryIdExcludeStreetFromSale    = 234

	categoryIdExcludeWhOrStagesFromSale = 454
	categoryIdInclusionWhOrStagesInSale = 455

	categoryIdExcludeWhOrStagesFromAssembly = 467
	categoryIdInclusionWhOrStagesInAssembly = 469

	categoryIdAdditionalTransport     = 323
	categoryIdAdditionalTransportNew  = 472
	categoryIdAdditionalTransportNew2 = 473
	categoryIdAdditionalTransportNew3 = 474
	categoryIdAdditionalTransportNew4 = 478

	categoryIdInclusionStreetInSale        = 183
	categoryIdCreateShkStatesOld           = 189
	categoryIdCreateShkStatesNew           = 338
	categoryIdCreateShkStatesNewV2         = 351
	categoryIdCreateInventTaskV322         = 322
	categoryIdCreateInventTaskV443         = 443
	categoryIdCreateOperations             = 50
	categoryIdCreateOperationsV2           = 207
	categoryIdcloseWhNew                   = 456
	categoryIdcloseWhNew2                  = 477
	categoryIdCloseWh                      = 157
	categoryIdCreateWh                     = 163
	categoryIdCreateStoragePlaceType       = 213
	categoryIdCreateTareState              = 236
	categoryIdRemoveRemainsWRR             = 288
	categoryIdDeleteStoragePlace           = 325
	categoryIdCreateStoragePlace           = 330
	categoryIdAssemblySheetTerminate       = 393
	categoryIdInternalResort               = 396
	categoryIdCreateStreets                = 398
	categoryIdTransferStreetsBetweenWh     = 390
	categoryIdCreateStoragePlaceInventTask = 464
	categoryIdCreateGoodsInventTask        = 460
	categoryIdCreateGoodsSearchInventTask  = 462
)

func main() {
	conf := env_config.GetCommonEnvConfigs()
	serviceWorker := app.NewServerScript()
	serviceWorker.InitConfigManagerDefault()
	sentry := gocore_sentry.NewSentry()

	financeTariffClusterMasterClient := clients.NewFinanceTariffClusterMasterClient()
	sprutAccessClients := sprut.NewSprutAccessClients()
	whSprutenvAdministrationApiClient := clients.NewWhSprutenvAdministrationApiClient()
	wbStorenvWarehouseInfoApiClient := clients.NewWbStorenvWarehouseInfoApiClient()
	ticketApiClient := clients.NewTicketApiClient()
	whWriteoffClusterApiClient := clients.NewWhWriteoffClusterApiClient()
	wbStorenvPlaceInfoApiClient := clients.NewWbStorenvPlaceInfoApiClient()
	wbStorenvTareAttachmentTypesApiClient := clients.NewWhStorenvTareAttachmentTypesApiClient()
	shkClaimsApiClient := clients.NewShkClaimsApiClient()
	bandBotClient := clients.NewBandBotClient()
	employeeInfoApiClient := clients.NewEmployeeInfoApiClient()
	storagePlaceApiClient := clients.NewStoragePlaceApiClient()
	storagePlaceAdministrationApiClient := clients.NewStoragePlaceAdministrationApiClient()
	cargoAggregatorLogisticClient := clients.NewCargoAggregatorLogisticClient()
	goodsIdentificationApiClient := clients.NewGoodsIdentificationApiClient()
	s3Client := clients.NewS3Client()
	gotenbergClient := clients.NewGotenbergClient()
	fileManagerApiClient := clients.NewSupportFileManagerApiClient()
	whPriceAggregatorApiClient := clients.NewWhPriceAggregatorApiClient()
	redisClient := redis.NewRedisClient()
	pdfSrv := docgen.NewService(gotenbergClient)

	ticketRepo := repo.NewTicketsRepo()

	closeWhNewHandler := close_wh_new.NewTicketHandlerCloseWhNew(
		ticketRepo,
		whSprutenvAdministrationApiClient,
		ticketApiClient,
		s3Client,
		pdfSrv,
		fileManagerApiClient,
		redisClient,
		close_wh_new.NewGoodsManager(
			storagePlaceApiClient,
			whPriceAggregatorApiClient,
		),
	)

	autoStatusService := services.NewAutoStatusService(ticketRepo)

	serviceWorker.Configuration(
		connection_control.InitConnections(conf),
		sentry.Configure,

		financeTariffClusterMasterClient.Configure,
		sprutAccessClients.Configure,
		whSprutenvAdministrationApiClient.Configure,
		wbStorenvWarehouseInfoApiClient.Configure,
		ticketApiClient.Configure,
		whWriteoffClusterApiClient.Configure,
		wbStorenvPlaceInfoApiClient.Configure,
		wbStorenvTareAttachmentTypesApiClient.Configure,
		shkClaimsApiClient.Configure,
		bandBotClient.Configure,
		employeeInfoApiClient.Configure,
		storagePlaceApiClient.Configure,
		storagePlaceAdministrationApiClient.Configure,
		cargoAggregatorLogisticClient.Configure,
		goodsIdentificationApiClient.Configure,
		s3Client.Configure,
		gotenbergClient.Configure,
		fileManagerApiClient.Configure,
		whPriceAggregatorApiClient.Configure,
		redisClient.Configure,
		closeWhNewHandler.Configure,
		autoStatusService.Configure,
	)

	commonHandlers := commonhandlers.NewCommonHandlers(ticketRepo, whSprutenvAdministrationApiClient, whSprutenvAdministrationApiClient, whSprutenvAdministrationApiClient)

	createOperationHandler := createoperations.NewTicketHandlerCreateOperation(ticketRepo, financeTariffClusterMasterClient)
	createWhHandler := createwh.NewTicketHandlerCreateWh(ticketRepo, sprutAccessClients, whSprutenvAdministrationApiClient, commonHandlers)
	createOfficeHandler := createoffice.NewTicketHandlerCreateOffice(ticketRepo, whSprutenvAdministrationApiClient)
	closeWhHanler := closewh.NewTicketHandlerCloseWh(ticketRepo, whSprutenvAdministrationApiClient)
	excludeTareHandler := excludetare.NewTicketHandlerExcludeTare(ticketRepo, ticketApiClient, whWriteoffClusterApiClient)
	createStagesHandler := createstages.NewTicketHandlerCreateStages(commonHandlers)

	excludeStreetFromSaleHandler := excludestreetfromsale.NewTicketHandlerExcludeStreetFromSale(ticketRepo, commonHandlers, whSprutenvAdministrationApiClient)
	inclusionStreetInSaleHandler := inclusionstreetinsale.NewTicketHandlerInclusionStreetInSale(ticketRepo, commonHandlers)

	excludeWhOrStagesFromSaleHandler := excludewhorstagesforsale.NewTicketHandlerExcludeWhOrStagesFromSale(ticketRepo, whSprutenvAdministrationApiClient)

	inclusionWhOrStagesInSaleHandler := inclusionwhorstagesinsale.NewTicketHandlerInclusionWhOrStagesInSale(ticketRepo, whSprutenvAdministrationApiClient)

	excludeWhOrStagesFromAssemblyHandler := excludewhorstagesfromassembly.NewTicketHandlerExcludeWhOrStagesFromAssembly(ticketRepo, commonHandlers)

	inclusionWhOrStagesInAssemblyHandler := inclusionwhorstagesinassembly.NewTicketHandlerInclusionWhOrStagesInAssembly(ticketRepo, commonHandlers)

	createShkStatesHandler := createshkstates.NewTicketHandlerCreateShkStates(ticketRepo, wbStorenvWarehouseInfoApiClient)

	createInventTask := createinventtask.NewTicketHandlerCreateInventTask(ticketRepo, whSprutenvAdministrationApiClient)

	createStoragePlaceType := createmxtype.NewTicketHandlerCreateStoragePlaceType(ticketRepo, wbStorenvPlaceInfoApiClient)

	createTareState := createtarestatecreate.NewTicketHandlerCreateTareState(ticketRepo, wbStorenvTareAttachmentTypesApiClient)

	removeRemainsWRR := removeremainswrr.NewTicketHandlerRemoveRemainsWRRHandler(ticketRepo, shkClaimsApiClient)

	additionalTransport := additionaltransport.NewTicketHandlerAdditionalTransport(
		ticketRepo,
		additionaltransport.NewExternalCargoLogistic(cargoAggregatorLogisticClient),
	)

	deleteStoragePlaceHandler := deletestorageplace.NewTicketHandlerDeleteStoragePlace(ticketRepo, storagePlaceApiClient)
	createStoragePlaceHandler := createstorageplace.NewTicketHandlerCreateStoragePlace(ticketRepo, storagePlaceApiClient)
	createStoragePlaceInventTaskHandler := createstorageplaceinventtask.NewTicketHandlerCreateStoragePlaceInventTask(ticketRepo, storagePlaceAdministrationApiClient)
	createGoodsInventTask := create_goods_invent_task.NewTicketHandlerCreateGoodsInventTask(ticketRepo, commonHandlers)
	createGoodsSearchInventTask := create_goods_search_invent_task.NewTicketHandlerCreateGoodsSearchInventTask(ticketRepo, commonHandlers)
	assemblySheetTerminateHandler := assemblysheetterminate.NewTicketHandlerAssemblySheetTerminate(ticketRepo, whSprutenvAdministrationApiClient)

	createStreetsHandler := createstreets.NewTicketHandlerCreateStreets(ticketRepo, whSprutenvAdministrationApiClient)

	internalResort := internalresort.NewTicketHandlerInternalResortHandler(ticketRepo, shkClaimsApiClient, goodsIdentificationApiClient)

	transferStreetsBetweenWhHandler := transferstreetsbetweenwh.NewTicketHandlerTransferStreetsBetweenWh(ticketRepo, whSprutenvAdministrationApiClient)

	autoStatusService.RegisterTicketHandler(categoryIdCreateOperations, createOperationHandler)
	autoStatusService.RegisterTicketHandler(categoryIdCreateOperationsV2, createOperationHandler)

	autoStatusService.RegisterTicketHandler(categoryIdCreateWh, createWhHandler)

	autoStatusService.RegisterTicketHandler(categoryIdCreateOffice, createOfficeHandler)
	autoStatusService.RegisterTicketHandler(categoryIdOldCreateOffice, createOfficeHandler)

	autoStatusService.RegisterTicketHandler(categoryIdCloseWh, closeWhHanler)

	autoStatusService.RegisterTicketHandler(categoryIdcloseWhNew, closeWhNewHandler)
	autoStatusService.RegisterTicketHandler(categoryIdcloseWhNew2, closeWhNewHandler)

	autoStatusService.RegisterTicketHandler(categoryIdExcludeTare, excludeTareHandler)

	autoStatusService.RegisterTicketHandler(categoryIdCreateStages, createStagesHandler)

	autoStatusService.RegisterTicketHandler(categoryIdExcludeStreetFromSaleOld, excludeStreetFromSaleHandler)
	autoStatusService.RegisterTicketHandler(categoryIdExcludeStreetFromSale, excludeStreetFromSaleHandler)

	autoStatusService.RegisterTicketHandler(categoryIdExcludeWhOrStagesFromSale, excludeWhOrStagesFromSaleHandler)

	autoStatusService.RegisterTicketHandler(categoryIdInclusionWhOrStagesInSale, inclusionWhOrStagesInSaleHandler)

	autoStatusService.RegisterTicketHandler(categoryIdExcludeWhOrStagesFromAssembly, excludeWhOrStagesFromAssemblyHandler)

	autoStatusService.RegisterTicketHandler(categoryIdInclusionWhOrStagesInAssembly, inclusionWhOrStagesInAssemblyHandler)

	autoStatusService.RegisterTicketHandler(categoryIdInclusionStreetInSale, inclusionStreetInSaleHandler)

	autoStatusService.RegisterTicketHandler(categoryIdCreateShkStatesOld, createShkStatesHandler)
	autoStatusService.RegisterTicketHandler(categoryIdCreateShkStatesNew, createShkStatesHandler)
	autoStatusService.RegisterTicketHandler(categoryIdCreateShkStatesNewV2, createShkStatesHandler)

	autoStatusService.RegisterTicketHandler(categoryIdCreateInventTaskV322, createInventTask)
	autoStatusService.RegisterTicketHandler(categoryIdCreateInventTaskV443, createInventTask)

	autoStatusService.RegisterTicketHandler(categoryIdCreateStoragePlaceType, createStoragePlaceType)

	autoStatusService.RegisterTicketHandler(categoryIdCreateTareState, createTareState)

	autoStatusService.RegisterTicketHandler(categoryIdRemoveRemainsWRR, removeRemainsWRR)

	autoStatusService.RegisterTicketHandler(categoryIdDeleteStoragePlace, deleteStoragePlaceHandler)

	autoStatusService.RegisterTicketHandler(categoryIdCreateStoragePlace, createStoragePlaceHandler)

	autoStatusService.RegisterTicketHandler(categoryIdAssemblySheetTerminate, assemblySheetTerminateHandler)

	autoStatusService.RegisterBatchTicketHandler(categoryIdAdditionalTransport, additionalTransport)
	autoStatusService.RegisterBatchTicketHandler(categoryIdAdditionalTransportNew, additionalTransport)
	autoStatusService.RegisterBatchTicketHandler(categoryIdAdditionalTransportNew2, additionalTransport)
	autoStatusService.RegisterBatchTicketHandler(categoryIdAdditionalTransportNew3, additionalTransport)
	autoStatusService.RegisterBatchTicketHandler(categoryIdAdditionalTransportNew4, additionalTransport)

	autoStatusService.RegisterTicketHandler(categoryIdCreateStreets, createStreetsHandler)

	autoStatusService.RegisterTicketHandler(categoryIdInternalResort, internalResort)

	autoStatusService.RegisterTicketHandler(categoryIdTransferStreetsBetweenWh, transferStreetsBetweenWhHandler)

	autoStatusService.RegisterTicketHandler(categoryIdCreateStoragePlaceInventTask, createStoragePlaceInventTaskHandler)

	autoStatusService.RegisterTicketHandler(categoryIdCreateGoodsInventTask, createGoodsInventTask)

	autoStatusService.RegisterTicketHandler(categoryIdCreateGoodsSearchInventTask, createGoodsSearchInventTask)

	serviceWorker.Tasks(func(ctx context.Context, conf configs.Config) {
		worker := cron_core.HandlerRegistration{Workers: map[string]cron_core.Worker{
			"auto_status_cron": cron.NewAutoStatusCron(autoStatusService),
		}}
		worker.InitCrons(ctx, conf)
	})

	app.StartServer(serviceWorker)
}
