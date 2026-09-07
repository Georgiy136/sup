package service

import (
	"context"
	"sort"
	"time"

	employeeactiongroups "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/resources_cache_cron/internal/models/employee_action_groups"
	timeutils "gitlab.wildberries.ru/wbwh/wh-core/gocore-utils-time.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_cron_core.git/cron_core"
	cron_models "gitlab.wildberries.ru/wbwh/wh-core/gocore_cron_core.git/cron_core/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	"github.com/sirupsen/logrus"
)

type employeeActionGroupsChangesCron struct {
	logId                     *cron_core.LogID
	cronCfg                   cron_models.CronCommonCfg
	employeeActionGroupsCache EmployeeActionGroupsCache
	employeeActionGroupsRepo  EmployeeActionGroupsRepo
}

func NewEmployeeActionGroupsChangesCron(employeeActionGroupsCache EmployeeActionGroupsCache, employeeActionGroupsRepo EmployeeActionGroupsRepo) cron_core.Worker {
	return &employeeActionGroupsChangesCron{
		logId:                     cron_core.NewInMemoryLogID(),
		employeeActionGroupsCache: employeeActionGroupsCache,
		employeeActionGroupsRepo:  employeeActionGroupsRepo,
	}
}

func (c *employeeActionGroupsChangesCron) Start(ctx context.Context, _ configs.Config, cronConf cron_models.CronCommonCfg) {
	c.cronCfg = cronConf
	c.startWorking(ctx)
}

func (c *employeeActionGroupsChangesCron) startWorking(ctx context.Context) {
	c.initCache(ctx)
}

func (c *employeeActionGroupsChangesCron) initCache(ctx context.Context) {
	var logID int64
	for {
		logID = c.logId.GetLastLogID()
		result, err := c.employeeActionGroupsRepo.GetEmployeeActionGroupsFromDB(logID)
		if err != nil {
			logrus.Errorf("[%s] can't get employee action groups from db, log_id = %v, err: %v", c.cronCfg.CronTaskName, logID, err)
			time.Sleep(c.cronCfg.CronTimeSleepOnErrorParsed)
			continue
		}

		logrus.Infof("data from db EmployeeActionGroupsFromDB: %v", result)

		if len(result.Data) == 0 {
			time.Sleep(c.cronCfg.CronTimeSleepOnNoDataParsed)
			continue
		}

		for i := range result.Data {
			parsedTime, err := timeutils.ParseDBTime(result.Data[i].ChDt)
			if err != nil {
				logrus.Errorf("[%s] can't parse ch_dt, log_id = %d, ch_dt = %s, err: %v", c.cronCfg.CronTaskName, result.LogId, result.Data[i].ChDt, err)
				time.Sleep(c.cronCfg.CronTimeSleepOnErrorParsed)
				continue
			}
			result.Data[i].ChDtParsed = parsedTime
		}

		sort.Slice(result.Data, func(i, j int) bool {
			return result.Data[i].ChDtParsed.Before(result.Data[j].ChDtParsed)
		})

		for i := range result.Data {
			var err error
			if result.Data[i].IsDel {
				err = c.employeeActionGroupsCache.RemoveGroupFromEmployee(ctx, result.Data[i].EmployeeId, result.Data[i].GroupId)
			} else {
				err = c.employeeActionGroupsCache.AddGroupToEmployee(ctx, result.Data[i].EmployeeId, result.Data[i].GroupId)
			}

			if err != nil {
				logrus.Errorf("[%s] can't process employee action groups events, log_id = %d, err: %v", c.cronCfg.CronTaskName, result.LogId, err)
				time.Sleep(c.cronCfg.CronTimeSleepOnErrorParsed)
				continue
			}
		}

		c.logId.SetLogID(result.LogId) //nolint:gosec
		logrus.Debugf("[%s] successfully processed employee action groups events: %+v", c.cronCfg.CronTaskName, result.Data)
		time.Sleep(time.Duration(result.SleepSeconds) * time.Second)
	}
}

type EmployeeActionGroupsCache interface {
	AddGroupToEmployee(ctx context.Context, employeeID, groupID int64) error
	RemoveGroupFromEmployee(ctx context.Context, employeeID, groupID int64) error
}

type EmployeeActionGroupsRepo interface {
	GetEmployeeActionGroupsFromDB(logID int64) (*employeeactiongroups.EmployeeActionGroupDBResp, error)
}
