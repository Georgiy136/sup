package service

import (
	jsoniter "github.com/json-iterator/go"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/publisher_from_wh_support/models"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/publisher_from_wh_support/sync_models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository"
	"gitlab.wildberries.ru/wbwh/wh-core/jetstream/sync_core.git/publisher/publisher_by_timer"
	"gitlab.wildberries.ru/wbwh/wh-core/jetstream/sync_core.git/publisher/publisher_core"
)

type TicketsStatusChanges struct{}

func (TicketsStatusChanges) GetDataForSync(syncRequestToDataBase publisher_by_timer.SyncRequestByTime) (publisher_by_timer.SyncRequestByTime, error) {
	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey("support_pgx")
	pg.SetStoredProcedureName("sync.ticketsstatuschange_exporttojson")
	pg.SetParams(syncRequestToDataBase.LogID)

	result, err := repository.ReadBytes(pg)
	if err != nil {
		return syncRequestToDataBase, err
	}

	if len(result) == 0 {
		return syncRequestToDataBase, nil
	}

	var ParsedTicketsFromDB models.TicketsFromDB

	if err = jsoniter.Unmarshal(result, &ParsedTicketsFromDB); err != nil {
		return syncRequestToDataBase, err
	}

	if len(ParsedTicketsFromDB.Data) == 0 {
		return syncRequestToDataBase, nil
	}
	if syncRequestToDataBase.LogID < ParsedTicketsFromDB.LogID {
		syncRequestToDataBase.LogID = ParsedTicketsFromDB.LogID
	}

	var tickets sync_models.Tickets
	tickets.Data = make([]sync_models.Tickets_TicketData, len(ParsedTicketsFromDB.Data))

	for i := range ParsedTicketsFromDB.Data {
		ext, err := jsoniter.Marshal(ParsedTicketsFromDB.Data[i].Ext)
		if err != nil {
			return syncRequestToDataBase, err
		}

		strExt := string(ext)

		tickets.Data[i] = sync_models.Tickets_TicketData{
			Ext:               &strExt,
			ChDt:              ParsedTicketsFromDB.Data[i].ChDt,
			ChEmployeeID:      ParsedTicketsFromDB.Data[i].ChEmployeeID,
			CreateDt:          ParsedTicketsFromDB.Data[i].CreateDt,
			TicketID:          ParsedTicketsFromDB.Data[i].TicketID,
			CategoryID:        ParsedTicketsFromDB.Data[i].CategoryID,
			StatusId:          ParsedTicketsFromDB.Data[i].StatusId,
			Comments:          ParsedTicketsFromDB.Data[i].Comments,
			RejectedComments:  ParsedTicketsFromDB.Data[i].RejectedComments,
			CreateEmployeeID:  ParsedTicketsFromDB.Data[i].CreateEmployeeID,
			ApproveEmployeeID: ParsedTicketsFromDB.Data[i].ApproveEmployeeID,
			PerformEmployeeID: ParsedTicketsFromDB.Data[i].PerformEmployeeID,
		}
	}

	resultData, err := tickets.Marshal()
	if err != nil {
		return syncRequestToDataBase, err
	}

	syncRequestToDataBase.ReaderRowCount = len(ParsedTicketsFromDB.Data)
	syncRequestToDataBase.ResponseData = []publisher_core.Msg{{BatchedData: [][]byte{resultData}}}
	if syncRequestToDataBase.ReaderRowCount >= syncRequestToDataBase.PackSize {
		syncRequestToDataBase.TimeSleepSeconds = ParsedTicketsFromDB.SleepSeconds
	}
	return syncRequestToDataBase, nil
}
