package statements

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	"encore.dev/beta/errs"
	"encore.dev/storage/objects"
)

// Upload accepts a multipart form with fields: file (PDF).
// Cards and period are extracted from the PDF by the AI parser.
//
//encore:api public raw method=POST path=/statements/upload
func (s *Service) Upload(w http.ResponseWriter, r *http.Request) {
	form, err := parseUploadForm(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	defer form.File.Close()

	// Read file into memory for hashing and uploading.
	pdfBytes, err := io.ReadAll(form.File)
	if err != nil {
		jsonError(w, "failed to read file", http.StatusInternalServerError)
		return
	}

	// Pre-validate: reject only if statement is fully parsed (all cards done).
	// If some cards were skipped last time, allow re-upload to process them.
	hash := sha256Hash(pdfBytes)
	fullyParsed, err := statementFullyParsedByHash(r.Context(), hash)
	if err != nil {
		jsonError(w, "failed to check for duplicate: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if fullyParsed {
		jsonError(w, "this statement has already been fully parsed", http.StatusConflict)
		return
	}

	// Upload PDF to bucket.
	filePath := fmt.Sprintf("statements/%s.pdf", hash)
	writer := statementFiles.Upload(r.Context(), filePath, objects.WithUploadAttrs(objects.UploadAttrs{
		ContentType: "application/pdf",
	}))
	if _, err := writer.Write(pdfBytes); err != nil {
		jsonError(w, "failed to store file", http.StatusInternalServerError)
		return
	}
	if err := writer.Close(); err != nil {
		jsonError(w, "failed to finalise upload", http.StatusInternalServerError)
		return
	}

	// Insert statement row with status=parsing.
	stmt, err := insertStatement(r.Context(), hash, filePath)
	if err != nil {
		writeErr(w, err)
		return
	}

	// Parse in background — passes bytes directly, no re-download needed.
	go s.parseStatement(stmt.ID, pdfBytes)

	jsonResponse(w, stmt, http.StatusCreated)
}

//encore:api public method=GET path=/statements/:id
func GetStatement(ctx context.Context, id string) (*Statement, error) {
	return getStatementByID(ctx, id)
}

type ListStatementsParams struct {
	CardID string `query:"card_id"`
}

// ListStatements returns all statements. If card_id is provided, filters to
// only those statements that include that card.
//
//encore:api public method=GET path=/statements
func ListStatements(ctx context.Context, p *ListStatementsParams) (*ListStatementsResponse, error) {
	var baseStmts []*Statement
	var err error

	if p.CardID != "" {
		baseStmts, err = queryStatementsByCard(ctx, p.CardID)
	} else {
		baseStmts, err = listAllStatements(ctx)
	}
	if err != nil {
		return nil, err
	}

	result := make([]*StatementWithCards, len(baseStmts))
	for i, s := range baseStmts {
		cards, _ := getCardStatementsByStatementID(ctx, s.ID)
		if cards == nil {
			cards = []*CardStatementInfo{}
		}

		// If the statement itself has no balance, derive it by summing
		// per-card balances stored in card_statement (covers pre-fix data).
		bal := s.StatementBal
		if bal == nil || *bal == "" {
			bal = sumCardBalances(cards)
		}

		result[i] = &StatementWithCards{
			ID:           s.ID,
			Status:       s.Status,
			Message:      s.Message,
			Year:         s.Year,
			Month:        s.Month,
			StatementBal: bal,
			FilePath:     s.FilePath,
			ParsedAt:     s.ParsedAt,
			CreatedAt:    s.CreatedAt,
			Cards:        cards,
		}
	}
	return &ListStatementsResponse{Statements: result}, nil
}

//encore:api public method=PATCH path=/statements/:id/balance
func UpdateBalance(ctx context.Context, id string, p *UpdateBalanceParams) (*Statement, error) {
	return updateStatementBalance(ctx, id, p.StatementBal)
}

// RetryStatement re-parses a statement that is in error state.
//
//encore:api public method=POST path=/statements/retry/:id
func (s *Service) RetryStatement(ctx context.Context, id string) (*Statement, error) {
	stmt, err := getStatementByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if stmt.Status != StatusError {
		return nil, errs.B().Code(errs.FailedPrecondition).Msg("only error statements can be retried").Err()
	}
	if stmt.FilePath == nil {
		return nil, errs.B().Code(errs.Internal).Msg("statement has no stored file").Err()
	}

	// Re-download file from object storage.
	reader := statementFiles.Download(ctx, *stmt.FilePath)
	defer reader.Close()
	pdfBytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "read statement file")
	}
	if err := reader.Err(); err != nil {
		return nil, errs.WrapCode(err, errs.Internal, "download statement file")
	}

	// Reset status to parsing and clear error message.
	stmt, err = resetStatementForRetry(ctx, id)
	if err != nil {
		return nil, err
	}

	// Re-parse in background.
	go s.parseStatement(stmt.ID, pdfBytes)

	return stmt, nil
}

// DeleteStatement permanently removes a statement that failed to parse.
// Only error-status statements can be deleted — parsed statements are permanent records.
//
//encore:api public method=DELETE path=/statements/:id
func DeleteStatement(ctx context.Context, id string) error {
	stmt, err := getStatementByID(ctx, id)
	if err != nil {
		return err
	}
	if stmt.Status != StatusError {
		return errs.B().Code(errs.FailedPrecondition).Msg("only error statements can be deleted").Err()
	}
	return deleteStatementByID(ctx, id)
}

func sha256Hash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
