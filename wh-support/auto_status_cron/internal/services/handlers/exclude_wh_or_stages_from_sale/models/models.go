package excludewhorstagesfromsalemodels

import (
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
)

type ExtTicketInfoForExcludeWh struct {
	WhID     WhID     `mapstructure:"wh_id"`
	OfficeID OfficeID `mapstructure:"office_id"`
}

type ExtTicketInfoForExcludeStages struct {
	WhID     WhID     `mapstructure:"wh_id"`
	OfficeID OfficeID `mapstructure:"office_id"`
	Stages   []Stage  `mapstructure:"stages"`
}

type Stage = models.ID
type WhID = models.NamedID[int64]
type OfficeID = models.NamedID[int64]
