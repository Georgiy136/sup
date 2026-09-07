package templates

import (
	"fmt"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/consts"
)

const (
	actionStatusTextFormat = "\n\n⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯\n" + "_%s_"

	approveStatusTextFormat   = "Сотрудник %d подтвердил(а) заявку ✅"
	bookedStatusTextFormat    = "Сотрудник %d взял(а) заявку в работу 💼"
	performedStatusTextFormat = "Сотрудник %d исполнил(а) заявку ✅"
	rejectedStatusTextFormat  = "Сотрудник %d отклонил(а) заявку ❌"
	completedStatusTextFormat = "Сотрудник %d выполнил(а) заявку ✅"
	returnedStatusTextFormat  = "Сотрудник %d вернул(а) заявку на предыдущий статус 🔶"

	unknownApproveStatusText   = "Заявка была подтверждена неизвестным пользователем ✅"
	unknownBookedStatusText    = "Заявка была взята в работу неизвестным пользователем 💼"
	unknownPerformedStatusText = "Заявка исполнена неизвестным пользователем ✅"
	unknownCompletedStatusText = "Заявка выполнена неизвестным пользователем ✅"
	unknownRejectedStatusText  = "Заявка была отклонена неизвестным пользователем ❌"
	unknownReturnedStatusText  = "Заявка была возвращена на предыдущий статус неизвестным пользователем 🔶"
)

// GetTicketStatusTextByOperationType возвращает форматированный текст статуса действия по заявке.
func GetTicketStatusTextByOperationType(operationType string, employeeID *int64) string {
	switch operationType {
	case consts.ApprovedOperation:
		if employeeID == nil {
			return fmt.Sprintf(actionStatusTextFormat, unknownApproveStatusText)
		}
		return fmt.Sprintf(actionStatusTextFormat, fmt.Sprintf(approveStatusTextFormat, *employeeID))
	case consts.BookedOperation:
		if employeeID == nil {
			return fmt.Sprintf(actionStatusTextFormat, unknownBookedStatusText)
		}
		return fmt.Sprintf(actionStatusTextFormat, fmt.Sprintf(bookedStatusTextFormat, *employeeID))
	case consts.PerformedOperation:
		if employeeID == nil {
			return fmt.Sprintf(actionStatusTextFormat, unknownPerformedStatusText)
		}
		return fmt.Sprintf(actionStatusTextFormat, fmt.Sprintf(performedStatusTextFormat, *employeeID))
	case consts.CompletedOperation:
		if employeeID == nil {
			return fmt.Sprintf(actionStatusTextFormat, unknownCompletedStatusText)
		}
		return fmt.Sprintf(actionStatusTextFormat, fmt.Sprintf(completedStatusTextFormat, *employeeID))
	case consts.RejectOperation:
		if employeeID == nil {
			return fmt.Sprintf(actionStatusTextFormat, unknownRejectedStatusText)
		}
		return fmt.Sprintf(actionStatusTextFormat, fmt.Sprintf(rejectedStatusTextFormat, *employeeID))
	case consts.ReturnedOperation:
		if employeeID == nil {
			return fmt.Sprintf(actionStatusTextFormat, unknownReturnedStatusText)
		}
		return fmt.Sprintf(actionStatusTextFormat, fmt.Sprintf(returnedStatusTextFormat, *employeeID))
	default:
		return ""
	}
}
