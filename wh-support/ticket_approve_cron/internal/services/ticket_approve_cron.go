package services

import (
	"context"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_approve_cron/internal/models"
	cron_models "gitlab.wildberries.ru/wbwh/wh-core/gocore_cron_core.git/cron_core/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_validators.git"

	"github.com/go-playground/validator/v10"
)

type TicketApproveCron struct {
	cronCfg   cron_models.CronCommonCfg
	validator *validator.Validate
	logID     logIDStore

	repo ticketsRepo

	clientTG tgBotClient
}

func NewTicketApproveCron(clientTG tgBotClient, repo ticketsRepo, logID logIDStore) *TicketApproveCron {
	return &TicketApproveCron{
		repo:     repo,
		clientTG: clientTG,
		logID:    logID,
	}
}

func (w *TicketApproveCron) Start(ctx context.Context, conf configs.Config, cronCfg cron_models.CronCommonCfg) {
	w.cronCfg = cronCfg

	valid := validator.New()
	gocore_validators.InitializeCustomValidatorsV10(valid)
	w.validator = valid

	w.clientTG.Configure(ctx, conf)

	w.sendNotificationsToTGBot()
}

type ticketsRepo interface {
	GetNotifications(logID int64) (*models.DBTicketsForNotificationResponse, int64, error)
}

type logIDStore interface {
	GetLastLogID() int64
	SetLogID(logID int64)
}
