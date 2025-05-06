// Package handlers provides HTTP handlers for the API endpoints
package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

// Response represents a standard API response.
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorData  `json:"error,omitempty"`
}

// ErrorData represents error details in an API response.
type ErrorData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewSuccessResponse creates a new success response.
func NewSuccessResponse(data interface{}) Response {
	return Response{
		Success: true,
		Data:    data,
	}
}

// NewErrorResponse creates a new error response.
func NewErrorResponse(code, message string) Response {
	return Response{
		Success: false,
		Error: &ErrorData{
			Code:    code,
			Message: message,
		},
	}
}

// WriteJSON writes a JSON response to the HTTP response writer.
func WriteJSON(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

// RespondWithError sends an error response.
func RespondWithError(w http.ResponseWriter, status int, code, message string) {
	if err := WriteJSON(w, status, NewErrorResponse(code, message)); err != nil {
		// If we can't write the error response, log it and write a simpler error
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// RespondWithJSON sends a success response.
func RespondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	if err := WriteJSON(w, status, NewSuccessResponse(data)); err != nil {
		// If we can't write the JSON response, log it and write a simpler error
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// Timestamp represents a custom time format for JSON serialization.
type Timestamp time.Time

// MarshalJSON implements the json.Marshaler interface for Timestamp.
func (t Timestamp) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(t).Unix())
}
