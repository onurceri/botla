package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func TestChatbot_DomainUpdates(t *testing.T) {
	// Setup
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)
	token := authToken(t, te.Server.URL, "domain_test@example.com")

	// Create Chatbot
	create := map[string]any{"name": "Domain Bot", "model": "gpt-4o-mini"}
	cb, _ := json.Marshal(create)
	req, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cb))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	res, _ := http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("setup create failed: %d", res.StatusCode)
	}
	var created struct {
		ID string `json:"id"`
	}
	json.NewDecoder(res.Body).Decode(&created)
	res.Body.Close()

	// 1. Update Basic Info
	basicInfo := map[string]any{
		"name":               "Updated Name",
		"description":        "Updated Desc",
		"custom_instruction": "Be helpful",
		"language":           "en-US",
	}
	bi, _ := json.Marshal(basicInfo)
	req1, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+created.ID+"/basic-info", bytes.NewReader(bi))
	req1.Header.Set("Authorization", "Bearer "+token)
	req1.Header.Set("Content-Type", "application/json")
	res1, _ := http.DefaultClient.Do(req1)
	if res1.StatusCode != http.StatusOK {
		var buf bytes.Buffer
		buf.ReadFrom(res1.Body)
		t.Fatalf("update basic info failed: %d, body: %s", res1.StatusCode, buf.String())
	}

	// 2. Update Appearance
	themeColor := "#FF0000"
	position := "bottom-left"
	appearance := map[string]any{
		"theme_color": themeColor,
		"position":    position,
	}
	ap, _ := json.Marshal(appearance)
	req2, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+created.ID+"/appearance", bytes.NewReader(ap))
	req2.Header.Set("Authorization", "Bearer "+token)
	req2.Header.Set("Content-Type", "application/json")
	res2, _ := http.DefaultClient.Do(req2)
	if res2.StatusCode != http.StatusOK {
		t.Fatalf("update appearance failed: %d", res2.StatusCode)
	}

	// 3. Update Model Settings
	modelSettings := map[string]any{
		"temperature": 0.5,
		"max_tokens":  100,
	}
	ms, _ := json.Marshal(modelSettings)
	req3, _ := http.NewRequest(http.MethodPut, te.Server.URL+"/api/v1/chatbots/"+created.ID+"/model", bytes.NewReader(ms))
	req3.Header.Set("Authorization", "Bearer "+token)
	req3.Header.Set("Content-Type", "application/json")
	res3, _ := http.DefaultClient.Do(req3)
	if res3.StatusCode != http.StatusOK {
		t.Fatalf("update model settings failed: %d", res3.StatusCode)
	}

	// Verify all changes via GET
	reqG, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/chatbots/"+created.ID, nil)
	reqG.Header.Set("Authorization", "Bearer "+token)
	resG, _ := http.DefaultClient.Do(reqG)
	if resG.StatusCode != http.StatusOK {
		t.Fatalf("get chatbot failed: %d", resG.StatusCode)
	}

	var updated struct {
		Name              string  `json:"name"`
		Description       *string `json:"description"`
		CustomInstruction string  `json:"custom_instruction"`
		Language          string  `json:"language"`
		ThemeColor        *string `json:"theme_color"`
		Temperature       float32 `json:"temperature"`
		MaxTokens         int     `json:"max_tokens"`
	}
	json.NewDecoder(resG.Body).Decode(&updated)
	resG.Body.Close()

	if updated.Name != "Updated Name" {
		t.Errorf("name mismatch: got %q", updated.Name)
	}
	if updated.Description == nil || *updated.Description != "Updated Desc" {
		t.Errorf("description mismatch")
	}
	if updated.CustomInstruction != "Be helpful" {
		t.Errorf("custom instruction mismatch: got %q", updated.CustomInstruction)
	}
	// Note: language normalizing might change en-US to en-US or similar, check if normalized
	// Assuming normalizeLocale just passes valid ones.
	if updated.Language != "en-US" {
		t.Errorf("language mismatch: got %q", updated.Language)
	}
	if updated.ThemeColor == nil || *updated.ThemeColor != "#FF0000" {
		t.Errorf("theme color mismatch")
	}
	if updated.Temperature != 0.5 {
		t.Errorf("temperature mismatch: got %f", updated.Temperature)
	}
	if updated.MaxTokens != 100 {
		t.Errorf("max tokens mismatch: got %d", updated.MaxTokens)
	}
}
