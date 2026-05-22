package cards

import (
	"encore.app/cards/db"
	"encore.dev/storage/sqldb"
)

var cardsDB = sqldb.NewDatabase("cards", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

var queries = db.New(cardsDB.Stdlib())
