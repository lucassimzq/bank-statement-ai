package statements

import (
	"strings"

	"encore.app/transactions"
)

func buildParsePrompt(categorySlugs []string, bankSlugs []string, mappings []*transactions.CategoryMapping) string {
	catList := strings.Join(categorySlugs, ", ")
	bankList := strings.Join(bankSlugs, ", ")

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

The statement may contain multiple cards (combined credit limit accounts). Each card has its own clearly labelled section showing the full card number.

Return a JSON object with this exact structure:
{
  "bank_slug": "maybank",
  "year": 2024,
  "month": 1,
  "cards": [
    {
      "card_last4": "1234",
      "statement_balance": "1100.49",
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
  ]
}

Rules:
- bank_slug: the issuing bank slug — must be one of: ` + bankList + `
- year: the statement year as an integer (e.g. 2024)
- month: the statement month as an integer 1-12 (e.g. 1 for January)
- cards: one entry per card found in the statement
- card_last4: the last 4 digits of the card number as shown in each section header
- statement_balance: the SUB TOTAL for this card section (string, 2 decimal places, no currency symbol)
- txn_date: ISO date format YYYY-MM-DD
- merchant_raw: the exact merchant string as it appears on the statement
- merchant: a clean, human-readable merchant name
- amount: positive number as string with 2 decimal places (no currency symbol); use negative for credits/refunds
- category_slug: must be one of: ` + catList + mappingSection + `
- When unsure of the category, use "others" rather than guessing
- Only include actual purchase/debit transactions. Skip interest charges, payment credits, balance transfers, and administrative lines.

Return ONLY the JSON object, no explanation or markdown.`
}
