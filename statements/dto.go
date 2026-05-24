package statements

import (
	"mime/multipart"
	"net/http"

	"encore.dev/beta/errs"
)

type ListStatementsResponse struct {
	Statements []*Statement `json:"statements"`
}

type UpdateBalanceParams struct {
	StatementBal string `json:"statement_bal"`
}

func (p *UpdateBalanceParams) Validate() error {
	if p.StatementBal == "" {
		return errs.B().Code(errs.InvalidArgument).Msg("statement_bal is required").Err()
	}
	return nil
}

type uploadForm struct {
	CardID string
	File   multipart.File
}

func parseUploadForm(r *http.Request) (*uploadForm, error) {
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		return nil, errBad("failed to parse form")
	}

	cardID := r.FormValue("card_id")
	if cardID == "" {
		return nil, errBad("card_id is required")
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, errBad("file is required")
	}

	return &uploadForm{CardID: cardID, File: file}, nil
}
