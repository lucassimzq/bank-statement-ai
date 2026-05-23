package transactions

import (
	"fmt"
	"time"

	"encore.dev/beta/errs"
)

type ListCategoriesResponse struct {
	Categories []*Category `json:"categories"`
}

type TransactionInput struct {
	TxnDate      time.Time `json:"txn_date"`
	MerchantRaw  string    `json:"merchant_raw"`
	Merchant     string    `json:"merchant"`
	Amount       string    `json:"amount"`
	CategorySlug string    `json:"category_slug"`
}

func (t *TransactionInput) Validate() error {
	eb := errs.B().Code(errs.InvalidArgument)
	if t.MerchantRaw == "" {
		return eb.Msg("merchant_raw is required").Err()
	}
	if t.Merchant == "" {
		return eb.Msg("merchant is required").Err()
	}
	if t.Amount == "" {
		return eb.Msg("amount is required").Err()
	}
	if t.CategorySlug == "" {
		return eb.Msg("category_slug is required").Err()
	}
	return nil
}

type BatchCreateParams struct {
	StatementID  string             `json:"statement_id"`
	CardID       string             `json:"card_id"`
	Transactions []TransactionInput `json:"transactions"`
}

func (p *BatchCreateParams) Validate() error {
	eb := errs.B().Code(errs.InvalidArgument)
	if p.StatementID == "" {
		return eb.Msg("statement_id is required").Err()
	}
	if p.CardID == "" {
		return eb.Msg("card_id is required").Err()
	}
	if len(p.Transactions) == 0 {
		return eb.Msg("transactions must not be empty").Err()
	}
	for i, t := range p.Transactions {
		if err := t.Validate(); err != nil {
			return errs.B().Code(errs.InvalidArgument).Msg(fmt.Sprintf("transaction[%d]: %s", i, err.Error())).Err()
		}
	}
	return nil
}

type BatchCreateResponse struct {
	Transactions []*Transaction `json:"transactions"`
}

type ListTransactionsResponse struct {
	Transactions []*Transaction `json:"transactions"`
}

type ListTransactionsParams struct {
	StatementID string `query:"statement_id"`
	CardID      string `query:"card_id"`
	Year        int    `query:"year"`
	Month       int    `query:"month"`
}

func (p *ListTransactionsParams) ValidateCardMonth() error {
	eb := errs.B().Code(errs.InvalidArgument)
	if p.Year < 2000 {
		return eb.Msg("invalid year").Err()
	}
	if p.Month < 1 || p.Month > 12 {
		return eb.Msg("invalid month (1-12)").Err()
	}
	return nil
}
