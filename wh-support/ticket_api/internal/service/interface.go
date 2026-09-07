package service

import (
	"context"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/models"
)

type RedisClientInterface interface {
	GetAccessPolicy(ctx context.Context, employeeID int64, typeAction string) (*models.AccessPolicyResult, error)
	GetAccessPolicies(ctx context.Context, employeeID int64, typeActions ...string) (*models.AccessPoliciesResult, error)

	GetAccessGroupsByEmployee(ctx context.Context, employeeID int64) ([]int64, error)
	GetResourceAccessPolicy(ctx context.Context, resource models.ResourceCompositeKey) (*models.ResourceAccessPolicy, error)

	GetChatActivities(ctx context.Context, employeeID int64, ticketIDs []int64) (map[int64]*models.ChatActivityData, error)
}

type ResourceEmployeeAccessClientInterface interface {
	GetAccessActionsByEmployeeID(employeeID int64) ([]string, error)
}

type FileManagerClientInterface interface {
	InitUpload(ctx context.Context, employeeID int64, body models.FileManagerInitUploadRequest) (*models.FileManagerInitUploadResponse, error)
	ConfirmUpload(ctx context.Context, employeeID int64, body models.FileManagerConfirmUploadRequest) error
	GetDownloadURL(ctx context.Context, employeeID, fileID int64) (*models.FileManagerDownloadURLResponse, error)
	CancelUpload(ctx context.Context, employeeID int64, body models.FileManagerCancelUploadRequest) error
	AttachFilesToTicket(ctx context.Context, employeeID int64, body models.AttachFilesToTicketRequest) error
	GetFileInfo(ctx context.Context, employeeID int64, body models.GetFileInfoRequest) (*models.GetFileInfoResponse, error)
}
