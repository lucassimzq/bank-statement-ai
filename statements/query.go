package statements

import (
	"context"
	"database/sql"

	stmtdb "encore.app/statements/db"
	"encore.dev/beta/errs"
	"github.com/google/uuid"
)

// ── Statement helpers ────────────────────────────────────────────────────────

func insertStatement(ctx context.Context, fileHash, filePath string) (*Statement, error) {
	row, err := queries.InsertStatement(ctx, stmtdb.InsertStatementParams{
		FileHash: sql.NullString{String: fileHash, Valid: true},
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
	rows, err := queries.QueryStatementsByCard(ctx, uuid.NullUUID{UUID: cID, Valid: true})
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "query statements")
	}
	stmts := make([]*Statement, len(rows))
	for i, r := range rows {
		stmts[i] = toStatement(r)
	}
	return stmts, nil
}

func updateStatementParsed(ctx context.Context, id string, year, month int, balance string) (*Statement, error) {
	sID, err := uuid.Parse(id)
	if err != nil {
		return nil, errs.B().Code(errs.InvalidArgument).Msg("invalid statement id").Err()
	}
	var bal sql.NullString
	if balance != "" {
		bal = sql.NullString{String: balance, Valid: true}
	}
	row, err := queries.UpdateStatementParsed(ctx, stmtdb.UpdateStatementParsedParams{
		ID:           sID,
		Year:         sql.NullInt32{Int32: int32(year), Valid: true},
		Month:        sql.NullInt32{Int32: int32(month), Valid: true},
		StatementBal: bal,
	})
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "update statement parsed")
	}
	return toStatement(row), nil
}

func updateStatementError(ctx context.Context, id, message string) {
	sID, err := uuid.Parse(id)
	if err != nil {
		return
	}
	_, _ = queries.UpdateStatementError(ctx, stmtdb.UpdateStatementErrorParams{
		ID:      sID,
		Message: sql.NullString{String: message, Valid: true},
	})
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

// statementFullyParsedByHash returns true only when the hash exists with
// status=1 AND every card_statement row is also status=1 (nothing was skipped).
// If any card was skipped (status=2), we allow re-upload.
func statementFullyParsedByHash(ctx context.Context, hash string) (bool, error) {
	exists, err := queries.StatementFullyParsedByHash(ctx, sql.NullString{String: hash, Valid: true})
	if err != nil {
		return false, errs.WrapCode(err, errs.Internal, "check hash: "+err.Error())
	}
	return exists, nil
}

// ── CardStatement helpers ────────────────────────────────────────────────────

// insertCardStatementParsed records a successfully processed card.
func insertCardStatementParsed(ctx context.Context, statementID, cardID, cardLast4, balance string) error {
	sID, err := uuid.Parse(statementID)
	if err != nil {
		return errs.B().Code(errs.InvalidArgument).Msg("invalid statement id").Err()
	}
	cID, err := uuid.Parse(cardID)
	if err != nil {
		return errs.B().Code(errs.InvalidArgument).Msg("invalid card id").Err()
	}
	var bal sql.NullString
	if balance != "" {
		bal = sql.NullString{String: balance, Valid: true}
	}
	_, err = queries.InsertCardStatement(ctx, stmtdb.InsertCardStatementParams{
		StatementID:  sID,
		CardLast4:    cardLast4,
		CardID:       uuid.NullUUID{UUID: cID, Valid: true},
		Status:       1,
		StatementBal: bal,
	})
	return err
}

// insertCardStatementSkipped records a card that was not found in the system.
func insertCardStatementSkipped(ctx context.Context, statementID, cardLast4 string) error {
	sID, err := uuid.Parse(statementID)
	if err != nil {
		return errs.B().Code(errs.InvalidArgument).Msg("invalid statement id").Err()
	}
	_, err = queries.InsertCardStatement(ctx, stmtdb.InsertCardStatementParams{
		StatementID: sID,
		CardLast4:   cardLast4,
		CardID:      uuid.NullUUID{Valid: false},
		Status:      2,
	})
	return err
}

// cardStatementExistsForPeriod checks whether a card already has a parsed
// card_statement entry for the given year+month (prevents double-importing).
func cardStatementExistsForPeriod(ctx context.Context, cardID string, year, month int) (bool, error) {
	cID, err := uuid.Parse(cardID)
	if err != nil {
		return false, errs.B().Code(errs.InvalidArgument).Msg("invalid card id").Err()
	}
	exists, err := queries.CardStatementExistsForPeriod(ctx, stmtdb.CardStatementExistsForPeriodParams{
		CardID: uuid.NullUUID{UUID: cID, Valid: true},
		Year:   sql.NullInt32{Int32: int32(year), Valid: true},
		Month:  sql.NullInt32{Int32: int32(month), Valid: true},
	})
	if err != nil {
		return false, errs.WrapCode(err, errs.Internal, "check card period")
	}
	return exists, nil
}

// ── Mapper ───────────────────────────────────────────────────────────────────

func toStatement(r stmtdb.Statement) *Statement {
	s := &Statement{
		ID:        r.ID.String(),
		Status:    int(r.Status),
		CreatedAt: r.CreatedAt,
	}
	if r.Year.Valid {
		y := int(r.Year.Int32)
		s.Year = &y
	}
	if r.Month.Valid {
		m := int(r.Month.Int32)
		s.Month = &m
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
	if r.Message.Valid {
		s.Message = &r.Message.String
	}
	return s
}
