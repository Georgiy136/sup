package models

const (
	EntityTypeTicketAttachment = "TICKET_ATTACHMENT"
	MaxFilesPerTicket          = 10
)

type FileManagerInitUploadRequest struct {
	FileName   string          `json:"file_name"`
	FileSize   int64           `json:"file_size"`
	MimeType   string          `json:"mime_type"`
	EntityType string          `json:"entity_type"`
	EntityData *FileEntityData `json:"entity_data,omitempty"`
}

type FileEntityData struct {
	TicketID   *int64 `json:"ticket_id"`
	CategoryID int64  `json:"category_id"`
}

type FileManagerInitUploadResponse struct {
	UploadID           int64             `json:"upload_id"`
	FileID             int64             `json:"file_id"`
	UploadURL          string            `json:"upload_url"`
	Method             string            `json:"method"`
	ExpiresAt          string            `json:"expires_at"`
	RequiredFormFields map[string]string `json:"required_form_fields"`
}

type FileManagerConfirmUploadRequest struct {
	UploadID int64 `json:"upload_id"`
	FileID   int64 `json:"file_id"`
}

type FileManagerCancelUploadRequest struct {
	UploadID int64 `json:"upload_id"`
	FileID   int64 `json:"file_id"`
}

type FileManagerDownloadURLRequest struct {
	FileID int64 `json:"file_id"`
}

type FileManagerDownloadURLResponse struct {
	FileID      int64  `json:"file_id"`
	DownloadURL string `json:"download_url"`
	FileName    string `json:"file_name"`
	ExpiresAt   string `json:"expires_at"`
}

type AttachFilesToTicketRequest struct {
	TicketID int64   `json:"ticket_id"`
	FileIDs  []int64 `json:"file_ids"`
}

type GetFileInfoRequest struct {
	FileID int64 `json:"file_id"`
}

type GetFileInfoResponse struct {
	FileID         int64   `json:"file_id"`
	SourceUploadID int64   `json:"source_upload_id"`
	EmployeeID     int64   `json:"employee_id"`
	TicketID       int64   `json:"ticket_id"`
	CategoryID     int64   `json:"category_id"`
	OriginalName   string  `json:"original_name"`
	MimeType       string  `json:"mime_type"`
	SizeBytes      int64   `json:"size_bytes"`
	Bucket         string  `json:"bucket"`
	ObjectKey      string  `json:"object_key"`
	Status         string  `json:"status"`
	CreatedAt      string  `json:"created_at"`
	DeletedAt      *string `json:"deleted_at"`
}
