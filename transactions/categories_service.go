package transactions

import "context"

//encore:api public method=GET path=/categories
func ListCategories(ctx context.Context) (*ListCategoriesResponse, error) {
	cats, err := queryCategories(ctx)
	if err != nil {
		return nil, err
	}
	return &ListCategoriesResponse{Categories: cats}, nil
}
