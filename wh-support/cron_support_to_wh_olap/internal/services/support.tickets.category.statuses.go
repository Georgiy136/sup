package services

import (
	"context"
	"fmt"
	"time"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/cron_support_to_wh_olap/internal/clients"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/cron_support_to_wh_olap/internal/models"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/cron_support_to_wh_olap/internal/repository"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/cron_support_to_wh_olap/internal/utils"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/cron_support_to_wh_olap/sync_models"
	cron_models "gitlab.wildberries.ru/wbwh/wh-core/gocore_cron_core.git/cron_core/models"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"

	timeutils "gitlab.wildberries.ru/wbwh/wh-core/gocore-utils-time.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
)

type TicketsCategoryStatusesExportCron struct {
	cronCfg cron_models.CronCommonCfg

	kafkaOlapClient *clients.WhKafkaOlapWebPublisherClient
}

func NewTicketsCategoryStatusesExportCron(kafkaOlapClient *clients.WhKafkaOlapWebPublisherClient) *TicketsCategoryStatusesExportCron {
	return &TicketsCategoryStatusesExportCron{
		kafkaOlapClient: kafkaOlapClient,
	}
}

func (s *TicketsCategoryStatusesExportCron) Start(ctx context.Context, conf configs.Config, cronConf cron_models.CronCommonCfg) {
	s.cronCfg = cronConf

	s.Work()
}

func (s *TicketsCategoryStatusesExportCron) Work() {
	const (
		maxBatchLen = 10
		apiKey      = "support.tickets.category.statuses"
		sp          = "sync.categorystatussync_exporttojson"
		ackSp       = "sync.categorystatussync_delsynced"
	)

	for {
		processedLenData := 0

		dbResponse, err := repository.ExecuteSupportDb(sp, false)
		if err != nil {
			logrus.Errorf("[%v] can't get category statuses from db, err %v, sleep %v", s.cronCfg.CronTaskName, err, s.cronCfg.CronTimeSleepOnErrorParsed)
			time.Sleep(s.cronCfg.CronTimeSleepOnErrorParsed)
			continue
		}

		var ticketsCategoryStatusesFromDb models.ResponseTicketsCategoryStatusesFromDB

		if err = jsoniter.Unmarshal(dbResponse, &ticketsCategoryStatusesFromDb); err != nil {
			logrus.Errorf("[%v] err unmarshal db response, err: %v, sleep %v", s.cronCfg.CronTaskName, err, s.cronCfg.CronTimeSleepOnErrorParsed)
			time.Sleep(s.cronCfg.CronTimeSleepOnErrorParsed)
			continue
		}

		if len(ticketsCategoryStatusesFromDb.CategoryStatusesData) == 0 {
			logrus.Infof("[%v] no data from db, sleep %v", s.cronCfg.CronTaskName, s.cronCfg.CronTimeSleepOnNoDataParsed)
			time.Sleep(s.cronCfg.CronTimeSleepOnNoDataParsed)
			continue
		}

		batch := make([]*sync_models.CategoryStatuses_CategoryStatus, 0, maxBatchLen)

		for _, categoryStatusData := range ticketsCategoryStatusesFromDb.CategoryStatusesData {
			dt, err := timeutils.ParseDBTime(categoryStatusData.ChDt)
			if err != nil {
				logrus.Errorf(
					"[%v] log_id %v cannot parse ChDt, format: %s, dt got: %s, err: %v; sleep: %v",
					s.cronCfg.CronTaskName,
					categoryStatusData.LogId,
					timeutils.LayoutDataBase,
					categoryStatusData.ChDt,
					err,
					s.cronCfg.CronTimeSleepOnErrorParsed,
				)
				time.Sleep(s.cronCfg.CronTimeSleepOnErrorParsed)
				continue
			}
			categoryStatusData.ChDt = timeutils.ToMoscowRFC3339Nano(dt)

			batch = append(batch, categoryStatusData)

			if len(batch) == maxBatchLen {
				if err = s.sendAndAcknowledgeBatch(batch, apiKey, ackSp); err != nil {
					batch = batch[:0]
					time.Sleep(s.cronCfg.CronTimeSleepOnErrorParsed)
					continue
				}
				processedLenData += len(batch)
				batch = batch[:0]
			}
		}

		if len(batch) > 0 {
			if err = s.sendAndAcknowledgeBatch(batch, apiKey, ackSp); err != nil {
				time.Sleep(s.cronCfg.CronTimeSleepOnErrorParsed)
				continue
			}
			processedLenData += len(batch)
		}

		if s.cronCfg.CountDataWithoutSleep > 0 && processedLenData >= s.cronCfg.CountDataWithoutSleep {
			continue
		}

		logrus.Infof("[%v] work complete, sleep %v", s.cronCfg.CronTaskName, s.cronCfg.CronTimeSleepOnOkParsed)
		time.Sleep(s.cronCfg.CronTimeSleepOnOkParsed)
	}
}

func (s *TicketsCategoryStatusesExportCron) sendAndAcknowledgeBatch(batch []*sync_models.CategoryStatuses_CategoryStatus, apiKey, ackSp string) error {
	if len(batch) == 0 {
		return nil
	}

	logIDs := make([]int64, len(batch))
	for i := range batch {
		logIDs[i] = batch[i].LogId
	}

	ticketsCategoryStatusesProto := sync_models.CategoryStatuses{Data: batch}
	ticketsCategoryStatusesConverted, err := utils.ConvertTicketsCategoryStatuses(&ticketsCategoryStatusesProto)
	if err != nil {
		logrus.Errorf("[%v] can't convert ticket category statuses to json, log_ids: %v, err: %v, sleep %v", s.cronCfg.CronTaskName, logIDs, err, s.cronCfg.CronTimeSleepOnErrorParsed)
		return err
	}

	if err := s.kafkaOlapClient.SendData(apiKey, ticketsCategoryStatusesConverted.Data); err != nil {
		logrus.Errorf("[%v] can't send batch to kafka olap, log_ids: %v, err: %v, sleep %v", s.cronCfg.CronTaskName, logIDs, err, s.cronCfg.CronTimeSleepOnErrorParsed)
		return fmt.Errorf("[%s] can't send batch to kafka olap: %w", s.cronCfg.CronTaskName, err)
	}

	if _, err := repository.ExecuteSupportDb(ackSp, true, logIDs); err != nil {
		logrus.Errorf("[%v] can't acknowledge synced data, log_ids: %v, err: %v, sleep %v", s.cronCfg.CronTaskName, logIDs, err, s.cronCfg.CronTimeSleepOnErrorParsed)
		return err
	}

	logrus.Infof("[%v] successfully acknowledged %d log_ids: %v", s.cronCfg.CronTaskName, len(logIDs), logIDs)
	return nil
}
