package db

import (
	"fmt"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_core.git/rest_data"
)

type EmployeesRepo struct {
}

func NewEmployeesRepo() *EmployeesRepo {
	return &EmployeesRepo{}
}

func (r *EmployeesRepo) GetAllowedEmployeesByGroup(groupID int64) ([]models.GroupDataResponse, error) {
	getNotificationsSP := "hr.employeegroups_getbygroup"

	db := new(postgresql.PgSpec)
	db.SetDatabaseKey(databaseKey)
	db.SetParams(groupID)
	db.SetStoredProcedureName(getNotificationsSP)

	var respDB []models.GroupDataResponse
	rd := repository.GetRestDataFromDbAndUnmarshal(db, &respDB)
	if rd.HasError() {
		return nil, fmt.Errorf("get allowed employees from db: %w", rest_data.ToError(rd))
	}
	if rd.DataEmpty() {
		return []models.GroupDataResponse{}, nil
	}
	return respDB, nil
}
