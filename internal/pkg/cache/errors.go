package cache

import "errors"

var (
	// ErrEmptyKey means "empty key is used".
	ErrEmptyKey = errors.New("empty key")

	// ErrEmptyData means "no data provided".
	ErrEmptyData = errors.New("empty data")

	ErrClosed = errors.New("cache closed")
)
