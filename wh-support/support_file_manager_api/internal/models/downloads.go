package models

import "time"

// GetDownloadURLRequest - запрос на получение ссылки для скачивания файла
type GetDownloadURLRequest struct {
	Bucket           string
	ObjectKey        string
	OriginalFileName string
	ContentType      string
	Expiration       time.Duration
}

// GetDownloadURLResponse - ответ со ссылкой для скачивания
type GetDownloadURLResponse struct {
	FileID      int64     `json:"file_id"`
	DownloadURL string    `json:"download_url"`
	FileName    string    `json:"file_name"`
	TicketID    *int64    `json:"ticket_id"`
	ExpiresAt   time.Time `json:"expires_at"`
}
