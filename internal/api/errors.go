package api

import "net/http"

// APIError represents a structured API error response
type APIError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	StatusCode int                    `json:"-"`
	Details    map[string]interface{} `json:"details,omitempty"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return e.Message
}

// NewError creates a new APIError
func NewError(code string, message string, statusCode int) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// WithDetails adds details to an error
func (e *APIError) WithDetails(details map[string]interface{}) *APIError {
	e.Details = details
	return e
}

// Predefined API errors
var (
	ErrBadRequest = NewError(
		ErrCodeInvalidRequest,
		MsgInvalidRequest,
		http.StatusBadRequest,
	)

	ErrNotFound = NewError(
		ErrCodeNotFound,
		MsgNotFound,
		http.StatusNotFound,
	)

	ErrUnauthorized = NewError(
		ErrCodeNotAuthenticated,
		MsgNotAuthenticated,
		http.StatusUnauthorized,
	)

	ErrForbidden = NewError(
		ErrCodeUnauthorized,
		MsgUnauthorized,
		http.StatusForbidden,
	)

	ErrInternalServer = NewError(
		ErrCodeInternalServer,
		MsgInternalServerError,
		http.StatusInternalServerError,
	)

	ErrConflict = NewError(
		ErrCodeConflict,
		MsgDuplicateUser,
		http.StatusConflict,
	)

	ErrValidationFailed = NewError(
		ErrCodeValidationFailed,
		MsgValidationFailed,
		http.StatusBadRequest,
	)

	ErrRateLimited = NewError(
		ErrCodeRateLimited,
		"Rate limit exceeded",
		http.StatusTooManyRequests,
	)
)
