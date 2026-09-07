package cache

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/telegram_approve_bot/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_utils.git/cache"
	"strconv"
)

type Cache struct {
	cache cache.Cache
}

func NewCache(config configs.Config) *Cache {
	confRaw := config.GetByServiceKeyRequired("cache_config")

	cache, err := cache.NewCache(confRaw)
	if err != nil {
		logrus.Fatalf("can't init cache: %v", err)
	}

	return &Cache{
		cache: cache,
	}
}

func (c *Cache) SaveTelegramExistInfoWithBuildInTTL(userID int64, isExist bool) error {
	tgExistInfo := models.TelegramExistInfo{
		IsExist: isExist,
	}

	rawData, err := jsoniter.Marshal(tgExistInfo)
	if err != nil {
		return fmt.Errorf("can't marshal telegram exist info: %v", err)
	}

	c.cache.Set(strconv.FormatInt(userID, 10), rawData)

	return nil
}

func (c *Cache) GetTelegramExistInfo(userID int64) (*models.TelegramExistInfo, error) {
	rawData, exist := c.cache.Get(strconv.FormatInt(userID, 10))
	if !exist {
		return nil, nil
	}

	var tgExistInfo models.TelegramExistInfo
	err := jsoniter.Unmarshal(rawData, &tgExistInfo)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal telegram exist info: %v", err)
	}

	return &tgExistInfo, nil
}
