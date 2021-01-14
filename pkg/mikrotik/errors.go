package mikrotik

// Error is a special type for package-specific errors.
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

// ErrEmptyFields means required fields does not filled.
const ErrEmptyFields Error = 1
