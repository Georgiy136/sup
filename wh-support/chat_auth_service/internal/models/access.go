package models

type AccessPoliciesResult struct {
	Policies       map[string]*AccessData
	EmployeeGroups []int64
}

type AccessData struct {
	ExternalActions []ExternalActionWithCategory
	ActionGroups    []ActionGroupWithCategory
}

type ExternalActionWithCategory struct {
	ExternalAction string  `json:"external_action"`
	CategoryID     int64   `json:"category_id"`
	StatusID       *string `json:"status_id"`
}

type ActionGroupWithCategory struct {
	GroupID    int64   `json:"group_id"`
	CategoryID int64   `json:"category_id"`
	StatusID   *string `json:"status_id"`
}
