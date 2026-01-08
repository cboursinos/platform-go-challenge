package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode represents a specific error type
type ErrorCode string

const (
	// Client errors (4xx)
	ErrorCodeValidation     ErrorCode = "VALIDATION_ERROR"
	ErrorCodeNotFound       ErrorCode = "NOT_FOUND"
	ErrorCodeConflict       ErrorCode = "CONFLICT"
	ErrorCodeBadRequest     ErrorCode = "BAD_REQUEST"
	ErrorCodeUnauthorized   ErrorCode = "UNAUTHORIZED"
	ErrorCodeForbidden      ErrorCode = "FORBIDDEN"
	ErrorCodeRequestTimeout ErrorCode = "REQUEST_TIMEOUT"

	// Server errors (5xx)
	ErrorCodeInternal     ErrorCode = "INTERNAL_ERROR"
	ErrorCodeUnavailable  ErrorCode = "SERVICE_UNAVAILABLE"
	ErrorCodeNotImplemented ErrorCode = "NOT_IMPLEMENTED"
)

// AppError represents an application error with HTTP status code
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	HTTPStatus int       `json:"-"`
	Err        error     `json:"-"` // Original error for logging
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewValidationError creates a validation error (400)
func NewValidationError(message string, err error) *AppError {
	return &AppError{
		Code:       ErrorCodeValidation,
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
		Err:        err,
	}
}

// NewNotFoundError creates a not found error (404)
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:       ErrorCodeNotFound,
		Message:    fmt.Sprintf("%s not found", resource),
		HTTPStatus: http.StatusNotFound,
	}
}

// NewConflictError creates a conflict error (409)
func NewConflictError(message string) *AppError {
	return &AppError{
		Code:       ErrorCodeConflict,
		Message:    message,
		HTTPStatus: http.StatusConflict,
	}
}

// NewBadRequestError creates a bad request error (400)
func NewBadRequestError(message string) *AppError {
	return &AppError{
		Code:       ErrorCodeBadRequest,
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
	}
}

// NewInternalError creates an internal server error (500)
func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Code:       ErrorCodeInternal,
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
		Err:        err,
	}
}

// NewServiceUnavailableError creates a service unavailable error (503)
func NewServiceUnavailableError(message string, err error) *AppError {
	return &AppError{
		Code:       ErrorCodeUnavailable,
		Message:    message,
		HTTPStatus: http.StatusServiceUnavailable,
		Err:        err,
	}
}

// NewRequestTimeoutError creates a request timeout error (408)
func NewRequestTimeoutError(message string) *AppError {
	return &AppError{
		Code:       ErrorCodeRequestTimeout,
		Message:    message,
		HTTPStatus: http.StatusRequestTimeout,
	}
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// AsAppError converts an error to AppError if possible
func AsAppError(err error) (*AppError, bool) {
	appErr, ok := err.(*AppError)
	return appErr, ok
}

// ToAppError converts any error to an AppError
// If it's already an AppError, returns it as-is
// Otherwise wraps it as an internal error
func ToAppError(err error) *AppError {
	if err == nil {
		return nil
	}

	if appErr, ok := AsAppError(err); ok {
		return appErr
	}

	// Default to internal error for unknown errors
	return NewInternalError("An unexpected error occurred", err)
}
