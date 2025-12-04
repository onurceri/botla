package scraper

import "testing"

func TestIsHTMLContentType(t *testing.T) {
    if !isHTMLContentType("text/html; charset=utf-8") { t.Fatal("text/html should be true") }
    if !isHTMLContentType("application/xhtml+xml") { t.Fatal("xhtml should be true") }
    if isHTMLContentType("application/json") { t.Fatal("json should be false") }
}
