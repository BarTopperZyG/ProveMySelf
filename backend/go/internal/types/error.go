package types

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error information
type ErrorDetail struct {
	Code    string  `json:"code"`
	Message string  `json:"message"`
	Details *string `json:"details,omitempty"`
}

// ValidationErrorResponse represents a validation error response
type ValidationErrorResponse struct {
	Error ValidationErrorDetail `json:"error"`
}

// ValidationErrorDetail represents detailed validation error information
type ValidationErrorDetail struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Errors  []ValidationError `json:"errors"`
}

// ValidationError represents a single field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Message string `json:"message"`
}