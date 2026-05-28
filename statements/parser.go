package statements

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"encore.app/cards"
	"encore.app/transactions"
	"encore.dev/rlog"
	"github.com/google/generative-ai-go/genai"
)

// parseStatement calls the AI parser and stores results for each card found.
// Runs as a background goroutine — errors update the statement row to status=error.
func (s *Service) parseStatement(statementID string, pdfBytes []byte) {
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

	// Load known bank slugs so the AI prompt can reference them and we can validate.
	banksResp, err := cards.ListBanks(ctx)
	if err != nil {
		fail("failed to load banks")
		return
	}
	bankSlugSet := make(map[string]bool, len(banksResp.Banks))
	bankSlugs := make([]string, len(banksResp.Banks))
	for i, b := range banksResp.Banks {
		bankSlugSet[b.Slug] = true
		bankSlugs[i] = b.Slug
	}

	// Call AI.
	parsed, err := s.callGemini(ctx, pdfBytes, slugs, bankSlugs, mappingsResp.Mappings)
	if err != nil {
		fail("AI parsing failed: " + err.Error())
		return
	}

	// Validate bank slug from AI response.
	if !bankSlugSet[parsed.BankSlug] {
		fail(fmt.Sprintf("bank not supported: %q (known banks: %v)", parsed.BankSlug, bankSlugs))
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

	// Process each card found in the statement.
	for _, pc := range parsed.Cards {
		if err := s.processCard(ctx, statementID, parsed.BankSlug, parsed.Year, parsed.Month, pc); err != nil {
			rlog.Error("parser: failed to process card",
				"statement_id", statementID,
				"card_last4", pc.CardLast4,
				"err", err,
			)
		}
	}

	// Sum card balances for the overall statement balance.
	var total float64
	for _, pc := range parsed.Cards {
		total += parseBalanceString(pc.StatementBalance)
	}
	balanceStr := ""
	if total > 0 {
		balanceStr = strconv.FormatFloat(total, 'f', 2, 64)
	}

	// Mark statement as parsed (sets year, month, balance, status=1).
	if _, err := updateStatementParsed(ctx, statementID, parsed.Year, parsed.Month, balanceStr); err != nil {
		rlog.Error("parser: failed to mark statement as parsed", "statement_id", statementID, "err", err)
	}
}

// parseBalanceString strips currency symbols / commas and returns a float.
// e.g. "RM 3,421.50" → 3421.50,  "1,200.00" → 1200.00
func parseBalanceString(s string) float64 {
	var b strings.Builder
	for _, r := range s {
		if (r >= '0' && r <= '9') || r == '.' || r == '-' {
			b.WriteRune(r)
		}
	}
	v, _ := strconv.ParseFloat(b.String(), 64)
	return v
}

// processCard handles one card section from the parsed statement.
func (s *Service) processCard(ctx context.Context, statementID, bankSlug string, year, month int, pc parsedCard) error {
	// Look up card by last4 + bank slug (last4 alone is not unique enough).
	card, err := cards.GetCardByLast4AndBank(ctx, &cards.GetCardByLast4AndBankParams{
		Last4:    pc.CardLast4,
		BankSlug: bankSlug,
	})
	if err != nil {
		// Card not found in system — record as skipped so re-upload can retry later.
		rlog.Info("parser: card not found, skipping", "card_last4", pc.CardLast4, "bank_slug", bankSlug)
		return insertCardStatementSkipped(ctx, statementID, pc.CardLast4)
	}

	// Check if this card already has a parsed entry for this period.
	exists, err := cardStatementExistsForPeriod(ctx, card.ID, year, month)
	if err != nil {
		return fmt.Errorf("check card period: %w", err)
	}
	if exists {
		rlog.Info("parser: card already parsed for period, skipping",
			"card_id", card.ID,
			"year", year,
			"month", month,
		)
		return nil
	}

	// Store transactions for this card.
	if len(pc.Transactions) > 0 {
		inputs := make([]transactions.TransactionInput, 0, len(pc.Transactions))
		for _, t := range pc.Transactions {
			date, err := time.Parse("2006-01-02", t.TxnDate)
			if err != nil {
				rlog.Warn("parser: skipping transaction with bad date",
					"txn_date", t.TxnDate,
					"card_last4", pc.CardLast4,
					"err", err,
				)
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
				CardID:       card.ID,
				Transactions: inputs,
			}); err != nil {
				return fmt.Errorf("store transactions: %w", err)
			}
		}
	}

	// Record successful card_statement entry.
	return insertCardStatementParsed(ctx, statementID, card.ID, pc.CardLast4, pc.StatementBalance)
}

func (s *Service) callGemini(ctx context.Context, pdfBytes []byte, categorySlugs, bankSlugs []string, mappings []*transactions.CategoryMapping) (*parsedStatement, error) {
	model := s.gemini.GenerativeModel("gemini-3.5-flash")
	model.ResponseMIMEType = "application/json"

	resp, err := model.GenerateContent(ctx,
		genai.Blob{MIMEType: "application/pdf", Data: pdfBytes},
		genai.Text(buildParsePrompt(categorySlugs, bankSlugs, mappings)),
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
