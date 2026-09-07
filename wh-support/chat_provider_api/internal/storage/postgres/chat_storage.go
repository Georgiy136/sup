package postgres

import (
	"fmt"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_provider_api/internal/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_core.git/rest_data"
)

const supportPgDatabaseKey = "support_pgx"

type ChatStorage struct{}

func NewChatStorage() *ChatStorage {
	return &ChatStorage{}
}

func (s *ChatStorage) GetTicketInfo(ticketID int64) (*models.TicketInfo, error) {
	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(supportPgDatabaseKey)
	pg.SetStoredProcedureName("tickets.tickets_getforchat")
	pg.SetParams(ticketID)

	info := models.TicketInfo{}
	rd := repository.GetRestDataFromDbAndUnmarshal(pg, &info)
	if rd.HasError() {
		return nil, fmt.Errorf("get ticket info db error: %w", rest_data.ToError(rd))
	}
	if rd.DataEmpty() {
		return nil, nil
	}

	return &info, nil
}
