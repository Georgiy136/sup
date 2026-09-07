package internalresort

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
	internalresortmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/internal_resort/models"
	handlerutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/utils"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/maps"
)

const (
	handlerName = "TicketHandlerInternalResort"

	maxGoodsIDPerTicket = 1000
	maxURLsPerTicket    = 5
	maxURLSize          = 150

	errMsgNotValidGoods         = "Нет валидных товаров"
	errMsgTooManyGoodsPerTicket = "Максимальное количество товаров в одной заявке не должно превышать: %d"
	errMsgInvalidComment        = "Размер комментария превышает допустимое значение"
	errMsgTooManyURLsPerTicket  = "Количество ссылок на заявку не должно превышать: %d"
	errMsgInvalidSizeURL        = "Размер ссылки превышает допустимое значение символов: %d"

	statusIDCheckUploadedGoodsIDBySeller    = "ECS"
	statusIDCheckUploadedGoodsIDByEmployee  = "ECE"
	statusIDSendTaskToProfileTeamBySeller   = "E1S"
	statusIDSendTaskToProfileTeamByEmployee = "E1E"
)

type TicketHandlerInternalResort struct {
	repo                   services.HandlerTicketsRepo
	shkClaimsApi           shkClaimsApi
	goodsIdentificationApi goodsIdentificationApi
}

func NewTicketHandlerInternalResortHandler(repo services.HandlerTicketsRepo, shkClaimsApi shkClaimsApi, goodsIdentificationApi goodsIdentificationApi) *TicketHandlerInternalResort {
	return &TicketHandlerInternalResort{
		repo:                   repo,
		shkClaimsApi:           shkClaimsApi,
		goodsIdentificationApi: goodsIdentificationApi,
	}
}

func (t *TicketHandlerInternalResort) Process(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	switch ticketInfo.StatusID {
	case statusIDCheckUploadedGoodsIDBySeller, statusIDCheckUploadedGoodsIDByEmployee:
		if err := t.checkUploadedGoodsIDs(ctx, ticketInfo); err != nil {
			return fmt.Errorf("[%s] can't check uploaded goods ids: %w", handlerName, err)
		}
	case statusIDSendTaskToProfileTeamBySeller:
		if err := t.sendTaskToProfileTeamBySeller(ctx, ticketInfo); err != nil {
			return fmt.Errorf("[%s] can't send task to profile team by seller: %w", handlerName, err)
		}

	case statusIDSendTaskToProfileTeamByEmployee:
		if err := t.sendTaskToProfileTeamByEmployee(ctx, ticketInfo); err != nil {
			return fmt.Errorf("[%s] can't send task to profile team by employee: %w", handlerName, err)
		}
	default:
		return fmt.Errorf("invalid status ID: %s", ticketInfo.StatusID)
	}

	return nil
}

