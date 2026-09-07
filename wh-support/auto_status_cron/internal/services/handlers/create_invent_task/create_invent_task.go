package createinventtasknew

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"

	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/localization"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services"
	createinventtaskmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_invent_task/models"
	handlerutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/utils"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"
)

const (
	statusIDCreateInventTaskNoWithdrawalBySections   = "EX1"
	statusIDCreateInventTaskWithWithdrawalBySections = "EX2"
	statusIDCreateInventTaskWithWithdrawalByPlaceIDs = "EX3"

	defaultPriorityInventTask = 2

	maxStreetsPerTicket        = 10
	errTooManyStreetsPerTicket = "Ограничение на создание задания составляет %d улиц"

	maxSectionsPerTicket        = 200
	errTooManySectionsPerTicket = "Ограничение на создание задания составляет суммарно %d секций"

	maxPlaceIDsPerTicket        = 100
	errTooManyPlaceIDsPerTicket = "Ограничение на создание задания составляет %d МХ. \nЕсли требуется больше МХ - воспользуйтесь функционалом создания заданий на снятие по секциям"
)

type TicketHandlerCreateInventTask struct {
	repo      services.HandlerTicketsRepo
	inventApi inventApi
}

func NewTicketHandlerCreateInventTask(repo services.HandlerTicketsRepo, inventApi inventApi) *TicketHandlerCreateInventTask {
	return &TicketHandlerCreateInventTask{
		repo:      repo,
		inventApi: inventApi,
	}
}

func (h *TicketHandlerCreateInventTask) Process(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	switch ticketInfo.StatusID {
	case statusIDCreateInventTaskNoWithdrawalBySections:
		err := h.createInventTaskNoWithdrawalBySections(ctx, ticketInfo)
		if err != nil {
			return fmt.Errorf("can't create invent task no withdrawal by sections: %w", err)
		}
	case statusIDCreateInventTaskWithWithdrawalBySections:
		err := h.createInventTaskWithWithdrawalBySections(ctx, ticketInfo)
		if err != nil {
			return fmt.Errorf("can't create invent task with withdrawal by sections: %w", err)
		}
	case statusIDCreateInventTaskWithWithdrawalByPlaceIDs:
		err := h.createInventTaskWithWithdrawalByPlaceIDs(ctx, ticketInfo)
		if err != nil {
			return fmt.Errorf("can't create invent task with withdrawal by place ids: %w", err)
		}
	default:
		return fmt.Errorf("invalid status ID: %s", ticketInfo.StatusID)
	}

	return nil
}

