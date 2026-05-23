package statements

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"encore.dev/beta/errs"
	"encore.dev/storage/objects"
)

// Upload accepts a multipart form with fields: card_id, year, month, file (PDF).
//
//encore:api public raw method=POST path=/statements/upload
func (s *Service) Upload(w http.ResponseWriter, r *http.Request) {
	form, err := parseUploadForm(r)
	if err != nil {
		writeErr(w, err)
		return
	}
	defer form.File.Close()

	filePath := fmt.Sprintf("statements/%s/%d-%02d.pdf", form.CardID, form.Year, form.Month)
	writer := statementFiles.Upload(r.Context(), filePath, objects.WithUploadAttrs(objects.UploadAttrs{ContentType: "application/pdf"}))
	if _, err := io.Copy(writer, form.File); err != nil {
		jsonError(w, "failed to store file", http.StatusInternalServerError)
		return
	}
	if err := writer.Close(); err != nil {
		jsonError(w, "failed to finalise upload", http.StatusInternalServerError)
		return
	}

	stmt, err := insertStatement(r.Context(), form.CardID, form.Year, form.Month, filePath)
	if err != nil {
		writeErr(w, err)
		return
	}

	go s.parseStatement(stmt.ID, stmt.CardID, filePath)

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
