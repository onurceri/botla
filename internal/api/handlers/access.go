package handlers

import (
	"context"

	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/services"
)

// checkChatbotAccess verifies if a user has access to a chatbot either by ownership or workspace membership
func checkChatbotAccess(ctx context.Context, c *models.Chatbot, userID string, workspaceService *services.WorkspaceService, orgService *services.OrganizationService) (bool, error) {
	if c.UserID == userID {
		return true, nil
	}

	if c.WorkspaceID != nil && *c.WorkspaceID != "" && workspaceService != nil && orgService != nil {
		ws, err := workspaceService.GetWorkspace(ctx, *c.WorkspaceID)
		if err == nil && ws != nil {
			mem, err := orgService.CheckMembership(ctx, userID, ws.OrganizationID)
			if err == nil && mem != nil {
				return true, nil
			}
		}
	}

	if c.OrganizationID != nil && *c.OrganizationID != "" && orgService != nil {
		mem, err := orgService.CheckMembership(ctx, userID, *c.OrganizationID)
		if err == nil && mem != nil {
			return true, nil
		}
	}

	return false, nil
}
