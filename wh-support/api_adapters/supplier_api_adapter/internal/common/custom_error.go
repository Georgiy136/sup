package common

import (
	"errors"

	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_core.git/rest_data"
)

var (
	SupplierNotFound = rest_data.CustomError{
		ErrorKey: "support.std.err_biz.supplier.not_found",
		Message:  "Поставщик не найден",
	}

	ErrResponseNil         = errors.New("response is nil")
	ErrWrongStatusCode     = errors.New("wrong status code")
	ErrTranslationNotFound = errors.New("translation not found")
)
