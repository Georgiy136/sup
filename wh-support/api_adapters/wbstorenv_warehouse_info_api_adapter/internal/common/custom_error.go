package common

import "gitlab.wildberries.ru/wbwh/wh-core/gocore_service_core.git/rest_data"

var (
	OfficeNotFound = rest_data.CustomError{
		ErrorKey: "deliverysettings.std.err_crit.office.not_found",
		Message:  "Офис не найден",
	}
)
