package constant

import "time"

const AttachPendingFileTTL = 7 * 24 * time.Hour

// File statuses - статусы записей в таблице files
const (
	FileStatusActive        = "ACTIVE"         // Файл активен
	FileStatusAttachPending = "ATTACH_PENDING" // Файл ожидает привязки к заявке
	FileStatusDeletePending = "DELETE_PENDING" // Файл ожидает удаления
	FileStatusDeleted       = "DELETED"        // Файл удален
)
