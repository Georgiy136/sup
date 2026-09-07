package repository

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/ticket_approve_cron/internal/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository"
)

type TicketsRepo struct {
}

func NewTicketsRepo() *TicketsRepo {
	return &TicketsRepo{}
}

func (w *TicketsRepo) GetNotifications(logID int64) (*models.DBTicketsForNotificationResponse, int64, error) {
	db := new(postgresql.PgSpec)
	db.SetDatabaseKey(models.DatabaseKey)
	db.SetStoredProcedureName(models.GetNotificationsSP)
	db.SetParams(logID)

	bytes, err := repository.ReadBytes(db)
	if err != nil {
		return nil, logID, fmt.Errorf("read from sp %s: %v", models.GetNotificationsSP, err)
	}

	var resp models.DBTicketsForNotificationResponse
	err = jsoniter.Unmarshal(bytes, &resp)
	if err != nil {
		return nil, logID, fmt.Errorf("unmarshal from sp %s bytes: %v", models.GetNotificationsSP, err)
	}

	if len(resp.Data) == 0 {
		return nil, logID, nil
	}

	logrus.Debugf("got from sp %s: %+v", models.GetNotificationsSP, resp)

	return &resp, resp.LogID, nil
}
