package models

type UpdateCategoryRequest struct {
	CategoryID       int64   `json:"category_id" validate:"required,gt=0"`
	CategoryName     string  `json:"category_name" validate:"omitempty,min=1,max=100"`
	ParentCategoryID *int64  `json:"parent_category_id" validate:"omitempty,gt=0"`
	ActionID         *string `json:"action_id" validate:"omitempty,max=50"`
	IsDel            *bool   `json:"is_del" validate:"required"`
	Icon             *string `json:"icon" validate:"omitempty,min=1,max=25"`
	Color            *string `json:"color" validate:"omitempty,min=1,max=9"`
}

type AddCategoryRequest struct {
	CategoryName     string  `json:"category_name" validate:"required,min=1,max=100"`
	ParentCategoryID *int64  `json:"parent_category_id" validate:"omitempty,gt=0"`
	ActionID         *string `json:"action_id" validate:"omitempty,max=50"`
	IsActive         *bool   `json:"is_active" validate:"required"`
}
type StatusInfo struct {
	AddedInformationModel []AddedInformationModel `json:"added_information_model"`
}

type AddedInformationModel struct {
	ScenarioOrderID               int64                           `json:"scenario_order_id"`
	ScenarioAddedInformationModel []ScenarioAddedInformationModel `json:"scenario_added_information_model"`
}

type ScenarioAddedInformationModel struct {
	FrontDatatype    string   `json:"front_data_type"`
	DataName         string   `json:"data_name"`
	FileAllowedTypes []string `json:"file_allowed_types"`
}

type CategoryStatuses struct {
	CategoryID int64    `json:"category_id"`
	StatusIDs  []string `json:"status_ids"`
}
