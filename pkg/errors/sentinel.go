// Package errors provides common sentinel errors for the application.
// These errors are designed to be used with errors.Is for type-safe
// error checking throughout the codebase, replacing magic string matching.
package errors

import "errors"

// ErrRateLimit indicates that an operation failed due to rate limiting
// (e.g., HTTP 429 responses from external APIs).
var ErrRateLimit = errors.New("rate limit exceeded")

// ErrTimeout indicates that an operation timed out.
// This is distinct from context cancellation.
var ErrTimeout = errors.New("operation timed out")

// ErrNetwork indicates a network-related error, such as connection refused
// or temporary network failures.
var ErrNetwork = errors.New("network error")

// ErrNotFound indicates that a requested resource was not found
// (e.g., HTTP 404 responses).
var ErrNotFound = errors.New("resource not found")

// ErrContextCancelled indicates that an operation was cancelled due to
// context cancellation (e.g., user-initiated cancellation or deadline).
var ErrContextCancelled = errors.New("context cancelled")
