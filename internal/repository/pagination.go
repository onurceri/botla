package repository

import (
	"errors"
)

// ValidatePagination validates pagination parameters.
// It ensures that limit and offset are non-negative to prevent integer overflow
// when converting to uint64 for database queries.
func ValidatePagination(limit, offset int) (uint64, uint64, error) {
	if limit < 0 {
		return 0, 0, errors.New("limit cannot be negative")
	}
	if offset < 0 {
		return 0, 0, errors.New("offset cannot be negative")
	}
	return uint64(limit), uint64(offset), nil
}

// ValidateLimit validates the limit parameter.
// It ensures that limit is non-negative to prevent integer overflow.
func ValidateLimit(limit int) (uint64, error) {
	if limit < 0 {
		return 0, errors.New("limit cannot be negative")
	}
	return uint64(limit), nil
}
