package statements

import (
	"context"

	"encore.app/statements/db"
	"encore.dev/storage/objects"
	"encore.dev/storage/sqldb"
	"github.com/google/generative-ai-go/genai"
	gooption "google.golang.org/api/option"
)

var secrets struct {
	GeminiAPIKey string
}

var statementsDB = sqldb.NewDatabase("statements", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

var queries = db.New(statementsDB.Stdlib())

var statementFiles = objects.NewBucket("statement-files", objects.BucketConfig{
	Versioned: false,
})

//encore:service
type Service struct {
	gemini *genai.Client
}

func initService() (*Service, error) {
	client, err := genai.NewClient(context.Background(), gooption.WithAPIKey(secrets.GeminiAPIKey))
	if err != nil {
		return nil, err
	}
	return &Service{gemini: client}, nil
}
