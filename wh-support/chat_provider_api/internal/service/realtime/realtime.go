package realtime

import (
	"context"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/chat_provider_api/internal/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
)

const eventChBuffer = 128

type Realtime struct {
	band            BandClient
	centrifugo      Centrifugo
	chatStorage     ChatStorage
	activityStorage ActivityStorage
	accessStorage   AccessStorage
	eventCh         chan models.RealtimePostEvent
}

func New(
	band BandClient,
	centrifugo Centrifugo,
	chatStorage ChatStorage,
	activityStorage ActivityStorage,
	accessStorage AccessStorage,
) *Realtime {
	return &Realtime{
		band:            band,
		centrifugo:      centrifugo,
		chatStorage:     chatStorage,
		activityStorage: activityStorage,
		accessStorage:   accessStorage,
		eventCh:         make(chan models.RealtimePostEvent, eventChBuffer),
	}
}

func (r *Realtime) Start(ctx context.Context, _ configs.Config) {
	go r.startListener(ctx)
	go r.startProcessor(ctx)
}
