package repository

import (
	"errors"
	"fmt"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/common"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_core.git/rest_data"
)

func dbExecute(pg *postgresql.PgSpec) (*rest_data.RestData, error) {
	data := repository.GetRestDataFromDb(pg)
	if data.HasError() {
		return nil, parseDBError(fmt.Errorf("DB: %w", &common.DBExecError{CustomError: data.Errors[0]}))
	}

	return &data, nil
}

func parseDBError(err error) error {
	var custom *common.DBExecError
	defaultErr := common.ErrGetDataFromDB
	if errors.As(err, &custom) {
		switch custom.ErrorKey {
		case common.ErrChatAlreadyExists.Error():
			return fmt.Errorf("database error: %w, details: %s", common.ErrChatAlreadyExists, err.Error())
		case common.ErrUserAlreadyExists.Error():
			return fmt.Errorf("database error: %w, details: %s", common.ErrUserAlreadyExists, err.Error())
		case common.ErrHasNoRights.Error():
			return fmt.Errorf("database error: %w, details: %s", common.ErrHasNoRights, err.Error())
		case common.ErrWrongTicketModel.Error():
			return fmt.Errorf("database error: %w, details: %s", common.ErrWrongTicketModel, err.Error())
		case common.ErrTicketNotNeedApprove.Error():
			return fmt.Errorf("database error: %w, details: %s", common.ErrTicketNotNeedApprove, err.Error())
		case common.ErrTicketNotFound.Error():
			return fmt.Errorf("database error: %w, details: %s", common.ErrTicketNotFound, err.Error())
		default:
			return fmt.Errorf("database error: %w, details: %s", errors.New(custom.ErrorKey), err.Error())
		}
	}

	return defaultErr
}
