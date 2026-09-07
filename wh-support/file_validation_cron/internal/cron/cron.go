package cron

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	cron_models "gitlab.wildberries.ru/wbwh/wh-core/gocore_cron_core.git/cron_core/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
)

type Cron struct {
	cronCfg cron_models.CronCommonCfg
	service Service
}

func NewCron(service Service) *Cron {
	return &Cron{
		service: service,
	}
}

func (f *Cron) Start(ctx context.Context, _ configs.Config, cronConf cron_models.CronCommonCfg) {
	f.cronCfg = cronConf

	for {
		select {
		case <-ctx.Done():
			logrus.Infof("cron %s stopped: %v", f.cronCfg.CronTaskName, ctx.Err())
			return
		default:
		}

		err := f.service.Process(ctx, f.cronCfg.CronTaskName)
		if err != nil {
			logrus.Errorf("can't cron %s process: %v", f.cronCfg.CronTaskName, err)
			time.Sleep(f.cronCfg.CronTimeSleepOnErrorParsed)
			continue
		}
		time.Sleep(f.cronCfg.CronTimeSleepOnOkParsed)
	}
}

type Service interface {
	Process(ctx context.Context, cronTaskName string) error
}
