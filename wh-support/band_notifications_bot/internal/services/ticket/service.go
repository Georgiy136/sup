package ticket

import (
	"context"
	"time"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/consts"
	customerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/errors"
	ticketmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/services/ticket/models"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/storage/memory"
	cronmodels "gitlab.wildberries.ru/wbwh/wh-core/gocore_cron_core.git/cron_core/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	"github.com/sirupsen/logrus"
)

type TicketNotificationsService struct {
	notificationBuilder *notificationBuilder
	ticketRepo          ticketActionRepo
	bandBot             bandBotClient
	ticketPostCache     ticketPostCache
	bandActions         bandActions
	logID               logIdRepo
	notificationRepo    notificationRepo
	modeChecker         modeChecker
	cronCfg             cronmodels.CronCommonCfg
}

func NewTicketNotificationsService(
	notificationBuilder *notificationBuilder,
	ticketRepo ticketActionRepo,
	bandBot bandBotClient,
	ticketPostCache ticketPostCache,
	notificationRepo notificationRepo,
	modeChecker modeChecker,
	bandActions bandActions,
) *TicketNotificationsService {
	return &TicketNotificationsService{
		notificationBuilder: notificationBuilder,
		ticketRepo:          ticketRepo,
		bandBot:             bandBot,
		ticketPostCache:     ticketPostCache,
		bandActions:         bandActions,
		logID:               memory.NewInMemoryLogIdStorage(),
		notificationRepo:    notificationRepo,
		modeChecker:         modeChecker,
	}
}

func (t *TicketNotificationsService) Configure(_ context.Context, cfg configs.Config, cronConf cronmodels.CronCommonCfg) {
	t.cronCfg = cronConf
}

func (t *TicketNotificationsService) ProcessNotifications(ctx context.Context) (sleep time.Duration, err error) {
	currentLogID := t.logID.GetLogID()

	notifications, err := t.notificationRepo.GetTicketNotifications(currentLogID)
	if err != nil {
		return t.cronCfg.CronTimeSleepOnErrorParsed, err
	}
	if len(notifications.Data) == 0 {
		return time.Duration(notifications.SleepSeconds) * time.Second, customerrors.ErrEmptyData
	}

	for idx := range notifications.Data {
		if t.modeChecker.IsTestMode() {
			if !t.modeChecker.IsAllowedInTestMode(notifications.Data[idx].EmployeeID, t.modeChecker.GetTestEmployees()) {
				continue
			}
		}
		if err = t.processNotification(ctx, notifications.Data[idx]); err != nil {
			logrus.Errorf("can't process notification for ticket %d, operation: %s: %v", notifications.Data[idx].TicketID, notifications.Data[idx].TicketInfo.OperationType, err)
		}
	}

	if currentLogID < notifications.LogID {
		currentLogID = notifications.LogID
	}
	t.logID.SetLogID(currentLogID)

	return time.Duration(notifications.SleepSeconds) * time.Second, nil
}

func (t *TicketNotificationsService) processNotification(ctx context.Context, notification ticketmodels.TicketNotification) error {
	var requiredSiteAction bool

	categoryStatusResp, err := t.ticketRepo.GetCategoryStatusByStatus(ctx, notification.TicketInfo.CategoryID, notification.TicketInfo.StatusID)
	if err != nil {
		logrus.Errorf("can't get category status for ticket from db, error: %v", err)
		requiredSiteAction = true
	}

	if categoryStatusResp != nil && len(categoryStatusResp.Data) != 0 {
		if !t.bandActions.CanAttachBandActions(&categoryStatusResp.Data[0]) {
			requiredSiteAction = true
		}
	}

	switch notification.TicketInfo.OperationType {
	case consts.ApproveOperation:
		return t.handleApproveOperation(ctx, notification, requiredSiteAction)
	case consts.ApprovedOperation:
		return t.handleApprovedOperation(ctx, notification)
	case consts.PerformOperation:
		return t.handlePerformOperation(ctx, notification, requiredSiteAction)
	case consts.PerformedOperation:
		return t.handlePerformedOperation(ctx, notification)
	case consts.BookedOperation:
		return t.handleBookedOperation(ctx, notification)
	case consts.UnbookedOperation:
		return t.handleUnbookedOperation(ctx, notification)
	case consts.CompletedOperation:
		return t.handleCompletedOperation(ctx, notification, requiredSiteAction)
	case consts.RejectOperation:
		return t.handleRejectedOperation(ctx, notification, requiredSiteAction)
	case consts.ReturnedOperation:
		return t.handleReturnedOperation(ctx, notification)
	default:
		logrus.Infof("unsupported operation type %s for ticket %d", notification.TicketInfo.OperationType, notification.TicketID)
		return nil
	}
}
