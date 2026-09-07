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

	"github.com/gin-gonic/gin/binding"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"

	timeutils "gitlab.wildberries.ru/wbwh/wh-core/gocore-utils-time.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
)

type TicketsCategoriesExportCron struct {
	cronCfg cron_models.CronCommonCfg

	kafkaOlapClient *clients.WhKafkaOlapWebPublisherClient
}

func NewTicketsCategoriesExportCron(kafkaOlapClient *clients.WhKafkaOlapWebPublisherClient) *TicketsCategoriesExportCron {
	return &TicketsCategoriesExportCron{
		kafkaOlapClient: kafkaOlapClient,
	}
}

func (s *TicketsCategoriesExportCron) Start(ctx context.Context, conf configs.Config, cronConf cron_models.CronCommonCfg) {
	s.cronCfg = cronConf

	s.Work()
}

func (s *TicketsCategoriesExportCron) Work() {
	const (
		maxBatchLen = 10
		apiKey      = "support.tickets.categories"
		sp          = "sync.categorysync_exporttojson"
		ackSp       = "sync.categorysync_delsynced"
	)

	for {
		processedLenData := 0

		dbResponse, err := repository.ExecuteSupportDb(sp, false)
		if err != nil {
			logrus.Errorf("[%v] can't get categories from db, err %v, sleep %v", s.cronCfg.CronTaskName, err, s.cronCfg.CronTimeSleepOnErrorParsed)
			time.Sleep(s.cronCfg.CronTimeSleepOnErrorParsed)
			continue
		}

		var ticketsCategoriesFromDb models.ResponseTicketsCategoriesFromDB

		if err = jsoniter.Unmarshal(dbResponse, &ticketsCategoriesFromDb); err != nil {
			logrus.Errorf("[%v] err unmarshal db response, err: %v, sleep %v", s.cronCfg.CronTaskName, err, s.cronCfg.CronTimeSleepOnErrorParsed)
			time.Sleep(s.cronCfg.CronTimeSleepOnErrorParsed)
			continue
		}

		if len(ticketsCategoriesFromDb.CategoriesData) == 0 {
			logrus.Infof("[%v] no data from db, sleep %v", s.cronCfg.CronTaskName, s.cronCfg.CronTimeSleepOnNoDataParsed)
			time.Sleep(s.cronCfg.CronTimeSleepOnNoDataParsed)
			continue
		}

		if err := binding.Validator.ValidateStruct(&ticketsCategoriesFromDb); err != nil {
			logrus.Errorf("[%v] err validate db response, err: %v, sleep %v", s.cronCfg.CronTaskName, err, s.cronCfg.CronTimeSleepOnErrorParsed)
			time.Sleep(s.cronCfg.CronTimeSleepOnErrorParsed)
			continue
		}

		batch := make([]*sync_models.Categories_Category, 0, maxBatchLen)
		for _, categoryData := range ticketsCategoriesFromDb.CategoriesData {
			dt, err := timeutils.ParseDBTime(categoryData.ChDt)
			if err != nil {
				logrus.Errorf(
					"[%v] log_id %v cannot parse ChDt, format: %s, dt got: %s, err: %v; sleep: %v",
					s.cronCfg.CronTaskName,
					categoryData.LogId,
					timeutils.LayoutDataBase,
					categoryData.ChDt,
					err,
					s.cronCfg.CronTimeSleepOnErrorParsed,
				)
				time.Sleep(s.cronCfg.CronTimeSleepOnErrorParsed)
				continue
			}
			categoryData.ChDt = timeutils.ToMoscowRFC3339Nano(dt)

			batch = append(batch, categoryData)

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

func (s *TicketsCategoriesExportCron) sendAndAcknowledgeBatch(batch []*sync_models.Categories_Category, apiKey, ackSp string) error {
	if len(batch) == 0 {
		return nil
	}

	logIDs := make([]int64, len(batch))
	for i := range batch {
		logIDs[i] = batch[i].LogId
	}

	ticketsCategoriesProto := sync_models.Categories{Data: batch}
	ticketsCategoriesConverted, err := utils.ConvertTicketsCategories(&ticketsCategoriesProto)
	if err != nil {
		logrus.Errorf("[%v] can't convert ticket categories to json, log_ids: %v, err: %v, sleep %v", s.cronCfg.CronTaskName, logIDs, err, s.cronCfg.CronTimeSleepOnErrorParsed)
		return err
	}

	if err := s.kafkaOlapClient.SendData(apiKey, ticketsCategoriesConverted.Data); err != nil {
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
