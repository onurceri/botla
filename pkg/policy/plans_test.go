package policy

import "testing"

func TestPlan_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		plan  Plan
		valid bool
	}{
		{"valid free", PlanFree, true},
		{"valid pro", PlanPro, true},
		{"valid ultra", PlanUltra, true},
		{"invalid empty", Plan(""), false},
		{"invalid unknown", Plan("enterprise"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.plan.IsValid(); got != tt.valid {
				t.Errorf("Plan.IsValid() = %v, want %v", got, tt.valid)
			}
		})
	}
}

func TestPlan_String(t *testing.T) {
	tests := []struct {
		name string
		plan Plan
		want string
	}{
		{"free plan", PlanFree, "free"},
		{"pro plan", PlanPro, "pro"},
		{"ultra plan", PlanUltra, "ultra"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.plan.String(); got != tt.want {
				t.Errorf("Plan.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAllPlans(t *testing.T) {
	plans := AllPlans()

	if len(plans) != 3 {
		t.Errorf("AllPlans() returned %d plans, want 3", len(plans))
	}

	// Verify all expected plans are present
	expected := map[Plan]bool{
		PlanFree:  false,
		PlanPro:   false,
		PlanUltra: false,
	}

	for _, p := range plans {
		if _, ok := expected[p]; ok {
			expected[p] = true
		}
	}

	for plan, found := range expected {
		if !found {
			t.Errorf("AllPlans() missing plan: %s", plan)
		}
	}
}
