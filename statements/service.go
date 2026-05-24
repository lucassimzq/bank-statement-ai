package statements

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	"encore.app/cards"
	"encore.dev/beta/errs"
	"encore.dev/storage/objects"
)

// Upload accepts a multipart form with fields: card_id and file (PDF).
// Year and month are extracted from the PDF by the AI parser.
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

	// Validate card exists.
	if _, err := cards.GetCard(r.Context(), form.CardID); err != nil {
		writeErr(w, errBad("card not found"))
		return
	}

	// Pre-validate: reject duplicate files before spending any AI tokens.
	hash := sha256Hash(pdfBytes)
	duplicate, err := statementExistsByHash(r.Context(), hash)
	if err != nil {
		jsonError(w, "failed to check for duplicate: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if duplicate {
		jsonError(w, "this statement has already been uploaded", http.StatusConflict)
		return
	}

	// Upload PDF to bucket.
	filePath := fmt.Sprintf("statements/%s/%s.pdf", form.CardID, hash)
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
	stmt, err := insertStatement(r.Context(), form.CardID, hash, filePath)
	if err != nil {
		writeErr(w, err)
		return
	}

	// Parse in background — passes bytes directly, no re-download needed.
	go s.parseStatement(stmt.ID, stmt.CardID, pdfBytes)

	jsonResponse(w, stmt, http.StatusCreated)
}

//encore:api public method=GET path=/statements/:id
func GetStatement(ctx context.Context, id string) (*Statement, error) {
	return getStatementByID(ctx, id)
}

type ListStatementsParams struct {
	CardID string `query:"card_id"`
}

//encore:api public method=GET path=/statements
func ListStatements(ctx context.Context, p *ListStatementsParams) (*ListStatementsResponse, error) {
	if p.CardID == "" {
		return nil, errs.B().Code(errs.InvalidArgument).Msg("card_id is required").Err()
	}
	stmts, err := queryStatementsByCard(ctx, p.CardID)
	if err != nil {
		return nil, err
	}
	return &ListStatementsResponse{Statements: stmts}, nil
}

//encore:api public method=PATCH path=/statements/:id/balance
func UpdateBalance(ctx context.Context, id string, p *UpdateBalanceParams) (*Statement, error) {
	return updateStatementBalance(ctx, id, p.StatementBal)
}

func sha256Hash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
