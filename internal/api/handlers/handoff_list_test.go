package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onurceri/botla-app/internal/api"
	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/services"
	"github.com/onurceri/botla-app/pkg/middleware"
)

const testHandoffOwnerID = "44444444-4444-4444-4444-444444444444"

// newOwnedHandoffHandler returns handlers whose chatbot is owned by testHandoffOwnerID,
// so getChatbotContext grants access without workspace/org services.
func newOwnedHandoffHandler(handoffRepo *repository.MockHandoffRepo) *HandoffHandlers {
	bots := repository.NewMockChatbotRepo()
	bots.GetByIDFunc = func(_ context.Context, id string) (*models.Chatbot, error) {
		return &models.Chatbot{ID: id, UserID: testHandoffOwnerID}, nil
	}
	return &HandoffHandlers{
		ChatbotRepo:    bots,
		HandoffRepo:    handoffRepo,
		HandoffService: services.NewHandoffService(handoffRepo, nil, nil, nil),
	}
}

func ownerRequest(method, target string) *http.Request {
	req := httptest.NewRequest(method, target, nil)
	req.SetPathValue("id", testIsPublicBotID)
	ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, testHandoffOwnerID)
	return req.WithContext(ctx)
}

func TestListHandoffRequests_PagingAndFilter(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantFilter models.HandoffListFilter
		wantPage   int
	}{
		{name: "defaults", query: "", wantFilter: models.HandoffListFilter{Limit: services.DefaultHandoffListLimit}, wantPage: 1},
		{name: "page and limit", query: "?page=3&limit=10", wantFilter: models.HandoffListFilter{Limit: 10, Offset: 20}, wantPage: 3},
		{name: "status filter", query: "?status=resolved&limit=5", wantFilter: models.HandoffListFilter{Status: models.HandoffStatusResolved, Limit: 5}, wantPage: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repository.NewMockHandoffRepo()
			repo.ListHandoffRequestsByBotIDFunc = func(context.Context, string, models.HandoffListFilter) ([]*models.HandoffRequest, int, error) {
				return []*models.HandoffRequest{{ID: "r1", ChatbotID: testIsPublicBotID}}, 42, nil
			}
			h := newOwnedHandoffHandler(repo)

			rr := httptest.NewRecorder()
			h.ListHandoffRequests(rr, ownerRequest(http.MethodGet, "/api/v1/chatbots/"+testIsPublicBotID+"/handoff-requests"+tt.query))

			if rr.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
			}
			if len(repo.Calls.ListHandoffRequestsByBotID) != 1 {
				t.Fatalf("expected 1 repo call")
			}
			call := repo.Calls.ListHandoffRequestsByBotID[0]
			if call.ChatbotID != testIsPublicBotID || call.Filter != tt.wantFilter {
				t.Fatalf("unexpected repo call: %+v", call)
			}
			var body services.HandoffRequestPage
			if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if body.Total != 42 || body.Page != tt.wantPage || body.Limit != tt.wantFilter.Limit || len(body.Requests) != 1 {
				t.Fatalf("unexpected body: %+v", body)
			}
		})
	}
}

func TestListHandoffRequests_BadParams(t *testing.T) {
	tests := []struct {
		query    string
		wantCode string
	}{
		{query: "?status=contacted", wantCode: api.ErrInvalidStatus},
		{query: "?page=abc", wantCode: api.ErrCodeBadRequest},
		{query: "?page=0", wantCode: api.ErrCodeBadRequest},
		{query: "?limit=-1", wantCode: api.ErrCodeBadRequest},
		{query: "?limit=101", wantCode: api.ErrCodeBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			repo := repository.NewMockHandoffRepo()
			h := newOwnedHandoffHandler(repo)
			rr := httptest.NewRecorder()
			h.ListHandoffRequests(rr, ownerRequest(http.MethodGet, "/api/v1/chatbots/"+testIsPublicBotID+"/handoff-requests"+tt.query))

			if rr.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", rr.Code)
			}
			var resp api.ErrorResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil || resp.Code != tt.wantCode {
				t.Fatalf("expected code %s, got %+v (%v)", tt.wantCode, resp, err)
			}
			if len(repo.Calls.ListHandoffRequestsByBotID) != 0 {
				t.Fatal("repo must not be called for bad params")
			}
		})
	}
}

func TestListHandoffRequests_RepoError(t *testing.T) {
	repo := repository.NewMockHandoffRepo()
	repo.ListHandoffRequestsByBotIDFunc = func(context.Context, string, models.HandoffListFilter) ([]*models.HandoffRequest, int, error) {
		return nil, 0, errors.New("db down")
	}
	h := newOwnedHandoffHandler(repo)
	rr := httptest.NewRecorder()
	h.ListHandoffRequests(rr, ownerRequest(http.MethodGet, "/api/v1/chatbots/"+testIsPublicBotID+"/handoff-requests"))
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}
