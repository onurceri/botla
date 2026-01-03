package integration

import (
	"fmt"
	"io"
	"net/http"
)

// testHTTPClient returns an HTTP client configured for testing.
// It disables keep-alives to prevent connection leaks in parallel tests.
//
//nolint:unused
func testHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}
}

// testHTTPPost performs an HTTP POST request using the test client.
// This prevents connection leaks in parallel tests.
//
//nolint:unused
func testHTTPPost(url, contentType string, body io.Reader) (*http.Response, error) {
	resp, err := testHTTPClient().Post(url, contentType, body)
	if err != nil {
		return nil, fmt.Errorf("test HTTP POST: %w", err)
	}
	return resp, nil
}

// testHTTPGet performs an HTTP GET request using the test client.
// This prevents connection leaks in parallel tests.
//
//nolint:unused
func testHTTPGet(url string) (*http.Response, error) {
	resp, err := testHTTPClient().Get(url)
	if err != nil {
		return nil, fmt.Errorf("test HTTP GET: %w", err)
	}
	return resp, nil
}

// drainBody fully reads and discards the response body to allow connection reuse.
// This prevents goroutine leaks when the response body is not fully consumed.
// Returns the response for chaining.
//
//nolint:unused
func drainBody(res *http.Response) *http.Response {
	if res != nil && res.Body != nil {
		_, _ = io.Copy(io.Discard, res.Body)
		_ = res.Body.Close()
	}
	return res
}
