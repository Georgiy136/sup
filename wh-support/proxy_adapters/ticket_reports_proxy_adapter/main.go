package main

import (
	"gitlab.wildberries.ru/wbwh/olap-bi/backend/reports.git/utils/clickhouse_logger"
	"gitlab.wildberries.ru/wbwh/olap-bi/backend/reports.git/utils/clickhouse_logger/client"
	clickhouse_handler "gitlab.wildberries.ru/wbwh/olap-bi/backend/reports.git/utils/clickhouse_logger/handler"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository/connection_control"
	httpClient "gitlab.wildberries.ru/wbwh/wh-core/gocore_http.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_proxy_adapter.git/service"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs/env_config"
)

func main() {
	conf := env_config.GetCommonEnvConfigs()
	httpClient := httpClient.NewHttpClient()
	click := client.NewClickhouse(httpClient)
	logger := clickhouse_logger.NewClickhouseLogger(click, conf)
	clickhouseLoggerHandler := clickhouse_handler.NewClickhouseLoggerHandler(logger)

	proxy := service.NewProxyAdapter()
	proxy.Run(
		service.InitConnections(connection_control.InitConnections(conf)),
		service.AddHandlerInApis(logger.Configure, clickhouseLoggerHandler.Handler, service.Start),
	)
}
