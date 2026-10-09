package services

import (
	"context"
	"errors"
	"testing"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
)

func TestListHandoffRequests_Validation(t *testing.T) {
	tests := []struct {
		name    string
		status  string
		page    int
		limit   int
		wantErr func(error) bool
	}{
		{name: "unknown status", status: "contacted", page: 1, limit: 10, wantErr: func(err error) bool {
			var invalid InvalidHandoffStatusError
			return errors.As(err, &invalid) && invalid.Status == "contacted"
		}},
		{name: "page zero", page: 0, limit: 10, wantErr: func(err error) bool { return errors.Is(err, ErrInvalidRequest) }},
		{name: "negative page", page: -1, limit: 10, wantErr: func(err error) bool { return errors.Is(err, ErrInvalidRequest) }},
		{name: "limit zero", page: 1, limit: 0, wantErr: func(err error) bool { return errors.Is(err, ErrInvalidRequest) }},
		{name: "limit above max", page: 1, limit: MaxHandoffListLimit + 1, wantErr: func(err error) bool { return errors.Is(err, ErrInvalidRequest) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repository.NewMockHandoffRepo()
			svc := NewHandoffService(repo, nil, nil, nil)
			_, err := svc.ListHandoffRequests(context.Background(), "bot-1", tt.status, tt.page, tt.limit)
			if err == nil || !tt.wantErr(err) {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(repo.Calls.ListHandoffRequestsByBotID) != 0 {
				t.Fatalf("repo must not be called on invalid input")
			}
		})
	}
}

func TestListHandoffRequests_PassesFilterAndPaging(t *testing.T) {
	repo := repository.NewMockHandoffRepo()
	repo.ListHandoffRequestsByBotIDFunc = func(context.Context, string, models.HandoffListFilter) ([]*models.HandoffRequest, int, error) {
		return []*models.HandoffRequest{{ID: "r1"}, {ID: "r2"}}, 7, nil
	}
	svc := NewHandoffService(repo, nil, nil, nil)

	page, err := svc.ListHandoffRequests(context.Background(), "bot-1", models.HandoffStatusAssigned, 3, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.Calls.ListHandoffRequestsByBotID) != 1 {
		t.Fatalf("expected 1 repo call")
	}
	call := repo.Calls.ListHandoffRequestsByBotID[0]
	want := models.HandoffListFilter{Status: models.HandoffStatusAssigned, Limit: 2, Offset: 4}
	if call.ChatbotID != "bot-1" || call.Filter != want {
		t.Fatalf("unexpected repo call: %+v", call)
	}
	if page.Total != 7 || page.Page != 3 || page.Limit != 2 || len(page.Requests) != 2 {
		t.Fatalf("unexpected page: %+v", page)
	}
}

func TestListHandoffRequests_EmptyIsNonNil(t *testing.T) {
	svc := NewHandoffService(repository.NewMockHandoffRepo(), nil, nil, nil)
	page, err := svc.ListHandoffRequests(context.Background(), "bot-1", "", 1, DefaultHandoffListLimit)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Requests == nil {
		t.Fatal("expected empty slice so JSON encodes [] not null")
	}
}

func TestListHandoffRequests_RepoError(t *testing.T) {
	repo := repository.NewMockHandoffRepo()
	boom := errors.New("db down")
	repo.ListHandoffRequestsByBotIDFunc = func(context.Context, string, models.HandoffListFilter) ([]*models.HandoffRequest, int, error) {
		return nil, 0, boom
	}
	svc := NewHandoffService(repo, nil, nil, nil)
	if _, err := svc.ListHandoffRequests(context.Background(), "bot-1", "", 1, 10); !errors.Is(err, boom) {
		t.Fatalf("expected wrapped repo error, got %v", err)
	}
}

func TestIsValidHandoffStatus(t *testing.T) {
	for _, s := range []string{models.HandoffStatusPending, models.HandoffStatusAssigned, models.HandoffStatusResolved} {
		if !models.IsValidHandoffStatus(s) {
			t.Fatalf("%q should be valid", s)
		}
	}
	for _, s := range []string{"", "contacted", "PENDING"} {
		if models.IsValidHandoffStatus(s) {
			t.Fatalf("%q should be invalid", s)
		}
	}
}
