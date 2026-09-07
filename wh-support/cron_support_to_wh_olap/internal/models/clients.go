package models

type DataWrapper[T any] struct {
	Data []T `json:"data"`
}

type Ticket struct {
	TicketID           int64   `json:"ticket_id"`
	CategoryID         int64   `json:"category_id"`
	CategoryName       string  `json:"category_name"`
	ParentCategoryName string  `json:"parent_category_name"`
	StatusID           string  `json:"status_id"`
	StatusDescription  string  `json:"status_description"`
	StatusEmployeeID   int64   `json:"status_employee_id"`
	StatusGroupName    *string `json:"status_group_name"`
	ScenarioOrderID    *int64  `json:"scenario_order_id"`
	ScenarioName       *string `json:"scenario_name"`
	RejectedComment    *string `json:"rejected_comment"`
	Ext                string  `json:"ext"`
	ChDt               string  `json:"dt"`
	CreateEmployeeID   int64   `json:"create_employee_id"`
	CreateDt           string  `json:"create_dt"`
}

type Category struct {
	CategoryID         int64   `json:"category_id"`
	CategoryName       string  `json:"category_name"`
	IsDel              *bool   `json:"is_del"`
	ParentCategoryID   int64   `json:"parent_category_id"`
	ParentCategoryName string  `json:"parent_category_name"`
	ActionID           *string `json:"action_id"`
	ChEmployeeID       int64   `json:"ch_employee_id"`
	ChDt               string  `json:"dt"`
	ChildCategoryIDs   []int64 `json:"child_category_ids"`
	IsProd             *bool   `json:"is_prod"`
	SyncDt             *string `json:"sync_dt"`
	SourceCategoryID   *int64  `json:"source_category_id"`
	ProdCategoryIDs    []int64 `json:"prod_category_ids"`
}

type CategoryStatus struct {
	CategoryID         int64  `json:"category_id" binding:"required"`
	StatusID           string `json:"status_id" binding:"required,len=3"`
	StatusDescription  string `json:"status_description" binding:"required"`
	GroupID            *int64 `json:"group_id" binding:"omitempty"`
	IsDel              *bool  `json:"is_del" binding:"required"`
	ChEmployeeID       int64  `json:"employee_id" binding:"required"`
	ChDt               string `json:"dt"`
	ConstructorOrderID int64  `json:"constructor_order_id" binding:"required,gt=0"`
}

type Group struct {
	GroupID      int64  `json:"group_id" binding:"required"`
	GroupName    string `json:"group_name" binding:"required"`
	IsDel        *bool  `json:"is_del" binding:"required"`
	ChEmployeeID int64  `json:"employee_id" binding:"required"`
	ChDt         string `json:"dt"`
}

type Scenario struct {
	CategoryID      int64  `json:"category_id" binding:"required"`
	StatusID        string `json:"status_id" binding:"required,len=3"`
	NextStatusID    string `json:"next_status_id" binding:"required,len=3"`
	ScenarioOrderID *int64 `json:"scenario_order_id" binding:"omitempty"`
	ScenarioName    string `json:"scenario_name" binding:"required"`
	IsDel           *bool  `json:"is_del" binding:"required"`
	ChEmployeeID    int64  `json:"employee_id" binding:"required"`
	ChDt            string `json:"dt"`
}
