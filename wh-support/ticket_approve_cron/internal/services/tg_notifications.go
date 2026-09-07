package services

import (
	"context"
	"github.com/sirupsen/logrus"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_approve_cron/internal/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	"time"
)

func (w *TicketApproveCron) sendNotificationsToTGBot() {
	for {
		data, logID, err := w.repo.GetNotifications(w.logID.GetLastLogID())
		if err != nil {
			logrus.Errorf("get tickets for action: %v", err)
			time.Sleep(w.cronCfg.CronTimeSleepOnErrorParsed)
			continue
		}

		if data == nil {
			logrus.Debugf("no data from %s, sleep...", models.GetNotificationsSP)
			time.Sleep(w.cronCfg.CronTimeSleepOnNoDataParsed)
			continue
		}

		var sendNotificationErr error
		for i := range data.Data {
			logrus.Debugf("sending notification about ticket %d for employee %d", data.Data[i].TicketID, data.Data[i].EmployeeID)

			req := models.NotificationRequest{
				TicketInfo:       data.Data[i].TicketInfo,
				Comments:         data.Data[i].Comments,
				TgChatID:         data.Data[i].TgChatID,
				TicketID:         data.Data[i].TicketID,
				EmployeeID:       data.Data[i].EmployeeID,
				TypeOfEmployeeID: data.Data[i].TypeOfEmployeeID,
			}

			logrus.Infof("sending notification to %d: ticket_id = %d", req.EmployeeID, req.TicketID)

			if err := w.clientTG.SendNotification(req); err != nil {
				sendNotificationErr = err
			}
		}

		if sendNotificationErr != nil {
			logrus.Errorf("send notification: %v", sendNotificationErr)
			time.Sleep(w.cronCfg.CronTimeSleepOnErrorParsed)
			continue
		}

		w.logID.SetLogID(logID)

		if len(data.Data) >= w.cronCfg.CountDataWithoutSleep {
			logrus.Debugf("current data from %s count %v is higher than %v, work without sleep", models.GetNotificationsSP, len(data.Data), w.cronCfg.CountDataWithoutSleep)
			continue
		}

		logrus.Debugf("work from %s complete, sleep...", models.GetNotificationsSP)
		time.Sleep(time.Duration(data.SleepSeconds) * time.Second)
	}
}

type tgBotClient interface {
	Configure(ctx context.Context, config configs.Config)
	SendNotification(in models.NotificationRequest) error
}
