package consts

const (
	MaxRejectCommentLength = 300
	MaxReturnCommentLength = 500
)

// Названия кнопок для действий Band
const (
	BandApproveActionName = "Подтверждение"
	BandApproveButtonName = "Подтвердить"
	BandPerformActionName = "Исполнение"
	BandPerformButtonName = "Исполнить"
	BandBookButtonName    = "Взять в работу"
	BandUnbookButtonName  = "Снять с себя"
	BandRejectButtonName  = "Отклонить"
	BandReturnButtonName  = "Вернуть на статус"
)

// Стили кнопок для действий Band
const (
	BandStyleSuccess = "success"
	BandStyleDanger  = "danger"
	BandStyleWarning = "warning"
	BandStylePrimary = "primary"
)

// Callback ID и имена полей для диалогов Band
const (
	BandAdditionalInfoDialogCallbackID = "ticket_additional_info"
	BandRejectDialogCallbackID         = "ticket_reject"
	BandCommentField                   = "comment"
	BandReturnDialogCallbackID         = "ticket_return"
	BandReturnToStatusIDField          = "return_to_status_id"
)

// Тексты диалогов Band
const (
	DatePlaceholder = "Введите дату в формате ДД.ММ.ГГГГ"

	BandRejectDialogTitle        = "Отклонение заявки"
	BandRejectDialogIntroFormat  = "Укажите причину отклонения заявки №%d"
	BandRejectSubmitLabel        = "Отклонить"
	BandRejectCommentDisplayName = "Комментарий"
	BandRejectCommentPlaceholder = "Введите причину отклонения"

	BandReturnDialogTitle        = "Возврат заявки"
	BandReturnDialogIntroFormat  = "Выберите статус и укажите причину возврата заявки №%d"
	BandReturnSubmitLabel        = "Вернуть"
	BandReturnStatusDisplayName  = "Статус"
	BandReturnCommentDisplayName = "Комментарий"
	BandReturnCommentPlaceholder = "Введите причину возврата"

	BandAdditionalInfoIntroductionText = "Заполните необходимые поля"

	BandActionDialogTitleFormat = "%s заявки №%d"
)
