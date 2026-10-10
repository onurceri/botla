package policy

import "testing"

func TestModel_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		model Model
		valid bool
	}{
		{"valid gpt-6-luna", ModelGPT6Luna, true},
		{"valid gpt-4o-mini", ModelGPT4oMini, true},
		{"valid gpt-4o", ModelGPT4o, true},
		{"valid gpt-5", ModelGPT5, true},
		{"invalid empty", Model(""), false},
		{"invalid unknown", Model("unknown-model"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.model.IsValid(); got != tt.valid {
				t.Errorf("Model.IsValid() = %v, want %v", got, tt.valid)
			}
		})
	}
}

func TestModel_String(t *testing.T) {
	tests := []struct {
		name  string
		model Model
		want  string
	}{
		{"gpt-6-luna", ModelGPT6Luna, "gpt-6-luna"},
		{"gpt-4o-mini", ModelGPT4oMini, "gpt-4o-mini"},
		{"gpt-4o", ModelGPT4o, "gpt-4o"},
		{"gpt-5", ModelGPT5, "gpt-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.model.String(); got != tt.want {
				t.Errorf("Model.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultChatModel(t *testing.T) {
	model := DefaultChatModel()
	if model != ModelGPT6Luna {
		t.Errorf("DefaultChatModel() = %v, want %v", model, ModelGPT6Luna)
	}

	// Ensure the default is a valid model
	if !model.IsValid() {
		t.Error("DefaultChatModel() returned an invalid model")
	}
}
