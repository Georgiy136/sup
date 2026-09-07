package models

type CategoryStatusResponse struct {
	Data []CategoryStatusData `json:"data"`
}

type CategoryStatusData struct {
	StatusID              string         `json:"status_id"`
	CategoryID            int64          `json:"category_id"`
	AddedInformationModel []ScenarioInfo `json:"added_information_model"`
	ReturnStatuses        []ReturnStatus `json:"return_statuses"`
	IsBlockedReject       bool           `json:"is_blocked_reject"`
	IsBlockedStatus       bool           `json:"is_blocked_status"`
}

type ScenarioInfo struct {
	ScenarioName           string          `json:"scenario_name"`
	ScenarioOrderID        int64           `json:"scenario_order_id"`
	NextStatusID           string          `json:"next_status_id,omitempty"`
	ScenarioAddedInfoModel []ScenarioField `json:"scenario_added_information_model"`
}

type ScenarioField struct {
	DataName             string             `json:"data_name"`
	FrontDataName        string             `json:"front_data_name"`
	DataType             string             `json:"data_type"`
	IsRequired           bool               `json:"is_required"`
	OrderID              int64              `json:"order_id"`
	FrontDataType        FieldDataType      `json:"front_data_type"`
	FrontDataDescription string             `json:"front_data_description"`
	IsURL                bool               `json:"is_url,omitempty"`
	LongText             bool               `json:"long_text,omitempty"`
	IsMultiple           bool               `json:"is_multiple,omitempty"`
	MaxThreshold         *string            `json:"front_max_threshold_data,omitempty"`
	MinThreshold         *string            `json:"front_min_threshold_data,omitempty"`
	RegularExpression    *RegularExpression `json:"regular_expression,omitempty"`
	DefaultValues        DefaultValues      `json:"default_values,omitempty"`
	Api                  *string            `json:"api,omitempty"`
	SelectorArray        []SelectorOption   `json:"selector_array,omitempty"`
}

type RegularExpression struct {
	RegExp      string `json:"reg_exp"`
	WarnMessage string `json:"warn_message"`
}

type DefaultValues struct {
	Values []any `json:"values"`
}

type SelectorOption struct {
	ID      any    `json:"id"`
	Name    string `json:"name"`
	OrderID int64  `json:"order_id"`
}
