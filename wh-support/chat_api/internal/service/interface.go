package service

import (
	"context"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_api/internal/models"
)

type ticketsRepo interface {
	GetTicketInfoByTicketID(ticketID int64) (*models.TicketInfo, error)
	AddBandChat(chatID string, ticketID, employeeID int64) error
	GetTicketEmployeeIDsForTag(ticketID int64) (*models.TicketEmployeeIDsForTag, error)
}

type chatCache interface {
	AddLastChatViewByUser(ctx context.Context, employeeID, ticketID int64, now string) error
	GetAccessPolicies(ctx context.Context, employeeID int64, typeActions ...string) (*models.AccessPoliciesResult, error)
}

type chatProviderApiClient interface {
	CreateNewChat(ctx context.Context, employeeID int64, body models.CreateNewChatRequest) (*models.CreateNewChatResponse, error)
	CreatePost(ctx context.Context, employeeID int64, body models.CreatePostRequest) (*models.CreatePostResponse, error)
	GetChatByChatID(ctx context.Context, employeeID int64, body models.GetChatByChatIDRequest) (*models.GetChatByChatIDResponse, error)
	GetHistoryChatByChatID(ctx context.Context, employeeID int64, body models.GetHistoryChatByChatIDRequest) (*models.GetHistoryChatByChatIDResponse, error)
}

type resourceEmployeeAccessApiClient interface {
	GetAccessActionsByEmployeeID(employeeID int64) ([]string, error)
}

type employeeInfoApiClient interface {
	GetEmployeesFullName(chEmployeeID int64, employeeIDs []int64) ([]models.EmployeeInfo, error)
}
