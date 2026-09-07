package repository

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/common"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql"
)

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (u *UserRepository) AddTgUser(chatID, employeeID int64) error {
	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey("support_pg")
	pg.SetStoredProcedureName("chats.sdtgchat_addtguser")
	pg.SetParams(chatID, employeeID)

	_, err := dbExecute(pg)
	if err != nil {
		return fmt.Errorf("can't add telegram user: %w", err)
	}

	return nil
}

type employees struct {
	EmpID int64 `json:"employee_id"`
}

func (u *UserRepository) GetEmpByUserID(userID int64) (int64, error) {
	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey("support_pg")
	pg.SetStoredProcedureName("chats.employee_getbysdtgchatid")
	pg.SetParams(userID)

	data, err := dbExecute(pg)
	if err != nil {
		return 0, fmt.Errorf("can't get emp_id by tg_user_id: %w", err)
	}

	if data.DataEmpty() {
		return 0, fmt.Errorf("can't get employeeID: %w", common.ErrEmptyData)
	}

	var emp []employees
	err = jsoniter.Unmarshal(data.GetData(), &emp)
	if err != nil {
		return 0, fmt.Errorf("can't unmarshal data from db: %w", err)
	}

	if len(emp) == 0 {
		return 0, fmt.Errorf("can't get employeeID: %w", common.ErrEmptyData)
	}

	return emp[0].EmpID, nil
}
