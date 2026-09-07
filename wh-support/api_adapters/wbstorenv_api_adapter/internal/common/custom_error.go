package common

import (
	"errors"

	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_core.git/rest_data"
)

var (
	WhNotFound = rest_data.CustomError{
		ErrorKey: "support.std.err_biz.wh.not_found",
		Message:  "Блок не найден или удален.",
	}

	WhHasActiveVirtual = rest_data.CustomError{
		ErrorKey: "support.std.err_biz.wh.has_active_virtual",
		Message:  "Нельзя деактивировать блок. Он имеет активные виртуальные блоки.",
	}

	WhAlreadyInactive = rest_data.CustomError{
		ErrorKey: "support.std.err_biz.wh.already_inactive",
		Message:  "Блок уже деактивирован.",
	}

	ErrResponseNil         = errors.New("response is nil")
	ErrWrongStatusCode     = errors.New("wrong status code")
	ErrTranslationNotFound = errors.New("translation not found")
)
