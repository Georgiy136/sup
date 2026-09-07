package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	internalactions "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/resources_cache_cron/internal/models/internal_actions"
	utils "gitlab.wildberries.ru/wbwh/wh-core/gocore-utils-time.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_cron_core.git/cron_core"
	cron_models "gitlab.wildberries.ru/wbwh/wh-core/gocore_cron_core.git/cron_core/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	"github.com/sirupsen/logrus"
)

var (
	ErrEmployeeNotFound = errors.New("employee not found")
)

type internalActionsChangesCron struct {
	logId                *cron_core.LogID
	cronCfg              cron_models.CronCommonCfg
	internalActionsCache InternalActionsCache
	internalActionsRepo  InternalActionsRepo
}

func NewInternalActionsChangesCron(internalActionsCache InternalActionsCache, internalActionsRepo InternalActionsRepo) cron_core.Worker {
	return &internalActionsChangesCron{
		logId:                cron_core.NewInMemoryLogID(),
		internalActionsCache: internalActionsCache,
		internalActionsRepo:  internalActionsRepo,
	}
}

func (c *internalActionsChangesCron) Start(ctx context.Context, _ configs.Config, cronConf cron_models.CronCommonCfg) {
	c.cronCfg = cronConf
	c.startWorking(ctx)
}

func (c *internalActionsChangesCron) startWorking(ctx context.Context) {
	c.initCache(ctx)
}

func (c *internalActionsChangesCron) initCache(ctx context.Context) {
	var logID int64
	for {
		logID = c.logId.GetLastLogID()
		result, err := c.internalActionsRepo.GetInternalActionsFromDB(logID)
		if err != nil {
			logrus.Errorf("[%s] can't get internal actions from db, log_id = %v, err: %v", c.cronCfg.CronTaskName, logID, err)
			time.Sleep(c.cronCfg.CronTimeSleepOnErrorParsed)
			continue
		}

		if len(result.Data) == 0 {
			time.Sleep(c.cronCfg.CronTimeSleepOnNoDataParsed)
			continue
		}

		changedInCache := make([]internalactions.EmployeeResourcesWithLogID, 0, len(result.Data))

		isSuccess := true

	resDataLoop:
		for i := range result.Data {
			if result.Data[i].IsDel {
				err := c.removeInternalActionFromEmployee(ctx, result.Data[i])
				if err != nil {
					if errors.Is(err, ErrEmployeeNotFound) {
						continue resDataLoop
					}

					logrus.Errorf("[%s] can't delete action from cache, log_id = %v, err: %v", c.cronCfg.CronTaskName, result.LogId, err)
					isSuccess = false
					time.Sleep(c.cronCfg.CronTimeSleepOnErrorParsed)
					break resDataLoop
				}
				changedInCache = append(changedInCache, result.Data[i])
				continue resDataLoop
			}

			cacheData, err := c.internalActionsCache.GetInternalActionsByEmployeeID(ctx, result.Data[i].EmployeeId)
			if err != nil {
				logrus.Errorf("[%s] can't get data from cache, log_id = %v, err: %v", c.cronCfg.CronTaskName, result.LogId, err)
				isSuccess = false
				time.Sleep(c.cronCfg.CronTimeSleepOnErrorParsed)
				break resDataLoop
			}

			if cacheData != nil {
				cacheChDt, err := utils.ParseDBTime(cacheData.ChDt)
				if err != nil {
					logrus.Errorf("[%s] cannot parse date in RFC3339Nano, err: %v", c.cronCfg.CronTaskName, err)
					isSuccess = false
					time.Sleep(c.cronCfg.CronTimeSleepOnErrorParsed)
					break resDataLoop
				}

				resultChDt, err := utils.ParseDBTime(result.Data[i].ChDt)
				if err != nil {
					logrus.Errorf("[%s] cannot parse date in RFC3339Nano, err: %v", c.cronCfg.CronTaskName, err)
					isSuccess = false
					time.Sleep(c.cronCfg.CronTimeSleepOnErrorParsed)
					break resDataLoop
				}
				if cacheChDt.Before(resultChDt) || cacheChDt.Equal(resultChDt) {
					for idx := range cacheData.ActionIds {
						if cacheData.ActionIds[idx] == result.Data[i].ActionId {
							err := c.setInternalActionsForEmployees(ctx, cacheData.EmployeeId, result.Data[i].ChEmployeeID, cacheData.ActionIds, result.Data[i].ChDt)
							if err != nil {
								logrus.Errorf("[%s] can't save data to cache, err: %v", c.cronCfg.CronTaskName, err)
								isSuccess = false
								time.Sleep(c.cronCfg.CronTimeSleepOnErrorParsed)
								break resDataLoop
							}

							changedInCache = append(changedInCache, result.Data[i])
							continue resDataLoop
						}
					}
					newActionIds := append([]string{result.Data[i].ActionId}, cacheData.ActionIds...)

					err := c.setInternalActionsForEmployees(ctx, result.Data[i].EmployeeId, result.Data[i].ChEmployeeID, newActionIds, result.Data[i].ChDt)
					if err != nil {
						logrus.Errorf("[%s] can't save data to cache, err: %v", c.cronCfg.CronTaskName, err)
						isSuccess = false
						time.Sleep(c.cronCfg.CronTimeSleepOnErrorParsed)
						break resDataLoop
					}
					changedInCache = append(changedInCache, result.Data[i])
				}
			} else {
				newActionIds := []string{result.Data[i].ActionId}

				err := c.setInternalActionsForEmployees(ctx, result.Data[i].EmployeeId, result.Data[i].ChEmployeeID, newActionIds, result.Data[i].ChDt)
				if err != nil {
					logrus.Errorf("[%s] can't save data to cache, err: %v", c.cronCfg.CronTaskName, err)
					isSuccess = false
					time.Sleep(c.cronCfg.CronTimeSleepOnErrorParsed)
					break resDataLoop
				}
				changedInCache = append(changedInCache, result.Data[i]) //nolint:gosec
			}
		}

		if !isSuccess {
			continue
		}

		c.logId.SetLogID(result.LogId)
		logrus.Debugf("[%s] need changed in cache: %+v", c.cronCfg.CronTaskName, *result)
		logrus.Debugf("[%s] successfully changed in cache: %+v", c.cronCfg.CronTaskName, changedInCache)
		time.Sleep(time.Duration(result.SleepSeconds) * time.Second)
	}
}

