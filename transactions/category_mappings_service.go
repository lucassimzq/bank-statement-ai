package transactions

import "context"

// ListCategoryMappings returns all manual merchant → category mappings.
//
//encore:api public method=GET path=/category-mappings
func ListCategoryMappings(ctx context.Context) (*ListCategoryMappingsResponse, error) {
	mappings, err := queryCategoryMappings(ctx)
	if err != nil {
		return nil, err
	}
	return &ListCategoryMappingsResponse{Mappings: mappings}, nil
}

// CreateCategoryMapping adds a new merchant pattern → category mapping.
//
//encore:api public method=POST path=/category-mappings
func CreateCategoryMapping(ctx context.Context, p *CreateCategoryMappingParams) (*CategoryMapping, error) {
	return insertCategoryMapping(ctx, p.MerchantPattern, p.CategorySlug)
}

// DeleteCategoryMapping removes a mapping by ID.
//
//encore:api public method=DELETE path=/category-mappings/:id
func DeleteCategoryMapping(ctx context.Context, id string) error {
	return deleteCategoryMapping(ctx, id)
}
