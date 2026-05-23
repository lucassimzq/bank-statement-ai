package seeds

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"runtime"

	cardsdb "encore.app/cards/db"
	"encore.dev/rlog"
	"encore.dev/storage/objects"
)

// BucketRef is the set of bucket permissions required by SeedBanks.
type BucketRef interface {
	objects.Uploader
	objects.Attrser
}

type bankDef struct {
	Name     string
	Slug     string
	LogoFile string // filename inside seeds/logos/, e.g. "maybank.png"
}

var banks = []bankDef{
	{Name: "Maybank", Slug: "maybank", LogoFile: "maybank.png"},
	{Name: "Alliance Bank", Slug: "alliance", LogoFile: "alliance.png"},
	{Name: "HSBC", Slug: "hsbc", LogoFile: "hsbc.png"},
	{Name: "Hong Leong Bank", Slug: "hong-leong", LogoFile: "hong-leong.png"},
}

// logosDir returns the absolute path to the seeds/logos directory.
func logosDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "logos")
}

func SeedBanks(ctx context.Context, queries *cardsdb.Queries, bucket BucketRef) {
	dir := logosDir()

	for _, b := range banks {
		logoURL := uploadLogo(ctx, bucket, dir, b.Slug, b.LogoFile)

		var nullLogoURL sql.NullString
		if logoURL != "" {
			nullLogoURL = sql.NullString{String: logoURL, Valid: true}
		}

		_, err := queries.UpsertBank(ctx, cardsdb.UpsertBankParams{
			Name:    b.Name,
			Slug:    b.Slug,
			LogoUrl: nullLogoURL,
		})
		if err != nil {
			rlog.Error("seed: failed to upsert bank", "slug", b.Slug, "err", err)
			continue
		}
		rlog.Info("seed: upserted bank", "slug", b.Slug)
	}
}

// uploadLogo uploads the logo file to the bucket if it doesn't already exist.
// Returns the object key (path) on success, empty string if skipped or failed.
func uploadLogo(ctx context.Context, bucket BucketRef, dir, slug, filename string) string {
	objectKey := "logos/" + slug + "/" + filename

	// Check if already uploaded
	if _, err := bucket.Attrs(ctx, objectKey); err == nil {
		rlog.Debug("seed: logo already in bucket, skipping upload", "key", objectKey)
		return objectKey
	}

	localPath := filepath.Join(dir, filename)
	data, err := os.ReadFile(localPath)
	if err != nil {
		if os.IsNotExist(err) {
			rlog.Warn("seed: logo file not found locally, skipping", "path", localPath)
		} else {
			rlog.Error("seed: failed to read logo file", "path", localPath, "err", err)
		}
		return ""
	}

	ext := filepath.Ext(filename)
	contentType := "image/png"
	if ext == ".jpg" || ext == ".jpeg" {
		contentType = "image/jpeg"
	} else if ext == ".svg" {
		contentType = "image/svg+xml"
	} else if ext == ".webp" {
		contentType = "image/webp"
	}

	writer := bucket.Upload(ctx, objectKey, objects.WithUploadAttrs(objects.UploadAttrs{
		ContentType: contentType,
	}))
	if _, err := writer.Write(data); err != nil {
		rlog.Error("seed: failed to write logo to bucket", "key", objectKey, "err", err)
		return ""
	}
	if err := writer.Close(); err != nil {
		rlog.Error("seed: failed to finalise logo upload", "key", objectKey, "err", err)
		return ""
	}

	rlog.Info("seed: uploaded logo", "key", objectKey)
	return objectKey
}
