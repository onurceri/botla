package models

import (
	"errors"
	"fmt"
	"math"
	"sort"
)

// ErrInvalidPlanLimitUpdate is returned when a partial plan_limits update
// contains an unknown column or a value of the wrong type.
var ErrInvalidPlanLimitUpdate = errors.New("invalid plan limit update")

// PlanLimitFieldKind describes the SQL type of an updatable plan_limits column.
type PlanLimitFieldKind int

const (
	// PlanLimitInt is a non-negative Postgres integer column.
	PlanLimitInt PlanLimitFieldKind = iota + 1
	// PlanLimitBool is a boolean column.
	PlanLimitBool
	// PlanLimitString is a text column.
	PlanLimitString
	// PlanLimitStringArray is a text[] column.
	PlanLimitStringArray
)

// updatablePlanLimitFields is the explicit allowlist of plan_limits columns that
// may be changed through a partial update. Keys are used verbatim as SQL column
// names, so only constants may ever be added here.
var updatablePlanLimitFields = map[string]PlanLimitFieldKind{
	"max_chatbots":                         PlanLimitInt,
	"max_monthly_ingestions":               PlanLimitInt,
	"max_monthly_embedding_tokens":         PlanLimitInt,
	"min_readd_cooldown_minutes":           PlanLimitInt,
	"scraping_dynamic_enabled":             PlanLimitBool,
	"scraping_max_urls_per_bot":            PlanLimitInt,
	"scraping_max_pages_per_crawl":         PlanLimitInt,
	"files_max_size_mb":                    PlanLimitInt,
	"files_max_files_per_bot":              PlanLimitInt,
	"files_max_files_total":                PlanLimitInt,
	"files_total_storage_mb":               PlanLimitInt,
	"files_max_text_length":                PlanLimitInt,
	"chat_default_model":                   PlanLimitString,
	"chat_allowed_models":                  PlanLimitStringArray,
	"chat_max_monthly_tokens":              PlanLimitInt,
	"chat_rag_top_k":                       PlanLimitInt,
	"chat_rag_max_context_tokens":          PlanLimitInt,
	"chat_max_suggested_questions":         PlanLimitInt,
	"chat_max_manual_questions":            PlanLimitInt,
	"chat_min_response_token_limit":        PlanLimitInt,
	"chat_max_response_token_limit":        PlanLimitInt,
	"refresh_enabled":                      PlanLimitBool,
	"refresh_max_monthly":                  PlanLimitInt,
	"security_secure_embed_enabled":        PlanLimitBool,
	"guardrails_can_customize_thresholds":  PlanLimitBool,
	"guardrails_can_use_smart_fallback":    PlanLimitBool,
	"guardrails_can_use_escalate_fallback": PlanLimitBool,
	"guardrails_can_manage_topics":         PlanLimitBool,
	"guardrails_can_customize_messages":    PlanLimitBool,
	"branding_can_hide_branding":           PlanLimitBool,
	"branding_can_custom_branding":         PlanLimitBool,
	"rate_limits_requests_per_minute":      PlanLimitInt,
	"rate_limits_window_seconds":           PlanLimitInt,
	"rate_limits_chat_rpm":                 PlanLimitInt,
	"rate_limits_chat_window":              PlanLimitInt,
	"rate_limits_sources_rpm":              PlanLimitInt,
	"rate_limits_sources_window":           PlanLimitInt,
}

// PlanLimitFieldKindOf reports the kind of an updatable plan_limits column and
// whether the column is on the allowlist.
func PlanLimitFieldKindOf(field string) (PlanLimitFieldKind, bool) {
	kind, ok := updatablePlanLimitFields[field]
	return kind, ok
}

// NormalizePlanLimitUpdates validates a partial plan_limits update decoded from
// JSON and returns a new map with values converted to their Go column types
// (int, bool, string, []string). Unknown fields and mistyped values produce an
// error wrapping ErrInvalidPlanLimitUpdate. Fields are checked in sorted order
// so the reported error is deterministic.
func NormalizePlanLimitUpdates(updates map[string]any) (map[string]any, error) {
	if len(updates) == 0 {
		return nil, fmt.Errorf("%w: no fields to update", ErrInvalidPlanLimitUpdate)
	}

	fields := make([]string, 0, len(updates))
	for field := range updates {
		fields = append(fields, field)
	}
	sort.Strings(fields)

	out := make(map[string]any, len(updates))
	for _, field := range fields {
		kind, ok := updatablePlanLimitFields[field]
		if !ok {
			return nil, fmt.Errorf("%w: unknown field %q", ErrInvalidPlanLimitUpdate, field)
		}
		v, err := normalizePlanLimitValue(kind, updates[field])
		if err != nil {
			return nil, fmt.Errorf("%w: field %q %s", ErrInvalidPlanLimitUpdate, field, err.Error())
		}
		out[field] = v
	}
	return out, nil
}

// planLimitTypeError is a plain error describing a type mismatch.
type planLimitTypeError string

func (e planLimitTypeError) Error() string { return string(e) }

func normalizePlanLimitValue(kind PlanLimitFieldKind, value any) (any, error) {
	switch kind {
	case PlanLimitInt:
		return normalizePlanLimitInt(value)
	case PlanLimitBool:
		if b, ok := value.(bool); ok {
			return b, nil
		}
		return nil, planLimitTypeError("must be a boolean")
	case PlanLimitString:
		if s, ok := value.(string); ok {
			return s, nil
		}
		return nil, planLimitTypeError("must be a string")
	case PlanLimitStringArray:
		return normalizePlanLimitStringArray(value)
	default:
		return nil, planLimitTypeError("has an unsupported type")
	}
}

func normalizePlanLimitInt(value any) (int, error) {
	var f float64
	switch n := value.(type) {
	case float64:
		f = n
	case int:
		f = float64(n)
	case int64:
		f = float64(n)
	default:
		return 0, planLimitTypeError("must be an integer")
	}
	if math.IsNaN(f) || math.IsInf(f, 0) || f != math.Trunc(f) {
		return 0, planLimitTypeError("must be an integer")
	}
	if f < 0 || f > math.MaxInt32 {
		return 0, planLimitTypeError(fmt.Sprintf("must be between 0 and %d", math.MaxInt32))
	}
	return int(f), nil
}

func normalizePlanLimitStringArray(value any) ([]string, error) {
	switch arr := value.(type) {
	case []string:
		return append([]string{}, arr...), nil
	case []any:
		out := make([]string, 0, len(arr))
		for _, item := range arr {
			s, ok := item.(string)
			if !ok {
				return nil, planLimitTypeError("must be an array of strings")
			}
			out = append(out, s)
		}
		return out, nil
	default:
		return nil, planLimitTypeError("must be an array of strings")
	}
}
