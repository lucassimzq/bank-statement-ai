package transactions

import "time"

type CategoryMapping struct {
	ID              string    `json:"id"`
	MerchantPattern string    `json:"merchant_pattern"`
	CategorySlug    string    `json:"category_slug"`
	CategoryName    string    `json:"category_name"`
	CreatedAt       time.Time `json:"created_at"`
}
