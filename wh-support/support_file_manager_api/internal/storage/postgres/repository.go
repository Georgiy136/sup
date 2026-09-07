package postgres

const (
	SupportPgDatabaseKey = "support_pgx"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}
