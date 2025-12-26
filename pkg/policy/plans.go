package policy

// Plan represents a typed plan code identifier.
// This eliminates magic strings like "free", "pro", "ultra" scattered across the codebase.
type Plan string

// Plan constants define all available plan types in the system.
const (
	PlanFree  Plan = "free"
	PlanPro   Plan = "pro"
	PlanUltra Plan = "ultra"
)

// String returns the string representation of the plan.
func (p Plan) String() string {
	return string(p)
}

// IsValid checks if the plan is one of the recognized plan types.
func (p Plan) IsValid() bool {
	switch p {
	case PlanFree, PlanPro, PlanUltra:
		return true
	default:
		return false
	}
}

// AllPlans returns a slice of all valid plan codes.
func AllPlans() []Plan {
	return []Plan{PlanFree, PlanPro, PlanUltra}
}
