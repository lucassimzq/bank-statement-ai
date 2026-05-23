package cards

import (
	"context"

	"encore.app/cards/db"
	"encore.app/cards/seeds"
	"encore.dev/storage/objects"
	"encore.dev/storage/sqldb"
)

var cardsDB = sqldb.NewDatabase("cards", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

var queries = db.New(cardsDB.Stdlib())

var bankLogosBucket = objects.NewBucket("bank-logos", objects.BucketConfig{
	Versioned: false,
	Public:    true,
})

//encore:service
type Service struct{}

func initService() (*Service, error) {
	ref := objects.BucketRef[seeds.BucketRef](bankLogosBucket)
	seeds.SeedBanks(context.Background(), queries, ref)
	return &Service{}, nil
}
