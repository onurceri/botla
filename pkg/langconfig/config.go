package langconfig

type LanguageConfig struct {
	Code              string
	Name              string
	Abbreviations     []string          // For sentence splitting (e.g., "Dr.", "Mr.")
	TokenMultiplier   float64           // For token estimation
	OCRLanguage       string            // Tesseract language code (e.g., "tur", "eng")
	TokenizerData     string            // Path to the JSON training data
	ResponseTemplates ResponseTemplates // Localized response strings
}

type ResponseTemplates struct {
    NoInfoFound                 string
    DefaultSystemPrompt         string
    ErrorMessage                string
    TopicExtractionSystemPrompt string
    TopicExtractionUserPrompt   string
    WelcomeMessage              string
    DefaultPersonaPrompt        string
    Errors                      map[string]string
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
            NoInfoFound:                 "Yeterli bilgi bulamadım.",
            DefaultSystemPrompt:         "Her zaman Türkçe yanıt ver ve sadece verilen bağlamı kullan.",
            ErrorMessage:                "Şu an bir hata oluştu, lütfen tekrar deneyin.",
            TopicExtractionSystemPrompt: TR_TopicExtractionSystemPrompt,
            TopicExtractionUserPrompt:   TR_TopicExtractionUserPrompt,
            WelcomeMessage:              "Merhaba! Size nasıl yardımcı olabilirim?",
            DefaultPersonaPrompt:        "Sen yararlı, kibar ve bilgili bir yapay zeka asistanısın.",
            Errors: map[string]string{
                "ERR_MONTHLY_TOKENS_EXCEEDED":      "Aylık token sınırı aşıldı",
                "ERR_NAME_AND_ACTION_TYPE_REQUIRED": "'name' ve 'action_type' alanları zorunludur",
                "ERR_PDF_LIMIT_REACHED":            "Sınır aşıldı: Chatbot başına en fazla PDF dosyası",
                "ERR_FILE_TOO_LARGE":               "Dosya çok büyük",
                "ERR_READD_COOLDOWN_ACTIVE":        "Yeniden ekleme bekleme süresi aktif",
                "ERR_DUPLICATE_URL":                "Yinelenen URL",
                "ERR_ONLY_URL_REFRESH":             "Yalnızca URL kaynakları yenilenebilir",
                "ERR_SOURCE_ALREADY_PROCESSING":    "Kaynak zaten işleniyor",
                "ERR_PLAN_REFRESH_UNAVAILABLE":     "Planınızda yenileme özelliği mevcut değil",
                "ERR_MONTHLY_REFRESH_EXCEEDED":     "Aylık yenileme sınırı aşıldı",
                "ERR_REFRESH_COOLDOWN_ACTIVE":      "Yenileme bekleme süresi aktif",
                "ERR_INVALID_REQUEST_BODY":         "Geçersiz istek gövdesi",
                "ERR_NO_URLS_PROVIDED":             "Herhangi bir URL sağlanmadı",
                "ERR_URL_LIMIT_REACHED":            "Bu chatbot için URL sınırı aşıldı",
                "ERR_MONTHLY_INGESTION_EXCEEDED":   "Aylık içe‑alma sınırı aşıldı",
                "ERR_SITEMAP_PARSE_FAILED":         "Site haritası ayrıştırılamadı",
                "CHAT_TIMEOUT_OR_INCOMPLETE":       "İşlem tamamlanamadı veya çok uzun sürdü.",
                "HANDOFF_NOT_ENABLED":              "Bu chatbot için devretme etkin değil",
                "HANDOFF_CREATE_FAILED":            "Devretme talebi oluşturulamadı",
                "HANDOFF_EMAIL_NOT_CONFIGURED":     "Devretme için e‑posta adresi yapılandırılmamış",
                "HANDOFF_CONVERSATION_LOAD_FAILED": "Konuşma yüklenemedi",
                "HANDOFF_EMAIL_SUBJECT":            "[Botla] Yeni Destek Talebi - %s",
                "HANDOFF_EMAIL_BODY_HEADER":        "Yeni bir destek talebi alındı.\n\n",
                "HANDOFF_EMAIL_LABEL_REQUEST_ID":   "Talep ID",
                "HANDOFF_EMAIL_LABEL_DATE":         "Tarih",
                "HANDOFF_EMAIL_LABEL_USER_NOTE":    "Kullanıcı Notu",
                "HANDOFF_EMAIL_LABEL_USER":         "Kullanıcı",
                "HANDOFF_EMAIL_LABEL_BOT":          "Bot",
                "HANDOFF_EMAIL_BODY_FOOTER":        "Bu e‑posta Botla tarafından otomatik olarak gönderilmiştir.\n",
                "ERR_INVALID_STATUS":               "Geçersiz durum: %s",
            },
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
            WelcomeMessage:              "Hello! How can I help you today?",
            DefaultPersonaPrompt:        "You are a helpful, polite, and knowledgeable AI assistant.",
            Errors: map[string]string{
                "ERR_MONTHLY_TOKENS_EXCEEDED":      "Monthly token limit exceeded",
                "ERR_NAME_AND_ACTION_TYPE_REQUIRED": "name and action_type are required",
                "ERR_PDF_LIMIT_REACHED":            "Limit reached: Max PDF files per chatbot",
                "ERR_FILE_TOO_LARGE":               "File too large",
                "ERR_READD_COOLDOWN_ACTIVE":        "Re-add cooldown active",
                "ERR_DUPLICATE_URL":                "Duplicate URL",
                "ERR_ONLY_URL_REFRESH":             "Only URL sources can be refreshed",
                "ERR_SOURCE_ALREADY_PROCESSING":    "Source is already being processed",
                "ERR_PLAN_REFRESH_UNAVAILABLE":     "Refresh feature is not available on your plan",
                "ERR_MONTHLY_REFRESH_EXCEEDED":     "Monthly refresh limit exceeded",
                "ERR_REFRESH_COOLDOWN_ACTIVE":      "Refresh cooldown active",
                "ERR_INVALID_REQUEST_BODY":         "Invalid request body",
                "ERR_NO_URLS_PROVIDED":             "No URLs provided",
                "ERR_URL_LIMIT_REACHED":            "URL limit reached for this chatbot",
                "ERR_MONTHLY_INGESTION_EXCEEDED":   "Monthly ingestion limit exceeded",
                "ERR_SITEMAP_PARSE_FAILED":         "Failed to parse sitemap",
                "CHAT_TIMEOUT_OR_INCOMPLETE":       "Operation could not be completed or took too long.",
                "HANDOFF_NOT_ENABLED":              "handoff is not enabled for this chatbot",
                "HANDOFF_CREATE_FAILED":            "failed to create handoff request",
                "HANDOFF_EMAIL_NOT_CONFIGURED":     "email address not configured for handoff",
                "HANDOFF_CONVERSATION_LOAD_FAILED": "failed to load conversation",
                "HANDOFF_EMAIL_SUBJECT":            "[Botla] New Support Request - %s",
                "HANDOFF_EMAIL_BODY_HEADER":        "A new support request has been received.\n\n",
                "HANDOFF_EMAIL_LABEL_REQUEST_ID":   "Request ID",
                "HANDOFF_EMAIL_LABEL_DATE":         "Date",
                "HANDOFF_EMAIL_LABEL_USER_NOTE":    "User Note",
                "HANDOFF_EMAIL_LABEL_USER":         "User",
                "HANDOFF_EMAIL_LABEL_BOT":          "Bot",
                "HANDOFF_EMAIL_BODY_FOOTER":        "This email was sent automatically by Botla.\n",
                "ERR_INVALID_STATUS":               "invalid status: %s",
            },
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
