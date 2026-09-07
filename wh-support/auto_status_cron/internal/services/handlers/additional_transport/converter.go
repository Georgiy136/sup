package additionaltransport

import (
	"fmt"
	"strconv"
	"strings"

	additionaltransportmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/additional_transport/models"
	supporttimeutils "gitlab.wildberries.ru/wbwh/support/utils.git/time_utils"
)

func prepareRequestForLogisticCargo(ticketId int64, in additionaltransportmodels.ExtTicketInfosForCreateLogisticTicket) (additionaltransportmodels.RequestForCreateLogisticCargo, error) {
	elems := make([]string, 0, len(in.Cargo))
	for _, elem := range in.Cargo {
		elems = append(elems, elem.Name)
	}
	contentList := strings.Join(elems, ",")

	dt, err := supporttimeutils.ConvertDateOnlyToRFC3339(in.Date)
	if err != nil {
		return additionaltransportmodels.RequestForCreateLogisticCargo{}, fmt.Errorf("error converting date to RFC3339: %w", err)
	}

	return additionaltransportmodels.RequestForCreateLogisticCargo{
		Cargo: additionaltransportmodels.CargoLogistic{
			TicketId:              ticketId,
			SrcOfficeId:           in.OfficeIdSrc.Id,
			DstOfficeId:           in.OfficeIdDst.Id,
			LoadDate:              dt,
			Description:           in.CargoDesc,
			Content:               contentList,
			ResponsibleEmployeeId: strconv.FormatInt(in.ResponsibleEmployee.Id, 10),
			DepartmentName:        in.Department.Name,
		},
	}, nil
}
