package models

import (
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/cron_support_to_wh_olap/sync_models"
)

type ResponseTicketsChangesFromDB struct {
	TicketsData []*sync_models.Tickets_Ticket `json:"data" binding:"required,gt=0,dive"`
}

type ResponseTicketsCategoriesFromDB struct {
	CategoriesData []*sync_models.Categories_Category `json:"data" binding:"required,gt=0,dive"`
}

type ResponseTicketsCategoryStatusesFromDB struct {
	CategoryStatusesData []*sync_models.CategoryStatuses_CategoryStatus `json:"data"`
}

type ResponseTicketsGroupsFromDB struct {
	GroupsData []*sync_models.Groups_Group `json:"data"`
}

type ResponseTicketsScenariosFromDB struct {
	ScenariosData []*sync_models.Scenarios_Scenario `json:"data"`
}
