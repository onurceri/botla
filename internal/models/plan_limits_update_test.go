package models

import (
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizePlanLimitUpdates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      map[string]any
		want    map[string]any
		wantErr string
	}{
		{name: "nil map", in: nil, wantErr: "no fields to update"},
		{name: "empty map", in: map[string]any{}, wantErr: "no fields to update"},
		{name: "integer from JSON float", in: map[string]any{"max_chatbots": float64(10)}, want: map[string]any{"max_chatbots": 10}},
		{name: "integer from Go int", in: map[string]any{"chat_rag_top_k": 5}, want: map[string]any{"chat_rag_top_k": 5}},
		{name: "boolean", in: map[string]any{"refresh_enabled": true}, want: map[string]any{"refresh_enabled": true}},
		{name: "string", in: map[string]any{"chat_default_model": "openai/gpt-4o"}, want: map[string]any{"chat_default_model": "openai/gpt-4o"}},
		{name: "string array from JSON", in: map[string]any{"chat_allowed_models": []any{"a", "b"}}, want: map[string]any{"chat_allowed_models": []string{"a", "b"}}},
		{name: "string array from Go", in: map[string]any{"chat_allowed_models": []string{"a"}}, want: map[string]any{"chat_allowed_models": []string{"a"}}},
		{name: "empty string array", in: map[string]any{"chat_allowed_models": []any{}}, want: map[string]any{"chat_allowed_models": []string{}}},
		{
			name: "multiple fields",
			in:   map[string]any{"max_chatbots": float64(3), "branding_can_hide_branding": false},
			want: map[string]any{"max_chatbots": 3, "branding_can_hide_branding": false},
		},
		{name: "unknown field", in: map[string]any{"not_a_column": 1}, wantErr: `unknown field "not_a_column"`},
		{name: "SQL-like key", in: map[string]any{"max_chatbots = 1, plan_id": float64(1)}, wantErr: "unknown field"},
		{name: "non-updatable column", in: map[string]any{"plan_id": "x"}, wantErr: `unknown field "plan_id"`},
		{name: "timestamp column", in: map[string]any{"updated_at": "2024-01-01"}, wantErr: `unknown field "updated_at"`},
		{name: "unknown among valid", in: map[string]any{"max_chatbots": float64(1), "zzz": 1}, wantErr: `unknown field "zzz"`},
		{name: "int given string", in: map[string]any{"max_chatbots": "10"}, wantErr: "must be an integer"},
		{name: "int given fraction", in: map[string]any{"max_chatbots": 1.5}, wantErr: "must be an integer"},
		{name: "int given NaN", in: map[string]any{"max_chatbots": math.NaN()}, wantErr: "must be an integer"},
		{name: "int given bool", in: map[string]any{"max_chatbots": true}, wantErr: "must be an integer"},
		{name: "int negative", in: map[string]any{"max_chatbots": float64(-1)}, wantErr: "must be between"},
		{name: "int overflow", in: map[string]any{"max_chatbots": float64(math.MaxInt32) + 1}, wantErr: "must be between"},
		{name: "int null", in: map[string]any{"max_chatbots": nil}, wantErr: "must be an integer"},
		{name: "bool given string", in: map[string]any{"refresh_enabled": "true"}, wantErr: "must be a boolean"},
		{name: "string given number", in: map[string]any{"chat_default_model": float64(1)}, wantErr: "must be a string"},
		{name: "array given string", in: map[string]any{"chat_allowed_models": "a"}, wantErr: "must be an array of strings"},
		{name: "array with non-string", in: map[string]any{"chat_allowed_models": []any{"a", float64(1)}}, wantErr: "must be an array of strings"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NormalizePlanLimitUpdates(tt.in)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.True(t, errors.Is(err, ErrInvalidPlanLimitUpdate), "error should wrap ErrInvalidPlanLimitUpdate")
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPlanLimitAllowlistMatchesModel(t *testing.T) {
	t.Parallel()

	// Every allowlisted column must be a known PlanLimits db column, and every
	// PlanLimits limit column must be updatable.
	expected := map[string]PlanLimitFieldKind{}
	for field, kind := range updatablePlanLimitFields {
		expected[field] = kind
	}

	modelColumns := planLimitsDBColumns(t)
	for col := range modelColumns {
		_, ok := PlanLimitFieldKindOf(col)
		assert.True(t, ok, "PlanLimits column %q missing from update allowlist", col)
		delete(expected, col)
	}
	assert.Empty(t, expected, "allowlist contains columns not present on PlanLimits")

	_, ok := PlanLimitFieldKindOf("plan_id")
	assert.False(t, ok)
}

// planLimitsDBColumns returns the db-tagged columns of PlanLimits that hold
// limit values (excluding keys and timestamps).
func planLimitsDBColumns(t *testing.T) map[string]struct{} {
	t.Helper()
	cols := map[string]struct{}{}
	typ := reflect.TypeOf(PlanLimits{})
	for i := 0; i < typ.NumField(); i++ {
		tag := typ.Field(i).Tag.Get("db")
		switch tag {
		case "", "-", "plan_id", "created_at", "updated_at":
			continue
		}
		cols[tag] = struct{}{}
	}
	return cols
}
