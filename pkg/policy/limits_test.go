package policy

import "testing"

func TestTokenLimitConstants(t *testing.T) {
	// Verify token limits are positive and in ascending order
	if TokenLimitFree <= 0 {
		t.Error("TokenLimitFree should be positive")
	}
	if TokenLimitPro <= TokenLimitFree {
		t.Error("TokenLimitPro should be greater than TokenLimitFree")
	}
	if TokenLimitUltra <= TokenLimitPro {
		t.Error("TokenLimitUltra should be greater than TokenLimitPro")
	}

	// Verify expected values (these are reference values)
	if TokenLimitFree != 100_000 {
		t.Errorf("TokenLimitFree = %d, want 100_000", TokenLimitFree)
	}
	if TokenLimitPro != 1_000_000 {
		t.Errorf("TokenLimitPro = %d, want 1_000_000", TokenLimitPro)
	}
	if TokenLimitUltra != 10_000_000 {
		t.Errorf("TokenLimitUltra = %d, want 10_000_000", TokenLimitUltra)
	}
}

func TestMaxChatbotsConstants(t *testing.T) {
	// Verify chatbot limits are positive and in ascending order
	if MaxChatbotsFree <= 0 {
		t.Error("MaxChatbotsFree should be positive")
	}
	if MaxChatbotsPro <= MaxChatbotsFree {
		t.Error("MaxChatbotsPro should be greater than MaxChatbotsFree")
	}
	if MaxChatbotsUltra <= MaxChatbotsPro {
		t.Error("MaxChatbotsUltra should be greater than MaxChatbotsPro")
	}

	// Verify expected values (these are reference values)
	if MaxChatbotsFree != 1 {
		t.Errorf("MaxChatbotsFree = %d, want 1", MaxChatbotsFree)
	}
	if MaxChatbotsPro != 5 {
		t.Errorf("MaxChatbotsPro = %d, want 5", MaxChatbotsPro)
	}
	if MaxChatbotsUltra != 100 {
		t.Errorf("MaxChatbotsUltra = %d, want 100", MaxChatbotsUltra)
	}
}
