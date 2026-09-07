package models

type PresignedURLSettings struct {
	UploadMinutes   int64 `json:"upload_url_expiration_minutes"`
	DownloadMinutes int64 `json:"download_url_expiration_minutes"`
}
