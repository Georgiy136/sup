package models

type DataWrapper[T any] struct {
	Data T `json:"data"`
}

const (
	SupportPgDatabaseKey = "support_pgx"
)
