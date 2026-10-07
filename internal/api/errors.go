package api

import (
	"encoding/json"
	"net/http"
)

// ErrorDetail represents the structured error payload.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse wraps the error detail.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// writeError writes a standardized structured JSON error response.
func writeError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}
