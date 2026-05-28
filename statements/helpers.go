package statements

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

// sumCardBalances adds up statement_bal across all parsed card entries.
// Returns nil if none of the cards carry a balance.
func sumCardBalances(cards []*CardStatementInfo) *string {
	var total float64
	found := false
	for _, c := range cards {
		if c.StatementBal == nil || *c.StatementBal == "" {
			continue
		}
		v, err := strconv.ParseFloat(*c.StatementBal, 64)
		if err != nil {
			continue
		}
		total += v
		found = true
	}
	if !found {
		return nil
	}
	s := fmt.Sprintf("%.2f", total)
	return &s
}

type badRequestError struct{ msg string }

func (e *badRequestError) Error() string { return e.msg }

func errBad(msg string) error { return &badRequestError{msg} }

func jsonResponse(w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, message string, status int) {
	jsonResponse(w, map[string]string{"error": message}, status)
}

func writeErr(w http.ResponseWriter, err error) {
	var bad *badRequestError
	if errors.As(err, &bad) {
		jsonError(w, bad.msg, http.StatusBadRequest)
		return
	}
	jsonError(w, err.Error(), http.StatusInternalServerError)
}