func (h *TicketHandlerCreateInventTask) createInventTaskNoWithdrawalBySections(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext createinventtaskmodels.ExtTicketInfoCreateInventTaskNoWithdrawalBySections

	err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext)
	if err != nil {
		return fmt.Errorf("can't decode map to structure: %w", err)
	}

	if len(ext.Streets) > maxStreetsPerTicket {
		locMsgs := localization.New(
			support_err_keys.KeyErrorStreetMaxCountExceeded,
			map[string]string{"max_count": strconv.Itoa(maxStreetsPerTicket)},
		)
		rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, fmt.Sprintf(errTooManyStreetsPerTicket, maxStreetsPerTicket), locMsgs)
		if rejectErr != nil {
			return fmt.Errorf("can't reject ticket %d: %w", ticketInfo.TicketID, rejectErr)
		}

		logrus.Infof("ticket %d reject: %v", ticketInfo.TicketID, fmt.Sprintf(errTooManyStreetsPerTicket, maxStreetsPerTicket))

		return nil
	}

	var sections []int64
	for _, s := range ext.Sections {
		sections = append(sections, s.SectionsRange.Value...)
	}

	if len(sections) > maxSectionsPerTicket {
		locMsgs := localization.New(
			support_err_keys.KeyErrorInventTaskSectionsLimit,
			map[string]string{"max_sections": strconv.Itoa(maxSectionsPerTicket)},
		)
		rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, fmt.Sprintf(errTooManySectionsPerTicket, maxSectionsPerTicket), locMsgs)
		if rejectErr != nil {
			return fmt.Errorf("can't reject ticket %d: %w", ticketInfo.TicketID, rejectErr)
		}

		logrus.Infof("ticket %d reject: %v", ticketInfo.TicketID, fmt.Sprintf(errTooManySectionsPerTicket, maxSectionsPerTicket))

		return nil
	}

	err = h.inventApi.CreateInventTaskNoWithdrawal(ctx, createinventtaskmodels.RequestDataForCreateInventTask{
		OfficeId:   ext.OfficeId.Id,
		WhId:       ext.WhId.Id,
		Stage:      ext.Stage.Id,
		Streets:    ext.Streets,
		Priority:   defaultPriorityInventTask,
		Sections:   sections,
		EmployeeId: ticketInfo.CreateEmployeeID,
	})
	if err != nil {
		if errWithMsg, ok := errors.AsType[internalerrors.ErrorWithMsg](err); ok {
			locMsgs := localization.New(errWithMsg.ErrKey, errWithMsg.MessageValues)
			rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, errWithMsg.Msg, locMsgs)
			if rejectErr != nil {
				return fmt.Errorf("can't reject ticket %d: %w", ticketInfo.TicketID, rejectErr)
			}

			logrus.Infof("ticket %d reject: %v", ticketInfo.TicketID, errWithMsg.Msg)

			return nil
		}
		return fmt.Errorf("can't create invent task no withdrawal by sections: %w", err)
	}

	err = h.repo.PerformTicket(ctx, ticketInfo.TicketID, nil)
	if err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
		if rejectErr != nil {
			logrus.Errorf("can't reject ticket %d: %v", ticketInfo.TicketID, rejectErr)
		}

		return fmt.Errorf("can't perform ticket %d: %w", ticketInfo.TicketID, err)
	}

	return nil
}

func (h *TicketHandlerCreateInventTask) createInventTaskWithWithdrawalBySections(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext createinventtaskmodels.ExtTicketInfoCreateInventTaskWithWithdrawalBySections

	err := handlerutils.DecodeMapToStructureWithoutErrorUnset(ticketInfo.Ext, &ext)
	if err != nil {
		return fmt.Errorf("can't decode map to structure: %w", err)
	}

	if len(ext.Streets) > maxStreetsPerTicket {
		locMsgs := localization.New(
			support_err_keys.KeyErrorStreetMaxCountExceeded,
			map[string]string{"max_count": strconv.Itoa(maxStreetsPerTicket)},
		)
		rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, fmt.Sprintf(errTooManyStreetsPerTicket, maxStreetsPerTicket), locMsgs)
		if rejectErr != nil {
			return fmt.Errorf("can't reject ticket %d: %w", ticketInfo.TicketID, rejectErr)
		}

		logrus.Infof("ticket %d reject: %v", ticketInfo.TicketID, fmt.Sprintf(errTooManyStreetsPerTicket, maxStreetsPerTicket))

		return nil
	}

	var sections []int64
	for _, s := range ext.Sections {
		sections = append(sections, s.SectionsRange.Value...)
	}

	if len(sections) > maxSectionsPerTicket {
		locMsgs := localization.New(
			support_err_keys.KeyErrorInventTaskSectionsLimit,
			map[string]string{"max_sections": strconv.Itoa(maxSectionsPerTicket)},
		)
		rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, fmt.Sprintf(errTooManySectionsPerTicket, maxSectionsPerTicket), locMsgs)
		if rejectErr != nil {
			return fmt.Errorf("can't reject ticket %d: %w", ticketInfo.TicketID, rejectErr)
		}

		logrus.Infof("ticket %d reject: %s", ticketInfo.TicketID, fmt.Sprintf(errTooManySectionsPerTicket, maxSectionsPerTicket))

		return nil
	}

	err = h.inventApi.CreateInventTaskWithWithdrawal(ctx, createinventtaskmodels.RequestDataForCreateInventTask{
		OfficeId:   ext.OfficeId.Id,
		WhId:       ext.WhId.Id,
		Stage:      ext.Stage.Id,
		Streets:    ext.Streets,
		Priority:   defaultPriorityInventTask,
		Sections:   sections,
		Racks:      emptySliceToNil(ext.Racks),
		EmployeeId: ticketInfo.CreateEmployeeID,
	})
	if err != nil {
		if errWithMsg, ok := errors.AsType[internalerrors.ErrorWithMsg](err); ok {
			locMsgs := localization.New(errWithMsg.ErrKey, errWithMsg.MessageValues)
			rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, errWithMsg.Msg, locMsgs)
			if rejectErr != nil {
				return fmt.Errorf("can't reject ticket %d: %w", ticketInfo.TicketID, rejectErr)
			}

			logrus.Infof("ticket %d reject: %v", ticketInfo.TicketID, errWithMsg.Msg)

			return nil
		}
		return fmt.Errorf("can't create invent task with withdrawal: %w", err)
	}

	err = h.repo.PerformTicket(ctx, ticketInfo.TicketID, nil)
	if err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
		if rejectErr != nil {
			logrus.Errorf("can't reject ticket %d: %v", ticketInfo.TicketID, rejectErr)
		}

		return fmt.Errorf("can't perform ticket %d: %w", ticketInfo.TicketID, err)
	}

	return nil
}

