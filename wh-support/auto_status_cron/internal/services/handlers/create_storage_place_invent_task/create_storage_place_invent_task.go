package createstorageplaceinventtask

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/localization"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services"
	createstorageplaceinventtaskmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_storage_place_invent_task/models"
	handlerutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/utils"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"
)

const (
	handlerName = "TicketHandlerCreateStoragePlaceInventTask"

	statusIDCreateBySections = "EX2"
	statusIDCreateByPlaceIDs = "EX3"

	commentTasksNotCreated = "Задания на инвентаризацию МХ не созданы."
	commentTasksCreatedFmt = "Создано заданий на инвентаризацию МХ: %d"
)

type TicketHandlerCreateStoragePlaceInventTask struct {
	repo      services.HandlerTicketsRepo
	inventApi inventApi
}

func NewTicketHandlerCreateStoragePlaceInventTask(repo services.HandlerTicketsRepo, inventApi inventApi) *TicketHandlerCreateStoragePlaceInventTask {
	return &TicketHandlerCreateStoragePlaceInventTask{
		repo:      repo,
		inventApi: inventApi,
	}
}

func (h *TicketHandlerCreateStoragePlaceInventTask) Process(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	switch ticketInfo.StatusID {
	case statusIDCreateBySections:
		if err := h.createBySections(ctx, ticketInfo); err != nil {
			return fmt.Errorf("[%s] can't create storage place invent task by sections: %w", handlerName, err)
		}
	case statusIDCreateByPlaceIDs:
		if err := h.createByPlaceIDs(ctx, ticketInfo); err != nil {
			return fmt.Errorf("[%s] can't create storage place invent task by place ids: %w", handlerName, err)
		}
	default:
		return fmt.Errorf("[%s] invalid status ID: %s", handlerName, ticketInfo.StatusID)
	}

	return nil
}

func (h *TicketHandlerCreateStoragePlaceInventTask) createBySections(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext createstorageplaceinventtaskmodels.ExtTicketInfoCreateBySections

	err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext)
	if err != nil {
		return fmt.Errorf("can't decode map to structure: %w", err)
	}

	createdQty, err := h.inventApi.CreateStoragePlaceInventTask(ctx, createstorageplaceinventtaskmodels.RequestDataForCreateStoragePlaceInventTask{
		OfficeID:   ext.OfficeID.ID,
		WhID:       ext.WhID.ID,
		Stage:      ext.Stage.ID,
		Street:     ext.Street.ID,
		EmployeeID: ext.EmployeeID.ID,
		Sections:   ext.Sections,
		TaskType:   ext.TaskType.ID,
	})
	if err != nil {
		if errWithMsg, ok := errors.AsType[internalerrors.ErrorWithMsg](err); ok {
			locMsgs := localization.New(errWithMsg.ErrKey, errWithMsg.MessageValues)
			rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, errWithMsg.Msg, locMsgs)
			if rejectErr != nil {
				return fmt.Errorf("can't reject ticket %d: %w", ticketInfo.TicketID, rejectErr)
			}

			logrus.Infof("[%s] ticket %d reject: %v", handlerName, ticketInfo.TicketID, errWithMsg.Msg)
			return nil
		}
		return fmt.Errorf("can't create storage place invent task: %w", err)
	}

	if createdQty == 0 {
		locMsgs := localization.New(support_err_keys.KeyErrorInventTaskStoragePlaceTasksNotCreated, nil)
		if rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, commentTasksNotCreated, locMsgs); rejectErr != nil {
			return fmt.Errorf("can't reject ticket %d: %w", ticketInfo.TicketID, rejectErr)
		}

		logrus.Infof("[%s] ticket %d reject: %s", handlerName, ticketInfo.TicketID, commentTasksNotCreated)
		return nil
	}

	rawPerformInfo, err := jsoniter.Marshal(createstorageplaceinventtaskmodels.AddedPerformInfoCount{
		Count: createdQty,
	})
	if err != nil {
		return fmt.Errorf("can't marshal perform info: %w", err)
	}

	if err = h.repo.PerformTicket(ctx, ticketInfo.TicketID, rawPerformInfo); err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
		if rejectErr != nil {
			logrus.Errorf("[%s] can't reject ticket %d: %v", handlerName, ticketInfo.TicketID, rejectErr)
		}

		return fmt.Errorf("can't perform ticket %d: %w", ticketInfo.TicketID, err)
	}

	return nil
}

