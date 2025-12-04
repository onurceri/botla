package handlers

import "testing"

func TestParseChatbotIDFromPath(t *testing.T) {
    id, ok := parseChatbotIDFromPath("/api/v1/chatbots/abc/sources")
    if !ok || id != "abc" {
        t.Fatalf("expected id 'abc' got ok=%v id=%q", ok, id)
    }
    if _, ok := parseChatbotIDFromPath("/api/v1/chatbots//sources"); ok {
        t.Fatalf("expected not ok for empty id")
    }
    if _, ok := parseChatbotIDFromPath("/api/v1/chatbots/abc/x"); ok {
        t.Fatalf("expected not ok for wrong suffix")
    }
}

func TestIsPDFContentType(t *testing.T) {
    if !isPDFContentType("application/pdf", "x.txt") { t.Fatal("ct should pass") }
    if !isPDFContentType("", "x.pdf") { t.Fatal("suffix should pass") }
    if isPDFContentType("text/plain", "x.txt") { t.Fatal("should fail") }
}
