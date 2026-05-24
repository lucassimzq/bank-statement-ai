package statements

import (
	"strings"

	"encore.app/transactions"
)

func buildParsePrompt(categorySlugs []string, mappings []*transactions.CategoryMapping) string {
	slugList := strings.Join(categorySlugs, ", ")

	var mappingSection string
	if len(mappings) > 0 {
		var sb strings.Builder
		sb.WriteString("\nKnown merchant mappings — when the merchant_raw contains any of these patterns, use the specified category_slug:\n")
		for _, m := range mappings {
			sb.WriteString("- \"")
			sb.WriteString(m.MerchantPattern)
			sb.WriteString("\" -> ")
			sb.WriteString(m.CategorySlug)
			sb.WriteString("\n")
		}
		mappingSection = sb.String()
	}

	return `You are a bank statement parser. Extract all transactions from this credit card PDF statement.

Return a JSON object with this exact structure:
{
  "year": 2024,
  "month": 1,
  "statement_balance": "1234.56",
  "transactions": [
    {
      "txn_date": "2024-01-15",
      "merchant_raw": "GRAB*FOOD 12345",
      "merchant": "Grab Food",
      "amount": "45.80",
      "category_slug": "dining"
    }
  ]
}

Rules:
- year: the statement year as an integer (e.g. 2024)
- month: the statement month as an integer 1-12 (e.g. 1 for January)
- statement_balance: the total/closing balance or amount due shown on the statement (string, 2 decimal places, no currency symbol)
- txn_date: ISO date format YYYY-MM-DD
- merchant_raw: the exact merchant string as it appears on the statement
- merchant: a clean, human-readable merchant name
- amount: positive number as string with 2 decimal places (no currency symbol); use negative for credits/refunds
- category_slug: must be one of: ` + slugList + mappingSection + `
- When unsure of the category, use "others" rather than guessing

Return ONLY the JSON object, no explanation or markdown.`
}
