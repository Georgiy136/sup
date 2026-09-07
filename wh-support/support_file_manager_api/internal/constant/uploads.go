package constant

const (
	SystemEmployeeID = 2542
)

const MaxUploadSize int64 = 20 * 1024 * 1024 // 20 МБ

const (
	UploadStatusInitiated   = "INITIATED"      // Загрузка инициирована, ожидание загрузки файла
	UploadStatusValidating  = "VALIDATING"     // Файл загружен во временное хранилище, идет валидация
	UploadStatusReady       = "READY"          // Файл провалидирован и готов к переносу
	UploadStatusInvalid     = "INVALID"        // Файл не прошел валидацию
	UploadStatusUserReject  = "USER_REJECT"    // Пользователь отменил загрузку
	UploadStatusTransferred = "TRANSFERRED"    // Файл перенесен в постоянное хранилище
	UploadStatusExpired     = "EXPIRED"        // Истек срок жизни аплоада
	UploadStatusService     = "SERVICE_UPLOAD" // Сервисная загрузка
)