func (h *TicketHandlerCreateInventTask) createInventTaskWithWithdrawalByPlaceIDs(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext createinventtaskmodels.ExtTicketInfoCreateInventTaskWithWithdrawalByPlaceIDs

	err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext)
	if err != nil {
		return fmt.Errorf("can't decode map to structure: %w", err)
	}

	placesIDs := make([]int64, len(ext.Places))
	for i, p := range ext.Places {
		placesIDs[i] = p.PlaceId
	}

	if len(placesIDs) > maxPlaceIDsPerTicket {
		locMsgs := localization.New(
			support_err_keys.KeyErrorInventTaskPlaceIDsLimit,
			map[string]string{"max_place_ids": strconv.Itoa(maxPlaceIDsPerTicket)},
		)
		rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, fmt.Sprintf(errTooManyPlaceIDsPerTicket, maxPlaceIDsPerTicket), locMsgs)
		if rejectErr != nil {
			return fmt.Errorf("can't reject ticket %d: %w", ticketInfo.TicketID, rejectErr)
		}

		logrus.Infof("ticket %d reject: %v", ticketInfo.TicketID, fmt.Sprintf(errTooManyPlaceIDsPerTicket, maxPlaceIDsPerTicket))

		return nil
	}

	err = h.inventApi.CreateInventTaskWithWithdrawal(ctx, createinventtaskmodels.RequestDataForCreateInventTask{
		OfficeId:   ext.OfficeId.Id,
		WhId:       ext.WhId.Id,
		Stage:      ext.Stage.Id,
		Streets:    []int64{ext.Street},
		Priority:   defaultPriorityInventTask,
		PlaceIDs:   placesIDs,
		EmployeeId: ticketInfo.CreateEmployeeID,
	})
	if err != nil {
		if errWithMsg, ok := errors.AsType[internalerrors.ErrorWithMsg](err); ok {
			locMsgs := localization.New(errWithMsg.ErrKey, errWithMsg.MessageValues)
			rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, errWithMsg.Msg, locMsgs)
			if rejectErr != nil {
				return fmt.Errorf("can't reject ticket %d: %w", ticketInfo.TicketID, rejectErr)
			}

			logrus.Infof("ticket %d reject: %v", ticketInfo.TicketID, errWithMsg.Msg)

			return nil
		}
		return fmt.Errorf("can't create invent task with withdrawal: %w", err)
	}

	err = h.repo.PerformTicket(ctx, ticketInfo.TicketID, nil)
	if err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
		if rejectErr != nil {
			logrus.Errorf("can't reject ticket %d: %v", ticketInfo.TicketID, rejectErr)
		}

		return fmt.Errorf("can't perform ticket %d: %w", ticketInfo.TicketID, err)
	}

	return nil
}

type inventApi interface {
	CreateInventTaskNoWithdrawal(ctx context.Context, reqBody createinventtaskmodels.RequestDataForCreateInventTask) error
	CreateInventTaskWithWithdrawal(ctx context.Context, body createinventtaskmodels.RequestDataForCreateInventTask) error
}

func emptySliceToNil[T any](s []T) []T {
	if len(s) == 0 {
		return nil
	}
	return s
}
