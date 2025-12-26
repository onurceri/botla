package errors

import "fmt"

// Wrapf wraps an error with a formatted message. If err is nil, returns nil.
// This helper ensures consistent error wrapping across the codebase following
// the "wrap at boundary" policy for all external package errors.
//
// Example:
//
//	resp, err := http.Get(url)
//	if err != nil {
//	    return nil, errors.Wrapf(err, "fetching sitemap from %s", url)
//	}
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}
