package files

import (
	"fmt"
)

// Cache error types
type ErrorType uint

const (
	ErrUnknown ErrorType = iota
	ErrFileOpening
	ErrFileReading
)

type Error struct {
	// The type of error
	Type ErrorType

	// The error message
	Message string

	// Previous error (for unwrapping)
	previous error
}

// String for representing error message.
func (e ErrorType) String() string {
	switch e {
	case ErrUnknown:
		return "unknown"
	case ErrFileOpening:
		return "cannot open file"
	case ErrFileReading:
		return "cannot read file"
	}

	return "unrecognized error type"
}

// Error returns the error's message.
func (e *Error) Error() string {
	return e.Message
}

// Unwrap returns previous error. If previous error does not exits, Unwrap returns nil.
func (e *Error) Unwrap() error {
	return e.previous
}

// Creates new error instance. Previous error can be nil.
func newError(tp ErrorType, message string, prev error) *Error {
	return &Error{Type: tp, Message: message, previous: prev}
}

// Creates new error instance with message formatting.
func newErrorf(tp ErrorType, prev error, format string, args ...interface{}) *Error {
	return newError(tp, fmt.Sprintf(format, args...), prev)
}
