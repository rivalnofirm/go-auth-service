package response

import (
	"encoding/json"
	"net/http"
)

// SuccessResponse defines the structure for successful responses.
type SuccessResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse defines the structure for error responses.
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func JSON(w http.ResponseWriter, statusCode int, status string, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if statusCode >= 400 {
		json.NewEncoder(w).Encode(ErrorResponse{Status: status, Message: message})
	} else {
		json.NewEncoder(w).Encode(SuccessResponse{Status: status, Message: message, Data: data})
	}
}
