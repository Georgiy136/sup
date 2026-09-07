package realtime

import (
	"context"
	"time"

	mattermost "github.com/mattermost/mattermost/server/public/model"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_provider_api/internal/models"
)

type BandClient interface {
	GetBotID() string
	NewWebSocketClient() (*mattermost.WebSocketClient, error)
	GetFileInfosForPost(ctx context.Context, postID string) ([]*mattermost.FileInfo, error)
}

type Centrifugo interface {
	Publish(ctx context.Context, channel string, data any) error
}

type ChatStorage interface {
	GetTicketInfo(ticketID int64) (*models.TicketInfo, error)
}

type ActivityStorage interface {
	SetTicketLastActivity(ctx context.Context, ticketID int64, at time.Time) error
}

type AccessStorage interface {
	GetResourceAccessPolicies(ctx context.Context, categoryID int64, statusID string, typeActions ...string) (map[string][]int64, error)
	GetEmployeeAccessGroups(ctx context.Context, employeeID int64) ([]int64, error)
}
