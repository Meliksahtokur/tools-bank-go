package utils

import (
	"fmt"
	"strings"
)

// Sentinel errors for common error conditions.
var (
	ErrNotFound    = fmt.Errorf("not found")
	ErrInvalidInput = fmt.Errorf("invalid input")
	ErrInternal    = fmt.Errorf("internal error")
)

// WError wraps an error with additional context.
// It implements the standard error interface and retains
// the underlying error for unwrapping.
type WError struct {
	msg    string
	err    error
	fields map[string]any
}

// Error returns the error message including wrapped context.
func (e *WError) Error() string {
	if e.err == nil {
		return e.msg
	}
	return fmt.Sprintf("%s: %s", e.msg, e.err.Error())
}

// Unwrap returns the underlying error for errors.Is/As.
func (e *WError) Unwrap() error {
	return e.err
}

// WithFields attaches additional fields to the wrapped error.
// Returns a new WError with the fields attached.
func (e *WError) WithFields(fields map[string]any) *WError {
	if e.fields == nil {
		e.fields = make(map[string]any)
	}
	for k, v := range fields {
		e.fields[k] = v
	}
	return e
}

// Wrap creates a new WError wrapping the given error with a message.
// If err is nil, Wrap returns nil.
func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return &WError{
		msg: msg,
		err: err,
	}
}

// WrapWithFields creates a new WError with additional fields.
func WrapWithFields(err error, msg string, fields map[string]any) error {
	if err == nil {
		return nil
	}
	return &WError{
		msg:    msg,
		err:    err,
		fields: fields,
	}
}

// IsNotFound checks if the error chain contains ErrNotFound.
// It uses errors.Is to properly handle wrapped errors.
func IsNotFound(err error) bool {
	return isErrorMatch(err, ErrNotFound)
}

// IsInvalidInput checks if the error chain contains ErrInvalidInput.
func IsInvalidInput(err error) bool {
	return isErrorMatch(err, ErrInvalidInput)
}

// IsInternal checks if the error chain contains ErrInternal.
func IsInternal(err error) bool {
	return isErrorMatch(err, ErrInternal)
}

// isErrorMatch checks if any error in the chain matches the target.
func isErrorMatch(err error, target error) bool {
	if err == nil {
		return false
	}

	// Check direct match
	if err == target {
		return true
	}

	// Check error message match for sentinel errors
	if werr, ok := err.(*WError); ok {
		if werr.err == target {
			return true
		}
		// Recursively check wrapped errors
		return isErrorMatch(werr.err, target)
	}

	// Try standard errors.Is for wrapped errors
	if strings.Contains(err.Error(), target.Error()) {
		return true
	}

	return false
}

// NewInvalidInput creates an ErrInvalidInput with a custom message.
func NewInvalidInput(msg string) error {
	return &WError{
		msg: msg,
		err: ErrInvalidInput,
	}
}

// NewNotFound creates an ErrNotFound with a custom message.
func NewNotFound(msg string) error {
	return &WError{
		msg: msg,
		err: ErrNotFound,
	}
}

// NewInternal creates an ErrInternal with a custom message.
func NewInternal(msg string) error {
	return &WError{
		msg: msg,
		err: ErrInternal,
	}
}
