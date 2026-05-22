package statements

import "time"

type Statement struct {
	ID           string     `json:"id"`
	CardID       string     `json:"card_id"`
	Year         int        `json:"year"`
	Month        int        `json:"month"`
	StatementBal *string    `json:"statement_bal"`
	FilePath     *string    `json:"file_path"`
	ParsedAt     *time.Time `json:"parsed_at"`
	CreatedAt    time.Time  `json:"created_at"`
}
