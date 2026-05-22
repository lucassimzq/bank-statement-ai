package statements

import (
	"mime/multipart"
	"net/http"
	"strconv"

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
	Year   int
	Month  int
	File   multipart.File
}

func parseUploadForm(r *http.Request) (*uploadForm, error) {
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		return nil, errBad("failed to parse form")
	}

	cardID := r.FormValue("card_id")
	yearStr := r.FormValue("year")
	monthStr := r.FormValue("month")

	if cardID == "" || yearStr == "" || monthStr == "" {
		return nil, errBad("card_id, year and month are required")
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 {
		return nil, errBad("invalid year")
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		return nil, errBad("invalid month (1-12)")
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, errBad("file is required")
	}

	return &uploadForm{CardID: cardID, Year: year, Month: month, File: file}, nil
}
