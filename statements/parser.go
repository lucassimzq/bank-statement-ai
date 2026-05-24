package statements

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"encore.app/transactions"
	"encore.dev/rlog"
	"github.com/google/generative-ai-go/genai"
)

// parseStatement calls the AI parser and stores results.
// Runs as a background goroutine — all errors update the statement row to status=error.
func (s *Service) parseStatement(statementID, cardID string, pdfBytes []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	fail := func(msg string) {
		rlog.Error("parser: "+msg, "statement_id", statementID)
		updateStatementError(ctx, statementID, msg)
	}

	// Load categories and mappings for the prompt.
	cats, err := transactions.ListCategories(ctx)
	if err != nil {
		fail("failed to load categories")
		return
	}
	slugs := make([]string, len(cats.Categories))
	for i, c := range cats.Categories {
		slugs[i] = c.Slug
	}

	mappingsResp, err := transactions.ListCategoryMappings(ctx)
	if err != nil {
		fail("failed to load category mappings")
		return
	}

	// Call AI.
	parsed, err := s.callGemini(ctx, pdfBytes, slugs, mappingsResp.Mappings)
	if err != nil {
		fail("AI parsing failed: " + err.Error())
		return
	}

	// Validate year/month from AI response.
	if parsed.Year < 2000 || parsed.Year > 2100 {
		fail("AI returned an invalid statement year")
		return
	}
	if parsed.Month < 1 || parsed.Month > 12 {
		fail("AI returned an invalid statement month")
		return
	}

	// Post-parse dedup: check if a statement for this card+period already exists.
	exists, err := statementExistsByCardPeriod(ctx, cardID, parsed.Year, parsed.Month)
	if err != nil {
		fail("failed to check for existing statement")
		return
	}
	if exists {
		fail(fmt.Sprintf("a statement for %d-%02d already exists for this card", parsed.Year, parsed.Month))
		return
	}

	// Store transactions.
	if len(parsed.Transactions) > 0 {
		inputs := make([]transactions.TransactionInput, 0, len(parsed.Transactions))
		for _, t := range parsed.Transactions {
			date, err := time.Parse("2006-01-02", t.TxnDate)
			if err != nil {
				rlog.Warn("parser: skipping transaction with bad date", "txn_date", t.TxnDate, "err", err)
				continue
			}
			slug := t.CategorySlug
			if slug == "" {
				slug = "others"
			}
			inputs = append(inputs, transactions.TransactionInput{
				TxnDate:      date,
				MerchantRaw:  t.MerchantRaw,
				Merchant:     t.Merchant,
				Amount:       t.Amount,
				CategorySlug: slug,
			})
		}
		if len(inputs) > 0 {
			if _, err := transactions.BatchCreate(ctx, &transactions.BatchCreateParams{
				StatementID:  statementID,
				CardID:       cardID,
				Transactions: inputs,
			}); err != nil {
				fail("failed to store transactions: " + err.Error())
				return
			}
		}
	}

	// Mark statement as parsed (sets year, month, balance, status=1).
	if _, err := updateStatementParsed(ctx, statementID, parsed.Year, parsed.Month, parsed.StatementBalance); err != nil {
		rlog.Error("parser: failed to mark statement as parsed", "statement_id", statementID, "err", err)
	}
}

func (s *Service) callGemini(ctx context.Context, pdfBytes []byte, categorySlugs []string, mappings []*transactions.CategoryMapping) (*parsedStatement, error) {
	model := s.gemini.GenerativeModel("gemini-3.5-flash")
	model.ResponseMIMEType = "application/json"

	resp, err := model.GenerateContent(ctx,
		genai.Blob{MIMEType: "application/pdf", Data: pdfBytes},
		genai.Text(buildParsePrompt(categorySlugs, mappings)),
	)
	if err != nil {
		return nil, err
	}
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("gemini returned empty response")
	}

	raw, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return nil, fmt.Errorf("unexpected response part type from gemini")
	}

	var result parsedStatement
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, err
	}
	return &result, nil
}
