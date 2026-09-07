package createstorageplace

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	internalerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/localization"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services"
	createstorageplacemodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/create_storage_place/models"
	handlerutils "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/handlers/utils"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
	"gitlab.wildberries.ru/wbwh/support/utils.git/support_err_keys"
)

const (
	handlerName = "TicketHandlerCreateStoragePlace"

	maxStoragePlacesPerTicket = 4920

	errIncorrectSectionInterval         = "Ошибка составления заявки: Проверьте правильность введенных секций\n"
	errIncorrectRackInterval            = "Ошибка составления заявки: Проверьте правильность введенных полок\n"
	errIncorrectFieldInterval           = "Ошибка составления заявки: Проверьте правильность введенных ячеек\n"
	errMsgTooManyStoragePlacesPerTicket = "Суммарное количество МХ не должно превышать %d\n"
)

var (
	errWrongSectionsInterval = errors.New("wrong sections interval")
	errWrongRacksInterval    = errors.New("wrong racks interval")
	errWrongFieldsInterval   = errors.New("wrong fields interval")
	errTooManyStoragePlaces  = errors.New("too many storage places")
)

type TicketHandlerCreateStoragePlace struct {
	repo            services.HandlerTicketsRepo
	storagePlaceApi storagePlaceApi
}

func NewTicketHandlerCreateStoragePlace(repo services.HandlerTicketsRepo, storagePlaceApi storagePlaceApi) *TicketHandlerCreateStoragePlace {
	return &TicketHandlerCreateStoragePlace{
		repo:            repo,
		storagePlaceApi: storagePlaceApi,
	}
}

func (h *TicketHandlerCreateStoragePlace) Process(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	if err := h.createStoragePlace(ctx, ticketInfo); err != nil {
		return fmt.Errorf("[%s] can't create storage place: %w", handlerName, err)
	}
	return nil
}

