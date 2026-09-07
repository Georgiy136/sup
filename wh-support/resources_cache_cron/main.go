package main

import (
	"context"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/resources_cache_cron/internal/service"
	postgresrepo "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/resources_cache_cron/internal/storage/postgres"
	redisrepo "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/resources_cache_cron/internal/storage/redis"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_cron_core.git/cron_core"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository/connection_control"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs/env_config"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_core.git/app"
)

func main() {
	serviceWorker := app.NewServerScript()
	serviceWorker.InitConfigManagerDefault()
	conf := env_config.GetCommonEnvConfigs()

	redis := redisrepo.NewRedisClient()

	pgRepo := postgresrepo.NewPostgresRepo()
	cacheInternalActionsRepo := redisrepo.NewInternalActionsRepository(redis)
	cacheEmployeeActionGroupsRepo := redisrepo.NewEmployeeActionGroupsRepository(redis)
	cacheAccessPolicyRepo := redisrepo.NewAccessPolicyRepository(redis)

	serviceWorker.Configuration(
		connection_control.InitConnections(conf),
		redis.Configure,
	)
	serviceWorker.Tasks(func(ctx context.Context, config configs.Config) {
		getHandler := cron_core.GetWorkerRegistration(map[string]cron_core.Worker{
			"internal_actions.changes":       service.NewInternalActionsChangesCron(cacheInternalActionsRepo, pgRepo),
			"employee_action_groups.changes": service.NewEmployeeActionGroupsChangesCron(cacheEmployeeActionGroupsRepo, pgRepo),
			"access_policy.changes":          service.NewAccessPolicyChangesCron(cacheAccessPolicyRepo, pgRepo),
		})
		getHandler.InitCrons(ctx, config)
	})

	app.StartServer(serviceWorker)
}
