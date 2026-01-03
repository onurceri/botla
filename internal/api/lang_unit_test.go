package api

import (
	"net/http/httptest"
	"testing"
)

func TestBaseLang(t *testing.T) {
	t.Run("returns_tr_for_empty_string", func(t *testing.T) {
		if got := BaseLang(""); got != "tr" {
			t.Errorf("BaseLang('') = %q, want %q", got, "tr")
		}
	})

	t.Run("returns_tr_for_whitespace_only", func(t *testing.T) {
		if got := BaseLang("   "); got != "tr" {
			t.Errorf("BaseLang('   ') = %q, want %q", got, "tr")
		}
	})

	t.Run("returns_language_code_from_simple_code", func(t *testing.T) {
		if got := BaseLang("en"); got != "en" {
			t.Errorf("BaseLang('en') = %q, want %q", got, "en")
		}
		if got := BaseLang("tr"); got != "tr" {
			t.Errorf("BaseLang('tr') = %q, want %q", got, "tr")
		}
		if got := BaseLang("de"); got != "de" {
			t.Errorf("BaseLang('de') = %q, want %q", got, "de")
		}
	})

	t.Run("extracts_base_from_locale_with_region", func(t *testing.T) {
		if got := BaseLang("en-US"); got != "en" {
			t.Errorf("BaseLang('en-US') = %q, want %q", got, "en")
		}
		if got := BaseLang("tr-TR"); got != "tr" {
			t.Errorf("BaseLang('tr-TR') = %q, want %q", got, "tr")
		}
		if got := BaseLang("zh-CN"); got != "zh" {
			t.Errorf("BaseLang('zh-CN') = %q, want %q", got, "zh")
		}
	})

	t.Run("handles_lowercase_locale", func(t *testing.T) {
		if got := BaseLang("en-us"); got != "en" {
			t.Errorf("BaseLang('en-us') = %q, want %q", got, "en")
		}
	})

	t.Run("handles_mixed_case_locale", func(t *testing.T) {
		if got := BaseLang("EN-us"); got != "EN" {
			t.Errorf("BaseLang('EN-us') = %q, want %q", got, "EN")
		}
	})

	t.Run("handles_code_with_extra_whitespace", func(t *testing.T) {
		if got := BaseLang("  en  "); got != "en" {
			t.Errorf("BaseLang('  en  ') = %q, want %q", got, "en")
		}
	})

	t.Run("returns_input_for_hyphen_only", func(t *testing.T) {
		if got := BaseLang("-"); got != "-" {
			t.Errorf("BaseLang('-') = %q, want %q", got, "-")
		}
	})

	t.Run("returns_input_for_region_without_language", func(t *testing.T) {
		if got := BaseLang("-US"); got != "-US" {
			t.Errorf("BaseLang('-US') = %q, want %q", got, "-US")
		}
	})
}

func TestLangFromRequest(t *testing.T) {
	t.Run("returns_default_tr_when_no_lang_params", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		if got := LangFromRequest(req); got != "tr" {
			t.Errorf("LangFromRequest() = %q, want %q", got, "tr")
		}
	})

	t.Run("uses_query_param_lang", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/?lang=en", nil)
		if got := LangFromRequest(req); got != "en" {
			t.Errorf("LangFromRequest(?lang=en) = %q, want %q", got, "en")
		}
	})

	t.Run("uses_query_param_lang_with_region", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/?lang=en-US", nil)
		if got := LangFromRequest(req); got != "en" {
			t.Errorf("LangFromRequest(?lang=en-US) = %q, want %q", got, "en")
		}
	})

	t.Run("prefers_query_param_over_header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/?lang=de", nil)
		req.Header.Set("Accept-Language", "fr")
		if got := LangFromRequest(req); got != "de" {
			t.Errorf("LangFromRequest with both query and header = %q, want %q", got, "de")
		}
	})

	t.Run("uses_accept_language_header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Language", "fr")
		if got := LangFromRequest(req); got != "fr" {
			t.Errorf("LangFromRequest with Accept-Language fr = %q, want %q", got, "fr")
		}
	})

	t.Run("extracts_base_from_accept_language_header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
		if got := LangFromRequest(req); got != "zh" {
			t.Errorf("LangFromRequest with Accept-Language zh-CN = %q, want %q", got, "zh")
		}
	})

	t.Run("ignores_empty_accept_language_header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Language", "")
		if got := LangFromRequest(req); got != "tr" {
			t.Errorf("LangFromRequest with empty Accept-Language = %q, want %q", got, "tr")
		}
	})

	t.Run("handles_multiple_accept_language_values", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Language", "es, ja, de")
		if got := LangFromRequest(req); got != "es" {
			t.Errorf("LangFromRequest with multiple Accept-Language = %q, want %q", got, "es")
		}
	})

	t.Run("handles_whitespace_in_accept_language", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Language", "  en  ")
		if got := LangFromRequest(req); got != "en" {
			t.Errorf("LangFromRequest with whitespace in Accept-Language = %q, want %q", got, "en")
		}
	})
}

func TestConfigFromBase(t *testing.T) {
	t.Run("returns_config_for_en", func(t *testing.T) {
		cfg := ConfigFromBase("en")
		if cfg.Code != "en" {
			t.Errorf("ConfigFromBase('en').Code = %q, want %q", cfg.Code, "en")
		}
		if cfg.Name != "English" {
			t.Errorf("ConfigFromBase('en').Name = %q, want %q", cfg.Name, "English")
		}
	})

	t.Run("returns_config_for_tr", func(t *testing.T) {
		cfg := ConfigFromBase("tr")
		if cfg.Code != "tr" {
			t.Errorf("ConfigFromBase('tr').Code = %q, want %q", cfg.Code, "tr")
		}
		if cfg.Name != "Turkish" {
			t.Errorf("ConfigFromBase('tr').Name = %q, want %q", cfg.Name, "Turkish")
		}
	})

	t.Run("falls_back_to_tr_for_unknown_language", func(t *testing.T) {
		cfg := ConfigFromBase("xx")
		if cfg.Code != "tr" {
			t.Errorf("ConfigFromBase('xx').Code = %q, want %q", cfg.Code, "tr")
		}
	})

	t.Run("falls_back_to_tr_for_empty_string", func(t *testing.T) {
		cfg := ConfigFromBase("")
		if cfg.Code != "tr" {
			t.Errorf("ConfigFromBase('').Code = %q, want %q", cfg.Code, "tr")
		}
	})
}
