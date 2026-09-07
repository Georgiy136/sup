package repository

import (
	"context"
	"fmt"
	"time"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/consts"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/localization"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/sentry"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_core.git/rest_data"

	jsoniter "github.com/json-iterator/go"
)

const (
	defaultScenarioOrderID = 0
	databaseKey            = "support_pgx"
)

type TicketsRepo struct{}

func NewTicketsRepo() *TicketsRepo { return &TicketsRepo{} }

func (t *TicketsRepo) PerformTicket(ctx context.Context, ticketID int64, ext []byte) error {
	return t.PerformTicketWithScenario(ctx, ticketID, ext, defaultScenarioOrderID)
}

func (t *TicketsRepo) PerformTicketWithScenario(ctx context.Context, ticketID int64, ext []byte, scenarioOrderID int64) error {
	return t.performTicket(ctx, ticketID, ext, scenarioOrderID, nil, nil)
}

func (t *TicketsRepo) PerformTicketWithScenarioAndErrComment(ctx context.Context, ticketID int64, ext []byte, scenarioOrderID int64, errComment []byte, locMsgs *localization.LocalizedErrors) error {
	return t.performTicket(ctx, ticketID, ext, scenarioOrderID, errComment, locMsgs)
}

func (t *TicketsRepo) PerformTicketWithErrComment(ctx context.Context, ticketID int64, ext []byte, errComment []byte, locMsgs *localization.LocalizedErrors) error {
	return t.performTicket(ctx, ticketID, ext, defaultScenarioOrderID, errComment, locMsgs)
}

