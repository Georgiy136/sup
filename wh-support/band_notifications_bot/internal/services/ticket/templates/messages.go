package templates

import "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/consts"

const (
	PerformOnSiteTicketMessage = "\n⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯\nДанную заявку можно исполнить на [сайте](" + consts.SupportSiteURL + ")"
	ConfirmOnSite              = "⚠️ Данную заявку можно исполнить только на сайте [Wh Support](" + consts.SupportSiteURL + ")"

	ApproveOnSiteTicketMessage = "\n⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯\nДанную заявку можно подтвердить на [сайте](" + consts.SupportSiteURL + ")"
	RejectedUnknownUser        = "\n⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯\nЗаявка была отклонена неизвестным пользователем ❌\nПодробная информация на сайте [Wh Support](" + consts.SupportSiteURL + ")"
	RejectedKnownUser          = "\n⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯\n%s (%d) отклонил(-а) заявку ❌\nПодробная информация на сайте [Wh Support](" + consts.SupportSiteURL + ")"
	DefaultSiteLink            = "\n⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯\nПодробная информация на сайте [Wh Support](" + consts.SupportSiteURL + ")"

	EmptyEmployeeName = "ФИО пользователя не определено"
	DefaultSystemName = "Система"

	ActionErrorMessageFormat   = "⚠️ Не удалось выполнить действие по заявке №%d:\n%s"
	ActionInternalErrorMessage = "Произошла техническая ошибка. Попробуйте позже."
)
