package ticket

import (
	"fmt"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/consts"
	ticketmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/services/ticket/models"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/services/ticket/templates"
	ticketinfofmt "gitlab.wildberries.ru/wbwh/support/utils.git/ticket_info_formatter"

	"github.com/sirupsen/logrus"
)

var formatMsgTextNotification = `📌 Уведомление по заявке *№ %d* — [Wh Support](%s)

*%s*
⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯
*Категория заявки:*
%s

*Тип заявки:*
%s

*Сценарий:*
%s

*Комментарий:*
%s

*Информация по заявке:*
`

var formatShortMsgTextNotificationForCreator = `📌 Уведомление по заявке *№ %d* — «*%s*». 
Заявка прошла статус: *%s*.
Подробную информацию по заявке смотрите на [Wh Support](` + consts.SupportSiteURL + `/my-tickets) в разделе «Мои заявки».
`

var formatMsgTextTicketStatusReturned = `📌 Уведомление по заявке *№ %d* — «*%s*».

*Комментарий:* %s

Подробная информация на сайте [Wh Support](` + consts.SupportSiteURL + `)
`

var formatMsgTextEmployees = `
*Ответственный:*
%s (%d)

*Заявитель:*
%s (%d)
`

var formatMsgTextEmployeesWithoutApprove = `
*Ответственный:*
Заявка не проходила согласование

*Заявитель:*
%s (%d)
`

type notificationBuilder struct {
	employeeInfoApi employeeInfoApi
}

func NewNotificationBuilder(employeeInfoApi employeeInfoApi) *notificationBuilder {
	return &notificationBuilder{employeeInfoApi: employeeInfoApi}
}

func (b *notificationBuilder) GenMsgNotification(notification ticketmodels.TicketNotification, requiredSiteAction bool) (string, error) {
	msgText := fmt.Sprintf(formatMsgTextNotification,
		notification.TicketID,
		consts.SupportSiteURL,
		notification.TicketInfo.TicketName,
		notification.TicketInfo.ParentCategoryName,
		notification.TicketInfo.CategoryName,
		notification.TicketInfo.ScenarioName,
		notification.Comments,
	)

	addTextMsg := ticketinfofmt.ParseTicketInfoAndGenMsgText(notification.TicketInfo.Ext)
	msgText += addTextMsg

	createEmployeeName, err := b.getEmployeeName(notification.TicketInfo.CreateEmployeeID)
	if err != nil || createEmployeeName == "" {
		logrus.Infof("can't get create employee name: %v", err)
		createEmployeeName = templates.EmptyEmployeeName
	}

	if notification.TicketInfo.ResponsibleEmployeeID != nil {
		approveEmployeeName, err := b.getEmployeeName(*notification.TicketInfo.ResponsibleEmployeeID)
		if err != nil || approveEmployeeName == "" {
			logrus.Infof("can't get create employee name: %v", err)
			approveEmployeeName = templates.EmptyEmployeeName
		}

		msgText += fmt.Sprintf(formatMsgTextEmployees,
			approveEmployeeName, *notification.TicketInfo.ResponsibleEmployeeID,
			createEmployeeName, notification.TicketInfo.CreateEmployeeID)
	} else {
		msgText += fmt.Sprintf(formatMsgTextEmployeesWithoutApprove,
			createEmployeeName,
			notification.TicketInfo.CreateEmployeeID)
	}

	switch notification.TicketInfo.OperationType {
	case consts.ApproveOperation:
		if requiredSiteAction {
			msgText += templates.ApproveOnSiteTicketMessage
		}
	case consts.PerformOperation:
		if requiredSiteAction {
			msgText += templates.PerformOnSiteTicketMessage
		}
	case consts.RejectOperation:
		if notification.TicketInfo.RejectedEmployeeID == nil {
			msgText += templates.RejectedUnknownUser
		} else {
			rejectEmployeeName, err := b.getEmployeeName(*notification.TicketInfo.RejectedEmployeeID)
			if err != nil || rejectEmployeeName == "" {
				logrus.Infof("can't get reject employee name: %v", err)
				rejectEmployeeName = templates.EmptyEmployeeName
			}
			msgText += fmt.Sprintf(templates.RejectedKnownUser, rejectEmployeeName, *notification.TicketInfo.RejectedEmployeeID)
		}
	default:
		msgText += templates.DefaultSiteLink
	}
	return msgText, nil
}

func (b *notificationBuilder) GenShortMsgTextForTicketCreator(notification ticketmodels.TicketNotification) string {
	return fmt.Sprintf(formatShortMsgTextNotificationForCreator,
		notification.TicketID,
		notification.TicketInfo.TicketName,
		notification.TicketInfo.StatusDescription,
	)
}

func (b *notificationBuilder) getEmployeeName(employeeID int64) (string, error) {
	if employeeID == 2542 {
		return templates.DefaultSystemName, nil
	}
	name, err := b.employeeInfoApi.GetEmployeeName(employeeID)
	if err != nil {
		return "", err
	}
	return name, nil
}
