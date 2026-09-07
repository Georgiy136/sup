package repository

import (
	"fmt"
	"net/http"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_api/internal/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_core.git/rest_data"
)

const (
	databaseKey = "support_pgx"
)

type TicketsRepo struct {
}

func NewTicketsRepo() *TicketsRepo {
	return &TicketsRepo{}
}

func (t *TicketsRepo) GetTicketInfoByTicketID(ticketID int64) (*models.TicketInfo, error) {
	getTicketInfoByTicketIDSP := "tickets.tickets_getforchat"

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(databaseKey)
	pg.SetParams(ticketID)
	pg.SetStoredProcedureName(getTicketInfoByTicketIDSP)

	var resp models.TicketInfo
	rd := repository.GetRestDataFromDbAndUnmarshal(pg, &resp)
	if rd.HasError() {
		return nil, fmt.Errorf("read from sp %s: %w", getTicketInfoByTicketIDSP, rest_data.ToError(rd))
	}

	if rd.DataEmpty() {
		return nil, nil
	}

	return &resp, nil
}

func (t *TicketsRepo) AddBandChat(chatID string, ticketID, employeeID int64) error {
	addBandChat := "chats.bandchat_add"

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(databaseKey)
	pg.SetParams(chatID, ticketID, employeeID)
	pg.SetStoredProcedureName(addBandChat)

	rd := repository.GetRestDataFromDb(pg)

	if rd.HasError() {
		if rd.HttpResultCode == http.StatusUnprocessableEntity {
			return fmt.Errorf("DB biz error: %w", &DBBizError{CustomError: rd.Errors[0]})
		}

		return fmt.Errorf("can't execute sp %s: %w", addBandChat, rest_data.ToError(rd))
	}
	return nil
}

func (t *TicketsRepo) GetTicketEmployeeIDsForTag(ticketID int64) (*models.TicketEmployeeIDsForTag, error) {
	getTicketEmployeeIDsForTag := "tickets.tickets_getforchattag"

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(databaseKey)
	pg.SetParams(ticketID)
	pg.SetStoredProcedureName(getTicketEmployeeIDsForTag)

	var resp models.TicketEmployeeIDsForTag
	rd := repository.GetRestDataFromDbAndUnmarshal(pg, &resp)
	if rd.HasError() {
		return nil, fmt.Errorf("read from sp %s: %w", getTicketEmployeeIDsForTag, rest_data.ToError(rd))
	}

	if rd.DataEmpty() {
		return nil, nil
	}

	return &resp, nil
}
