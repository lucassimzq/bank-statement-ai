package transactions

import "time"

type Category struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

type Transaction struct {
	ID           string    `json:"id"`
	StatementID  string    `json:"statement_id"`
	CardID       string    `json:"card_id"`
	TxnDate      time.Time `json:"txn_date"`
	MerchantRaw  string    `json:"merchant_raw"`
	Merchant     string    `json:"merchant"`
	Amount       string    `json:"amount"`
	CategoryID   string    `json:"category_id"`
	CategoryName string    `json:"category_name"`
	CategorySlug string    `json:"category_slug"`
	CreatedAt    time.Time `json:"created_at"`
}
