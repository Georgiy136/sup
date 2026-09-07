package consts

// Сообщения валидации полей
const (
	ValidationErrorIntegerOnly     = "Допускаются только целые числа"
	ValidationErrorIntegerTooLarge = "Число слишком большое"

	ValidationErrorMinValueFormat = "Минимальное значение поля - %d"
	ValidationErrorMaxValueFormat = "Максимальное значение поля - %d"

	ValidationErrorMinLengthFormat = "Минимально допустимая длина поля - %d символов"
	ValidationErrorMaxLengthFormat = "Максимально допустимая длина поля - %d символов"

	ValidationErrorInvalidFormat = "Некорректный формат"
	ValidationErrorFieldRequired = "Заполните поле"

	URLSchemeSeparator = "://"
	URLSchemeHTTP      = "http"
	URLSchemeHTTPS     = "https"
	DefaultURLPattern  = `^https?://([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}(/\S*)?$`

	ValidationErrorLinkRequired = "Укажите ссылку в формате " + URLSchemeHTTP + URLSchemeSeparator + " или " + URLSchemeHTTPS + URLSchemeSeparator
)
