package postgres

import (
	"fmt"

	jsoniter "github.com/json-iterator/go"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/support_file_manager_api/internal/local_errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/support_file_manager_api/internal/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository"
)

// GetUploadID - получение нового ID для аплоада
func (r *Repository) GetNewUploadID() (int64, error) {
	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(SupportPgDatabaseKey)
	pg.SetStoredProcedureName("files.uploads_getid")

	rd := repository.GetRestDataFromDb(pg)
	if err := local_errors.HandleRestDataError(rd, "failed to get new upload ID"); err != nil {
		return 0, err
	}

	var result struct {
		UploadID int64 `json:"upload_id"`
	}
	if err := jsoniter.Unmarshal(rd.GetData(), &result); err != nil {
		return 0, fmt.Errorf("failed to unmarshal upload ID response: %w", err)
	}

	return result.UploadID, nil
}

// AddUpload - создание записи аплоада
func (r *Repository) AddUpload(params models.UploadAddParams) error {
	paramsArray := []models.UploadAddParams{params}

	paramsJSON, err := jsoniter.Marshal(paramsArray)
	if err != nil {
		return fmt.Errorf("failed to marshal upload params: %w", err)
	}

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(SupportPgDatabaseKey)
	pg.SetStoredProcedureName("files.uploads_add")
	pg.SetParams(string(paramsJSON))

	rd := repository.GetRestDataFromDb(pg)
	if err = local_errors.HandleRestDataError(rd, "failed to add upload"); err != nil {
		return err
	}

	return nil
}

// UpdateUploadStatus - обновление статуса аплоада
func (r *Repository) UpdateUploadStatus(uploadID int64, employeeID int64, status string) error {
	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(SupportPgDatabaseKey)
	pg.SetStoredProcedureName("files.uploads_upd")
	pg.SetParams(uploadID, employeeID, status)

	rd := repository.GetRestDataFromDb(pg)
	if err := local_errors.HandleRestDataError(rd, "failed to update upload"); err != nil {
		return err
	}

	return nil
}

// GetUploadByID - получение аплоада по ID
func (r *Repository) GetUploadByID(uploadID int64) (*models.Upload, error) {
	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(SupportPgDatabaseKey)
	pg.SetStoredProcedureName("files.uploads_getbyid")
	pg.SetParams(uploadID)

	rd := repository.GetRestDataFromDb(pg)
	if err := local_errors.HandleRestDataError(rd, "failed to get upload"); err != nil {
		return nil, err
	}

	if rd.DataEmpty() {
		return nil, local_errors.ErrUploadNotFound
	}

	var upload models.Upload
	if err := jsoniter.Unmarshal(rd.GetData(), &upload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal upload: %w", err)
	}

	return &upload, nil
}

// GetUploadsByStatus - получение аплоадов по статусу
func (r *Repository) GetUploadsByStatus(status string) ([]models.Upload, error) {
	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(SupportPgDatabaseKey)
	pg.SetStoredProcedureName("files.uploads_getbystatus")
	pg.SetParams(status)

	rd := repository.GetRestDataFromDb(pg)
	if err := local_errors.HandleRestDataError(rd, "failed to get uploads by status"); err != nil {
		return nil, err
	}

	if rd.DataEmpty() {
		return []models.Upload{}, nil
	}

	var uploads []models.Upload
	if err := jsoniter.Unmarshal(rd.GetData(), &uploads); err != nil {
		return nil, fmt.Errorf("failed to unmarshal uploads: %w", err)
	}

	return uploads, nil
}

// GetExpiredUploads - получение аплоадов с истекшим сроком жизни
func (r *Repository) GetExpiredUploads() ([]models.Upload, error) {
	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(SupportPgDatabaseKey)
	pg.SetStoredProcedureName("files.uploads_getbyexpires")

	rd := repository.GetRestDataFromDb(pg)
	if err := local_errors.HandleRestDataError(rd, "failed to get expired uploads"); err != nil {
		return nil, err
	}

	if rd.DataEmpty() {
		return []models.Upload{}, nil
	}

	var uploads []models.Upload
	if err := jsoniter.Unmarshal(rd.GetData(), &uploads); err != nil {
		return nil, fmt.Errorf("failed to unmarshal expired uploads: %w", err)
	}

	return uploads, nil
}
