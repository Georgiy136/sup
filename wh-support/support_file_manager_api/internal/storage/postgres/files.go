package postgres

import (
	"fmt"

	jsoniter "github.com/json-iterator/go"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/support_file_manager_api/internal/local_errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/support_file_manager_api/internal/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository"
)

// GetFileID - получение нового ID для файла
func (r *Repository) GetNewFileID() (int64, error) {
	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(SupportPgDatabaseKey)
	pg.SetStoredProcedureName("files.files_getid")

	rd := repository.GetRestDataFromDb(pg)
	if err := local_errors.HandleRestDataError(rd, "failed to get new file ID"); err != nil {
		return 0, err
	}

	var result struct {
		FileID int64 `json:"file_id"`
	}
	if err := jsoniter.Unmarshal(rd.GetData(), &result); err != nil {
		return 0, fmt.Errorf("failed to unmarshal file ID response: %w", err)
	}

	return result.FileID, nil
}

func (r *Repository) GetFileInfoByID(fileID int64) (*models.File, error) {
	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(SupportPgDatabaseKey)
	pg.SetStoredProcedureName("files.files_getbyid")
	pg.SetParams(fileID)

	rd := repository.GetRestDataFromDb(pg)
	if err := local_errors.HandleRestDataError(rd, "failed to get file"); err != nil {
		return nil, err
	}

	if rd.DataEmpty() {
		return nil, local_errors.ErrFileNotFound
	}

	var file models.File
	if err := jsoniter.Unmarshal(rd.GetData(), &file); err != nil {
		return nil, fmt.Errorf("failed to unmarshal file: %w", err)
	}

	return &file, nil
}

// UpdateFileStatus - обновление статуса файла
func (r *Repository) UpdateFileStatus(fileID int64, status string) error {
	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(SupportPgDatabaseKey)
	pg.SetStoredProcedureName("files.files_upd")
	pg.SetParams(fileID, status)

	rd := repository.GetRestDataFromDb(pg)
	if err := local_errors.HandleRestDataError(rd, "failed to update file status"); err != nil {
		return err
	}
	return nil
}

// AttachTicketIDToFiles - добавляет ticket_id у файлов
func (r *Repository) AttachTicketIDToFiles(fileIDs []int64, ticketID int64) error {
	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(SupportPgDatabaseKey)
	pg.SetStoredProcedureName("files.files_updbyticketid")
	pg.SetParams(fileIDs, ticketID)

	rd := repository.GetRestDataFromDb(pg)
	if err := local_errors.HandleRestDataError(rd, "failed to attach files to ticket"); err != nil {
		return err
	}

	return nil
}

// GetFilesByStatus - получение файлов по статусу
func (r *Repository) GetFilesByStatus(status string) ([]models.File, error) {
	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(SupportPgDatabaseKey)
	pg.SetStoredProcedureName("files.files_getbystatus")
	pg.SetParams(status)

	rd := repository.GetRestDataFromDb(pg)
	if err := local_errors.HandleRestDataError(rd, "failed to get files by status"); err != nil {
		return nil, err
	}

	if rd.DataEmpty() {
		return []models.File{}, nil
	}

	var files []models.File
	if err := jsoniter.Unmarshal(rd.GetData(), &files); err != nil {
		return nil, fmt.Errorf("failed to unmarshal files: %w", err)
	}

	return files, nil
}

// AddFiles - добавление файлов в постоянное хранилище
func (r *Repository) AddFiles(params []models.FileAddParams) error {
	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(SupportPgDatabaseKey)
	pg.SetStoredProcedureName("files.files_add")
	pg.SetParams(params)

	rd := repository.GetRestDataFromDb(pg)
	if err := local_errors.HandleRestDataError(rd, "failed to add files"); err != nil {
		return err
	}

	return nil
}
