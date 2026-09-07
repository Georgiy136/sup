package repository

import (
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_database.git/postgresql/repository"
)

func ExecuteSupportDb(sp string, isProcedure bool, params ...interface{}) ([]byte, error) {
	const dbKey = "support_pgx"

	pg := new(postgresql.PgSpec)
	pg.SetDatabaseKey(dbKey)
	pg.SetStoredProcedureName(sp)
	pg.SetParams(params...)

	if isProcedure {
		pg.SetUseProcedure()
		return nil, repository.Execute(pg)
	} else {
		return repository.ReadBytes(pg)
	}
}
