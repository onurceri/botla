package langconfig

// LanguageConfig holds configuration for a specific language.
type LanguageConfig struct {
	Code            string
	Name            string
	Abbreviations   []string // For sentence splitting (e.g., "Dr.", "Mr.")
	TokenMultiplier float64  // For token estimation
	OCRLanguage     string   // Tesseract language code (e.g., "tur", "eng")
	TokenizerData   string   // Path to the JSON training data
	// UserMessages contains localized strings shown to end users
	UserMessages UserMessages
}

// UserMessages contains all user-facing localized strings.
// LLM prompts are NOT included here - they are always in English (see services/chat_prompts.go).
type UserMessages struct {
	// Fallback and error messages shown to users
	NoInfoFound       string // "I couldn't find information on this topic"
	ErrorMessage      string // "An error occurred, please try again"
	WelcomeMessage    string // Default welcome message
	ConfidenceWarning string // Warning appended to medium-confidence responses
	HandoffSuggestion string // Message suggesting human support
	EmptyStateMessage string // Softer message when bot has no knowledge sources

	// Organization/Workspace defaults
	DefaultOrgName       string // Default org name when no user name
	DefaultOrgNameFormat string // Format: "%s's Workspace"
	DefaultWorkspaceName string // Default workspace name

	// API Error messages
	Errors map[string]string
}

