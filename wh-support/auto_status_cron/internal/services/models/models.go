package models

import "encoding/json"

type DataTickets struct {
	Tickets []TicketCommonInfo `json:"data"`
}

type TicketCommonInfo struct {
	Ext          map[string]any `json:"ext"`
	StatusID     string         `json:"status_id"`
	TicketID     int64          `json:"ticket_id"`
	CategoryID   int64          `json:"category_id"`
	StatusFields []struct {
		StatusID          string `json:"status_id"`
		ApproveEmployeeID int64  `json:"approve_employee_id"`
		PerformEmployeeID int64  `json:"perform_employee_id"`
	} `json:"status_fields"`
	CreateEmployeeID   int64  `json:"create_employee_id"`
	PrerejectComment   string `json:"prereject_comment"`
	IsPrereject        bool   `json:"is_prereject"`
	ReturnFromStatusID string `json:"return_from_status_id"`
}

type PerformErrValue struct {
	Comment string `json:"comment"`
	Values  []any  `json:"value"`
}

type OverrideCategoriesConfig struct {
	IsOverrideCategory bool    `json:"is_override_category"`
	EnabledCategories  []int64 `json:"enabled_categories"`
}

type RequestPerformTicketV2 struct {
	TicketID        int64           `json:"ticket_id"`
	ScenarioOrderID int64           `json:"scenario_order_id"`
	Ext             json.RawMessage `json:"ext"`
	Files           json.RawMessage `json:"files,omitempty"`
}

type Files struct {
	FileID   int64  `json:"file_id"`
	FileType string `json:"file_type"`
	FileSize int64  `json:"file_size"`
	FileName string `json:"file_name"`
}

type FieldValue struct {
	OrderId       int64  `json:"order_id"`
	FrontDataName string `json:"front_data_name"`
	Value         any    `json:"value"`
}
