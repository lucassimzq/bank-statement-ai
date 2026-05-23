package transactions

import (
	"context"

	txdb "encore.app/transactions/db"
	"encore.dev/beta/errs"
	"github.com/google/uuid"
)

func queryCategoryMappings(ctx context.Context) ([]*CategoryMapping, error) {
	rows, err := queries.QueryCategoryMappings(ctx)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "query category mappings")
	}
	result := make([]*CategoryMapping, len(rows))
	for i, r := range rows {
		result[i] = &CategoryMapping{
			ID:              r.ID.String(),
			MerchantPattern: r.MerchantPattern,
			CategorySlug:    r.CategorySlug,
			CategoryName:    r.CategoryName,
			CreatedAt:       r.CreatedAt,
		}
	}
	return result, nil
}

func insertCategoryMapping(ctx context.Context, pattern, categorySlug string) (*CategoryMapping, error) {
	row, err := queries.InsertCategoryMapping(ctx, txdb.InsertCategoryMappingParams{
		MerchantPattern: pattern,
		Slug:            categorySlug,
	})
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "insert category mapping")
	}
	mappings, err := queryCategoryMappings(ctx)
	if err != nil {
		return nil, err
	}
	for _, m := range mappings {
		if m.ID == row.ID.String() {
			return m, nil
		}
	}
	return nil, errs.B().Code(errs.Internal).Msg("mapping not found after insert").Err()
}

func deleteCategoryMapping(ctx context.Context, id string) error {
	mID, err := uuid.Parse(id)
	if err != nil {
		return errs.B().Code(errs.InvalidArgument).Msg("invalid mapping id").Err()
	}
	if err := queries.DeleteCategoryMapping(ctx, mID); err != nil {
		return errs.WrapCode(err, errs.Internal, "delete category mapping")
	}
	return nil
}