func (t *TicketHandlerInternalResort) checkUploadedGoodsIDs(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext internalresortmodels.ExtTicketInfoCheckUploadedGoodsIDs

	if err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext); err != nil {
		return fmt.Errorf("[%s] can't decode map to structure: %w", handlerName, err)
	}

	goodsIDs := make(map[int64]struct{}, len(ext.GoodsIDs))
	for i := range ext.GoodsIDs {
		goodsIDs[ext.GoodsIDs[i].GoodsID.Value] = struct{}{}
	}

	if len(goodsIDs) > maxGoodsIDPerTicket {
		locMsgs := localization.New(
			support_err_keys.KeyErrorGoodsListLimitExceeded,
			map[string]string{"limit": strconv.Itoa(maxGoodsIDPerTicket)},
		)

		if rejectErr := t.repo.RejectTicket(ctx, ticketInfo.TicketID, fmt.Sprintf(errMsgTooManyGoodsPerTicket, maxGoodsIDPerTicket), locMsgs); rejectErr != nil {
			return fmt.Errorf("[%s] can't reject ticket %d: %w", handlerName, ticketInfo.TicketID, rejectErr)
		}

		logrus.Infof("[%s] ticket %d reject: %s", handlerName, ticketInfo.TicketID, fmt.Sprintf(errMsgTooManyGoodsPerTicket, maxGoodsIDPerTicket))
		return nil
	}

	reqBody := internalresortmodels.RequestDataForGoodsValidation{
		GoodsIDs: maps.Keys(goodsIDs),
	}
	respGoodsValidation, err := t.shkClaimsApi.GoodsValidation(ctx, reqBody, ticketInfo.CreateEmployeeID)
	if err != nil {
		return fmt.Errorf("[%s] can't validate goods ids err: %w", handlerName, err)
	}

	notValidGoodsIDs := make(map[string][]int64, len(respGoodsValidation))
	for i := range respGoodsValidation {
		comment := respGoodsValidation[i].Comment
		notValidGoodsIDs[comment] = append(notValidGoodsIDs[comment], respGoodsValidation[i].GoodsIDs...)

		for _, v := range respGoodsValidation[i].GoodsIDs {
			delete(goodsIDs, v)
		}
	}

	if len(goodsIDs) == 0 {
		rejectValues := map[string]any{
			"values_map": notValidGoodsIDs,
		}

		locMsgs := localization.New(support_err_keys.KeyErrorInternalResortGoodsNotValid, nil)
		if rejectErr := t.repo.RejectTicketWithValues(ctx, ticketInfo.TicketID, errMsgNotValidGoods, locMsgs, rejectValues); rejectErr != nil {
			return fmt.Errorf("[%s] can't reject ticket %d: %w", handlerName, ticketInfo.TicketID, rejectErr)
		}

		logrus.Infof("[%s] ticket %d reject: %s", handlerName, ticketInfo.TicketID, errMsgNotValidGoods)
		return nil
	}

	performInfo := internalresortmodels.PerformInfoCheckUploadedGoodsIDs{
		ValidGoodsIDs:    getPerformValidGoodsIDs(goodsIDs),
		NotValidGoodsIDs: getPerformNotValidGoodsIDs(notValidGoodsIDs),
	}

	rawPerformInfo, err := jsoniter.Marshal(performInfo)
	if err != nil {
		return fmt.Errorf("[%s] can't marshal performInfo: %w", handlerName, err)
	}

	if err = t.repo.PerformTicket(ctx, ticketInfo.TicketID, rawPerformInfo); err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		if rejectErr := t.repo.RejectTicket(ctx, ticketInfo.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs); rejectErr != nil {
			logrus.Errorf("[%s] can't reject ticket %d: %v", handlerName, ticketInfo.TicketID, rejectErr)
		}

		return fmt.Errorf("[%s] can't perform ticket %d: %v", handlerName, ticketInfo.TicketID, err)
	}

	return nil
}

func (t *TicketHandlerInternalResort) sendTaskToProfileTeamBySeller(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	const contragentCode = "SEL"

	var ext internalresortmodels.ExtTicketInfoSendTaskToProfileTeamBySeller
	if err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext); err != nil {
		return fmt.Errorf("[%s] can't decode map to structure: %w", handlerName, err)
	}
	infoForSendTask := internalresortmodels.InfoForSendTaskToProfileTeam{
		Proof:          ext.Proof,
		ContragentCode: contragentCode,
		ContragentID:   ext.NmId.ID,
		NmIDNew:        ext.NmIDNew,
		ValidGoodsIDs:  ext.ValidGoodsIDs,
		URLs:           ext.URLs,
	}

	return t.sendTaskToProfileTeam(ctx, ticketInfo, infoForSendTask)
}

