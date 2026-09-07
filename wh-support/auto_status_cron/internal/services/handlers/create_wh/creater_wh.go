package createwh

import (
	"context"
	"errors"
	"fmt"

	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/localization"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services"
	commonhandlersmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/common_handlers/models"
	createwhmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_wh/models"
	handlerutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/utils"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

const (
	statusIDAddBuilding                     = "AU1"
	statusIDAddPhysicalWhWithCreateBuilding = "AW1"
	statusIDAddPhysicalWhWithSelectBuilding = "AW2"
	statusIDAddVirtualWhAW3                 = "AW3"
	statusIDAddVirtualWhAW4                 = "AW4"
	statusIDCreateStages                    = "CFL"
	statusIDCreateParts                     = "CPR"
)

type TicketHandlerCreateWh struct {
	buildingApi  buildingApi
	whApi        whApi
	stageBuilder stageBuilder

	repo services.HandlerTicketsRepo
}

func NewTicketHandlerCreateWh(repo services.HandlerTicketsRepo, buildingApi buildingApi, whApi whApi, stageBuilder stageBuilder) *TicketHandlerCreateWh {
	return &TicketHandlerCreateWh{
		repo:         repo,
		buildingApi:  buildingApi,
		whApi:        whApi,
		stageBuilder: stageBuilder,
	}
}

func (h *TicketHandlerCreateWh) Process(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	switch ticketInfo.StatusID {
	case statusIDAddBuilding:
		err := h.addBuilding(ctx, ticketInfo)
		if err != nil {
			return fmt.Errorf("can't add building on physical wh: %w", err)
		}
	case statusIDAddPhysicalWhWithCreateBuilding:
		err := h.addPhysicalWhWithCreateBuilding(ctx, ticketInfo)
		if err != nil {
			return fmt.Errorf("can't add physical wh with create building: %w", err)
		}
	case statusIDAddPhysicalWhWithSelectBuilding:
		err := h.addPhysicalWhWithSelectBuilding(ctx, ticketInfo)
		if err != nil {
			return fmt.Errorf("can't add physical wh with select building: %w", err)
		}
	case statusIDAddVirtualWhAW3:
		err := h.addVirtualWhAW3(ctx, ticketInfo)
		if err != nil {
			return fmt.Errorf("can't add virtual wh AW3: %w", err)
		}
	case statusIDAddVirtualWhAW4:
		err := h.addVirtualWhAW4(ctx, ticketInfo)
		if err != nil {
			return fmt.Errorf("can't add virtual wh AW4: %w", err)
		}
	case statusIDCreateStages:
		err := h.createStages(ctx, ticketInfo)
		if err != nil {
			return fmt.Errorf("can't create stages: %w", err)
		}
	case statusIDCreateParts:
		err := h.createParts(ctx, ticketInfo)
		if err != nil {
			return fmt.Errorf("can't create parts: %w", err)
		}
	default:
		return fmt.Errorf("invalid status ID: %s", ticketInfo.StatusID)
	}

	return nil
}

func (h *TicketHandlerCreateWh) createParts(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext createwhmodels.ExtTicketInfoCreateParts

	err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext)
	if err != nil {
		return fmt.Errorf("can't decode request body: %w", err)
	}

	if len(ext.Stages) == 0 {
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

	err = h.stageBuilder.CreateParts(ctx, commonhandlersmodels.HandlerRequestForCreateParts{
		OfficeID:   ext.OfficeId.Id,
		WhID:       ext.WhId,
		Stages:     ext.Stages,
		TicketID:   ticketInfo.TicketID,
		EmployeeID: ticketInfo.CreateEmployeeID,
		PartName:   ext.PartName,
	})
	if err != nil {
		return fmt.Errorf("can't create parts: %w", err)
	}

	return nil
}

func (h *TicketHandlerCreateWh) createStages(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext createwhmodels.ExtTicketInfoCreateStages

	err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext)
	if err != nil {
		return fmt.Errorf("can't decode request body: %w", err)
	}

	if len(ext.Stages) == 0 {
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

	err = h.stageBuilder.CreateStages(ctx, commonhandlersmodels.HandlerRequestForCreateStages{
		OfficeID:   ext.OfficeId.Id,
		WhID:       ext.WhId,
		Stages:     ext.Stages,
		TicketID:   ticketInfo.TicketID,
		EmployeeID: ticketInfo.CreateEmployeeID,
	})
	if err != nil {
		return fmt.Errorf("can't create stages: %w", err)
	}

	return nil
}

func (h *TicketHandlerCreateWh) addBuilding(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext createwhmodels.ExtTicketInfoAddBuilding

	err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext)
	if err != nil {
		return fmt.Errorf("can't decode request body: %w", err)
	}

	bodyRequestAddNewBuilding := prepareRequestDataForAddNewBuilding(ext)

	bodyRequestAddNewBuilding.EmployeeID = ticketInfo.CreateEmployeeID

	err = h.buildingApi.AddNewBuilding(ctx, bodyRequestAddNewBuilding)
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
		return fmt.Errorf("can't add new building err: %w", err)
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

func (h *TicketHandlerCreateWh) addPhysicalWhWithCreateBuilding(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext createwhmodels.ExtTicketInfoAddPhysicalWhWithCreateBuilding

	err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext)
	if err != nil {
		return fmt.Errorf("can't decode request body: %w", err)
	}

	bodyRequestDataForAddNewWh := convertExtTicketInfoPhysicalWhWithCreateBuildingToRequestDataForAddNewWh(ext)

	bodyRequestDataForAddNewWh.EmployeeId = ticketInfo.CreateEmployeeID

	whId, err := h.whApi.AddNewWh(ctx, bodyRequestDataForAddNewWh)
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
		return fmt.Errorf("can't add new wh err: %w", err)
	}

	var addedInfoWh createwhmodels.AddedPerformInfoWh

	addedInfoWh.WhID = whId

	rawAddedInfoWh, err := jsoniter.Marshal(addedInfoWh)
	if err != nil {
		return fmt.Errorf("can't marshal addedInfoWh: %w", err)
	}

	err = h.repo.PerformTicket(ctx, ticketInfo.TicketID, rawAddedInfoWh)
	if err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
		if rejectErr != nil {
			logrus.Errorf("can't reject ticket %d: %v", ticketInfo.TicketID, rejectErr)
		}

		return fmt.Errorf("can't perform ticket %d: %v", ticketInfo.TicketID, err)
	}

	return nil
}

