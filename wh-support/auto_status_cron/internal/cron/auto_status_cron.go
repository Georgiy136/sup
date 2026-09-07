package cron

import (
	"context"
	"time"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services"
	cron_models "gitlab.wildberries.ru/wbwh/wh-core/gocore_cron_core.git/cron_core/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	"github.com/sirupsen/logrus"
)

type AutoStatusCron struct {
	cronCfg cron_models.CronCommonCfg
	service *services.AutoStatusService
}

func NewAutoStatusCron(service *services.AutoStatusService) *AutoStatusCron {
	return &AutoStatusCron{
		service: service,
	}
}

func (a *AutoStatusCron) Start(ctx context.Context, _ configs.Config, cronConf cron_models.CronCommonCfg) {
	a.cronCfg = cronConf

	for {
		err := a.service.ProcessTicketsOnAutoStatus(ctx)
		if err != nil {
			logrus.Errorf("can't process ticket on auto status error: %v", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(a.cronCfg.CronTimeSleepOnErrorParsed):
			}
			continue
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(a.cronCfg.CronTimeSleepOnOkParsed):
		}
	}
}

type Service interface {
	ProcessTicketsOnAutoStatus(ctx context.Context) error
}
