package redis

import (
	"context"
	"fmt"
	"time"
)

type ActivityStorage struct {
	cache *Cache
}

func NewActivityStorage(cache *Cache) *ActivityStorage {
	return &ActivityStorage{cache: cache}
}

const lastActivityTTL = 7 * 24 * time.Hour

func (a *ActivityStorage) SetTicketLastActivity(ctx context.Context, ticketID int64, at time.Time) error {
	key := fmt.Sprintf("ticket_chat:%d:last_activity", ticketID)
	if err := a.cache.Client.Set(ctx, key, at.Format(time.RFC3339Nano), lastActivityTTL).Err(); err != nil {
		return fmt.Errorf("set ticket last activity: %w", err)
	}
	return nil
}
