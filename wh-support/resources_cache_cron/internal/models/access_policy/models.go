package accesspolicy

import (
	"time"
)

type AccessPolicyDBResp struct {
	LogId        int64          `json:"log_id"`
	SleepSeconds int64          `json:"sleep_seconds"`
	Data         []AccessPolicy `json:"data"`
}

type AccessPolicy struct {
	CategoryId     int64     `json:"category_id"`
	TypeAction     string    `json:"typeaction_id"`
	StatusId       *string   `json:"status_id"`
	ExternalAction *string   `json:"external_action"`
	GroupId        int64     `json:"group_id"`
	IsDel          bool      `json:"is_del"`
	ChDt           string    `json:"ch_dt"`
	ChDtParsed     time.Time `json:"-"`
}

type ResourceCompositeKey struct {
	CategoryID int64
	TypeAction string
	StatusID   *string
}

type ExternalActionWithCategory struct {
	ExternalAction string  `json:"external_action"`
	CategoryId     int64   `json:"category_id"`
	StatusId       *string `json:"status_id"`
}

type GroupIdWithCategory struct {
	GroupId    int64   `json:"group_id"`
	CategoryId int64   `json:"category_id"`
	StatusId   *string `json:"status_id"`
}
