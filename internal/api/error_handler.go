package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gwi/platform-go-challenge/internal/errors"
)

// ErrorResponse represents the error response structure
type ErrorResponse struct {
	Error   errors.ErrorCode `json:"error"`
	Message string           `json:"message"`
}

// HandleError handles errors and returns appropriate HTTP responses
func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}

	// Check for context cancellation errors first
	if err == context.Canceled {
		// Client disconnected - don't log as error, just return
		respondWithError(w, http.StatusRequestTimeout, errors.ErrorCodeRequestTimeout, "Request was cancelled")
		return
	}

	if err == context.DeadlineExceeded {
		// Timeout - log for monitoring
		var ctx context.Context
		if r != nil {
			ctx = r.Context()
		} else {
			ctx = context.Background()
		}
		logError(ctx, errors.NewRequestTimeoutError("Request timed out"), err)
		respondWithError(w, http.StatusRequestTimeout, errors.ErrorCodeRequestTimeout, "Request timed out")
		return
	}

	// Try to extract AppError
	appErr, ok := errors.AsAppError(err)
	if !ok {
		// Check for common database errors and context errors
		appErr = mapDatabaseError(err)
	}

	// Log error details (including stack trace in production)
	// Handle nil request gracefully for legacy error handling
	var ctx context.Context
	if r != nil {
		ctx = r.Context()
	} else {
		ctx = context.Background()
	}
	logError(ctx, appErr, err)

	// Don't expose internal error details to clients in production
	message := appErr.Message
	if appErr.HTTPStatus >= 500 {
		// In production, don't expose internal error details
		// You might want to check an environment variable here
		message = "An internal error occurred. Please try again later."
	}

	respondWithError(w, appErr.HTTPStatus, appErr.Code, message)
}

// mapDatabaseError maps database errors to appropriate AppErrors
func mapDatabaseError(err error) *errors.AppError {
	if err == nil {
		return nil
	}

	// Check for context cancellation errors
	if err == context.Canceled {
		return errors.NewRequestTimeoutError("Request was cancelled")
	}

	if err == context.DeadlineExceeded {
		return errors.NewRequestTimeoutError("Request timed out")
	}

	// Check for sql.ErrNoRows
	if err == sql.ErrNoRows {
		return errors.NewNotFoundError("Resource")
	}

	errStr := strings.ToLower(err.Error())

	// Check for context-related error messages
	if strings.Contains(errStr, "context canceled") || strings.Contains(errStr, "request cancelled") {
		return errors.NewRequestTimeoutError("Request was cancelled")
	}

	if strings.Contains(errStr, "context deadline exceeded") || strings.Contains(errStr, "request timed out") {
		return errors.NewRequestTimeoutError("Request timed out")
	}

	// Database connection errors
	if strings.Contains(errStr, "connection") || strings.Contains(errStr, "timeout") {
		return errors.NewServiceUnavailableError("Database temporarily unavailable", err)
	}

	// Constraint violations
	if strings.Contains(errStr, "duplicate") || strings.Contains(errStr, "unique constraint") {
		return errors.NewConflictError("Resource already exists")
	}

	// Foreign key violations
	if strings.Contains(errStr, "foreign key") || strings.Contains(errStr, "constraint") {
		return errors.NewBadRequestError("Invalid reference: related resource does not exist")
	}

	// Default to internal error
	return errors.NewInternalError("Database operation failed", err)
}

// logError logs error details for monitoring
func logError(ctx context.Context, appErr *errors.AppError, originalErr error) {
	// In production, you'd want to use structured logging
	// and include request ID, user ID, etc.

	if appErr.HTTPStatus >= 500 {
		// Log full error details for server errors
		log.Printf("[ERROR] %s (HTTP %d): %v", appErr.Code, appErr.HTTPStatus, originalErr)
	} else {
		// Log client errors at a lower level
		log.Printf("[WARN] %s (HTTP %d): %s", appErr.Code, appErr.HTTPStatus, appErr.Message)
	}
}

// respondWithError sends an error JSON response
func respondWithError(w http.ResponseWriter, code int, errorCode errors.ErrorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := ErrorResponse{
		Error:   errorCode,
		Message: message,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Fallback if JSON encoding fails
		http.Error(w, fmt.Sprintf(`{"error":"%s","message":"%s"}`, errorCode, message), code)
	}
}