func (c *internalActionsChangesCron) setInternalActionsForEmployees(ctx context.Context, employeeId, chEmployeeId int64, actionIds []string, chDt string) error {
	resToSave := internalactions.EmployeeResources{
		EmployeeId:   employeeId,
		ActionIds:    actionIds,
		ChEmployeeID: chEmployeeId,
		ChDt:         chDt,
	}

	err := c.internalActionsCache.SetInternalActionsByEmployeeIDs(ctx, resToSave)
	if err != nil {
		return fmt.Errorf("can't set internal actions by employee ids, err: %v", err)
	}

	return nil
}

func (c *internalActionsChangesCron) removeInternalActionFromEmployee(ctx context.Context, employeeResource internalactions.EmployeeResourcesWithLogID) error {
	var newActionIds []string

	cacheData, err := c.internalActionsCache.GetInternalActionsByEmployeeID(ctx, employeeResource.EmployeeId)
	if err != nil {
		return fmt.Errorf("can't get data from cache, err: %w", err)
	}

	if cacheData == nil {
		return fmt.Errorf("cashe data is nil: %w", ErrEmployeeNotFound)
	}

	for i := range cacheData.ActionIds {
		if cacheData.ActionIds[i] != employeeResource.ActionId {
			newActionIds = append(newActionIds, cacheData.ActionIds[i])
		}
	}

	if newActionIds == nil {
		err := c.internalActionsCache.DelInternalActionByEmployeeID(ctx, employeeResource.EmployeeId)
		if err != nil {
			return fmt.Errorf("can't delete data from cache, err: %w", err)
		}

		return nil
	}

	err = c.setInternalActionsForEmployees(ctx, employeeResource.EmployeeId, employeeResource.ChEmployeeID, newActionIds, employeeResource.ChDt)
	if err != nil {
		return fmt.Errorf("can't save data to cache, err: %w", err)
	}

	return nil
}

type InternalActionsCache interface {
	DelInternalActionByEmployeeID(ctx context.Context, employeeID int64) error
	GetInternalActionsByEmployeeID(ctx context.Context, employeeID int64) (*internalactions.EmployeeResources, error)
	SetInternalActionsByEmployeeIDs(ctx context.Context, data ...internalactions.EmployeeResources) error
}

type InternalActionsRepo interface {
	GetInternalActionsFromDB(logID int64) (*internalactions.EmployeeResourcesDBResp, error)
}
