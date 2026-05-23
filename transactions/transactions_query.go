package transactions

import (
	"context"

	txdb "encore.app/transactions/db"
	"encore.dev/beta/errs"
	"github.com/google/uuid"
)

func insertTransaction(ctx context.Context, statementID, cardID string, t TransactionInput) (*Transaction, error) {
	sID, err := uuid.Parse(statementID)
	if err != nil {
		return nil, errs.B().Code(errs.InvalidArgument).Msg("invalid statement_id").Err()
	}
	cID, err := uuid.Parse(cardID)
	if err != nil {
		return nil, errs.B().Code(errs.InvalidArgument).Msg("invalid card_id").Err()
	}
	cat, err := queries.GetCategoryBySlug(ctx, t.CategorySlug)
	if err != nil {
		return nil, errs.B().Code(errs.InvalidArgument).Msgf("unknown category_slug: %s", t.CategorySlug).Err()
	}

	row, err := queries.InsertTransaction(ctx, txdb.InsertTransactionParams{
		StatementID: sID,
		CardID:      cID,
		TxnDate:     t.TxnDate,
		MerchantRaw: t.MerchantRaw,
		Merchant:    t.Merchant,
		Amount:      t.Amount,
		CategoryID:  cat.ID,
	})
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "insert transaction")
	}

	return getTransactionByID(ctx, row.ID.String())
}

func getTransactionByID(ctx context.Context, id string) (*Transaction, error) {
	tID, err := uuid.Parse(id)
	if err != nil {
		return nil, errs.B().Code(errs.InvalidArgument).Msg("invalid transaction id").Err()
	}
	row, err := queries.GetTransactionByID(ctx, tID)
	if err != nil {
		return nil, errs.B().Code(errs.NotFound).Msg("transaction not found").Err()
	}
	return &Transaction{
		ID:           row.ID.String(),
		StatementID:  row.StatementID.String(),
		CardID:       row.CardID.String(),
		TxnDate:      row.TxnDate,
		MerchantRaw:  row.MerchantRaw,
		Merchant:     row.Merchant,
		Amount:       row.Amount,
		CategoryID:   row.CategoryID.String(),
		CategoryName: row.CategoryName,
		CategorySlug: row.CategorySlug,
		CreatedAt:    row.CreatedAt,
	}, nil
}

func queryTransactionsByStatement(ctx context.Context, statementID string) ([]*Transaction, error) {
	sID, err := uuid.Parse(statementID)
	if err != nil {
		return nil, errs.B().Code(errs.InvalidArgument).Msg("invalid statement_id").Err()
	}
	rows, err := queries.QueryTransactionsByStatement(ctx, sID)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "query transactions by statement")
	}
	result := make([]*Transaction, len(rows))
	for i, r := range rows {
		result[i] = &Transaction{
			ID:           r.ID.String(),
			StatementID:  r.StatementID.String(),
			CardID:       r.CardID.String(),
			TxnDate:      r.TxnDate,
			MerchantRaw:  r.MerchantRaw,
			Merchant:     r.Merchant,
			Amount:       r.Amount,
			CategoryID:   r.CategoryID.String(),
			CategoryName: r.CategoryName,
			CategorySlug: r.CategorySlug,
			CreatedAt:    r.CreatedAt,
		}
	}
	return result, nil
}

func queryTransactionsByCardAndMonth(ctx context.Context, cardID string, year, month int) ([]*Transaction, error) {
	cID, err := uuid.Parse(cardID)
	if err != nil {
		return nil, errs.B().Code(errs.InvalidArgument).Msg("invalid card_id").Err()
	}
	rows, err := queries.QueryTransactionsByCardAndMonth(ctx, txdb.QueryTransactionsByCardAndMonthParams{
		CardID:  cID,
		Column2: int32(year),
		Column3: int32(month),
	})
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "query transactions by card and month")
	}
	result := make([]*Transaction, len(rows))
	for i, r := range rows {
		result[i] = &Transaction{
			ID:           r.ID.String(),
			StatementID:  r.StatementID.String(),
			CardID:       r.CardID.String(),
			TxnDate:      r.TxnDate,
			MerchantRaw:  r.MerchantRaw,
			Merchant:     r.Merchant,
			Amount:       r.Amount,
			CategoryID:   r.CategoryID.String(),
			CategoryName: r.CategoryName,
			CategorySlug: r.CategorySlug,
			CreatedAt:    r.CreatedAt,
		}
	}
	return result, nil
}

func deleteTransactionsByStatement(ctx context.Context, statementID string) error {
	sID, err := uuid.Parse(statementID)
	if err != nil {
		return errs.B().Code(errs.InvalidArgument).Msg("invalid statement_id").Err()
	}
	if err := queries.DeleteTransactionsByStatement(ctx, sID); err != nil {
		return errs.WrapCode(err, errs.Internal, "delete transactions")
	}
	return nil
}
