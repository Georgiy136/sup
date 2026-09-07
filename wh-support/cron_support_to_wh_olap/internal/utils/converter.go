package utils

import (
	"fmt"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/cron_support_to_wh_olap/internal/models"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/cron_support_to_wh_olap/sync_models"

	jsoniter "github.com/json-iterator/go"
)

func ConvertTicketsCategories(categoriesProto *sync_models.Categories) (*models.DataWrapper[models.Category], error) {
	if categoriesProto == nil {
		return nil, fmt.Errorf("tickets categories for convert cannot be nil")
	}

	data := make([]models.Category, len(categoriesProto.Data))
	for i, categoryProto := range categoriesProto.Data {
		data[i] = models.Category{
			CategoryID:         categoryProto.GetCategoryID(),
			CategoryName:       categoryProto.GetCategoryName(),
			IsDel:              categoryProto.IsDel,
			ParentCategoryID:   categoryProto.GetParentCategoryID(),
			ParentCategoryName: categoryProto.GetParentCategoryName(),
			ActionID:           categoryProto.ActionID,
			ChEmployeeID:       categoryProto.GetChEmployeeID(),
			ChDt:               categoryProto.GetChDt(),
			ChildCategoryIDs:   categoryProto.GetChildCategoryIDs(),
			IsProd:             categoryProto.IsProd,
			SyncDt:             categoryProto.SyncDt,
			SourceCategoryID:   categoryProto.SourceCategoryID,
			ProdCategoryIDs:    categoryProto.GetProdCategoryIDs(),
		}
	}

	return &models.DataWrapper[models.Category]{Data: data}, nil
}

func ConvertTicketsChanges(ticketsProto *sync_models.Tickets) (*models.DataWrapper[models.Ticket], error) {
	if ticketsProto == nil {
		return nil, fmt.Errorf("tickets for convert cannot be nil")
	}

	data := make([]models.Ticket, len(ticketsProto.Data))
	for i, ticketProto := range ticketsProto.Data {
		data[i] = models.Ticket{
			TicketID:           ticketProto.GetTicketID(),
			CategoryID:         ticketProto.GetCategoryID(),
			CategoryName:       ticketProto.GetCategoryName(),
			ParentCategoryName: ticketProto.GetParentCategoryName(),
			StatusID:           ticketProto.GetStatusID(),
			StatusDescription:  ticketProto.GetStatusDescription(),
			StatusEmployeeID:   ticketProto.GetStatusEmployeeID(),
			StatusGroupName:    ticketProto.StatusGroupName,
			ScenarioOrderID:    ticketProto.ScenarioOrderID,
			ScenarioName:       ticketProto.ScenarioName,
			RejectedComment:    ticketProto.RejectedComment,
			ChDt:               ticketProto.GetChDt(),
			CreateEmployeeID:   ticketProto.GetCreateEmployeeID(),
			CreateDt:           ticketProto.GetCreateDt(),
		}

		extLocal, err := toJSONString(ticketProto.Ext)
		if err != nil {
			return nil, fmt.Errorf("cannot convert ext to string: %w", err)
		}
		data[i].Ext = extLocal
	}

	return &models.DataWrapper[models.Ticket]{Data: data}, nil
}

func ConvertTicketsCategoryStatuses(categoryStatusesProto *sync_models.CategoryStatuses) (*models.DataWrapper[models.CategoryStatus], error) {
	if categoryStatusesProto == nil {
		return nil, fmt.Errorf("tickets category statuses for convert cannot be nil")
	}

	data := make([]models.CategoryStatus, len(categoryStatusesProto.Data))
	for i, categoryStatusProto := range categoryStatusesProto.Data {
		data[i] = models.CategoryStatus{
			CategoryID:         categoryStatusProto.GetCategoryID(),
			StatusID:           categoryStatusProto.GetStatusID(),
			StatusDescription:  categoryStatusProto.GetStatusDescription(),
			GroupID:            categoryStatusProto.GroupID,
			IsDel:              categoryStatusProto.IsDel,
			ChEmployeeID:       categoryStatusProto.GetChEmployeeID(),
			ChDt:               categoryStatusProto.GetChDt(),
			ConstructorOrderID: categoryStatusProto.GetConstructorOrderID(),
		}
	}

	return &models.DataWrapper[models.CategoryStatus]{Data: data}, nil
}

func ConvertTicketsGroups(groupsProto *sync_models.Groups) (*models.DataWrapper[models.Group], error) {
	if groupsProto == nil {
		return nil, fmt.Errorf("tickets groups for convert cannot be nil")
	}

	data := make([]models.Group, len(groupsProto.Data))
	for i, groupProto := range groupsProto.Data {
		data[i] = models.Group{
			GroupID:      groupProto.GetGroupID(),
			GroupName:    groupProto.GetGroupName(),
			IsDel:        groupProto.IsDel,
			ChEmployeeID: groupProto.GetChEmployeeID(),
			ChDt:         groupProto.GetChDt(),
		}
	}

	return &models.DataWrapper[models.Group]{Data: data}, nil
}

func ConvertTicketsScenarios(scenariosProto *sync_models.Scenarios) (*models.DataWrapper[models.Scenario], error) {
	if scenariosProto == nil {
		return nil, fmt.Errorf("tickets scenarios for convert cannot be nil")
	}

	data := make([]models.Scenario, len(scenariosProto.Data))
	for i, scenarioProto := range scenariosProto.Data {
		data[i] = models.Scenario{
			CategoryID:      scenarioProto.GetCategoryID(),
			StatusID:        scenarioProto.GetStatusID(),
			NextStatusID:    scenarioProto.GetNextStatusID(),
			ScenarioOrderID: scenarioProto.ScenarioOrderID,
			ScenarioName:    scenarioProto.GetScenarioName(),
			IsDel:           scenarioProto.IsDel,
			ChEmployeeID:    scenarioProto.GetChEmployeeID(),
			ChDt:            scenarioProto.GetChDt(),
		}
	}

	return &models.DataWrapper[models.Scenario]{Data: data}, nil
}

func toJSONString(s any) (string, error) {
	if s == nil {
		return "", nil
	}
	b, err := jsoniter.Marshal(s)
	if err != nil {
		return "", err
	}

	str := string(b)
	if str == "null" || str == "[]" {
		return "", nil
	}

	return str, nil
}
