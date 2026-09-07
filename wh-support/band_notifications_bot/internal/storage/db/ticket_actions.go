package db

import (
	"context"
	"fmt"

	customerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/errors"
	ticketmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/services/ticket/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository"

	jsoniter "github.com/json-iterator/go"
)

type TicketActionsRepo struct{}

func NewTicketActionsRepo() *TicketActionsRepo {
	return &TicketActionsRepo{}
}

func (r *TicketActionsRepo) Approve(ctx context.Context, ticketID, employeeID int64, fields map[string]any, scenarioOrderID int64) error {
	const approveSP = "tickets.tickets_approve"
	if err := r.executeAction(ctx, approveSP, ticketID, employeeID, fields, scenarioOrderID); err != nil {
		return fmt.Errorf("execute %s for ticket %d: %w", approveSP, ticketID, err)
	}
	return nil
}

func (r *TicketActionsRepo) Reject(ctx context.Context, ticketID, employeeID int64, comment string) error {
	const rejectSP = "tickets.tickets_reject"
	if err := r.execute(ctx, rejectSP, ticketID, employeeID, comment); err != nil {
		return fmt.Errorf("execute %s for ticket %d: %w", rejectSP, ticketID, err)
	}
	return nil
}

func (r *TicketActionsRepo) ReturnToStatus(ctx context.Context, ticketID int64, returnStatusID, comment string, employeeID int64) error {
	const returnSP = "tickets.tickets_ticketsreturnstatus"
	if err := r.execute(ctx, returnSP, ticketID, returnStatusID, comment, employeeID); err != nil {
		return fmt.Errorf("execute %s for ticket %d: %w", returnSP, ticketID, err)
	}
	return nil
}

func (r *TicketActionsRepo) Book(ctx context.Context, ticketID, employeeID int64) error {
	const bookSP = "tickets.tickets_booking"
	if err := r.execute(ctx, bookSP, ticketID, employeeID); err != nil {
		return fmt.Errorf("execute %s for ticket %d: %w", bookSP, ticketID, err)
	}
	return nil
}

func (r *TicketActionsRepo) Perform(ctx context.Context, ticketID, employeeID int64, fields map[string]any, scenarioOrderID int64) error {
	const performSP = "tickets.tickets_perform"
	if err := r.executeAction(ctx, performSP, ticketID, employeeID, fields, scenarioOrderID); err != nil {
		return fmt.Errorf("execute %s for ticket %d: %w", performSP, ticketID, err)
	}
	return nil
}

func (r *TicketActionsRepo) Unbook(ctx context.Context, ticketID, employeeID int64) error {
	const unbookSP = "tickets.tickets_unbooking"
	if err := r.execute(ctx, unbookSP, ticketID, employeeID); err != nil {
		return fmt.Errorf("execute %s for ticket %d: %w", unbookSP, ticketID, err)
	}
	return nil
}

func (r *TicketActionsRepo) GetCategoryStatusByStatus(ctx context.Context, categoryID int64, statusID string) (*ticketmodels.CategoryStatusResponse, error) {
	const getCategoryStatusSP = "tickets.categorystatus_getbystatus"

	db := new(postgresql.PgSpec)
	db.SetDatabaseKey(databaseKey)
	db.SetStoredProcedureName(getCategoryStatusSP)
	db.SetParams(categoryID, statusID)

	bytes, err := repository.ReadBytesWithCtx(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("read from sp %s: %w", getCategoryStatusSP, err)
	}

	var resp ticketmodels.CategoryStatusResponse
	if err = jsoniter.Unmarshal(bytes, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal from sp %s, db params: %+v, err :%w", getCategoryStatusSP, db.GetParameters(), err)
	}

	return &resp, nil
}

func (r *TicketActionsRepo) executeAction(ctx context.Context, sp string, ticketID, employeeID int64, fields map[string]any, scenarioOrderID int64) error {
	var rawFields []byte
	if len(fields) > 0 {
		var err error
		rawFields, err = jsoniter.Marshal(fields)
		if err != nil {
			return fmt.Errorf("marshal fields for %s: %w", sp, err)
		}
	}
	return r.execute(ctx, sp, ticketID, rawFields, employeeID, scenarioOrderID)
}

func (r *TicketActionsRepo) execute(ctx context.Context, sp string, params ...interface{}) error {
	db := new(postgresql.PgSpec)
	db.SetDatabaseKey(databaseKey)
	db.SetStoredProcedureName(sp)
	db.SetParams(params...)

	restData := repository.GetRestDataFromDbWithCtx(ctx, db)
	if restData.HasError() {
		dbErr := restData.Errors[0]
		return customerrors.NewDBActionError(restData.HttpResultCode, dbErr.ErrorKey, dbErr.Message, dbErr.Detail)
	}
	return nil
}