func (h *TicketHandlerCreateStoragePlaceInventTask) createByPlaceIDs(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext createstorageplaceinventtaskmodels.ExtTicketInfoCreateByPlaceIDs

	err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext)
	if err != nil {
		return fmt.Errorf("can't decode map to structure: %w", err)
	}

	placesIDs := make([]int64, len(ext.Places))
	for i, p := range ext.Places {
		placesIDs[i] = p.ID
	}

	requestedQty := int64(len(placesIDs))

	createdQty, err := h.inventApi.CreateStoragePlaceInventTask(ctx, createstorageplaceinventtaskmodels.RequestDataForCreateStoragePlaceInventTask{
		OfficeID:   ext.OfficeID.ID,
		WhID:       ext.WhID.ID,
		Stage:      ext.Stage.ID,
		Street:     ext.Street.ID,
		EmployeeID: ext.EmployeeID.ID,
		PlaceIDs:   placesIDs,
		TaskType:   ext.TaskType.ID,
	})
	if err != nil {
		if errWithMsg, ok := errors.AsType[internalerrors.ErrorWithMsg](err); ok {
			locMsgs := localization.New(errWithMsg.ErrKey, errWithMsg.MessageValues)
			rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, errWithMsg.Msg, locMsgs)
			if rejectErr != nil {
				return fmt.Errorf("can't reject ticket %d: %w", ticketInfo.TicketID, rejectErr)
			}

			logrus.Infof("[%s] ticket %d reject: %v", handlerName, ticketInfo.TicketID, errWithMsg.Msg)
			return nil
		}
		return fmt.Errorf("can't create storage place invent task: %w", err)
	}

	if createdQty == 0 {
		locMsgs := localization.New(support_err_keys.KeyErrorInventTaskStoragePlaceTasksNotCreated, nil)
		if rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, commentTasksNotCreated, locMsgs); rejectErr != nil {
			return fmt.Errorf("can't reject ticket %d: %w", ticketInfo.TicketID, rejectErr)
		}

		logrus.Infof("[%s] ticket %d reject: %s", handlerName, ticketInfo.TicketID, commentTasksNotCreated)
		return nil
	}

	if createdQty < requestedQty {
		comment := fmt.Sprintf(commentTasksCreatedFmt, createdQty)
		locMsgs := localization.New(support_err_keys.KeyTextInventTaskStoragePlaceTasksCreated, map[string]string{
			"created_qty": strconv.FormatInt(createdQty, 10),
		})

		rawComment, err := jsoniter.Marshal(models.PerformErrValue{Comment: comment})
		if err != nil {
			return fmt.Errorf("can't marshal perform err comment: %w", err)
		}

		if err = h.repo.PerformTicketWithErrComment(ctx, ticketInfo.TicketID, nil, rawComment, locMsgs); err != nil {
			locMsgs = localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
			rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
			if rejectErr != nil {
				logrus.Errorf("[%s] can't reject ticket %d: %v", handlerName, ticketInfo.TicketID, rejectErr)
			}

			return fmt.Errorf("can't perform ticket with err comment %d: %w", ticketInfo.TicketID, err)
		}
		return nil
	}

	rawPerformInfo, err := jsoniter.Marshal(createstorageplaceinventtaskmodels.AddedPerformInfoCount{
		Count: createdQty,
	})
	if err != nil {
		return fmt.Errorf("can't marshal perform info: %w", err)
	}

	if err = h.repo.PerformTicket(ctx, ticketInfo.TicketID, rawPerformInfo); err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
		if rejectErr != nil {
			logrus.Errorf("[%s] can't reject ticket %d: %v", handlerName, ticketInfo.TicketID, rejectErr)
		}

		return fmt.Errorf("can't perform ticket %d: %w", ticketInfo.TicketID, err)
	}
	return nil
}

type inventApi interface {
	CreateStoragePlaceInventTask(ctx context.Context, body createstorageplaceinventtaskmodels.RequestDataForCreateStoragePlaceInventTask) (int64, error)
}
