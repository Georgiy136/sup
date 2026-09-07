package common

import (
	"errors"

	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_core.git/rest_data"
)

var (
	OfficeNotFound = rest_data.CustomError{
		ErrorKey: "deliverysettings.std.err_crit.office.not_found",
		Message:  "Офис не найден",
	}

	OfficeExists = rest_data.CustomError{
		ErrorKey: "std.std.err_biz.office.already_exists",
		Message:  "Офис уже существует в складской системе",
	}

	ErrResponseNil         = errors.New("response is nil")
	ErrWrongStatusCode     = errors.New("wrong status code")
	ErrTranslationNotFound = errors.New("translation not found")
)
