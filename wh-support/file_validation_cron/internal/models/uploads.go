package models

const (
	UploadStatusInitiated   string = "INITIATED"
	UploadStatusValidating  string = "VALIDATING"
	UploadStatusReady       string = "READY"
	UploadStatusInvalid     string = "INVALID"
	UploadStatusUserReject  string = "USER_REJECT"
	UploadStatusTransferred string = "TRANSFERRED"
	UploadStatusExpired     string = "EXPIRED"
)

type Upload struct {
	UploadID          int64  `json:"upload_id"`
	FileID            int64  `json:"file_id"`
	ChEmployeeID      int64  `json:"ch_employee_id"`
	TicketID          *int64 `json:"ticket_id"`
	EntityType        string `json:"entity_type"`
	OriginalName      string `json:"original_name"`
	DeclaredMimeType  string `json:"declared_mime_type"`
	DeclaredSizeBytes int64  `json:"declared_size_bytes"`
	TmpBucket         string `json:"tmp_bucket"`
	TmpObjectKey      string `json:"tmp_object_key"`
	Status            string `json:"status"`
	ExpiresAt         string `json:"expires_at"`
	CreatedAt         string `json:"created_at"`
}

type ChangeStatusUploadRequest struct {
	UploadID int64  `json:"upload_id"`
	Status   string `json:"status"`
}