func (t *TicketsRepo) performTicket(ctx context.Context, ticketID int64, ext []byte, scenarioOrderID int64, errComment []byte, locMsgs *localization.LocalizedErrors) (err error) {
	const procedureName = "tickets.tickets_perform"
	span := sentry.StartDBQuerySpan(ctx, procedureName)
	defer func() {
		sentry.FinishDBQuerySpan(span, err)
	}()

	var locMsgsBytes []byte
	if locMsgs != nil {
		locMsgsBytes, err = jsoniter.Marshal(locMsgs)
		if err != nil {
			return fmt.Errorf("can't marshal localization for perform ticket: %w", err)
		}
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(databaseKey)
	pg.SetParams(ticketID, ext, consts.SystemEmployeeID, scenarioOrderID, errComment, locMsgsBytes)
	pg.SetStoredProcedureName(procedureName)
	if err = dbExecute(ctx, pg); err != nil {
		return fmt.Errorf("can't execute request perform ticket to db, err: %w", err)
	}
	return nil
}

func (t *TicketsRepo) RejectTicket(ctx context.Context, ticketID int64, comment string, locMsgs *localization.LocalizedErrors) (err error) {
	const procedureName = "tickets.tickets_reject"
	span := sentry.StartDBQuerySpan(ctx, procedureName)
	defer func() {
		sentry.FinishDBQuerySpan(span, err)
	}()

	rejectCommentsLocBytes, err := jsoniter.Marshal(locMsgs)
	if err != nil {
		return fmt.Errorf("can't marshal reject ticket with localization: %w", err)
	}
	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(databaseKey)
	pg.SetParams(ticketID, consts.SystemEmployeeID, comment, nil, rejectCommentsLocBytes, nil)
	pg.SetStoredProcedureName(procedureName)
	if err = dbExecute(ctx, pg); err != nil {
		return fmt.Errorf("can't execute request reject ticket to db: %w", err)
	}
	return nil
}

func (t *TicketsRepo) RejectTicketWithEmployee(ctx context.Context, ticketID int64, employeeID int64, comment string, locMsgs *localization.LocalizedErrors) (err error) {
	const procedureName = "tickets.tickets_reject"
	span := sentry.StartDBQuerySpan(ctx, procedureName)
	defer func() {
		sentry.FinishDBQuerySpan(span, err)
	}()

	rejectCommentsLocBytes, err := jsoniter.Marshal(locMsgs)
	if err != nil {
		return fmt.Errorf("can't marshal reject ticket with localization: %w", err)
	}
	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(databaseKey)
	pg.SetParams(ticketID, employeeID, comment, nil, rejectCommentsLocBytes, nil)
	pg.SetStoredProcedureName(procedureName)
	if err = dbExecute(ctx, pg); err != nil {
		return fmt.Errorf("can't execute request reject ticket to db: %w", err)
	}
	return nil
}

func (t *TicketsRepo) CancelPreReject(ctx context.Context, ticketID int64, preRejectAnswer map[string]any) (err error) {
	const procedureName = "tickets.tickets_prereject_cancel"
	span := sentry.StartDBQuerySpan(ctx, procedureName)
	defer func() {
		sentry.FinishDBQuerySpan(span, err)
	}()

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(databaseKey)
	pg.SetParams(ticketID, consts.SystemEmployeeID, preRejectAnswer)
	pg.SetStoredProcedureName(procedureName)
	if err = dbExecute(ctx, pg); err != nil {
		return fmt.Errorf("can't execute request cancel prereject ticket to db: %w", err)
	}
	return nil
}

func (t *TicketsRepo) RejectTicketWithValues(ctx context.Context, ticketID int64, comment string, locMsgs *localization.LocalizedErrors, values map[string]any) (err error) {
	const procedureName = "tickets.tickets_reject"
	span := sentry.StartDBQuerySpan(ctx, procedureName)
	defer func() {
		sentry.FinishDBQuerySpan(span, err)
	}()

	rejectCommentsLocBytes, err := jsoniter.Marshal(locMsgs)
	if err != nil {
		return fmt.Errorf("can't marshal reject ticket with localization: %w", err)
	}
	rejectValues, err := jsoniter.Marshal(values)
	if err != nil {
		return fmt.Errorf("can't marshal reject ticket with values: %w", err)
	}
	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(databaseKey)
	pg.SetParams(ticketID, consts.SystemEmployeeID, comment, rejectValues, rejectCommentsLocBytes, rejectValues)
	pg.SetStoredProcedureName(procedureName)
	if err = dbExecute(ctx, pg); err != nil {
		return fmt.Errorf("can't execute request reject ticket to db: %w", err)
	}
	return nil
}

func (t *TicketsRepo) GetTicketsWithAutoStatus(ctx context.Context) (tickets *models.DataTickets, err error) {
	const procedureName = "cron.tickets_getforautoapi"
	span := sentry.StartDBQuerySpan(ctx, procedureName)
	defer func() {
		sentry.FinishDBQuerySpan(span, err)
	}()

	db := new(postgresql.PgSpec)
	db.SetDatabaseKey(databaseKey)
	db.SetStoredProcedureName(procedureName)
	bytes, err := repository.ReadBytes(db)
	if err != nil {
		return nil, fmt.Errorf("read from sp %s: %w", procedureName, err)
	}
	var response models.DataTickets
	if err = jsoniter.Unmarshal(bytes, &response); err != nil {
		return nil, fmt.Errorf("unmarshal from sp %s bytes: %w", procedureName, err)
	}
	if len(response.Tickets) == 0 {
		return nil, nil
	}
	return &response, nil
}

func (t *TicketsRepo) UpdateNextCheckAt(ctx context.Context, ticketID int64, nextCheckDt time.Time) (err error) {
	const procedureName = "tickets.tickets_updnextbycheckdt"
	span := sentry.StartDBQuerySpan(ctx, procedureName)
	defer func() {
		sentry.FinishDBQuerySpan(span, err)
	}()

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(databaseKey)
	pg.SetParams(ticketID, nextCheckDt.Format(time.RFC3339Nano))
	pg.SetStoredProcedureName(procedureName)
	if err = dbExecute(ctx, pg); err != nil {
		return fmt.Errorf("can't execute request update next check at to db: %w", err)
	}
	return nil
}

func (t *TicketsRepo) ReturnToStatus(ctx context.Context, ticketID int64, returnStatusID, returnComment string, employeeID int64) (err error) {
	const procedureName = "tickets.tickets_ticketsreturnstatus"
	span := sentry.StartDBQuerySpan(ctx, procedureName)
	defer func() {
		sentry.FinishDBQuerySpan(span, err)
	}()

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(databaseKey)
	pg.SetParams(ticketID, returnStatusID, returnComment, employeeID)
	pg.SetStoredProcedureName(procedureName)
	if err = dbExecute(ctx, pg); err != nil {
		return fmt.Errorf("can't execute request return ticket to status to db: %w", err)
	}
	return nil
}

func dbExecute(ctx context.Context, pg *postgresql.PgSpec) error {
	data := repository.GetRestDataFromDbWithCtx(ctx, pg)
	if data.HasError() {
		return fmt.Errorf("DB: %w", &DBExecError{CustomError: data.Errors[0]})
	}
	return nil
}

type DBExecError struct{ rest_data.CustomError }

func (e *DBExecError) Error() string {
	return fmt.Sprintf("code: %s, msg: %s", e.CustomError.ErrorKey, e.CustomError.Message)
}
