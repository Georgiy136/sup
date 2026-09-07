package validators

import (
	"testing"
)

func TestFindURL(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedURL bool
		foundURL    string
	}{
		{
			name:        "Пустая строка",
			input:       "",
			expectedURL: false,
			foundURL:    "",
		},
		{
			name:        "Обычный текст без URL",
			input:       "Это обычный текст без ссылок",
			expectedURL: false,
			foundURL:    "",
		},
		{
			name:        "Стандартное название заявки на подключение",
			input:       "Заявка на подключение клиента",
			expectedURL: false,
			foundURL:    "",
		},
		{
			name:        "Стандартное название заявки на изменение данных",
			input:       "Заявка на изменение данных клиента",
			expectedURL: false,
			foundURL:    "",
		},
		{
			name:        "Стандартное название заявки на проверку договора",
			input:       "Заявка на проверку договора",
			expectedURL: false,
			foundURL:    "",
		},
		{
			name:        "Стандартное название заявки на корректировку статуса",
			input:       "Заявка на корректировку статуса",
			expectedURL: false,
			foundURL:    "",
		},
		{
			name:        "Стандартное название заявки на разблокировку пользователя",
			input:       "Заявка на разблокировку пользователя",
			expectedURL: false,
			foundURL:    "",
		},
		{
			name:        "Стандартное название заявки с номером",
			input:       "Заявка №12345",
			expectedURL: false,
			foundURL:    "",
		},
		{
			name:        "Стандартное название заявки с кодом задачи",
			input:       "Заявка TASK-123",
			expectedURL: false,
			foundURL:    "",
		},
		{
			name:        "Стандартное название заявки с кодом инцидента",
			input:       "Инцидент INC-789",
			expectedURL: false,
			foundURL:    "",
		},
		{
			name:        "Текст с http URL",
			input:       "Перейдите по ссылке http://example.com",
			expectedURL: true,
			foundURL:    "http://example.com",
		},
		{
			name:        "Текст с https URL",
			input:       "Перейдите по ссылке https://example.com",
			expectedURL: true,
			foundURL:    "https://example.com",
		},
		{
			name:        "Текст с ftp URL",
			input:       "Скачайте с ftp://files.example.com",
			expectedURL: true,
			foundURL:    "ftp://files.example.com",
		},
		{
			name:        "Текст с www URL",
			input:       "Посетите www.example.com",
			expectedURL: true,
			foundURL:    "www.example.com",
		},
		{
			name:        "Текст с URL в верхнем регистре",
			input:       "Перейдите по HTTP://EXAMPLE.COM",
			expectedURL: true,
			foundURL:    "http://example.com",
		},
		{
			name:        "Текст с HTTPS URL в верхнем регистре",
			input:       "Перейдите по HTTPS://EXAMPLE.COM/path",
			expectedURL: true,
			foundURL:    "https://example.com/path",
		},
		{
			name:        "Текст с WWW URL в верхнем регистре",
			input:       "Посетите WWW.EXAMPLE.COM",
			expectedURL: true,
			foundURL:    "www.example.com",
		},
		{
			name:        "Текст с несколькими URL",
			input:       "Ссылки: http://example.com и https://test.com",
			expectedURL: true,
			foundURL:    "http://example.com",
		},
		{
			name:        "Текст с доменом .com без протокола",
			input:       "Посетите example.com для информации",
			expectedURL: true,
			foundURL:    "example.com",
		},
		{
			name:        "Текст с доменом .org без протокола",
			input:       "Сайт организации: nonprofit.org",
			expectedURL: true,
			foundURL:    "nonprofit.org",
		},
		{
			name:        "Текст с доменом .ru без протокола",
			input:       "Сайт организации example.ru",
			expectedURL: true,
			foundURL:    "example.ru",
		},
		{
			name:        "Текст с поддоменом без протокола",
			input:       "Документация находится на docs.example.com",
			expectedURL: true,
			foundURL:    "docs.example.com",
		},
		{
			name:        "Текст с несколькими поддоменами без протокола",
			input:       "Сервис расположен на test.stage.example.com",
			expectedURL: true,
			foundURL:    "test.stage.example.com",
		},
		{
			name:        "Текст с доменом и путем",
			input:       "Описание доступно на example.com/path",
			expectedURL: true,
			foundURL:    "example.com/path",
		},
		{
			name:        "Текст с доменом и вложенным путем",
			input:       "Описание доступно на example.com/path/to/page",
			expectedURL: true,
			foundURL:    "example.com/path/to/page",
		},
		{
			name:        "Текст с доменом и query параметром",
			input:       "Откройте example.com?param=value",
			expectedURL: true,
			foundURL:    "example.com?param=value",
		},
		{
			name:        "Текст с доменом, путем и query параметрами",
			input:       "Откройте example.com/path?param=value&x=1",
			expectedURL: true,
			foundURL:    "example.com/path?param=value&x=1",
		},
		{
			name:        "Текст с доменом и anchor",
			input:       "Откройте example.com#section",
			expectedURL: true,
			foundURL:    "example.com#section",
		},
		{
			name:        "Текст с доменом и портом",
			input:       "Сервис доступен на example.com:8080",
			expectedURL: true,
			foundURL:    "example.com:8080",
		},
		{
			name:        "Текст с доменом, портом и путем",
			input:       "Сервис доступен на example.com:8080/path",
			expectedURL: true,
			foundURL:    "example.com:8080/path",
		},
		{
			name:        "Текст с github ссылкой без протокола",
			input:       "Репозиторий github.com/test/repo",
			expectedURL: true,
			foundURL:    "github.com/test/repo",
		},
		{
			name:        "Текст с telegram ссылкой без протокола",
			input:       "Канал t.me/channel",
			expectedURL: true,
			foundURL:    "t.me/channel",
		},
		{
			name:        "URL с точкой в конце предложения",
			input:       "См. example.com.",
			expectedURL: true,
			foundURL:    "example.com",
		},
		{
			name:        "URL с запятой после ссылки",
			input:       "См. example.com, затем создайте заявку",
			expectedURL: true,
			foundURL:    "example.com",
		},
		{
			name:        "URL в круглых скобках",
			input:       "Ссылка указана в описании (example.com)",
			expectedURL: true,
			foundURL:    "example.com",
		},
		{
			name:        "URL в кавычках",
			input:       `Ссылка указана в описании "example.com"`,
			expectedURL: true,
			foundURL:    "example.com",
		},
		{
			name:        "Версия из трех чисел",
			input:       "Версия 1.2.3",
			expectedURL: false,
			foundURL:    "",
		},
		{
			name:        "Версия с префиксом v",
			input:       "Версия v1.2.3",
			expectedURL: false,
			foundURL:    "",
		},
		{
			name:        "Номер релиза",
			input:       "Релиз 2.15.0",
			expectedURL: false,
			foundURL:    "",
		},
		{
			name:        "Пункт договора",
			input:       "Пункт 1.2.3 договора",
			expectedURL: false,
			foundURL:    "",
		},
		{
			name:        "Дата с точками",
			input:       "Дата обработки 03.06.2026",
			expectedURL: false,
			foundURL:    "",
		},
		{
			name:        "Период с датами",
			input:       "Период 01.01.2024 - 31.12.2024",
			expectedURL: false,
			foundURL:    "",
		},
		{
			name:        "Сумма с точкой",
			input:       "Сумма 10000.50",
			expectedURL: false,
			foundURL:    "",
		},
		{
			name:        "Процент с точкой",
			input:       "Процент 12.5%",
			expectedURL: false,
			foundURL:    "",
		},
		{
			name:        "Сокращение и так далее",
			input:       "Нужно проверить статус, сумму, дату и т.д.",
			expectedURL: false,
			foundURL:    "",
		},
		{
			name:        "Город с сокращением",
			input:       "г. Москва",
			expectedURL: false,
			foundURL:    "",
		},
		{
			name:        "Организационно-правовая форма",
			input:       "ООО Ромашка",
			expectedURL: false,
			foundURL:    "",
		},
		{
			name:        "Техническое название поля",
			input:       "request_id",
			expectedURL: false,
			foundURL:    "",
		},
		{
			name:        "Техническое название сервиса",
			input:       "service_name",
			expectedURL: false,
			foundURL:    "",
		},
		{
			name:        "Обычное слово test",
			input:       "test",
			expectedURL: false,
			foundURL:    "",
		},
		{
			name:        "Доменная зона из одной буквы",
			input:       "test.c",
			expectedURL: false,
			foundURL:    "",
		},
		{
			name:        "Название заявки без ссылки со словом support",
			input:       "Заявка на проверку обращения support Wildberries",
			expectedURL: false,
			foundURL:    "",
		},
		{
			name:        "Название заявки без ссылки со словом ticket",
			input:       "Тестовое название заявки ticket name",
			expectedURL: false,
			foundURL:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			foundURL, hasURL := FindURL(tt.input)

			if hasURL != tt.expectedURL {
				t.Errorf("FindURL(%q) hasURL = %v, expected %v", tt.input, hasURL, tt.expectedURL)
			}

			if foundURL != tt.foundURL {
				t.Errorf("FindURL(%q) foundURL = %q, expected %q", tt.input, foundURL, tt.foundURL)
			}
		})
	}
}
