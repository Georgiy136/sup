package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	accesspolicy "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/resources_cache_cron/internal/models/access_policy"
	timeutils "gitlab.wildberries.ru/wbwh/wh-core/gocore-utils-time.git"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_cron_core.git/cron_core"
	cron_models "gitlab.wildberries.ru/wbwh/wh-core/gocore_cron_core.git/cron_core/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	"github.com/sirupsen/logrus"
)

type accessPolicyChangesCron struct {
	logId             *cron_core.LogID
	cronCfg           cron_models.CronCommonCfg
	accessPolicyCache AccessPolicyCache
	accessPolicyRepo  AccessPolicyRepo

	accessPolicyHandlers []*accessPolicyHandler
}

type accessPolicyHandler struct {
	localLogID  int64
	handlerName string
	handle      func(ctx context.Context, policies []accesspolicy.AccessPolicy) error
}

func NewAccessPolicyChangesCron(accessPolicyCache AccessPolicyCache, accessPolicyRepo AccessPolicyRepo) cron_core.Worker {
	return &accessPolicyChangesCron{
		logId:             cron_core.NewInMemoryLogID(),
		accessPolicyCache: accessPolicyCache,
		accessPolicyRepo:  accessPolicyRepo,
	}
}

func (c *accessPolicyChangesCron) Start(ctx context.Context, _ configs.Config, cronConf cron_models.CronCommonCfg) {
	c.cronCfg = cronConf

	c.accessPolicyHandlers = append(c.accessPolicyHandlers, &accessPolicyHandler{
		handlerName: "accessPoliciesForActionGroupsHandler",
		handle:      c.accessPoliciesForActionGroupsHandler,
	})
	c.accessPolicyHandlers = append(c.accessPolicyHandlers, &accessPolicyHandler{
		handlerName: "accessPoliciesForExternalActionsHandler",
		handle:      c.accessPoliciesForExternalActionsHandler,
	})
	c.accessPolicyHandlers = append(c.accessPolicyHandlers, &accessPolicyHandler{
		handlerName: "accessPoliciesExternalActionToTypesActionHandler",
		handle:      c.accessPoliciesExternalActionToTypesActionHandler,
	})
	c.accessPolicyHandlers = append(c.accessPolicyHandlers, &accessPolicyHandler{
		handlerName: "accessPoliciesGroupIdToTypesActionHandler",
		handle:      c.accessPoliciesGroupIdToTypesActionHandler,
	})

	c.startWorking(ctx)
}

func (c *accessPolicyChangesCron) startWorking(ctx context.Context) {
	c.worker(ctx)
}

func (c *accessPolicyChangesCron) worker(ctx context.Context) {
	var logID int64
	for {
		logID = c.logId.GetLastLogID()
		result, err := c.accessPolicyRepo.GetAccessPolicyFromDB(logID)
		if err != nil {
			logrus.Errorf("[%s] can't get access policy from db, log_id = %v, err: %v", c.cronCfg.CronTaskName, logID, err)
			time.Sleep(c.cronCfg.CronTimeSleepOnErrorParsed)
			continue
		}

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

		sortAccessPoliciesForApply(result.Data)

		var isErr bool
		for _, handlerData := range c.accessPolicyHandlers {
			if handlerData.localLogID >= result.LogId {
				logrus.Infof("[%s] handler %s skipped, localLogID = %d, db logID = %d", c.cronCfg.CronTaskName, handlerData.handlerName, handlerData.localLogID, result.LogId)
				continue
			}

			err := handlerData.handle(ctx, result.Data)
			if err != nil {
				logrus.Errorf("[%s] can't execute handler %s, err: %v", c.cronCfg.CronTaskName, handlerData.handlerName, err)
				isErr = true
				continue
			}

			handlerData.localLogID = result.LogId
		}

		if isErr {
			time.Sleep(c.cronCfg.CronTimeSleepOnErrorParsed)
			continue
		}

		c.logId.SetLogID(result.LogId) //nolint:gosec
		logrus.Infof("[%s] successfully processed access policy changes events: %+v", c.cronCfg.CronTaskName, result.Data)
		time.Sleep(time.Duration(result.SleepSeconds) * time.Second)
	}
}

// sortAccessPoliciesForApply упорядочивает изменения перед применением в кэш.
//
// Основной ключ — время изменения ch_dt. При равном ch_dt (ресинк категории
// пишет удаление старой и вставку новой строки одной транзакцией, поэтому ch_dt
// у пары совпадает) "удаления" ставятся перед "добавлениями": тк изменение категории
// несете за собой два события "удаление старых значений политики" и "добавление новых значений политики"
// с одинаковым временем изменения
func sortAccessPoliciesForApply(policies []accesspolicy.AccessPolicy) {
	sort.Slice(policies, func(i, j int) bool {
		if !policies[i].ChDtParsed.Equal(policies[j].ChDtParsed) {
			return policies[i].ChDtParsed.Before(policies[j].ChDtParsed)
		}

		// i идёт раньше j, только если i — это удаление, а j — добавление
		return policies[i].IsDel && !policies[j].IsDel
	})
}

