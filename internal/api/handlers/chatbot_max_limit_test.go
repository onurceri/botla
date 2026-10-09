package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onurceri/botla-app/internal/api"
	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/pkg/middleware"
)

func newMaxChatbotsHandler(limits *models.PlanLimits, existing int) (*ChatbotHandlers, *repository.MockChatbotRepo) {
	plans := repository.NewMockPlanRepo()
	plans.GetByUserIDFunc = func(context.Context, string) (*models.Plan, error) {
		return &models.Plan{Code: "free", Limits: limits}, nil
	}
	bots := repository.NewMockChatbotRepo()
	bots.CountByUserIDFunc = func(context.Context, string) (int, error) { return existing, nil }
	bots.CreateFunc = func(context.Context, *models.Chatbot) (string, error) { return testIsPublicBotID, nil }
	bots.GetByIDFunc = func(_ context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, Name: "New Bot"}, nil
	}
	return &ChatbotHandlers{PlanRepo: plans, ChatbotRepo: bots}, bots
}

func postCreateChatbot(h *ChatbotHandlers) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots", bytes.NewReader([]byte(`{"name":"New Bot"}`)))
	req = req.WithContext(context.WithValue(req.Context(), middleware.ContextKeyUserID, testHandoffOwnerID))
	rr := httptest.NewRecorder()
	h.ListOrCreate(rr, req)
	return rr
}

func TestCreateChatbot_MaxChatbotsExceeded(t *testing.T) {
	for _, existing := range []int{2, 3} {
		h, bots := newMaxChatbotsHandler(&models.PlanLimits{MaxChatbots: 2}, existing)
		rr := postCreateChatbot(h)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("existing=%d: expected 403, got %d: %s", existing, rr.Code, rr.Body.String())
		}
		var resp api.ErrorResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil || resp.Code != api.ErrMaxChatbotsExceeded {
			t.Fatalf("expected %s, got %+v (%v)", api.ErrMaxChatbotsExceeded, resp, err)
		}
		if len(bots.Calls.Create) != 0 {
			t.Fatal("chatbot must not be created over the limit")
		}
	}
}

func TestCreateChatbot_UnderLimitOrUnlimited(t *testing.T) {
	tests := []struct {
		name     string
		limits   *models.PlanLimits
		existing int
	}{
		{name: "below limit", limits: &models.PlanLimits{MaxChatbots: 2}, existing: 1},
		{name: "zero means unlimited", limits: &models.PlanLimits{MaxChatbots: 0}, existing: 50},
		{name: "plan without limits row", limits: nil, existing: 50},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, bots := newMaxChatbotsHandler(tt.limits, tt.existing)
			rr := postCreateChatbot(h)
			if rr.Code != http.StatusCreated {
				t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
			}
			if len(bots.Calls.Create) != 1 {
				t.Fatalf("expected chatbot to be created")
			}
		})
	}
}
