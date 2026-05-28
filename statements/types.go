package statements

import "time"

const (
	StatusParsing = iota
	StatusParsed  = 1
	StatusError   = 2
)

// CardStatementStatus values for card_statement.status
const (
	CardStatusParsed  = 1
	CardStatusSkipped = 2
)

type Statement struct {
	ID           string     `json:"id"`
	Status       int        `json:"status"`
	Message      *string    `json:"message"`
	Year         *int       `json:"year"`
	Month        *int       `json:"month"`
	StatementBal *string    `json:"statement_bal"`
	FilePath     *string    `json:"file_path"`
	ParsedAt     *time.Time `json:"parsed_at"`
	CreatedAt    time.Time  `json:"created_at"`
}

// CardStatementInfo is a summary of one card found within a statement.
type CardStatementInfo struct {
	CardLast4    string  `json:"card_last4"`
	CardID       *string `json:"card_id"`       // null when the card was not in our system
	Status       int     `json:"status"`        // 1 = parsed, 2 = skipped
	StatementBal *string `json:"statement_bal"` // per-card balance from the statement
}

// StatementWithCards embeds a Statement plus the cards detected in it.
type StatementWithCards struct {
	ID           string               `json:"id"`
	Status       int                  `json:"status"`
	Message      *string              `json:"message"`
	Year         *int                 `json:"year"`
	Month        *int                 `json:"month"`
	StatementBal *string              `json:"statement_bal"`
	FilePath     *string              `json:"file_path"`
	ParsedAt     *time.Time           `json:"parsed_at"`
	CreatedAt    time.Time            `json:"created_at"`
	Cards        []*CardStatementInfo `json:"cards"`
}