// Configs holds all supported language configurations.
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
		UserMessages: UserMessages{
			NoInfoFound:       "Yeterli bilgi bulamadım.",
			ErrorMessage:      "Şu an bir hata oluştu, lütfen tekrar deneyin.",
			WelcomeMessage:    "Merhaba! Size nasıl yardımcı olabilirim?",
			ConfidenceWarning: "\n\n⚠️ *Bu yanıt, sınırlı bilgi kaynaklarına dayanmaktadır ve kesin doğruluğu garanti edilemez.*",
			HandoffSuggestion: "Bu konuda size en iyi şekilde yardımcı olabilmem için bir uzmanımızla görüşmenizi öneririm. 'İnsan Desteği İste' butonunu kullanabilirsiniz.",
			EmptyStateMessage: "Henüz bilgi kaynaklarım yüklenmedi, ama yardımcı olmaya hazırım!",

			DefaultOrgName:       "Kişisel Organizasyon",
			DefaultOrgNameFormat: "%s Organizasyonu",
			DefaultWorkspaceName: "Varsayılan",

			Errors: map[string]string{
				"ERR_MONTHLY_TOKENS_EXCEEDED":       "Aylık token sınırı aşıldı",
				"ERR_NAME_AND_ACTION_TYPE_REQUIRED": "'name' ve 'action_type' alanları zorunludur",
				"ERR_PDF_LIMIT_REACHED":             "Sınır aşıldı: Chatbot başına en fazla PDF dosyası",
				"ERR_FILE_TOO_LARGE":                "Dosya çok büyük",
				"ERR_READD_COOLDOWN_ACTIVE":         "Yeniden ekleme bekleme süresi aktif",
				"ERR_DUPLICATE_URL":                 "Yinelenen URL",
				"ERR_ONLY_URL_REFRESH":              "Yalnızca URL kaynakları yenilenebilir",
				"ERR_SOURCE_ALREADY_PROCESSING":     "Kaynak zaten işleniyor",
				"ERR_PLAN_REFRESH_UNAVAILABLE":      "Planınızda yenileme özelliği mevcut değil",
				"ERR_MONTHLY_REFRESH_EXCEEDED":      "Aylık yenileme sınırı aşıldı",
				"ERR_REFRESH_COOLDOWN_ACTIVE":       "Yenileme bekleme süresi aktif",
				"ERR_INVALID_REQUEST_BODY":          "Geçersiz istek gövdesi",
				"ERR_NO_URLS_PROVIDED":              "Herhangi bir URL sağlanmadı",
				"ERR_URL_LIMIT_REACHED":             "Bu chatbot için URL sınırı aşıldı",
				"ERR_MONTHLY_INGESTION_EXCEEDED":    "Aylık içe‑alma sınırı aşıldı",
				"ERR_SITEMAP_PARSE_FAILED":          "Site haritası ayrıştırılamadı",
				"CHAT_TIMEOUT_OR_INCOMPLETE":        "İşlem tamamlanamadı veya çok uzun sürdü.",
				"handoff_exists":                    "Bu konuşma için zaten açık bir destek talebi var",
				"handoff_not_found":                 "Destek talebi bulunamadı",
				"handoff_expired":                   "Bu destek talebinin süresi doldu",
				"handoff_closed":                    "Bu destek talebi zaten kapatıldı",
				"handoff_rate_limited":              "Çok fazla destek talebi. Lütfen daha sonra tekrar deneyin.",
				"handoff_not_enabled":               "Bu chatbot için devretme etkin değil",
				"HANDOFF_NOT_ENABLED":               "Bu chatbot için devretme etkin değil",
				"HANDOFF_CREATE_FAILED":             "Devretme talebi oluşturulamadı",
				"HANDOFF_RECEIVED":                  "Talebiniz alındı. En kısa sürede bir temsilcimiz sizinle iletişime geçecektir.",
				"HANDOFF_EMAIL_NOT_CONFIGURED":      "Devretme için e‑posta adresi yapılandırılmamış",
				"HANDOFF_CONVERSATION_LOAD_FAILED":  "Konuşma yüklenemedi",
				"HANDOFF_EMAIL_SUBJECT":             "[Botla] Yeni Destek Talebi - %s",
				"HANDOFF_EMAIL_BODY_HEADER":         "Yeni bir destek talebi alındı.\n\n",
				"HANDOFF_EMAIL_LABEL_REQUEST_ID":    "Talep ID",
				"HANDOFF_EMAIL_LABEL_DATE":          "Tarih",
				"HANDOFF_EMAIL_LABEL_USER_NOTE":     "Kullanıcı Notu",
				"HANDOFF_EMAIL_LABEL_USER":          "Kullanıcı",
				"HANDOFF_EMAIL_LABEL_BOT":           "Bot",
				"HANDOFF_EMAIL_BODY_FOOTER":         "Bu e‑posta Botla tarafından otomatik olarak gönderilmiştir.\n",
				"ERR_INVALID_STATUS":                "Geçersiz durum: %s",
				"ERR_MAX_CHATBOTS_EXCEEDED":         "Maksimum chatbot limitine ulaştınız.",
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
		UserMessages: UserMessages{
			NoInfoFound:       "I could not find enough information.",
			ErrorMessage:      "An error occurred, please try again later.",
			WelcomeMessage:    "Hello! How can I help you today?",
			ConfidenceWarning: "\n\n⚠️ *This response is based on limited sources and accuracy cannot be guaranteed.*",
			HandoffSuggestion: "For the best assistance on this topic, I recommend speaking with one of our specialists. You can use the 'Request Human Support' button.",
			EmptyStateMessage: "My knowledge sources haven't been set up yet, but I'm ready to help!",

			DefaultOrgName:       "Personal Workspace",
			DefaultOrgNameFormat: "%s's Workspace",
			DefaultWorkspaceName: "Default",

			Errors: map[string]string{
				"ERR_MONTHLY_TOKENS_EXCEEDED":       "Monthly token limit exceeded",
				"ERR_NAME_AND_ACTION_TYPE_REQUIRED": "name and action_type are required",
				"ERR_PDF_LIMIT_REACHED":             "Limit reached: Max PDF files per chatbot",
				"ERR_FILE_TOO_LARGE":                "File too large",
				"ERR_READD_COOLDOWN_ACTIVE":         "Re-add cooldown active",
				"ERR_DUPLICATE_URL":                 "Duplicate URL",
				"ERR_ONLY_URL_REFRESH":              "Only URL sources can be refreshed",
				"ERR_SOURCE_ALREADY_PROCESSING":     "Source is already being processed",
				"ERR_PLAN_REFRESH_UNAVAILABLE":      "Refresh feature is not available on your plan",
				"ERR_MONTHLY_REFRESH_EXCEEDED":      "Monthly refresh limit exceeded",
				"ERR_REFRESH_COOLDOWN_ACTIVE":       "Refresh cooldown active",
				"ERR_INVALID_REQUEST_BODY":          "Invalid request body",
				"ERR_NO_URLS_PROVIDED":              "No URLs provided",
				"ERR_URL_LIMIT_REACHED":             "URL limit reached for this chatbot",
				"ERR_MONTHLY_INGESTION_EXCEEDED":    "Monthly ingestion limit exceeded",
				"ERR_SITEMAP_PARSE_FAILED":          "Failed to parse sitemap",
				"CHAT_TIMEOUT_OR_INCOMPLETE":        "Operation could not be completed or took too long.",
				"handoff_exists":                    "A support request already exists for this conversation",
				"handoff_not_found":                 "Support request not found",
				"handoff_expired":                   "This support request has expired",
				"handoff_closed":                    "This support request has already been closed",
				"handoff_rate_limited":              "Too many support requests. Please try again later.",
				"handoff_not_enabled":               "handoff is not enabled for this chatbot",
				"HANDOFF_NOT_ENABLED":               "handoff is not enabled for this chatbot",
				"HANDOFF_CREATE_FAILED":             "failed to create handoff request",
				"HANDOFF_RECEIVED":                  "Your request has been received. Our team will contact you shortly.",
				"HANDOFF_EMAIL_NOT_CONFIGURED":      "email address not configured for handoff",
				"HANDOFF_CONVERSATION_LOAD_FAILED":  "failed to load conversation",
				"HANDOFF_EMAIL_SUBJECT":             "[Botla] New Support Request - %s",
				"HANDOFF_EMAIL_BODY_HEADER":         "A new support request has been received.\n\n",
				"HANDOFF_EMAIL_LABEL_REQUEST_ID":    "Request ID",
				"HANDOFF_EMAIL_LABEL_DATE":          "Date",
				"HANDOFF_EMAIL_LABEL_USER_NOTE":     "User Note",
				"HANDOFF_EMAIL_LABEL_USER":          "User",
				"HANDOFF_EMAIL_LABEL_BOT":           "Bot",
				"HANDOFF_EMAIL_BODY_FOOTER":         "This email was sent automatically by Botla.\n",
				"ERR_INVALID_STATUS":                "invalid status: %s",
				"ERR_MAX_CHATBOTS_EXCEEDED":         "Max chatbots limit exceeded.",
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
