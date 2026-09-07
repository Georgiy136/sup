package bandaction

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
	OpenInteractiveDialog(ctx context.Context, triggerID string, config models.BandInteractiveDialogConfig, state models.BandDialogState) error
	GetEmployeeIDByUserID(ctx context.Context, userID string) (int64, error)
	SendChannelMessage(ctx context.Context, channelID, message string) error
}

type ticketActionRepo interface {
	Approve(ctx context.Context, ticketID, employeeID int64, fields map[string]any, scenarioOrderID int64) error
	Reject(ctx context.Context, ticketID, employeeID int64, comment string) error
	ReturnToStatus(ctx context.Context, ticketID int64, returnStatusID, comment string, employeeID int64) error
	Book(ctx context.Context, ticketID, employeeID int64) error
	Perform(ctx context.Context, ticketID, employeeID int64, fields map[string]any, scenarioOrderID int64) error
	Unbook(ctx context.Context, ticketID, employeeID int64) error
	GetCategoryStatusByStatus(ctx context.Context, categoryID int64, statusID string) (*ticketmodels.CategoryStatusResponse, error)
}

type ticketPostCache interface {
	SaveBandTicketPost(ctx context.Context, post ticketmodels.TicketPost) error
	GetBandTicketPostsByChannel(ctx context.Context, ticketID int64, channelID string) ([]ticketmodels.TicketPost, error)
	DeleteBandTicketPosts(ctx context.Context, ticketID int64, channelID string) error
	DeleteBandTicketPostsByOperation(ctx context.Context, ticketID int64, channelID string, operation string) error
}

type actionSigner interface {
	Sign(parts ...string) string
	Compare(signature string, parts ...string) bool
}
