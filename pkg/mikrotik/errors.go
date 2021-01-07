package mikrotik

// Special type for package-specific errors.
type Error uint8

// Error returns error in a string representation.
func (err Error) Error() string {
	switch err {
	case ErrEmptyFields:
		return "required fields does not filled"

	default:
		return "unknown error"
	}
}

// Package-specific error constants.
const (
	ErrEmptyFields Error = iota + 1 // required fields does not filled
)
