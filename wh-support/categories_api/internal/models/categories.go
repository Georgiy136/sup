package models

type AccessData struct {
	ExternalActions []ExternalActionWithCategory
	ActionGroups    []ActionGroupInfoWithCategory
	EmployeeGroups  []int64
}

type ExternalActionWithCategory struct {
	ExternalAction string  `json:"external_action"`
	CategoryID     int64   `json:"category_id"`
	StatusID       *string `json:"status_id"`
}

type ActionGroupInfoWithCategory struct {
	GroupID    int64   `json:"group_id"`
	CategoryID int64   `json:"category_id"`
	StatusID   *string `json:"status_id"`
}