func (h *TicketHandlerCreateStoragePlace) createStoragePlace(ctx context.Context, ticketInfo models.TicketCommonInfo) error {
	var ext createstorageplacemodels.ExtTicketInfoForCreateStoragePlace

	err := handlerutils.DecodeMapToStructureWithErrorUnset(ticketInfo.Ext, &ext)
	if err != nil {
		return fmt.Errorf("[%s] can't decode map to structure: %w", handlerName, err)
	}

	err = h.validateBody(ext)
	if err != nil {
		var errMsgSB strings.Builder
		locMsgs := localization.LocalizedErrors{}

		if errors.Is(err, errWrongSectionsInterval) {
			errMsgSB.WriteString(errIncorrectSectionInterval)
			locMsgs.Add(support_err_keys.KeyErrorStreetSectionsInvalid, nil)
		}

		if errors.Is(err, errWrongRacksInterval) {
			errMsgSB.WriteString(errIncorrectRackInterval)
			locMsgs.Add(support_err_keys.KeyErrorStoragePlaceRacksInvalid, nil)
		}

		if errors.Is(err, errWrongFieldsInterval) {
			errMsgSB.WriteString(errIncorrectFieldInterval)
			locMsgs.Add(support_err_keys.KeyErrorStoragePlaceFieldsInvalid, nil)
		}

		if errors.Is(err, errTooManyStoragePlaces) {
			errMsgSB.WriteString(fmt.Sprintf(errMsgTooManyStoragePlacesPerTicket, maxStoragePlacesPerTicket))
			locMsgs.Add(support_err_keys.KeyErrorStoragePlaceLimitExceeded, map[string]string{
				"max_count": strconv.Itoa(maxStoragePlacesPerTicket),
			})
		}

		rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, errMsgSB.String(), &locMsgs)
		if rejectErr != nil {
			return fmt.Errorf("[%s] can't reject ticket %d: %w", handlerName, ticketInfo.TicketID, rejectErr)
		}

		logrus.Infof("ticket %d reject: %v", ticketInfo.TicketID, errMsgSB.String())
		return nil
	}

	storagePlaces, err := h.storagePlaceApi.CreateStoragePlace(ctx, createstorageplacemodels.RequestDataForCreateStoragePlace{
		OfficeID:     ext.OfficeID.ID,
		WhID:         ext.WhID.ID,
		PlaceTypeID:  ext.PlaceType.ID,
		Part:         ext.Part.ID,
		Stage:        ext.Stage.ID,
		Street:       ext.Street,
		SectionFirst: ext.Sections[0],
		SectionLast:  ext.Sections[len(ext.Sections)-1],
		RackFirst:    ext.Racks[0],
		RackLast:     ext.Racks[len(ext.Racks)-1],
		FieldFirst:   ext.Fields[0],
		FieldLast:    ext.Fields[len(ext.Fields)-1],
		EmployeeID:   ticketInfo.CreateEmployeeID,
	})
	if err != nil {
		if errWithMsg, ok := errors.AsType[internalerrors.ErrorWithMsg](err); ok {
			locMsgs := localization.New(errWithMsg.ErrKey, errWithMsg.MessageValues)
			rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, errWithMsg.Msg, locMsgs)
			if rejectErr != nil {
				return fmt.Errorf("[%s] can't reject ticket %d: %w", handlerName, ticketInfo.TicketID, rejectErr)
			}

			logrus.Infof("[%s] ticket %d reject: %v", handlerName, ticketInfo.TicketID, errWithMsg.Msg)

			return nil
		}
		return fmt.Errorf("[%s] can't create storage place err: %w", handlerName, err)
	}

	addedInfoPlaces := createstorageplacemodels.AddedPerformInfoPlaces{
		Places: make([]createstorageplacemodels.Place, len(storagePlaces)),
	}
	placeIDFrontDataName := "place_id"
	placeNameFrontDataName := "place_name"

	for i := range storagePlaces {

		addedInfoPlaces.Places[i].PlaceID = createstorageplacemodels.PlaceID{
			Value:         storagePlaces[i].PlaceID,
			OrderID:       1,
			FrontDataName: placeIDFrontDataName,
		}
		addedInfoPlaces.Places[i].PlaceName = createstorageplacemodels.PlaceName{
			Value:         storagePlaces[i].PlaceName,
			OrderID:       2,
			FrontDataName: placeNameFrontDataName,
		}
	}

	rawAddedInfoPlaces, err := jsoniter.Marshal(addedInfoPlaces)
	if err != nil {
		return fmt.Errorf("[%s] can't marshal addedInfoPlaces: %w", handlerName, err)
	}

	err = h.repo.PerformTicket(ctx, ticketInfo.TicketID, rawAddedInfoPlaces)
	if err != nil {
		locMsgs := localization.New(support_err_keys.KeyErrorExecutionFailed, nil)
		rejectErr := h.repo.RejectTicket(ctx, ticketInfo.TicketID, internalerrors.CommentErrorDuringPerformTicket, locMsgs)
		if rejectErr != nil {
			logrus.Errorf("[%s] can't reject ticket %d: %v", handlerName, ticketInfo.TicketID, rejectErr)
		}

		return fmt.Errorf("[%s] can't perform ticket %d: %w", handlerName, ticketInfo.TicketID, err)
	}
	return nil
}

func (h *TicketHandlerCreateStoragePlace) validateBody(ext createstorageplacemodels.ExtTicketInfoForCreateStoragePlace) error {
	var errs []error

	if len(ext.Sections) == 0 || ext.Sections[0] > ext.Sections[len(ext.Sections)-1] {
		errs = append(errs, errWrongSectionsInterval)
	}

	if len(ext.Racks) == 0 || ext.Racks[0] > ext.Racks[len(ext.Racks)-1] {
		errs = append(errs, errWrongRacksInterval)
	}

	if len(ext.Fields) == 0 || ext.Fields[0] > ext.Fields[len(ext.Fields)-1] {
		errs = append(errs, errWrongFieldsInterval)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	sectionFirst := ext.Sections[0]
	sectionLast := ext.Sections[len(ext.Sections)-1]
	rackFirst := ext.Racks[0]
	rackLast := ext.Racks[len(ext.Racks)-1]
	fieldFirst := ext.Fields[0]
	fieldLast := ext.Fields[len(ext.Fields)-1]

	if sumStoragePlaces := (sectionLast - sectionFirst + 1) * (rackLast - rackFirst + 1) * (fieldLast - fieldFirst + 1); sumStoragePlaces > maxStoragePlacesPerTicket {
		return errTooManyStoragePlaces
	}

	return nil
}

type storagePlaceApi interface {
	CreateStoragePlace(ctx context.Context, body createstorageplacemodels.RequestDataForCreateStoragePlace) ([]createstorageplacemodels.StoragePlace, error)
}
