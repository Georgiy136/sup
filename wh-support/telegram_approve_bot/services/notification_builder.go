package services

import (
	"fmt"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/common"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/models"
	ticketinfofmt "gitlab.wildberries.ru/wbwh/support/utils.git/ticket_info_formatter"
)

func (b *Bot) genCommonMsgText(notification models.Notification) string {
	notification = replaceSpecSymbolsNotification(notification)

	msgText := fmt.Sprintf(common.FormatMsgTextNotification,
		notification.TicketID,
		notification.TicketInfo.TicketName,
		notification.TicketInfo.ParentCategoryName,
		notification.TicketInfo.CategoryName,
		notification.TicketInfo.ScenarioName,
		notification.Comments,
	)

	msgText += ticketinfofmt.ParseTicketInfoAndGenMsgText(notification.TicketInfo.Ext)

	createEmployeeName := b.getEmployeeNameByEmployeeID(notification.TicketInfo.CreateEmployeeID)

	if notification.TicketInfo.ResponsibleEmployeeID != nil {
		approveEmployeeName := b.getEmployeeNameByEmployeeID(*notification.TicketInfo.ResponsibleEmployeeID)

		msgText += fmt.Sprintf(common.FormatMsgTextEmployees,
			approveEmployeeName, *notification.TicketInfo.ResponsibleEmployeeID,
			createEmployeeName, notification.TicketInfo.CreateEmployeeID)
	} else {
		msgText += fmt.Sprintf(common.FormatMsgTextEmployeesWithoutApprove,
			createEmployeeName,
			notification.TicketInfo.CreateEmployeeID)
	}

	return msgText
}

func (b *Bot) genShortMsgTex(notification models.Notification) string {
	notification = replaceSpecSymbolsNotification(notification)

	msgText := fmt.Sprintf(common.FormatMsgTextNotification,
		notification.TicketID,
		notification.TicketInfo.TicketName,
		notification.TicketInfo.ParentCategoryName,
		notification.TicketInfo.CategoryName,
		notification.TicketInfo.ScenarioName,
		notification.Comments,
	)

	msgText += common.TooLongMessageWarning

	createEmployeeName := b.getEmployeeNameByEmployeeID(notification.TicketInfo.CreateEmployeeID)

	if notification.TicketInfo.ResponsibleEmployeeID != nil {
		approveEmployeeName := b.getEmployeeNameByEmployeeID(*notification.TicketInfo.ResponsibleEmployeeID)

		msgText += fmt.Sprintf(common.FormatMsgTextEmployees,
			approveEmployeeName, *notification.TicketInfo.ResponsibleEmployeeID,
			createEmployeeName, notification.TicketInfo.CreateEmployeeID)
	} else {
		msgText += fmt.Sprintf(common.FormatMsgTextEmployeesWithoutApprove,
			createEmployeeName,
			notification.TicketInfo.CreateEmployeeID)
	}

	return msgText
}

func (b *Bot) genCommonMsgTextForTicketCreator(notification models.Notification) string {
	notification = replaceSpecSymbolsNotification(notification)

	msgText := fmt.Sprintf(common.FormatMsgTextNotificationForCreator,
		notification.TicketID,
		notification.TicketInfo.TicketName,
		notification.TicketInfo.StatusDescription,
	)

	return msgText
}
