package crons

import (
	"context"
	"errors"
	"time"

	customerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/errors"
	cronmodels "gitlab.wildberries.ru/wbwh/wh-core/gocore_cron_core.git/cron_core/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	"github.com/sirupsen/logrus"
)

type SendNotificationsCron struct {
	cronCfg         cronmodels.CronCommonCfg
	notifierService Notifier
}

func NewSendNotificationsCron(notifierService Notifier) *SendNotificationsCron {
	return &SendNotificationsCron{
		notifierService: notifierService,
	}
}

func (a *SendNotificationsCron) Start(ctx context.Context, cfg configs.Config, cronConf cronmodels.CronCommonCfg) {
	a.cronCfg = cronConf
	a.notifierService.Configure(ctx, cfg, cronConf)

	logrus.Infof("[%s] started...", a.cronCfg.CronTaskName)

	for {
		select {
		case <-ctx.Done():
			logrus.Infof("[%s] stopped", a.cronCfg.CronTaskName)
			return
		default:
			sleepDuration, err := a.notifierService.ProcessNotifications(ctx)
			if err != nil {
				if errors.Is(err, customerrors.ErrEmptyData) {
					logrus.Infof("[%s] no notifications... sleep %s", a.cronCfg.CronTaskName, sleepDuration.String())
					time.Sleep(sleepDuration)
					continue
				}

				logrus.Errorf("[%s] can't process notifications; sleep: %s: err: %v", a.cronCfg.CronTaskName, sleepDuration.String(), err)
				time.Sleep(sleepDuration)
				continue
			}

			logrus.Infof("[%s] processed notifications... sleep: %s", a.cronCfg.CronTaskName, sleepDuration.String())
			time.Sleep(sleepDuration)
		}
	}
}

type Notifier interface {
	ProcessNotifications(ctx context.Context) (sleep time.Duration, err error)
	Configure(ctx context.Context, cfg configs.Config, cronConf cronmodels.CronCommonCfg)
}
