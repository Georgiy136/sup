package models

type CategoryStatusAccess struct {
	CategoryID int64   `json:"category_id"`
	StatusID   *string `json:"status_id"`
}

type PaginationModel struct {
	TicketIDs []int64 `json:"ticket_ids"`
}

type CheckAccessForSubTabData struct {
	HasAccess bool `json:"has_access"`
}

type TicketCommonInfo struct {
	TicketID             int64   `json:"ticket_id"`
	CategoryID           int64   `json:"category_id"`
	StatusID             string  `json:"status_id"`
	CreateEmployeeID     int64   `json:"create_employee_id"`
	FavouriteEmployeeIDs []int64 `json:"favourite_employee_ids"`
	GroupEmployeeIDs     []int64 `json:"group_employee_ids"`
}

type TicketFiles struct {
	FileID   int64  `json:"file_id"`
	FileType string `json:"file_type"`
	FileSize int64  `json:"file_size"`
	FileName string `json:"file_name"`
}

type TiketsGetIDDataFromDB struct {
	TiketID int64 `json:"ticket_id"`
}
