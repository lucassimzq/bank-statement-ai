package transactions

import (
	"context"

	"encore.dev/beta/errs"
)

func queryCategories(ctx context.Context) ([]*Category, error) {
	rows, err := queries.QueryCategories(ctx)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "query categories")
	}
	cats := make([]*Category, len(rows))
	for i, r := range rows {
		cats[i] = &Category{ID: r.ID.String(), Name: r.Name, Slug: r.Slug, CreatedAt: r.CreatedAt}
	}
	return cats, nil
}