func (c *accessPolicyChangesCron) accessPoliciesForActionGroupsHandler(ctx context.Context, policies []accesspolicy.AccessPolicy) error {
	var errs []error
	for _, policy := range policies {
		if !policy.IsDel {
			err := c.accessPolicyCache.AddAccessPolicyForActionGroup(ctx, accesspolicy.ResourceCompositeKey{
				CategoryID: policy.CategoryId,
				TypeAction: policy.TypeAction,
				StatusID:   policy.StatusId,
			}, policy.GroupId)
			if err != nil {
				errs = append(errs, fmt.Errorf("can't add access policy for action group, err: %v", err))
			}

			continue
		}

		err := c.accessPolicyCache.DelAccessPolicyForActionGroup(ctx, accesspolicy.ResourceCompositeKey{
			CategoryID: policy.CategoryId,
			TypeAction: policy.TypeAction,
			StatusID:   policy.StatusId,
		}, policy.GroupId)
		if err != nil {
			errs = append(errs, fmt.Errorf("can't del access policy for action group, err: %v", err))
		}
	}

	return errors.Join(errs...)
}

func (c *accessPolicyChangesCron) accessPoliciesForExternalActionsHandler(ctx context.Context, policies []accesspolicy.AccessPolicy) error {
	var errs []error
	for _, policy := range policies {
		if policy.ExternalAction == nil {
			continue
		}

		if !policy.IsDel {
			err := c.accessPolicyCache.AddAccessPolicyForExternalAction(ctx, accesspolicy.ResourceCompositeKey{
				CategoryID: policy.CategoryId,
				TypeAction: policy.TypeAction,
				StatusID:   policy.StatusId,
			}, *policy.ExternalAction)
			if err != nil {
				errs = append(errs, fmt.Errorf("can't add access policy for external action, err: %v", err))
			}

			continue
		}

		err := c.accessPolicyCache.DelAccessPolicyForExternalAction(ctx, accesspolicy.ResourceCompositeKey{
			CategoryID: policy.CategoryId,
			TypeAction: policy.TypeAction,
			StatusID:   policy.StatusId,
		}, *policy.ExternalAction)
		if err != nil {
			errs = append(errs, fmt.Errorf("can't del access policy for external action, err: %v", err))
		}
	}

	return errors.Join(errs...)
}

func (c *accessPolicyChangesCron) accessPoliciesExternalActionToTypesActionHandler(ctx context.Context, policies []accesspolicy.AccessPolicy) error {
	var errs []error
	for _, policy := range policies {
		if policy.ExternalAction == nil {
			continue
		}

		if !policy.IsDel {
			err := c.accessPolicyCache.AddExternalActionToTypeAction(ctx, policy.TypeAction, accesspolicy.ExternalActionWithCategory{
				CategoryId:     policy.CategoryId,
				ExternalAction: *policy.ExternalAction,
				StatusId:       policy.StatusId,
			})
			if err != nil {
				errs = append(errs, fmt.Errorf("can't add external action to type action, err: %v", err))
			}

			continue
		}

		err := c.accessPolicyCache.DelExternalActionToTypeAction(ctx, policy.TypeAction, accesspolicy.ExternalActionWithCategory{
			CategoryId:     policy.CategoryId,
			ExternalAction: *policy.ExternalAction,
			StatusId:       policy.StatusId,
		})
		if err != nil {
			errs = append(errs, fmt.Errorf("can't del external action to type action, err: %v", err))
		}
	}

	return errors.Join(errs...)
}

func (c *accessPolicyChangesCron) accessPoliciesGroupIdToTypesActionHandler(ctx context.Context, policies []accesspolicy.AccessPolicy) error {
	var errs []error
	for _, policy := range policies {
		if !policy.IsDel {
			err := c.accessPolicyCache.AddGroupIdToTypeAction(ctx, policy.TypeAction, accesspolicy.GroupIdWithCategory{
				CategoryId: policy.CategoryId,
				GroupId:    policy.GroupId,
				StatusId:   policy.StatusId,
			})
			if err != nil {
				errs = append(errs, fmt.Errorf("can't add group id to type action, err: %v", err))
			}

			continue
		}

		err := c.accessPolicyCache.DelGroupIdToTypeAction(ctx, policy.TypeAction, accesspolicy.GroupIdWithCategory{
			CategoryId: policy.CategoryId,
			GroupId:    policy.GroupId,
			StatusId:   policy.StatusId,
		})
		if err != nil {
			errs = append(errs, fmt.Errorf("can't del group id to type action, err: %v", err))
		}
	}

	return errors.Join(errs...)
}

type AccessPolicyCache interface {
	AddAccessPolicyForActionGroup(ctx context.Context, resource accesspolicy.ResourceCompositeKey, groupID int64) error
	DelAccessPolicyForActionGroup(ctx context.Context, resource accesspolicy.ResourceCompositeKey, groupID int64) error

	AddAccessPolicyForExternalAction(ctx context.Context, resource accesspolicy.ResourceCompositeKey, externalAction string) error
	DelAccessPolicyForExternalAction(ctx context.Context, resource accesspolicy.ResourceCompositeKey, externalAction string) error

	AddExternalActionToTypeAction(ctx context.Context, typeAction string, externalActionWithCategory accesspolicy.ExternalActionWithCategory) error
	DelExternalActionToTypeAction(ctx context.Context, typeAction string, groupIdWithCategory accesspolicy.ExternalActionWithCategory) error

	AddGroupIdToTypeAction(ctx context.Context, typeAction string, groupId accesspolicy.GroupIdWithCategory) error
	DelGroupIdToTypeAction(ctx context.Context, typeAction string, groupId accesspolicy.GroupIdWithCategory) error
}

type AccessPolicyRepo interface {
	GetAccessPolicyFromDB(logID int64) (*accesspolicy.AccessPolicyDBResp, error)
}
