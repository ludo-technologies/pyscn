package domain

import (
	"fmt"
)

// DomainError represents errors in the domain layer
type DomainError struct {
	Code    string
	Message string
	Cause   error
}

func (e DomainError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e DomainError) Unwrap() error {
	return e.Cause
}

// Domain error codes
const (
	ErrCodeInvalidInput      = "INVALID_INPUT"
	ErrCodeFileNotFound      = "FILE_NOT_FOUND"
	ErrCodeParseError        = "PARSE_ERROR"
	ErrCodeAnalysisError     = "ANALYSIS_ERROR"
	ErrCodeConfigError       = "CONFIG_ERROR"
	ErrCodeOutputError       = "OUTPUT_ERROR"
	ErrCodeUnsupportedFormat = "UNSUPPORTED_FORMAT"
	ErrCodeTimeout           = "TIMEOUT"
	ErrCodeCancelled         = "CANCELLED"
	ErrCodeNotImplemented    = "NOT_IMPLEMENTED"
	ErrCodeInternal          = "INTERNAL"
)

// NewDomainError creates a new domain error
func NewDomainError(code, message string, cause error) error {
	return DomainError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// NewInvalidInputError creates an invalid input error
func NewInvalidInputError(message string, cause error) error {
	return NewDomainError(ErrCodeInvalidInput, message, cause)
}

// NewFileNotFoundError creates a file not found error
func NewFileNotFoundError(path string, cause error) error {
	return NewDomainError(ErrCodeFileNotFound, fmt.Sprintf("file not found: %s", path), cause)
}

// NewParseError creates a parse error
func NewParseError(file string, cause error) error {
	return NewDomainError(ErrCodeParseError, fmt.Sprintf("failed to parse file: %s", file), cause)
}

// NewAnalysisError creates an analysis error
func NewAnalysisError(message string, cause error) error {
	return NewDomainError(ErrCodeAnalysisError, message, cause)
}

// NewConfigError creates a configuration error
func NewConfigError(message string, cause error) error {
	return NewDomainError(ErrCodeConfigError, message, cause)
}

// NewOutputError creates an output error
func NewOutputError(message string, cause error) error {
	return NewDomainError(ErrCodeOutputError, message, cause)
}

// NewUnsupportedFormatError creates an unsupported format error
func NewUnsupportedFormatError(format string) error {
	return NewDomainError(ErrCodeUnsupportedFormat, fmt.Sprintf("unsupported format: %s", format), nil)
}

// NewValidationError creates a validation error
func NewValidationError(message string) error {
	return NewDomainError(ErrCodeInvalidInput, message, nil)
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(message string, cause error) error {
	return NewDomainError(ErrCodeTimeout, message, cause)
}

// NewCancelledError creates a cancellation error
func NewCancelledError(message string, cause error) error {
	return NewDomainError(ErrCodeCancelled, message, cause)
}

// NewNotImplementedError creates a not implemented error
func NewNotImplementedError(feature string) error {
	return NewDomainError(ErrCodeNotImplemented, fmt.Sprintf("not implemented: %s", feature), nil)
}

// NewInternalError creates an internal error
func NewInternalError(message string, cause error) error {
	return NewDomainError(ErrCodeInternal, message, cause)
}
