package removeremainswrr

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"

	"github.com/sirupsen/logrus"
	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/localization"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services"
	removeremainswrrmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/remove_remains_wrr/models"
	handlerutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/utils"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"
)

const (
	handlerName = "TicketHandlerRemoveRemainsWRR"

	maxShkListPerTicket        = 1000
	errTooManyShkListPerTicket = "Ограничение на количество ШК в одной заявке: %d"
)

type TicketHandlerRemoveRemainsWRR struct {
	shkClaimsApi shkClaimsApi

	repo services.HandlerTicketsRepo
}

func NewTicketHandlerRemoveRemainsWRRHandler(repo services.HandlerTicketsRepo, shkClaimsApi shkClaimsApi) *TicketHandlerRemoveRemainsWRR {
	return &TicketHandlerRemoveRemainsWRR{
		repo:         repo,
		shkClaimsApi: shkClaimsApi,
	}
}

func (r *TicketHandlerRemoveRemainsWRR) Process(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext removeremainswrrmodels.ExtTicketInfoForRemoveRemainsWRR

	err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext)
	if err != nil {
		return fmt.Errorf("[%s] can't decode map to structure with err unset: %w", handlerName, err)
	}

	shkIDMap := make(map[int64]removeremainswrrmodels.ShkList, len(ext.ShkList))
	for _, shk := range ext.ShkList {
		shkIDMap[shk.ShkID.Value] = shk
	}
	ext.ShkList = slices.Collect(maps.Values(shkIDMap))

	if len(ext.ShkList) > maxShkListPerTicket {
		locMsgs := localization.New(
			support_err_keys.KeyErrorGoodsListLimitExceeded,
			map[string]string{"limit": strconv.Itoa(maxShkListPerTicket)},
		)
		rejectErr := r.repo.RejectTicket(ctx, ticketInfo.TicketID, fmt.Sprintf(errTooManyShkListPerTicket, maxShkListPerTicket), locMsgs)
		if rejectErr != nil {
			logrus.Errorf("[%s] can't reject ticket %d: %v", handlerName, ticketInfo.TicketID, rejectErr)
		}

		logrus.Infof("[%s] ticket %d reject: %v", handlerName, ticketInfo.TicketID, fmt.Sprintf(errTooManyShkListPerTicket, maxShkListPerTicket))
		return nil
	}

	requestShkList := make([]removeremainswrrmodels.RequestShkList, len(ext.ShkList))
	for i := range ext.ShkList {
		requestShkList[i].ShkID = ext.ShkList[i].ShkID.Value
		requestShkList[i].ChrtID = ext.ShkList[i].ChrtID.Value
		requestShkList[i].NmID = ext.ShkList[i].NmID.Value
	}

	err = r.shkClaimsApi.ShksRelease(ctx, removeremainswrrmodels.RequestDataForShksRelease{
		RequestShkList: requestShkList,
	}, ticketInfo.CreateEmployeeID)
	if err != nil {
		if errWithMsg, ok := errors.AsType[internalerrors.ErrorWithMsg](err); ok {
			locMsgs := localization.New(errWithMsg.ErrKey, errWithMsg.MessageValues)
			rejectErr := r.repo.RejectTicket(ctx, ticketInfo.TicketID, errWithMsg.Msg, locMsgs)
			if rejectErr != nil {
				return fmt.Errorf("[%s] can't reject ticket %d: %w", handlerName, ticketInfo.TicketID, rejectErr)
			}

			logrus.Infof("[%s] ticket %d reject: %v", handlerName, ticketInfo.TicketID, errWithMsg.Msg)

			return nil
		}
		return fmt.Errorf("[%s] can't shks release: %w", handlerName, err)
	}

	err = r.repo.PerformTicket(ctx, ticketInfo.TicketID, nil)
	if err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		rejectErr := r.repo.RejectTicket(ctx, ticketInfo.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
		if rejectErr != nil {
			logrus.Errorf("[%s] can't reject ticket %d: %v", handlerName, ticketInfo.TicketID, rejectErr)
		}

		return fmt.Errorf("[%s] can't perform ticket %d: %v", handlerName, ticketInfo.TicketID, err)
	}

	return nil
}

type shkClaimsApi interface {
	ShksRelease(ctx context.Context, body removeremainswrrmodels.RequestDataForShksRelease, employeeID int64) (err error)
}
