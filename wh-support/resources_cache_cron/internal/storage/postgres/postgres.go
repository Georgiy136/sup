package postgres

import (
	"fmt"

	accesspolicy "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/resources_cache_cron/internal/models/access_policy"
	employeeactiongroups "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/resources_cache_cron/internal/models/employee_action_groups"
	internalactions "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/resources_cache_cron/internal/models/internal_actions"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository"

	jsoniter "github.com/json-iterator/go"
)

type PostgresRepo struct{}

func NewPostgresRepo() *PostgresRepo {
	return &PostgresRepo{}
}

func (*PostgresRepo) GetInternalActionsFromDB(logID int64) (*internalactions.EmployeeResourcesDBResp, error) {
	var res internalactions.EmployeeResourcesDBResp

	db := new(postgresql.PgSpec)
	db.SetDatabaseKey("support_pgx")
	db.SetStoredProcedureName("sync.employeeactions_getforcache")
	db.SetParams(logID)

	bytes, err := repository.ReadBytes(db)
	if err != nil {
		return nil, fmt.Errorf("read from sp %s: %v", "sync.employeeactions_getforcache", err)
	}

	err = jsoniter.Unmarshal(bytes, &res)
	if err != nil {
		return nil, fmt.Errorf("error get internal actions from DB unmarshal error: %v", err)
	}

	return &res, nil
}

func (*PostgresRepo) GetEmployeeActionGroupsFromDB(logID int64) (*employeeactiongroups.EmployeeActionGroupDBResp, error) {
	var res employeeactiongroups.EmployeeActionGroupDBResp

	db := new(postgresql.PgSpec)
	db.SetDatabaseKey("support_pgx")
	db.SetStoredProcedureName("sync.employeegroups_getforcache")
	db.SetParams(logID)

	bytes, err := repository.ReadBytes(db)
	if err != nil {
		return nil, fmt.Errorf("read from sp %s: %v", "sync.employeeactiongroups_getforcache", err)
	}

	err = jsoniter.Unmarshal(bytes, &res)
	if err != nil {
		return nil, fmt.Errorf("error get group actions from DB unmarshal error: %v", err)
	}

	return &res, nil
}

func (*PostgresRepo) GetAccessPolicyFromDB(logID int64) (*accesspolicy.AccessPolicyDBResp, error) {
	var res accesspolicy.AccessPolicyDBResp

	db := new(postgresql.PgSpec)
	db.SetDatabaseKey("support_pgx")
	db.SetStoredProcedureName("sync.categorygroup_getforcache")
	db.SetParams(logID)

	bytes, err := repository.ReadBytes(db)
	if err != nil {
		return nil, fmt.Errorf("read from sp %s: %v", "sync.employeeactions_exporttojson", err)
	}

	err = jsoniter.Unmarshal(bytes, &res)
	if err != nil {
		return nil, fmt.Errorf("error getRecoursesFromDB unmarshal error: %v", err)
	}

	return &res, nil
}
