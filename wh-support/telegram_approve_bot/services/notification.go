package services

import (
	"fmt"
	"unicode/utf8"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/common"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/models"

	"github.com/sirupsen/logrus"
)

var (
	emptyEmployeeName = "ФИО пользователя не определено"
)

const (
	ParseModeMarkdown        = "Markdown"
	TelegramMessageMaxLength = 4000
)

const (
	ApproveOperation   = "approve"   // Заявка требует подтверждения
	PerformOperation   = "perform"   // Заявка требует исполнения
	RejectedOperation  = "rejected"  // Заявка отклонена юзером
	ApprovedOperation  = "approved"  // Заявка подтверждена юзером
	PerformedOperation = "performed" // Заявка исполнена юзером
	BookedOperation    = "booked"    // Заявка взята в работу юзером
	UnbookedOperation  = "unbooked"  // Заявка снята с работы юзером
	CompletedOperation = "completed" // Заявка выполнена юзером
	FavoriteOperation  = "favourite" // Стали наблюдателем по заявке
)

const (
	isCreator  = "CRT" //  Создатель заявки
	isGroup    = "GRP" // Группа исполнения или подтверждения
	isEmployee = "EMP" // Сотрудник, работающий с заявкой
	isObserver = "FVR" // Наблюдатель заявки
)

const (
	DefaultSystemName = "Система"
)

func (b *Bot) SendNotification(notification models.Notification) error {
	if err := b.handleByOperationType(notification); err != nil {
		return fmt.Errorf("can not handle ticket %d by operation type %s: %w", notification.TicketID, notification.TicketInfo.OperationType, err)
	}
	return nil
}

func (b *Bot) handleByOperationType(notification models.Notification) error {
	msgText := b.genCommonMsgText(notification)
	msgLength := utf8.RuneCountInString(msgText)

	if msgLength > TelegramMessageMaxLength {
		return b.handleTooLongMessage(notification)
	}

	switch notification.TicketInfo.OperationType {
	case ApproveOperation:
		if err := b.handleApproveOperation(msgText, notification); err != nil {
			return fmt.Errorf("handle approve operation error for ticket %d: %w", notification.TicketID, err)
		}
	case ApprovedOperation:
		if err := b.handleApprovedOperation(notification); err != nil {
			return fmt.Errorf("handle approved operation error for ticket %d: %w", notification.TicketID, err)
		}
	case PerformOperation:
		if err := b.handlePerformOperation(msgText, notification); err != nil {
			return fmt.Errorf("handle perform operation error for ticket %d: %w", notification.TicketID, err)
		}
	case PerformedOperation:
		if err := b.handlePerformedOperation(notification); err != nil {
			return fmt.Errorf("handle performed operation error for ticket %d: %w", notification.TicketID, err)
		}
	case RejectedOperation:
		if err := b.handleRejectedOperation(msgText, notification); err != nil {
			return fmt.Errorf("handle rejected operation error for ticket %d: %w", notification.TicketID, err)
		}
	case CompletedOperation:
		if err := b.handleCompletedOperation(msgText, notification); err != nil {
			return fmt.Errorf("handle completed operation error for ticket %d: %w", notification.TicketID, err)
		}
	case BookedOperation:
		if err := b.handleBookedOperation(notification); err != nil {
			return fmt.Errorf("handle booked operation error for ticket %d: %w", notification.TicketID, err)
		}
	case UnbookedOperation:
		if err := b.handleUnBookedOperation(notification); err != nil {
			return fmt.Errorf("handle unbooked operation error for ticket %d: %w", notification.TicketID, err)
		}
	default:
		if err := b.sendCustomMsg(notification.ChatID, msgText); err != nil {
			return fmt.Errorf("cant send text msg error for ticket %d: %w", notification.TicketID, err)
		}
	}

	logrus.Infof("notification on ticket № %d success send", notification.TicketID)
	return nil
}

func (b *Bot) handleTooLongMessage(notification models.Notification) error {
	redirectMsgText := b.genShortMsgTex(notification)
	if err := b.sendCustomMsg(notification.ChatID, redirectMsgText); err != nil {
		return fmt.Errorf("can't send short message: %w", err)
	}
	logrus.Infof("notification on ticket № %d sent short message", notification.TicketID)
	return nil
}

func (b *Bot) sendTicketMsgForCreator(notification models.Notification) error {
	msgText := b.genCommonMsgTextForTicketCreator(notification)

	return b.sendCustomMsg(notification.ChatID, msgText)
}

func (b *Bot) getEmployeeNameByEmployeeID(employeeID int64) string {
	if employeeID == 2542 {
		return DefaultSystemName
	}

	var employeeName string

	employeeInfo, err := b.employeeInfoApiClient.GetEmployeeSelfFullName(employeeID)
	switch {
	case err != nil:
		logrus.Warnf("nil from employee info by %d: %v", employeeID, common.ErrEmployeeUnknown)
		employeeName = emptyEmployeeName
	case employeeInfo == nil:
		logrus.Errorf("can't get employee info by %d: %v", employeeID, err)
		employeeName = emptyEmployeeName
	default:
		employeeName = employeeInfo.Name
	}

	return employeeName
}
