package statements

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"encore.app/transactions"
	"encore.dev/rlog"
	"github.com/google/generative-ai-go/genai"
)

// parseStatement downloads the PDF, sends it to Claude, and stores results.
// Intended to run as a background goroutine — errors are logged, not returned.
func (s *Service) parseStatement(statementID, cardID, filePath string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cats, err := transactions.ListCategories(ctx)
	if err != nil {
		rlog.Error("parser: failed to load categories", "statement_id", statementID, "err", err)
		return
	}
	slugs := make([]string, len(cats.Categories))
	for i, c := range cats.Categories {
		slugs[i] = c.Slug
	}

	mappingsResp, err := transactions.ListCategoryMappings(ctx)
	if err != nil {
		rlog.Error("parser: failed to load category mappings", "statement_id", statementID, "err", err)
		return
	}

	pdfBytes, err := downloadFile(ctx, filePath)
	if err != nil {
		rlog.Error("parser: download failed", "statement_id", statementID, "err", err)
		return
	}

	parsed, err := s.callGemini(ctx, pdfBytes, slugs, mappingsResp.Mappings)
	if err != nil {
		rlog.Error("parser: gemini call failed", "statement_id", statementID, "err", err)
		return
	}

	if err := storeParsedResults(ctx, statementID, cardID, parsed); err != nil {
		rlog.Error("parser: store failed", "statement_id", statementID, "err", err)
	}
}

func downloadFile(ctx context.Context, filePath string) ([]byte, error) {
	reader := statementFiles.Download(ctx, filePath)
	defer reader.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
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

func storeParsedResults(ctx context.Context, statementID, cardID string, parsed *parsedStatement) error {
	if len(parsed.Transactions) > 0 {
		inputs := make([]transactions.TransactionInput, 0, len(parsed.Transactions))
		for _, t := range parsed.Transactions {
			date, err := time.Parse("2006-01-02", t.TxnDate)
			if err != nil {
				rlog.Warn("parser: skipping bad date", "txn_date", t.TxnDate, "err", err)
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
				return err
			}
		}
	}

	if parsed.StatementBalance != "" {
		if _, err := updateStatementBalance(ctx, statementID, parsed.StatementBalance); err != nil {
			return err
		}
	}

	return nil
}
