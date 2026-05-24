package statements

type parsedStatement struct {
	Year             int                 `json:"year"`
	Month            int                 `json:"month"`
	StatementBalance string              `json:"statement_balance"`
	Transactions     []parsedTransaction `json:"transactions"`
}

type parsedTransaction struct {
	TxnDate      string `json:"txn_date"`
	MerchantRaw  string `json:"merchant_raw"`
	Merchant     string `json:"merchant"`
	Amount       string `json:"amount"`
	CategorySlug string `json:"category_slug"`
}
