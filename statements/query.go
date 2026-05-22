package statements

import (
	"context"
	"database/sql"

	stmtdb "encore.app/statements/db"
	"encore.dev/beta/errs"
	"github.com/google/uuid"
)

func insertStatement(ctx context.Context, cardID string, year, month int, filePath string) (*Statement, error) {
	cID, err := uuid.Parse(cardID)
	if err != nil {
		return nil, errs.B().Code(errs.InvalidArgument).Msg("invalid card_id").Err()
	}
	row, err := queries.InsertStatement(ctx, stmtdb.InsertStatementParams{
		CardID:   cID,
		Year:     int32(year),
		Month:    int32(month),
		FilePath: sql.NullString{String: filePath, Valid: true},
	})
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "insert statement")
	}
	return toStatement(row), nil
}

func getStatementByID(ctx context.Context, id string) (*Statement, error) {
	sID, err := uuid.Parse(id)
	if err != nil {
		return nil, errs.B().Code(errs.InvalidArgument).Msg("invalid statement id").Err()
	}
	row, err := queries.GetStatementByID(ctx, sID)
	if err != nil {
		return nil, errs.B().Code(errs.NotFound).Msg("statement not found").Err()
	}
	return toStatement(row), nil
}

func queryStatementsByCard(ctx context.Context, cardID string) ([]*Statement, error) {
	cID, err := uuid.Parse(cardID)
	if err != nil {
		return nil, errs.B().Code(errs.InvalidArgument).Msg("invalid card_id").Err()
	}
	rows, err := queries.QueryStatementsByCard(ctx, cID)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "query statements")
	}
	stmts := make([]*Statement, len(rows))
	for i, r := range rows {
		stmts[i] = toStatement(r)
	}
	return stmts, nil
}

func updateStatementBalance(ctx context.Context, id, balance string) (*Statement, error) {
	sID, err := uuid.Parse(id)
	if err != nil {
		return nil, errs.B().Code(errs.InvalidArgument).Msg("invalid statement id").Err()
	}
	row, err := queries.UpdateStatementBalance(ctx, stmtdb.UpdateStatementBalanceParams{
		ID:           sID,
		StatementBal: sql.NullString{String: balance, Valid: true},
	})
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "update statement balance")
	}
	return toStatement(row), nil
}

func toStatement(r stmtdb.Statement) *Statement {
	s := &Statement{
		ID:        r.ID.String(),
		CardID:    r.CardID.String(),
		Year:      int(r.Year),
		Month:     int(r.Month),
		CreatedAt: r.CreatedAt,
	}
	if r.StatementBal.Valid {
		s.StatementBal = &r.StatementBal.String
	}
	if r.FilePath.Valid {
		s.FilePath = &r.FilePath.String
	}
	if r.ParsedAt.Valid {
		s.ParsedAt = &r.ParsedAt.Time
	}
	return s
}