func (t *TicketHandlerInternalResort) sendTaskToProfileTeamByEmployee(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	const contragentCode = "EMP"

	var ext internalresortmodels.ExtTicketInfoSendTaskToProfileTeamByEmployee
	if err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext); err != nil {
		return fmt.Errorf("[%s] can't decode map to structure: %w", handlerName, err)
	}
	infoForSendTask := internalresortmodels.InfoForSendTaskToProfileTeam{
		Proof:          ext.Proof,
		ContragentCode: contragentCode,
		ContragentID:   ext.EmployeeID.ID,
		NmIDNew:        ext.NmIDNew,
		ValidGoodsIDs:  ext.ValidGoodsIDs,
		URLs:           ext.URLs,
	}

	return t.sendTaskToProfileTeam(ctx, ticketInfo, infoForSendTask)
}

func (t *TicketHandlerInternalResort) sendTaskToProfileTeam(ctx context.Context, ticketInfo models.TicketCommonInfo, infoForSendTask internalresortmodels.InfoForSendTaskToProfileTeam) error {
	if err := validateCommentSize(infoForSendTask.Proof); err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorCommentSizeExceeded, map[string]string{
			"limit": strconv.Itoa(maxCommentSizeSymbol),
		})
		rejectErr := t.repo.RejectTicket(ctx, ticketInfo.TicketID, errMsgInvalidComment, locMsgs)
		if rejectErr != nil {
			return fmt.Errorf("[%s] can't reject ticket %d: %w", handlerName, ticketInfo.TicketID, rejectErr)
		}

		logrus.Infof("[%s] ticket %d reject: %s", handlerName, ticketInfo.TicketID, errMsgInvalidComment)
		return nil
	}

	goodsIDs := make([]int64, 0, len(infoForSendTask.ValidGoodsIDs))
	for i := range infoForSendTask.ValidGoodsIDs {
		goodsIDs = append(goodsIDs, infoForSendTask.ValidGoodsIDs[i].GoodsID.Value)
	}

	if len(infoForSendTask.URLs) > maxURLsPerTicket {
		errMsg := fmt.Sprintf(errMsgTooManyURLsPerTicket, maxURLsPerTicket)
		locMsgs := localization.New(support_err_keys.KeyErrorURLLimitExceeded, map[string]string{
			"limit": strconv.Itoa(maxURLsPerTicket),
		})

		rejectErr := t.repo.RejectTicket(ctx, ticketInfo.TicketID, errMsg, locMsgs)
		if rejectErr != nil {
			return fmt.Errorf("[%s] can't reject ticket %d: %w", handlerName, ticketInfo.TicketID, rejectErr)
		}

		logrus.Infof("[%s] ticket %d reject: %s", handlerName, ticketInfo.TicketID, errMsg)
		return nil
	}

	urls := make([]string, 0, len(infoForSendTask.URLs))
	for i := range infoForSendTask.URLs {
		if len(infoForSendTask.URLs[i].URL.Value) == 0 {
			continue
		}

		if len(infoForSendTask.URLs[i].URL.Value) > maxURLSize {
			errMsg := fmt.Sprintf(errMsgInvalidSizeURL, maxURLSize)
			locMsgs := localization.New(support_err_keys.KeyErrorURLSizeExceeded, map[string]string{
				"limit": strconv.Itoa(maxURLSize),
			})

			rejectErr := t.repo.RejectTicket(ctx, ticketInfo.TicketID, errMsg, locMsgs)
			if rejectErr != nil {
				return fmt.Errorf("[%s] can't reject ticket %d: %w", handlerName, ticketInfo.TicketID, rejectErr)
			}

			logrus.Infof("[%s] ticket %d reject: %s", handlerName, ticketInfo.TicketID, errMsg)
			return nil
		}

		urls = append(urls, infoForSendTask.URLs[i].URL.Value)
	}

	reqBody := internalresortmodels.RequestCreateInternalOrder{
		GoodsIDs:       goodsIDs,
		NmID:           infoForSendTask.NmIDNew,
		ContragentCode: infoForSendTask.ContragentCode,
		ContragentID:   infoForSendTask.ContragentID,
		Comment:        infoForSendTask.Proof,
		PhotoURLs:      urls,
	}
	resp, err := t.goodsIdentificationApi.CreateInternalOrder(ctx, reqBody, ticketInfo.CreateEmployeeID)
	if err != nil {
		if errWithMsg, ok := errors.AsType[internalerrors.ErrorWithMsg](err); ok {
			locMsgs := localization.New(errWithMsg.ErrKey, errWithMsg.MessageValues)
			rejectErr := t.repo.RejectTicket(ctx, ticketInfo.TicketID, errWithMsg.Msg, locMsgs)
			if rejectErr != nil {
				return fmt.Errorf("[%s] can't reject ticket %d: %w", handlerName, ticketInfo.TicketID, rejectErr)
			}

			logrus.Infof("ticket %d reject: %v", ticketInfo.TicketID, errWithMsg.Msg)
			return nil
		}

		return fmt.Errorf("[%s] can't create internal order by %s err: %w", handlerName, infoForSendTask.ContragentCode, err)
	}

	if resp == nil {
		if err := t.repo.PerformTicket(ctx, ticketInfo.TicketID, nil); err != nil {
			locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
			if rejectErr := t.repo.RejectTicket(ctx, ticketInfo.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs); rejectErr != nil {
				logrus.Errorf("[%s] can't reject ticket %d: %v", handlerName, ticketInfo.TicketID, rejectErr)
			}

			return fmt.Errorf("[%s] can't perform ticket %d: %v", handlerName, ticketInfo.TicketID, err)
		}

		return nil
	}

	goodsIDsErr := make(map[string][]int64)
	for i := range resp.GoodsIDsErrProcessing {
		comm := resp.GoodsIDsErrProcessing[i].Error
		goodsIDsErr[comm] = append(goodsIDsErr[comm], resp.GoodsIDsErrProcessing[i].GoodsID)
	}

	if len(goodsIDsErr) == len(goodsIDs) {
		rejectValues := map[string]any{
			"values_map": goodsIDsErr,
		}

		locMsgs := localization.New(support_err_keys.KeyErrorInternalResortGoodsNotValid, nil)
		if rejectErr := t.repo.RejectTicketWithValues(ctx, ticketInfo.TicketID, errMsgNotValidGoods, locMsgs, rejectValues); rejectErr != nil {
			return fmt.Errorf("[%s] can't reject ticket %d: %w", handlerName, ticketInfo.TicketID, rejectErr)
		}

		logrus.Infof("[%s] ticket %d reject: %s", handlerName, ticketInfo.TicketID, errMsgNotValidGoods)
		return nil
	}

	notValidGoodsIDs := getPerformNotValidGoodsIDs(goodsIDsErr)
	performInfo := internalresortmodels.PerformInfoSendTaskToProfileTeam{
		NotValidGoodsIDs: notValidGoodsIDs,
	}

	rawPerformInfo, err := jsoniter.Marshal(performInfo)
	if err != nil {
		return fmt.Errorf("[%s] can't marshal perform info: %w", handlerName, err)
	}

	if err := t.repo.PerformTicket(ctx, ticketInfo.TicketID, rawPerformInfo); err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		if rejectErr := t.repo.RejectTicket(ctx, ticketInfo.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs); rejectErr != nil {
			logrus.Errorf("[%s] can't reject ticket %d: %v", handlerName, ticketInfo.TicketID, rejectErr)
		}

		return fmt.Errorf("[%s] can't perform ticket %d: %v", handlerName, ticketInfo.TicketID, err)
	}

	return nil
}

type shkClaimsApi interface {
	GoodsValidation(ctx context.Context, body internalresortmodels.RequestDataForGoodsValidation, employeeID int64) ([]internalresortmodels.ResponseDataForGoodsValidation, error)
}

type goodsIdentificationApi interface {
	CreateInternalOrder(ctx context.Context, body internalresortmodels.RequestCreateInternalOrder, employeeID int64) (*internalresortmodels.ResponseCreateInternalOrder, error)
}
