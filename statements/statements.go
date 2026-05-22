package statements

import (
	"encore.app/statements/db"
	"encore.dev/storage/objects"
	"encore.dev/storage/sqldb"
)

var statementsDB = sqldb.NewDatabase("statements", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

var queries = db.New(statementsDB.Stdlib())

var statementFiles = objects.NewBucket("statement-files", objects.BucketConfig{
	Versioned: false,
})
