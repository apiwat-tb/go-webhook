package domain

import "errors"

var (
	// ErrNotFound is returned when an entity does not exist in storage.
	ErrNotFound = errors.New("not found")

	// ErrInvalidInput is returned when input validation fails.
	ErrInvalidInput = errors.New("invalid input")
)
