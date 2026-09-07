package repository

import (
	"fmt"

	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_core.git/rest_data"
)

type DBBizError struct {
	rest_data.CustomError
}

func (e *DBBizError) Error() string {
	return fmt.Sprintf("code: %s, msg: %s", e.CustomError.ErrorKey, e.CustomError.Message)
}
