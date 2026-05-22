package statements

import (
	"encoding/json"
	"errors"
	"net/http"
)

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
