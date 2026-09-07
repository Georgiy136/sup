package db

import (
	"fmt"

	ticketmodels "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/band_notifications_bot/internal/services/ticket/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

const databaseKey = "support_pgx"

type NotificationsRepo struct {
}

func NewNotificationsRepo() *NotificationsRepo {
	return &NotificationsRepo{}
}

func (r *NotificationsRepo) GetTicketNotifications(logID int64) (*ticketmodels.TicketNotifications, error) {
	getNotificationsSP := "sync.notificationband_exporttojson"

	db := new(postgresql.PgSpec)
	db.SetDatabaseKey(databaseKey)
	db.SetParams(logID)
	db.SetStoredProcedureName(getNotificationsSP)

	bytes, err := repository.ReadBytes(db)
	if err != nil {
		return nil, fmt.Errorf("read from sp %s: %w", getNotificationsSP, err)
	}

	logrus.Infof("Size data from db: %d KB", cap(bytes)/1024)

	var resp ticketmodels.TicketNotifications
	err = jsoniter.Unmarshal(bytes, &resp)
	if err != nil {
		return nil, fmt.Errorf("unmarshal from sp %s bytes: %w", getNotificationsSP, err)
	}

	return &resp, nil
}
