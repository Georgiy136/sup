package repository

import (
	"fmt"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql"

	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_core.git/rest_data"
)

const defaultScenarioOrderID = 0

type TicketRepository struct{}

func NewTicketRepository() *TicketRepository {
	return &TicketRepository{}
}

func (t *TicketRepository) TicketsProcessing(sp string, params ...interface{}) (*rest_data.RestData, error) {
	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey("support_pg")
	pg.SetStoredProcedureName(sp)
	pg.SetParams(params...)

	return dbExecute(pg)
}

func (t *TicketRepository) TicketsApproveWithoutExt(ticketID, employeeID int64) error {
	const sp = "tickets.tickets_approve"
	_, err := t.TicketsProcessing(sp, ticketID, nil, employeeID, defaultScenarioOrderID)
	if err != nil {
		return fmt.Errorf("can't approve ticket № %d from db: %w", ticketID, err)
	}

	return nil
}

func (t *TicketRepository) TicketsReject(ticketID, employeeID int64, comment string) error {
	const sp = "tickets.tickets_reject"
	_, err := t.TicketsProcessing(sp, ticketID, employeeID, comment)
	if err != nil {
		return fmt.Errorf("can't reject ticket: %w", err)
	}

	return nil
}

func (t *TicketRepository) TicketsBook(ticketID, employeeID int64) error {
	const sp = "tickets.tickets_booking"
	_, err := t.TicketsProcessing(sp, ticketID, employeeID)
	if err != nil {
		return fmt.Errorf("can't booking ticket: %w", err)
	}

	return nil
}

func (t *TicketRepository) TicketsPerformWithoutAddInfo(ticketID int64, employeeID int64) error {
	const sp = "tickets.tickets_perform"
	_, err := t.TicketsProcessing(sp, ticketID, nil, employeeID, defaultScenarioOrderID)
	if err != nil {
		return fmt.Errorf("can't perform ticket: %w", err)
	}

	return nil
}

func (t *TicketRepository) TicketsUnbook(ticketID int64, employeeID int64) error {
	const sp = "tickets.tickets_unbooking"
	_, err := t.TicketsProcessing(sp, ticketID, employeeID)
	if err != nil {
		return fmt.Errorf("can't unbook ticket: %w", err)
	}

	return nil
}
