package models

type File struct {
	FileID           int64   `json:"file_id"`
	SourceUploadID   int64   `json:"source_upload_id"`
	CreateEmployeeID int64   `json:"employee_id"`
	TicketID         *int64  `json:"ticket_id"`
	CategoryID       int64   `json:"category_id"`
	OriginalName     string  `json:"original_name"`
	MimeType         string  `json:"mime_type"`
	SizeBytes        int64   `json:"size_bytes"`
	Bucket           string  `json:"bucket"`
	ObjectKey        string  `json:"object_key"`
	Status           string  `json:"status"`
	CreatedAt        string  `json:"created_at"`
	DeletedAt        *string `json:"deleted_at"`
}

type MarkDeletedFileRequest struct {
	FileID int64 `json:"file_id"`
}

type TransferFileRequest struct {
	UploadID int64 `json:"upload_id"`
}

type DeleteFileRequest struct {
	FileID int64 `json:"file_id"`
}
