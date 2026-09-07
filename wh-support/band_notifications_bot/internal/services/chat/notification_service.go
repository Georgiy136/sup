package chat

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/consts"
	customerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/models"
	cronmodels "gitlab.wildberries.ru/wbwh/wh-core/gocore_cron_core.git/cron_core/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

type EventHandler interface {
	Event() string
	Handle(ctx context.Context, publications []models.Publication) error
}

type ChatNotificationsService struct {
	historyProvider historyProvider
	cache           offsetCache
	handlers        map[string]EventHandler
	channel         string
	cronCfg         cronmodels.CronCommonCfg
}

func (s *ChatNotificationsService) Configure(_ context.Context, _ configs.Config, cronConf cronmodels.CronCommonCfg) {
	s.cronCfg = cronConf
}

func NewChatNotificationsService(historyProvider historyProvider, cache offsetCache, eventHandlers ...EventHandler) *ChatNotificationsService {
	handlersByEvent := make(map[string]EventHandler, len(eventHandlers))
	for _, h := range eventHandlers {
		handlersByEvent[h.Event()] = h
	}

	return &ChatNotificationsService{
		historyProvider: historyProvider,
		cache:           cache,
		handlers:        handlersByEvent,
		channel:         consts.ChatNotificationsChannel,
	}
}

func (s *ChatNotificationsService) ProcessNotifications(ctx context.Context) (time.Duration, error) {
	var since *models.StreamPosition

	savedOffset, err := s.cache.GetSavedOffset(ctx)
	if err != nil {
		if !errors.Is(err, customerrors.ErrKeyNotFound) {
			return s.cronCfg.CronTimeSleepOnErrorParsed, fmt.Errorf("can't get saved offset: %w", err)
		}
	}
	if savedOffset != nil {
		since = &models.StreamPosition{Epoch: savedOffset.Epoch, Offset: savedOffset.Offset}
	}

	historyResult, err := s.historyProvider.History(ctx, s.channel, since)
	if err != nil {
		if !errors.Is(err, customerrors.ErrHistoryExpired) {
			return s.cronCfg.CronTimeSleepOnErrorParsed, fmt.Errorf("can't get history: %w", err)
		}
		logrus.Warnf("history expired, fetch without offset: %v", err)
		historyResult, err = s.historyProvider.History(ctx, s.channel, nil)
		if err != nil {
			return s.cronCfg.CronTimeSleepOnErrorParsed, fmt.Errorf("can't get history: %w", err)
		}
	}

	if len(historyResult.Publications) == 0 {
		if err = s.saveOffset(ctx, historyResult.Epoch, historyResult.Offset); err != nil {
			return s.cronCfg.CronTimeSleepOnErrorParsed, fmt.Errorf("can't save offset: %w", err)
		}
		return s.cronCfg.CronTimeSleepOnNoDataParsed, customerrors.ErrEmptyData
	}

	for event, pubs := range s.groupPublicationsByEvent(historyResult.Publications) {
		handler, ok := s.handlers[event]
		if !ok {
			logrus.Errorf("no handler registered for event: %s, skip %d publications", event, len(pubs))
			continue
		}
		if err = handler.Handle(ctx, pubs); err != nil {
			logrus.Errorf("event: %s, handler error: %v", event, err)
		}
	}

	lastPub := historyResult.Publications[len(historyResult.Publications)-1]
	if err = s.saveOffset(ctx, historyResult.Epoch, lastPub.Offset); err != nil {
		return s.cronCfg.CronTimeSleepOnErrorParsed, fmt.Errorf("can't save offset: %w", err)
	}

	return s.cronCfg.CronTimeSleepOnOkParsed, nil
}

func (s *ChatNotificationsService) groupPublicationsByEvent(publications []models.Publication) map[string][]models.Publication {
	grouped := make(map[string][]models.Publication, len(s.handlers))

	for idx := range publications {
		event := extractEvent(publications[idx].Data)
		if event == "" {
			logrus.Errorf("can't extract event from publication, offset: %d", publications[idx].Offset)
			continue
		}
		grouped[event] = append(grouped[event], publications[idx])
	}
	return grouped
}

func extractEvent(rawData jsoniter.RawMessage) string {
	return jsoniter.Get(rawData, "event").ToString()
}

func (s *ChatNotificationsService) saveOffset(ctx context.Context, epoch string, offset int64) error {
	return s.cache.SaveOffset(ctx, models.HistoryOffset{
		Epoch:     epoch,
		Offset:    offset,
		UpdatedAt: time.Now(),
	})
}

type historyProvider interface {
	History(ctx context.Context, channel string, since *models.StreamPosition) (*models.HistoryResult, error)
}

type offsetCache interface {
	GetSavedOffset(ctx context.Context) (*models.HistoryOffset, error)
	SaveOffset(ctx context.Context, pos models.HistoryOffset) error
}
