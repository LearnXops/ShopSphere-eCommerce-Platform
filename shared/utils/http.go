package utils

import (
	"context"
	"encoding/json"
	"net/http"
)

// WriteJSONResponse writes a JSON response with the given status code and data
func WriteJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			Logger.Error(context.TODO(), "Failed to encode JSON response", err, map[string]interface{}{
				"status_code": statusCode,
			})
		}
	}
}

// WriteErrorResponse writes an error response with the given status code, error code, and message
func WriteErrorResponse(w http.ResponseWriter, statusCode int, errorCode, message string) {
	errorResponse := ErrorResponse{
		Error: ErrorDetail{
			Code:    errorCode,
			Message: message,
		},
	}
	
	WriteJSONResponse(w, statusCode, errorResponse)
}

// WriteAppErrorResponse writes an error response from an AppError
func WriteAppErrorResponse(w http.ResponseWriter, err *AppError) {
	errorResponse := ErrorResponse{
		Error: ErrorDetail{
			Code:    string(err.Code),
			Message: err.Message,
			Details: err.Details,
		},
	}
	
	WriteJSONResponse(w, err.HTTPStatusCode(), errorResponse)
}

// WriteValidationErrorResponse writes a validation error response
func WriteValidationErrorResponse(w http.ResponseWriter, message string) {
	WriteErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", message)
}

// WriteNotFoundResponse writes a not found error response
func WriteNotFoundResponse(w http.ResponseWriter, resource string) {
	WriteErrorResponse(w, http.StatusNotFound, "NOT_FOUND", resource+" not found")
}

// WriteInternalErrorResponse writes an internal server error response
func WriteInternalErrorResponse(w http.ResponseWriter, message string) {
	WriteErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}

// WriteConflictResponse writes a conflict error response
func WriteConflictResponse(w http.ResponseWriter, message string) {
	WriteErrorResponse(w, http.StatusConflict, "CONFLICT", message)
}

// WriteUnauthorizedResponse writes an unauthorized error response
func WriteUnauthorizedResponse(w http.ResponseWriter, message string) {
	WriteErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// WriteForbiddenResponse writes a forbidden error response
func WriteForbiddenResponse(w http.ResponseWriter, message string) {
	WriteErrorResponse(w, http.StatusForbidden, "FORBIDDEN", message)
}
