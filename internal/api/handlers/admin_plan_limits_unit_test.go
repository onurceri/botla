package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminPlanHandlers_UpdateLimits_Validation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		body        string
		repoErr     error
		wantStatus  int
		wantMessage string
		wantUpdates map[string]interface{}
	}{
		{
			name:        "valid partial update",
			body:        `{"max_chatbots": 10, "refresh_enabled": true, "chat_allowed_models": ["a","b"]}`,
			wantStatus:  http.StatusOK,
			wantUpdates: map[string]interface{}{"max_chatbots": 10, "refresh_enabled": true, "chat_allowed_models": []string{"a", "b"}},
		},
		{name: "unknown field", body: `{"evil_column": 1}`, wantStatus: http.StatusBadRequest, wantMessage: `unknown field "evil_column"`},
		{name: "key with SQL fragment", body: `{"max_chatbots = 0 --": 1}`, wantStatus: http.StatusBadRequest, wantMessage: "unknown field"},
		{name: "non-updatable column", body: `{"plan_id": "x"}`, wantStatus: http.StatusBadRequest, wantMessage: `unknown field "plan_id"`},
		{name: "wrong type", body: `{"max_chatbots": "ten"}`, wantStatus: http.StatusBadRequest, wantMessage: "must be an integer"},
		{name: "negative integer", body: `{"max_chatbots": -5}`, wantStatus: http.StatusBadRequest, wantMessage: "must be between"},
		{name: "empty object", body: `{}`, wantStatus: http.StatusBadRequest, wantMessage: "no fields to update"},
		{name: "malformed JSON", body: `{`, wantStatus: http.StatusBadRequest},
		{name: "repository failure", body: `{"max_chatbots": 2}`, repoErr: errors.New("db down"), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := repository.NewMockPlanRepo()
			repo.GetByIDFunc = func(_ context.Context, id string) (*models.Plan, error) {
				return &models.Plan{ID: id, Code: "pro"}, nil
			}
			repo.UpdatePlanLimitsFunc = func(_ context.Context, _ string, _ map[string]interface{}) error {
				return tt.repoErr
			}
			h := NewAdminPlanHandlers(services.NewPlanService(repo, nil), repo)

			req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/plans/p1/limits", strings.NewReader(tt.body))
			req.SetPathValue("id", "p1")
			rr := httptest.NewRecorder()
			h.UpdateLimits(rr, req)

			require.Equal(t, tt.wantStatus, rr.Code, rr.Body.String())

			if tt.wantMessage != "" {
				var resp struct {
					Code    string            `json:"code"`
					Details map[string]string `json:"details"`
				}
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
				assert.Equal(t, "BAD_REQUEST", resp.Code)
				assert.Contains(t, resp.Details["message"], tt.wantMessage)
			}

			if tt.wantStatus == http.StatusBadRequest {
				assert.Empty(t, repo.Calls.UpdatePlanLimits, "repository must not be called for invalid input")
			}
			if tt.wantUpdates != nil {
				require.Len(t, repo.Calls.UpdatePlanLimits, 1)
				assert.Equal(t, "p1", repo.Calls.UpdatePlanLimits[0].PlanID)
				assert.Equal(t, tt.wantUpdates, repo.Calls.UpdatePlanLimits[0].Updates)
			}
		})
	}
}
