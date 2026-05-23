package transactions

import (
	"encore.app/transactions/db"
	"encore.dev/storage/sqldb"
)

var transactionsDB = sqldb.NewDatabase("transactions", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

var queries = db.New(transactionsDB.Stdlib())
