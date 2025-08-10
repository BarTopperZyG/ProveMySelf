package types

import (
	"errors"
	"fmt"
	"net/http"
)

// Common error codes used across the application
const (
	// Generic errors
	ErrorCodeInternalError      = "internal_error"
	ErrorCodeBadRequest         = "bad_request"
	ErrorCodeNotFound           = "not_found"
	ErrorCodeUnauthorized       = "unauthorized"
	ErrorCodeForbidden          = "forbidden"
	ErrorCodeConflict           = "conflict"
	ErrorCodeValidationFailed   = "validation_failed"
	ErrorCodeInvalidContentType = "invalid_content_type"
	ErrorCodeRequestTooLarge    = "request_too_large"
	ErrorCodeRateLimited        = "rate_limited"

	// Project-specific errors
	ErrorCodeProjectNotFound     = "project_not_found"
	ErrorCodeProjectTitleTooShort = "project_title_too_short"
	ErrorCodeProjectTitleTooLong  = "project_title_too_long"
	ErrorCodeProjectExists       = "project_exists"

	// File upload errors
	ErrorCodeFileNotFound     = "file_not_found"
	ErrorCodeFileTooBig       = "file_too_big"
	ErrorCodeInvalidFileType  = "invalid_file_type"
	ErrorCodeStorageUnavailable = "storage_unavailable"

	// Authentication errors
	ErrorCodeInvalidToken     = "invalid_token"
	ErrorCodeTokenExpired     = "token_expired"
	ErrorCodeInvalidCredentials = "invalid_credentials"

	// Authorization errors
	ErrorCodeInsufficientPermissions = "insufficient_permissions"
	ErrorCodeResourceAccessDenied     = "resource_access_denied"
)

// APIError represents a structured API error
type APIError struct {
	Code       string
	Message    string
	Details    string
	StatusCode int
	Cause      error
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *APIError) Unwrap() error {
	return e.Cause
}

// ToErrorResponse converts APIError to ErrorResponse
func (e *APIError) ToErrorResponse() ErrorResponse {
	var details *string
	if e.Details != "" {
		details = &e.Details
	}

	return ErrorResponse{
		Error: ErrorDetail{
			Code:    e.Code,
			Message: e.Message,
			Details: details,
		},
	}
}

// NewAPIError creates a new API error
func NewAPIError(code, message string, statusCode int) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// NewAPIErrorWithDetails creates a new API error with details
func NewAPIErrorWithDetails(code, message, details string, statusCode int) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		Details:    details,
		StatusCode: statusCode,
	}
}

// NewAPIErrorWithCause creates a new API error with an underlying cause
func NewAPIErrorWithCause(code, message string, statusCode int, cause error) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Cause:      cause,
	}
}

// Predefined API errors for common scenarios

var (
	ErrInternalServer = &APIError{
		Code:       ErrorCodeInternalError,
		Message:    "An unexpected error occurred",
		StatusCode: http.StatusInternalServerError,
	}

	ErrBadRequest = &APIError{
		Code:       ErrorCodeBadRequest,
		Message:    "Invalid request",
		StatusCode: http.StatusBadRequest,
	}

	ErrNotFound = &APIError{
		Code:       ErrorCodeNotFound,
		Message:    "Resource not found",
		StatusCode: http.StatusNotFound,
	}

	ErrUnauthorized = &APIError{
		Code:       ErrorCodeUnauthorized,
		Message:    "Authentication required",
		StatusCode: http.StatusUnauthorized,
	}

	ErrForbidden = &APIError{
		Code:       ErrorCodeForbidden,
		Message:    "Access forbidden",
		StatusCode: http.StatusForbidden,
	}

	ErrValidationFailed = &APIError{
		Code:       ErrorCodeValidationFailed,
		Message:    "Validation failed",
		StatusCode: http.StatusBadRequest,
	}

	ErrProjectNotFound = &APIError{
		Code:       ErrorCodeProjectNotFound,
		Message:    "Project not found",
		StatusCode: http.StatusNotFound,
	}

	ErrProjectTitleTooShort = &APIError{
		Code:       ErrorCodeProjectTitleTooShort,
		Message:    "Project title is too short",
		StatusCode: http.StatusUnprocessableEntity,
	}

	ErrProjectTitleTooLong = &APIError{
		Code:       ErrorCodeProjectTitleTooLong,
		Message:    "Project title is too long",
		StatusCode: http.StatusUnprocessableEntity,
	}

	ErrFileTooBig = &APIError{
		Code:       ErrorCodeFileTooBig,
		Message:    "File size exceeds the maximum allowed limit",
		StatusCode: http.StatusRequestEntityTooLarge,
	}

	ErrInvalidFileType = &APIError{
		Code:       ErrorCodeInvalidFileType,
		Message:    "File type is not allowed",
		StatusCode: http.StatusUnsupportedMediaType,
	}

	ErrStorageUnavailable = &APIError{
		Code:       ErrorCodeStorageUnavailable,
		Message:    "Storage service is currently unavailable",
		StatusCode: http.StatusServiceUnavailable,
	}
)

// MapDomainError maps domain errors to API errors
func MapDomainError(err error) *APIError {
	if err == nil {
		return nil
	}

	// If it's already an API error, return it
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}

	// Map common domain errors
	switch {
	case errors.Is(err, errors.New("project not found")):
		return ErrProjectNotFound
	case errors.Is(err, errors.New("project title too short")):
		return ErrProjectTitleTooShort
	case errors.Is(err, errors.New("project title too long")):
		return ErrProjectTitleTooLong
	case errors.Is(err, errors.New("file not found")):
		return NewAPIError(ErrorCodeFileNotFound, "File not found", http.StatusNotFound)
	case errors.Is(err, errors.New("file too big")):
		return ErrFileTooBig
	case errors.Is(err, errors.New("invalid file type")):
		return ErrInvalidFileType
	case errors.Is(err, errors.New("storage unavailable")):
		return ErrStorageUnavailable
	default:
		// For unknown errors, return internal server error
		return NewAPIErrorWithCause(ErrorCodeInternalError, "An unexpected error occurred", 
			http.StatusInternalServerError, err)
	}
}

// IsClientError returns true if the error is a client error (4xx)
func IsClientError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode >= 400 && apiErr.StatusCode < 500
	}
	return false
}

// IsServerError returns true if the error is a server error (5xx)
func IsServerError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode >= 500
	}
	return false
}

// GetHTTPStatusCode returns the HTTP status code for an error
func GetHTTPStatusCode(err error) int {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode
	}
	return http.StatusInternalServerError
}