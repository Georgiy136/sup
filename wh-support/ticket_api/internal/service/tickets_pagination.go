package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/common"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/models"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/service/chat_mapper"
	internal_utils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_api/internal/utils"
	"gitlab.wildberries.ru/wbwh/support/utils.git/access_actions"
	"gitlab.wildberries.ru/wbwh/support/utils.git/subtab"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_rest_auto_api.git/validators"
	utils "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git"
	core_errors "gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/errors/errors_keys"
)

func (t TicketsService) CheckTicketAccessForSubTab(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TicketID int64  `json:"ticket_id" binding:"required,gt=0"`
		Subtab   string `json:"subtab" binding:"required,min=1,max=50"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	dbTab, err := subtab.ToDBTab(body.Subtab)
	if err != nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, err)
		return
	}

	var categoryStatusJSON []byte
	if subtab.RequiresCategoryStatusIDs(dbTab) {
		filtered, err := t.ticketAccess.GetCategoryStatusAccessForTicket(ctx.Request.Context(), employeeID, body.TicketID, access_actions.TypeActionStatusApproveCategory, access_actions.TypeActionStatusPerformCategory)
		if err != nil {
			if errors.Is(err, common.ErrTicketNotFound) || errors.Is(err, common.ErrNoCategoryStatusAccess) {
				utils.BindObjectToRestData(ctx, models.CheckAccessForSubTabData{HasAccess: false})
				return
			}
			t.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, err)
			return
		}

		categoryStatusJSON, err = jsoniter.Marshal(filtered)
		if err != nil {
			t.errBuilder.BindError(ctx, errors_keys.Err500MarshalWrong, fmt.Errorf("can't marshal category status access: %w", err))
			return
		}
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(body.TicketID, employeeID, dbTab, categoryStatusJSON)
	pg.SetStoredProcedureName("tickets.tickets_checkaccessfortab")
	utils.BindRestData(ctx, repository.GetRestDataFromDb(pg))
}

func (t TicketsService) GetCategoryStatusesForMyTickets(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(employeeID)
	pg.SetStoredProcedureName("tickets.categorystatus_getbycreator")
	utils.BindRestData(ctx, repository.GetRestDataFromDb(pg))
}

func (t TicketsService) GetCategoryStatusesForWorkTickets(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	categoryStatusAccess, err := t.ticketAccess.GetEmployeeCategoryStatusAccess(ctx.Request.Context(), employeeID, access_actions.TypeActionStatusApproveCategory, access_actions.TypeActionStatusPerformCategory)
	if err != nil {
		t.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, err)
		return
	}
	if len(categoryStatusAccess) == 0 {
		utils.BindNoContent(ctx)
		return
	}

	categoryStatusIDsJSON, err := jsoniter.Marshal(groupCategoryStatusesByCategory(categoryStatusAccess))
	if err != nil {
		t.errBuilder.BindError(ctx, errors_keys.Err500MarshalWrong, fmt.Errorf("can't marshal category status access: %w", err))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)
	pg.SetParams(string(categoryStatusIDsJSON))
	pg.SetStoredProcedureName("tickets.categorystatus_getfornames")
	utils.BindRestData(ctx, repository.GetRestDataFromDb(pg))
}

func (t TicketsService) GetTicketsByIDs(ctx *gin.Context, params map[string]interface{}) {
	employeeID, ok := params["employee_id"].(int64)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("can't parse employeeID parameter"))
		return
	}

	body, ok := params[validators.BodyValidatorBODY].(*struct {
		TicketIDs []int64 `json:"ticket_ids" binding:"required,min=1,dive,gt=0"`
		Subtab    string  `json:"subtab" binding:"required,min=1,max=50"`
	})
	if !ok {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("unsupported type"))
		return
	}
	if body == nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, errors.New("passed empty body"))
		return
	}

	dbTab, err := subtab.ToDBTab(body.Subtab)
	if err != nil {
		t.errBuilder.BindError(ctx, errors_keys.ErrVldRequestBodyWrong, err)
		return
	}

	tabQuery, ok := subtab.GetTabTicketsQuerySpec(dbTab)
	if !ok {
		t.errBuilder.BindError(ctx, core_errors.ErrorKey(errors_keys.ErrorKeyValidation), fmt.Errorf("no tab tickets query for tab: %s", dbTab))
		return
	}

	paginationModelJSON, err := jsoniter.Marshal(models.PaginationModel{TicketIDs: body.TicketIDs})
	if err != nil {
		t.errBuilder.BindError(ctx, errors_keys.Err500MarshalWrong, fmt.Errorf("can't marshal pagination model: %w", err))
		return
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(models.SupportPgDatabaseKey)

	if subtab.RequiresCategoryStatusIDs(dbTab) {
		categoryStatusAccess, err := t.ticketAccess.GetEmployeeCategoryStatusAccess(ctx.Request.Context(), employeeID, access_actions.TypeActionStatusApproveCategory, access_actions.TypeActionStatusPerformCategory)
		if err != nil {
			t.errBuilder.BindError(ctx, errors_keys.Err500ReadDbResponseWrong, err)
			return
		}
		if len(categoryStatusAccess) == 0 {
			utils.BindNoContent(ctx)
			return
		}

		categoryStatusIDsJSON, err := jsoniter.Marshal(categoryStatusAccess)
		if err != nil {
			t.errBuilder.BindError(ctx, errors_keys.Err500MarshalWrong, fmt.Errorf("can't marshal category status access: %w", err))
			return
		}

		pg.SetParams(string(categoryStatusIDsJSON), employeeID, string(paginationModelJSON))
	} else {
		pg.SetParams(employeeID, string(paginationModelJSON))
	}

	pg.SetStoredProcedureName(tabQuery.ProcedureName)
	rd := repository.GetRestDataFromDb(pg)

	if rd.HasError() || rd.DataEmpty() {
		utils.BindRestData(ctx, rd)
		return
	}

	res, err := mapTicketsSectionWithChatInfo(ctx.Request.Context(), t.chatMapper, rd.GetData(), employeeID, tabQuery.ResponseJSONKey)
	if err != nil {
		logrus.Errorf("[GetTicketsByIDs] mapTicketsSectionWithChatInfo error: %v", err)
		utils.BindRestData(ctx, rd)
		return
	}
	rd.Data = &res

	utils.BindRestData(ctx, rd)
}

func mapTicketsSectionWithChatInfo(
	ctx context.Context,
	chatMapper *chat_mapper.ChatMapper,
	rawData json.RawMessage,
	employeeID int64,
	sectionKey string,
) (json.RawMessage, error) {
	const op = "mapTicketsSectionWithChatInfo"

	var sections []map[string]json.RawMessage
	if err := jsoniter.Unmarshal(rawData, &sections); err != nil {
		return rawData, fmt.Errorf("[%s]: unmarshal rawData error: %w", op, err)
	}
	if len(sections) == 0 {
		return rawData, nil
	}

	ticketsRaw, ok := sections[0][sectionKey]
	if !ok || internal_utils.IsJSONEmpty(ticketsRaw) {
		return rawData, nil
	}

	mappedTickets, err := chatMapper.MapTicketListWithChatInfo(ctx, ticketsRaw, employeeID)
	if err != nil {
		return rawData, fmt.Errorf("[%s]: MapTicketListWithChatInfo error: %w", op, err)
	}

	sections[0][sectionKey] = mappedTickets
	return jsoniter.Marshal(sections)
}

func groupCategoryStatusesByCategory(accessList []models.CategoryStatusAccess) []models.CategoryStatuses {
	statusesByCategory := make(map[int64][]string)
	for _, item := range accessList {
		if item.StatusID == nil {
			continue
		}
		statusesByCategory[item.CategoryID] = append(statusesByCategory[item.CategoryID], *item.StatusID)
	}

	result := make([]models.CategoryStatuses, 0, len(statusesByCategory))
	for categoryID, statusIDs := range statusesByCategory {
		result = append(result, models.CategoryStatuses{
			CategoryID: categoryID,
			StatusIDs:  statusIDs,
		})
	}
	return result
}