func (h *TicketHandlerCreateWh) addPhysicalWhWithSelectBuilding(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext createwhmodels.ExtTicketInfoAddPhysicalWhWithSelectBuilding

	err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext)
	if err != nil {
		return fmt.Errorf("can't decode request body: %w", err)
	}

	bodyRequestDataForAddNewWh := convertExtTicketInfoPhysicalWhWithSelectBuildingToRequestDataForAddNewWh(ext)

	bodyRequestDataForAddNewWh.EmployeeId = ticketInfo.CreateEmployeeID

	whId, err := h.whApi.AddNewWh(ctx, bodyRequestDataForAddNewWh)
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
		return fmt.Errorf("can't add new wh err: %w", err)
	}

	var addedInfoWh createwhmodels.AddedPerformInfoWh

	addedInfoWh.WhID = whId

	rawAddedInfoWh, err := jsoniter.Marshal(addedInfoWh)
	if err != nil {
		return fmt.Errorf("can't marshal addedInfoWh: %w", err)
	}

	err = h.repo.PerformTicket(ctx, ticketInfo.TicketID, rawAddedInfoWh)
	if err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
		if rejectErr != nil {
			logrus.Errorf("can't reject ticket %d: %v", ticketInfo.TicketID, rejectErr)
		}

		return fmt.Errorf("can't perform ticket %d: %v", ticketInfo.TicketID, err)
	}

	return nil
}

func (h *TicketHandlerCreateWh) addVirtualWhAW3(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext createwhmodels.ExtTicketInfoAddVirtualWhAW3

	err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext)
	if err != nil {
		return fmt.Errorf("can't decode request body: %w", err)
	}

	bodyRequestDataForAddNewWh := convertExtTicketInfoAddVirtualWhAW3ToRequestDataForAddNewWh(ext)

	bodyRequestDataForAddNewWh.EmployeeId = ticketInfo.CreateEmployeeID

	whId, err := h.whApi.AddNewWh(ctx, bodyRequestDataForAddNewWh)
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
		return fmt.Errorf("can't add new wh err: %w", err)
	}

	var addedInfoWh createwhmodels.AddedPerformInfoWh
	addedInfoWh.WhID = whId

	rawAddedInfoWh, err := jsoniter.Marshal(addedInfoWh)
	if err != nil {
		return fmt.Errorf("can't marshal addedInfoWh: %w", err)
	}

	err = h.repo.PerformTicket(ctx, ticketInfo.TicketID, rawAddedInfoWh)
	if err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
		if rejectErr != nil {
			logrus.Errorf("can't reject ticket %d: %v", ticketInfo.TicketID, rejectErr)
		}

		return fmt.Errorf("can't perform ticket %d: %v", ticketInfo.TicketID, err)
	}

	return nil
}

func (h *TicketHandlerCreateWh) addVirtualWhAW4(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext createwhmodels.ExtTicketInfoAddVirtualWhAW4

	err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext)
	if err != nil {
		return fmt.Errorf("can't decode request body: %w", err)
	}

	bodyRequestDataForAddNewWh := convertExtTicketInfoAddVirtualWhAW4ToRequestDataForAddNewWh(ext)

	bodyRequestDataForAddNewWh.EmployeeId = ticketInfo.CreateEmployeeID

	whId, err := h.whApi.AddNewWh(ctx, bodyRequestDataForAddNewWh)
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
		return fmt.Errorf("can't add new wh err: %w", err)
	}

	var addedInfoWh createwhmodels.AddedPerformInfoWh
	addedInfoWh.WhID = whId

	rawAddedInfoWh, err := jsoniter.Marshal(addedInfoWh)
	if err != nil {
		return fmt.Errorf("can't marshal addedInfoWh: %w", err)
	}

	err = h.repo.PerformTicket(ctx, ticketInfo.TicketID, rawAddedInfoWh)
	if err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
		if rejectErr != nil {
			logrus.Errorf("can't reject ticket %d: %v", ticketInfo.TicketID, rejectErr)
		}

		return fmt.Errorf("can't perform ticket %d: %v", ticketInfo.TicketID, err)
	}

	return nil
}

type buildingApi interface {
	AddNewBuilding(ctx context.Context, reqBody createwhmodels.RequestDataForAddNewBuilding) (err error)
}

type whApi interface {
	AddNewWh(ctx context.Context, reqBody createwhmodels.RequestDataForAddNewWh) (whId int64, err error)
}

type stageBuilder interface {
	CreateStages(ctx context.Context, req commonhandlersmodels.HandlerRequestForCreateStages) error
	CreateParts(ctx context.Context, req commonhandlersmodels.HandlerRequestForCreateParts) error
}
