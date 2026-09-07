package service

import (
	"errors"
	"fmt"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/models"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/utils"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_core.git/rest_data"
)

type bodyCreateTicketWithReserveTicketID struct {
	categoryID      int64
	employeeID      int64
	infoForCreate   []byte
	scenarioOrderID *int64
	ticketName      string
	ticketID        int64
}

type bodyCreateTicketWithPublicCategory struct {
	categoryID       int64
	employeeID       int64
	infoForCreate    []byte
	scenarioOrderID  *int64
	ticketName       string
	isPublicCategory bool
}

func createTicketWithReserveTicketID(body bodyCreateTicketWithReserveTicketID) rest_data.RestData {
	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(body.categoryID, body.employeeID, utils.ToJSONObjectOrDefault(body.infoForCreate), body.scenarioOrderID, body.ticketName, body.ticketID)
	pg.SetStoredProcedureName("tickets.tickets_create_v1")

	return repository.GetRestDataFromDb(pg)
}

func createTicketWithPublicCategory(body bodyCreateTicketWithPublicCategory) rest_data.RestData {
	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(body.categoryID, body.employeeID, utils.ToJSONObjectOrDefault(body.infoForCreate), body.scenarioOrderID, body.ticketName, body.isPublicCategory)
	pg.SetStoredProcedureName("tickets.tickets_create_v2")

	return repository.GetRestDataFromDb(pg)
}

func reserveTicketID() (int64, error) {
	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetStoredProcedureName("tickets.tickets_getid")

	var ticketIDFromDB models.TiketsGetIDDataFromDB
	rd := repository.GetRestDataFromDbAndUnmarshal(pg, &ticketIDFromDB)
	if rd.HasError() {
		return 0, fmt.Errorf("error get ticket id from db: %w", rest_data.ToError(rd))
	}
	if rd.DataEmpty() {
		return 0, errors.New("can't create ticket id in db")
	}
	return ticketIDFromDB.TiketID, nil
}

func getCategoryStatusInfoForCreate(categoryID int64) ([]models.StatusInfo, error) {
	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(categoryID)
	pg.SetStoredProcedureName("tickets.categorystatus_getbystatus")

	var categoryStatusInfo []models.StatusInfo
	rd := repository.GetRestDataFromDbAndUnmarshal(pg, &categoryStatusInfo)
	if rd.HasError() {
		return nil, fmt.Errorf("can't get category status: %w", rest_data.ToError(rd))
	}
	return categoryStatusInfo, nil
}

func getCategoryStatusInfo(categoryID int64, statusID string) ([]models.StatusInfo, error) {
	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(categoryID, statusID)
	pg.SetStoredProcedureName("tickets.categorystatus_getbystatus")

	var categoryStatusInfo []models.StatusInfo
	rd := repository.GetRestDataFromDbAndUnmarshal(pg, &categoryStatusInfo)
	if rd.HasError() {
		return nil, fmt.Errorf("can't get category status: %w", rest_data.ToError(rd))
	}
	return categoryStatusInfo, nil
}

func getTicketCommonInfo(ticketID int64) (*models.TicketCommonInfo, error) {
	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(ticketID)
	pg.SetStoredProcedureName("tickets.tickets_getforchat")

	var ticketCommonInfo models.TicketCommonInfo
	rd := repository.GetRestDataFromDbAndUnmarshal(pg, &ticketCommonInfo)
	if rd.HasError() {
		return &ticketCommonInfo, fmt.Errorf("can't get ticket info: %w", rest_data.ToError(rd))
	}
	if rd.DataEmpty() {
		return nil, nil
	}
	return &ticketCommonInfo, nil
}
