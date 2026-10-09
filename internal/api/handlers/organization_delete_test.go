package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/onurceri/botla-app/internal/api"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/services"
	"github.com/onurceri/botla-app/internal/testdb"
	"github.com/onurceri/botla-app/pkg/logger"
	"github.com/onurceri/botla-app/pkg/middleware"
)

func deleteOrgAs(h *OrganizationHandlers, ownerID, orgID string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/organizations/"+orgID, nil)
	ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, ownerID)
	ctx = context.WithValue(ctx, middleware.ContextKeyOrgID, orgID)
	rr := httptest.NewRecorder()
	h.DeleteOrganization(rr, req.WithContext(ctx))
	return rr
}

func TestDeleteOrganization_WithChatbots_Returns409(t *testing.T) {
	db := setupTestDB(t)
	svc := services.NewOrganizationService(db, logger.New("test"))
	h := &OrganizationHandlers{OrgService: svc, UserRepo: repository.NewPostgresUserRepo(db)}

	tests := []struct {
		name string
		// detachOrg leaves the chatbot linked only through its workspace.
		detachOrg bool
	}{
		{name: "chatbot linked to org"},
		{name: "chatbot linked only via workspace", detachOrg: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot := testdb.CreateChatbot(t, db)
			orgID := bot.Organization.ID
			ownerID := bot.User.ID
			if tt.detachOrg {
				if _, err := db.Exec("UPDATE chatbots SET organization_id = NULL WHERE id = $1", bot.Chatbot.ID); err != nil {
					t.Fatalf("detach chatbot: %v", err)
				}
			}
			// A second org so the last-organization guard does not fire first.
			if _, err := svc.CreateOrganization(context.Background(), "Other", fmt.Sprintf("other-%d", time.Now().UnixNano()), ownerID); err != nil {
				t.Fatalf("create second org: %v", err)
			}

			rr := deleteOrgAs(h, ownerID, orgID)
			if rr.Code != http.StatusConflict {
				t.Fatalf("expected 409, got %d: %s", rr.Code, rr.Body.String())
			}
			var resp api.ErrorResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil || resp.Code != api.ErrOrgHasChatbots {
				t.Fatalf("expected code %s, got %+v (%v)", api.ErrOrgHasChatbots, resp, err)
			}
			var exists bool
			if err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM chatbots WHERE id=$1)", bot.Chatbot.ID).Scan(&exists); err != nil || !exists {
				t.Fatalf("chatbot must survive a refused delete (exists=%v, err=%v)", exists, err)
			}

			// Once the chatbot is (soft-)deleted the organization can go.
			if _, err := db.Exec("UPDATE chatbots SET deleted_at = NOW() WHERE id = $1", bot.Chatbot.ID); err != nil {
				t.Fatalf("soft-delete chatbot: %v", err)
			}
			if rr := deleteOrgAs(h, ownerID, orgID); rr.Code != http.StatusNoContent {
				t.Fatalf("expected 204 after removing chatbots, got %d: %s", rr.Code, rr.Body.String())
			}
		})
	}
}

func TestDeleteOrganization_LastOrgCheckedFirst(t *testing.T) {
	db := setupTestDB(t)
	svc := services.NewOrganizationService(db, logger.New("test"))
	bot := testdb.CreateChatbot(t, db)

	err := svc.DeleteOrganization(context.Background(), bot.Organization.ID)
	if !errors.Is(err, services.ErrLastOrganization) {
		t.Fatalf("expected ErrLastOrganization, got %v", err)
	}
}
