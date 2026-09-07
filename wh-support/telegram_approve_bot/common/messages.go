package common

const (
	HasNoRightsMsgError        = "Ошибка регистрации. У Вас недостаточно прав для использования бота"
	RegistrationMsgError       = "Ошибка регистрации. Попробуйте позже"
	AlreadyRegisteredMsg       = "Вы уже зарегистрированы. Если проблема сохраняется - обратитесь в поддержку и попробуйте позже"
	TryAgainLaterMsg           = "Повторите попытку позже"
	UnexpectedErrorMsg         = "Что-то пошло не так"
	CommandNotExistErrorMsg    = "Команда телеграм бота не найдена"
	FiredEmployeeErrorMsg      = "Вы не можете пользоваться этим ботом так как вы уволены"
	InvalidNumberErrorMsg      = "Неверный номер"
	TelegramNotFoundInWhPortal = "Данный аккаунт телеграм не найден в сервисе Wh Portal. Обратитесь в поддержку и попробуйте позже"
	UseContactButtonMsgText    = "Чтобы прислать номер нажмите на кнопку отправки номера"

	TicketNotNeedApproveMessage = "\n-------------\nЗаявка не требует подтверждения. Необходимо сразу перейти к исполнению."
	PerformOnSiteTicketMessage  = "\n-------------\nДанная заявка может быть исполнена только на [сайте](https://support.wbwh.tech)"
	ApproveOnSiteTicketMessage  = "\n-------------\nДанная заявка может быть подтверждена только на [сайте](https://support.wbwh.tech)"
	TicketNotFoundMessage       = "\n-------------\nЗаявка не найдена (выполнена или удалена)"
	TooLongMessageWarning       = "ℹ️ Уведомление сокращено из-за большого объёма данных\n\nДля просмотра полной информации перейдите на [WH Support](https://support.wbwh.tech/)\n"
)

var FormatMsgTextNotification = `📌 У Вас новое уведомление по заявке *№ %d*

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
var FormatMsgTextEmployees = `
*Ответственный:*
%s (%d)

*Заявитель:*
%s (%d)
`

var FormatMsgTextNotificationForCreator = `📌 Уведомление по заявке *№ %d* — «*%s*». 
Заявка прошла статус: *%s*.
Подробную информацию по заявке смотрите на [Wh Support](https://support.wbwh.tech/my-tickets) в разделе «Мои заявки».
`

var FormatMsgTextEmployeesWithoutApprove = `
*Ответственный:*
Заявка не проходила согласование

*Заявитель:*
%s (%d)
`
