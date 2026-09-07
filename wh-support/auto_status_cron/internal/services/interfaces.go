package services

import (
	"context"
	"time"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/localization"
)

type HandlerTicketsRepo interface {
	PerformTicket(ctx context.Context, ticketID int64, ext []byte) error
	PerformTicketWithScenario(ctx context.Context, ticketID int64, ext []byte, scenarioOrderID int64) error
	PerformTicketWithScenarioAndErrComment(ctx context.Context, ticketID int64, ext []byte, scenarioOrderID int64, errComment []byte, locMsgs *localization.LocalizedErrors) error
	PerformTicketWithErrComment(ctx context.Context, ticketID int64, ext []byte, errComment []byte, locMsgs *localization.LocalizedErrors) error
	RejectTicket(ctx context.Context, ticketID int64, comment string, locMsgs *localization.LocalizedErrors) error
	RejectTicketWithEmployee(ctx context.Context, ticketID int64, employeeID int64, comment string, locMsgs *localization.LocalizedErrors) error
	RejectTicketWithValues(ctx context.Context, ticketID int64, comment string, locMsgs *localization.LocalizedErrors, values map[string]any) error
	CancelPreReject(ctx context.Context, ticketID int64, preRejectAnswer map[string]any) error
	UpdateNextCheckAt(ctx context.Context, ticketID int64, nextCheckDt time.Time) error
	ReturnToStatus(ctx context.Context, ticketID int64, returnStatusID string, returnComment string, employeeID int64) error
}
