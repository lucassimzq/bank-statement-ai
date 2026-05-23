package transactions

import "encore.dev/beta/errs"

type ListCategoryMappingsResponse struct {
	Mappings []*CategoryMapping `json:"mappings"`
}

type CreateCategoryMappingParams struct {
	MerchantPattern string `json:"merchant_pattern"`
	CategorySlug    string `json:"category_slug"`
}

func (p *CreateCategoryMappingParams) Validate() error {
	eb := errs.B().Code(errs.InvalidArgument)
	if p.MerchantPattern == "" {
		return eb.Msg("merchant_pattern is required").Err()
	}
	if p.CategorySlug == "" {
		return eb.Msg("category_slug is required").Err()
	}
	return nil
}
