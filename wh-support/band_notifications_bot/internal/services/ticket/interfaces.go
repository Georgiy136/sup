package ticket

import (
	"context"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/models"
	ticketmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/services/ticket/models"
)

type bandBotClient interface {
	SendNotification(ctx context.Context, recipient models.BandRecipient, msgNotification string, attachments ...models.BandAttachment) (*models.BandSentNotification, error)
	GetPostMessage(ctx context.Context, postID string) (string, error)
	UpdatePost(ctx context.Context, postID, message string, attachments ...models.BandAttachment) error
	GetDirectChannelID(ctx context.Context, employeeID int64) (string, error)
}

type ticketActionRepo interface {
	GetCategoryStatusByStatus(ctx context.Context, categoryID int64, statusID string) (*ticketmodels.CategoryStatusResponse, error)
}

type ticketPostCache interface {
	SaveBandTicketPost(ctx context.Context, post ticketmodels.TicketPost) error
	GetBandTicketPostsByChannel(ctx context.Context, ticketID int64, channelID string) ([]ticketmodels.TicketPost, error)
	DeleteBandTicketPosts(ctx context.Context, ticketID int64, channelID string) error
	DeleteBandTicketPostsByOperation(ctx context.Context, ticketID int64, channelID string, operation string) error
}

type bandActions interface {
	BuildApproveActions(actionCtx models.BandActionContext, opts ticketmodels.ActionOptions) []models.BandAction
	BuildBookActions(actionCtx models.BandActionContext, opts ticketmodels.ActionOptions) []models.BandAction
	BuildPerformActions(actionCtx models.BandActionContext, opts ticketmodels.ActionOptions) []models.BandAction
	CanAttachBandActions(data *ticketmodels.CategoryStatusData) bool
}

type logIdRepo interface {
	GetLogID() int64
	SetLogID(id int64)
}

type notificationRepo interface {
	GetTicketNotifications(logID int64) (*ticketmodels.TicketNotifications, error)
}

type employeeInfoApi interface {
	GetEmployeeName(employeeID int64) (string, error)
}

type modeChecker interface {
	IsTestMode() bool
	IsAllowedInTestMode(employeeID int64, employeesList []int64) bool
	GetTestEmployees() []int64
}
