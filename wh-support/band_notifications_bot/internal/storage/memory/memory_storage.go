package memory

import (
	"context"
	"strconv"

	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/cache"

	"github.com/sirupsen/logrus"
)

type InMemoryCache struct {
	cache cache.Cache
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{}
}

func (i *InMemoryCache) Configure(_ context.Context, config configs.Config) {
	confRaw := config.GetByServiceKeyRequired("cache_config")

	usersCache, err := cache.NewCache(confRaw)
	if err != nil {
		logrus.Panicf("can't init users cache: %v", err)
	}

	i.cache = usersCache
}

func (i *InMemoryCache) GetBandUserID(employeeID int64) (string, bool) {
	raw, ok := i.cache.Get(strconv.FormatInt(employeeID, 10))
	if !ok {
		return "", false
	}
	return string(raw), true
}

func (i *InMemoryCache) SetBandUserID(employeeID int64, userID string) {
	i.cache.Set(strconv.FormatInt(employeeID, 10), []byte(userID))
}
