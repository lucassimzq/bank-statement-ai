package transactions

import (
	"context"

	"encore.dev/beta/errs"
)

// BatchCreate inserts multiple transactions for a statement in one call.
// Only callable internally — invoked by the parsing worker, not external clients.
//
//encore:api private method=POST path=/transactions
func BatchCreate(ctx context.Context, p *BatchCreateParams) (*BatchCreateResponse, error) {
	result := make([]*Transaction, 0, len(p.Transactions))
	for _, t := range p.Transactions {
		tx, err := insertTransaction(ctx, p.StatementID, p.CardID, t)
		if err != nil {
			return nil, err
		}
		result = append(result, tx)
	}
	return &BatchCreateResponse{Transactions: result}, nil
}

//encore:api public method=GET path=/transactions/:id
func GetTransaction(ctx context.Context, id string) (*Transaction, error) {
	return getTransactionByID(ctx, id)
}

// GetMonthlySpending returns total spending grouped by card and month, across all time.
// Used by the frontend trend chart.
//
//encore:api public method=GET path=/monthly-spending
func GetMonthlySpending(ctx context.Context) (*GetMonthlySpendingResponse, error) {
	data, err := queryMonthlySpending(ctx)
	if err != nil {
		return nil, err
	}
	return &GetMonthlySpendingResponse{Data: data}, nil
}

// ListTransactions returns transactions filtered by either statement_id
// or card_id+year+month (at least one filter is required).
//
//encore:api public method=GET path=/transactions
func ListTransactions(ctx context.Context, p *ListTransactionsParams) (*ListTransactionsResponse, error) {
	switch {
	case p.StatementID != "":
		txns, err := queryTransactionsByStatement(ctx, p.StatementID)
		if err != nil {
			return nil, err
		}
		return &ListTransactionsResponse{Transactions: txns}, nil

	case p.CardID != "":
		if err := p.ValidateCardMonth(); err != nil {
			return nil, err
		}
		txns, err := queryTransactionsByCardAndMonth(ctx, p.CardID, p.Year, p.Month)
		if err != nil {
			return nil, err
		}
		return &ListTransactionsResponse{Transactions: txns}, nil

	default:
		return nil, errs.B().Code(errs.InvalidArgument).Msg("statement_id or card_id is required").Err()
	}
}
