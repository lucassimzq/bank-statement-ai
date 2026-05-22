package cards

import "time"

type Bank struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

type Card struct {
	ID        string    `json:"id"`
	BankID    string    `json:"bank_id"`
	BankName  string    `json:"bank_name"`
	Label     string    `json:"label"`
	Purpose   string    `json:"purpose"`
	Last4     string    `json:"last4"`
	CreatedAt time.Time `json:"created_at"`
}
