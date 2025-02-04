package errors

import (
	"fmt"
	"net/http"
)

// ErrorType represents different categories of errors.
type ErrorType int

const (
	ErrorTypeInternal ErrorType = iota
	ErrorTypeValidation
	ErrorTypeExternal
)

// ServiceError encapsulates error details for the service.
type ServiceError struct {
	Type    ErrorType // The category of the error.
	Message string    // A human-readable error message.
	Err     error     // The underlying error.
}

// Error returns the formatted error string.
func (e *ServiceError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// IsRetryable determines if the error is retryable (typically for external errors).
func (e *ServiceError) IsRetryable() bool {
	return e.Type == ErrorTypeExternal
}

// NewInternalError creates a new internal error.
func NewInternalError(message string, err error) *ServiceError {
	return &ServiceError{
		Type:    ErrorTypeInternal,
		Message: message,
		Err:     err,
	}
}

// NewValidationError creates a new validation error.
func NewValidationError(message string) *ServiceError {
	return &ServiceError{
		Type:    ErrorTypeValidation,
		Message: message,
	}
}

// NewExternalError creates a new error related to external systems.
func NewExternalError(message string, err error) *ServiceError {
	return &ServiceError{
		Type:    ErrorTypeExternal,
		Message: message,
		Err:     err,
	}
}

// ToHTTPError maps a ServiceError to an HTTP status code and a message.
// This can be extended for structured logging or external error reporting.
func ToHTTPError(err error) (int, string) {
	if svcErr, ok := err.(*ServiceError); ok {
		switch svcErr.Type {
		case ErrorTypeValidation:
			return http.StatusBadRequest, svcErr.Message
		case ErrorTypeExternal:
			return http.StatusServiceUnavailable, "External service error"
		default:
			return http.StatusInternalServerError, "Internal server error"
		}
	}
	return http.StatusInternalServerError, "Unknown error"
}
