package langconfig

type LanguageConfig struct {
	Code            string
	Name            string
	Abbreviations   []string // For sentence splitting (e.g., "Dr.", "Mr.")
	TokenMultiplier float64  // For token estimation
	OCRLanguage     string   // Tesseract language code (e.g., "tur", "eng")
	TokenizerData       string            // Path to the JSON training data
	ResponseTemplates   ResponseTemplates // Localized response strings
}

type ResponseTemplates struct {
	NoInfoFound                 string
	DefaultSystemPrompt         string
	ErrorMessage                string
	TopicExtractionSystemPrompt string
	TopicExtractionUserPrompt   string
}

var Configs = map[string]LanguageConfig{
	"tr": {
		Code: "tr",
		Name: "Turkish",
		Abbreviations: []string{
			"Dr.", "Prof.", "vb.", "Av.", "Ecz.", "Doç.", "Yrd.", "Cad.", "Sok.", "Mah.",
		},
		TokenMultiplier: 1.3,
		OCRLanguage:     "tur",
		TokenizerData:   "data/sentences/turkish.json",
		ResponseTemplates: ResponseTemplates{
			NoInfoFound:         "Yeterli bilgi bulamadım.",
			DefaultSystemPrompt: "Her zaman Türkçe yanıt ver ve sadece verilen bağlamı kullan.",
			ErrorMessage:        "Şu an bir hata oluştu, lütfen tekrar deneyin.",
			TopicExtractionSystemPrompt: TR_TopicExtractionSystemPrompt,
			TopicExtractionUserPrompt:   TR_TopicExtractionUserPrompt,
		},
	},
	"en": {
		Code: "en",
		Name: "English",
		Abbreviations: []string{
			"Mr.", "Mrs.", "Ms.", "Dr.", "Prof.", "Inc.", "Ltd.", "Jr.", "Sr.", "St.",
		},
		TokenMultiplier: 1.0,
		OCRLanguage:     "eng",
		TokenizerData:   "data/sentences/english.json",
		ResponseTemplates: ResponseTemplates{
			NoInfoFound:                 "I could not find enough information.",
			DefaultSystemPrompt:         "Always answer in English and use only the provided context.",
			ErrorMessage:                "An error occurred, please try again later.",
			TopicExtractionSystemPrompt: EN_TopicExtractionSystemPrompt,
			TopicExtractionUserPrompt:   EN_TopicExtractionUserPrompt,
		},
	},
}

// Get returns the configuration for the given language code.
// Defaults to "tr" (Turkish) if the code is not found or empty.
func Get(langCode string) LanguageConfig {
	if config, ok := Configs[langCode]; ok {
		return config
	}
	return Configs["tr"]
}

// IsSupported checks if a language code is supported.
func IsSupported(langCode string) bool {
	_, ok := Configs[langCode]
	return ok
}
