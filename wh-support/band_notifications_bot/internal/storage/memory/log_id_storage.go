package memory

import (
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_cron_core.git/cron_core"
)

type InMemoryLogIdStorage struct {
	logID *cron_core.LogID
}

func NewInMemoryLogIdStorage() *InMemoryLogIdStorage {
	return &InMemoryLogIdStorage{
		logID: cron_core.NewInMemoryLogID(),
	}
}

func (l *InMemoryLogIdStorage) GetLogID() int64 {
	return l.logID.GetLastLogID()
}

func (l *InMemoryLogIdStorage) SetLogID(id int64) {
	l.logID.SetLogID(id)
}
