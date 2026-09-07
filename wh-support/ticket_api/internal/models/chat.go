package models

type ChatInfo struct {
	Exists             bool    `json:"exists"`
	HasUnread          bool    `json:"has_unread"`
	LastActivityAt     *string `json:"last_activity_at"`
	LastUserActivityAt *string `json:"last_user_activity_at"`
}

type ChatActivityData struct {
	LastActivityAt     string
	LastUserActivityAt string
}
