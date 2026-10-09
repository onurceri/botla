package services

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
)

func TestNormalizeContactEmail(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "plain", in: "user@example.com", want: "user@example.com"},
		{name: "trims whitespace", in: "  user@example.com\n", want: "user@example.com"},
		{name: "plus and subdomain", in: "a.b+tag@mail.example.co.uk", want: "a.b+tag@mail.example.co.uk"},
		{name: "empty", in: "", wantErr: true},
		{name: "whitespace only", in: "   ", wantErr: true},
		{name: "no at", in: "userexample.com", wantErr: true},
		{name: "no domain dot", in: "user@localhost", wantErr: true},
		{name: "trailing dot domain", in: "user@example.", wantErr: true},
		{name: "display name", in: "Bob <bob@example.com>", wantErr: true},
		{name: "angle brackets only", in: "<bob@example.com>", wantErr: true},
		{name: "two addresses", in: "a@example.com, b@example.com", wantErr: true},
		{name: "header injection", in: "a@example.com\r\nBcc: x@example.com", wantErr: true},
		{name: "too long", in: strings.Repeat("a", 250) + "@example.com", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeContactEmail(tt.in)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidContactEmail) {
					t.Fatalf("expected ErrInvalidContactEmail, got %q, %v", got, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSubmitContactEmail(t *testing.T) {
	const botID = "bot-1"
	const reqID = "req-1"

	existing := func(chatbotID, status string) func(context.Context, string) (*models.HandoffRequest, error) {
		return func(_ context.Context, id string) (*models.HandoffRequest, error) {
			return &models.HandoffRequest{ID: id, ChatbotID: chatbotID, Status: status}, nil
		}
	}

	tests := []struct {
		name       string
		email      string
		getFn      func(context.Context, string) (*models.HandoffRequest, error)
		setErr     error
		wantErr    error
		wantSet    bool
		wantStored string
	}{
		{name: "persists trimmed email", email: " user@example.com ", getFn: existing(botID, models.HandoffStatusPending), wantSet: true, wantStored: "user@example.com"},
		{name: "assigned request accepts email", email: "user@example.com", getFn: existing(botID, models.HandoffStatusAssigned), wantSet: true, wantStored: "user@example.com"},
		{name: "invalid email", email: "nope", getFn: existing(botID, models.HandoffStatusPending), wantErr: ErrInvalidContactEmail},
		{name: "missing request", email: "user@example.com", getFn: func(context.Context, string) (*models.HandoffRequest, error) { return nil, nil }, wantErr: ErrHandoffNotFound},
		{name: "request of another bot", email: "user@example.com", getFn: existing("other-bot", models.HandoffStatusPending), wantErr: ErrHandoffNotFound},
		{name: "resolved request", email: "user@example.com", getFn: existing(botID, models.HandoffStatusResolved), wantErr: ErrHandoffClosed},
		{name: "row vanished", email: "user@example.com", getFn: existing(botID, models.HandoffStatusPending), setErr: sql.ErrNoRows, wantErr: ErrHandoffNotFound, wantSet: true, wantStored: "user@example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repository.NewMockHandoffRepo()
			repo.GetHandoffRequestByIDFunc = tt.getFn
			repo.SetHandoffContactEmailFunc = func(context.Context, string, string, string) error { return tt.setErr }
			svc := NewHandoffService(repo, nil, nil, nil)

			err := svc.SubmitContactEmail(context.Background(), botID, reqID, tt.email)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			calls := repo.Calls.SetHandoffContactEmail
			if !tt.wantSet {
				if len(calls) != 0 {
					t.Fatalf("expected no write, got %+v", calls)
				}
				return
			}
			if len(calls) != 1 {
				t.Fatalf("expected 1 write, got %d", len(calls))
			}
			if calls[0].ID != reqID || calls[0].ChatbotID != botID || calls[0].Email != tt.wantStored {
				t.Fatalf("unexpected write: %+v", calls[0])
			}
		})
	}
}

func TestSubmitContactEmail_LookupError(t *testing.T) {
	repo := repository.NewMockHandoffRepo()
	boom := errors.New("db down")
	repo.GetHandoffRequestByIDFunc = func(context.Context, string) (*models.HandoffRequest, error) { return nil, boom }
	svc := NewHandoffService(repo, nil, nil, nil)

	err := svc.SubmitContactEmail(context.Background(), "bot", "req", "user@example.com")
	if !errors.Is(err, boom) {
		t.Fatalf("expected wrapped lookup error, got %v", err)
	}
}
