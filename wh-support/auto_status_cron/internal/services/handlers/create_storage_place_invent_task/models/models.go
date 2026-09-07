package createstorageplaceinventtaskmodels

import (
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
)

type ExtTicketInfoCreateBySections struct {
	Stage      Stage      `mapstructure:"stage"`
	Street     Street     `mapstructure:"street"`
	WhID       WhID       `mapstructure:"wh_id"`
	Sections   []int64    `mapstructure:"sections"`
	OfficeID   OfficeID   `mapstructure:"office_id"`
	TaskType   TaskType   `mapstructure:"task_type"`
	EmployeeID EmployeeID `mapstructure:"employee_id"`
}

type ExtTicketInfoCreateByPlaceIDs struct {
	Stage      Stage      `mapstructure:"stage"`
	Street     Street     `mapstructure:"street"`
	WhID       WhID       `mapstructure:"wh_id"`
	Places     []PlaceID  `mapstructure:"places"`
	OfficeID   OfficeID   `mapstructure:"office_id"`
	TaskType   TaskType   `mapstructure:"task_type"`
	EmployeeID EmployeeID `mapstructure:"employee_id"`
}

type Stage = models.ID
type Street = models.ID
type WhID = models.NamedID[int64]
type OfficeID = models.NamedID[int64]
type TaskType = models.NamedID[string]
type EmployeeID = models.NamedID[int64]
type PlaceID = models.ID

type RequestDataForCreateStoragePlaceInventTask struct {
	OfficeID   int64   `json:"office_id"`
	WhID       int64   `json:"wh_id"`
	Stage      int64   `json:"stage"`
	Street     int64   `json:"street"`
	EmployeeID int64   `json:"employee_id"`
	Sections   []int64 `json:"sections,omitempty"`
	PlaceIDs   []int64 `json:"place_ids,omitempty"`
	TaskType   string  `json:"task_type"`
}

type AddedPerformInfoCount struct {
	Count int64 `json:"count"`
}
