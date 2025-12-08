## Durum Özeti
- `pkg/langconfig` mevcut ve TR/EN için metin/prompt, tokenizer, OCR ayarları sağlıyor. Varsayılan dil `tr`.
- Servis ve RAG katmanlarında `langconfig` kullanılmaya başlamış: prompt fallback’ları, tokenizer, konu çıkarımı, OCR.
- API hataları ve bazı servis çıktıları hâlâ sabit metinler ile yazılıyor; isteğe göre dil seçimi yapılmıyor.
- Dil seçimi şu an chatbot’un `language` alanından (ör. `tr-TR`/`en-US`) normalize edilerek yapılıyor; `Accept-Language`/`?lang=` kullanılmıyor.

## Düzeltilmesi Gerekenler
- `internal/api/errors.go`: hatalar yalnızca verilen `message` string ile yazılıyor; yerelleştirme yok.
- Aşağıdaki sabit kullanıcı mesajları `langconfig` ile yerelleştirilmeli:
  - `internal/api/handlers/chat.go:97` — "Monthly token limit exceeded"
  - `internal/api/handlers/public.go:174` — "Monthly token limit exceeded"
  - `internal/api/handlers/action.go:124` — "name and action_type are required"
  - `internal/api/handlers/source_create.go:65` — "Limit reached: Max PDF files per chatbot"
  - `internal/api/handlers/source_create.go:82` — "File too large"
  - `internal/api/handlers/source_create.go:163` — "Re-add cooldown active"
  - `internal/api/handlers/source_create.go:170` — "Duplicate URL"
  - `internal/api/handlers/source_refresh.go:59` — "Only URL sources can be refreshed"
  - `internal/api/handlers/source_refresh.go:65` — "Source is already being processed"
  - `internal/api/handlers/source_refresh.go:77` — "Refresh feature is not available on your plan"
  - `internal/api/handlers/source_refresh.go:84` — "Monthly refresh limit exceeded"
  - `internal/api/handlers/source_refresh.go:95` — "Refresh cooldown active"
  - `internal/api/handlers/source_bulk.go:63` — "Invalid request body"
  - `internal/api/handlers/source_bulk.go:68` — "No URLs provided"
  - `internal/api/handlers/source_bulk.go:87` — "URL limit reached for this chatbot"
  - `internal/api/handlers/source_bulk.go:99` — "Monthly ingestion limit exceeded"
  - `internal/api/handlers/source_sitemap.go:54` — "Invalid request body"
  - `internal/api/handlers/source_sitemap.go:69` — "Failed to parse sitemap: "
  - `internal/services/chat_service.go:312` — "İşlem tamamlanamadı veya çok uzun sürdü."
  - `internal/services/handoff_service.go:42` — "handoff is not enabled for this chatbot"
  - `internal/services/handoff_service.go:54` — "failed to create handoff request: "
  - `internal/services/handoff_service.go:90` — "Talebiniz alındı..." (e‑posta içerikleri)
  - `internal/services/handoff_service.go:99` — "email address not configured for handoff"
  - `internal/services/handoff_service.go:105` — "failed to load conversation: "
  - `internal/services/handoff_service.go:114` — "[Botla] Yeni Destek Talebi - %s"
  - `internal/services/handoff_service.go:140–160` — e‑posta gövdesindeki TR sabit metinler ("Kullanıcı", "Bot" dâhil)
  - `internal/services/handoff_service.go:179` — "invalid status: %s"

## Teknik Plan
1. `langconfig` genişletme
   - `ResponseTemplates` içine `Errors map[string]string` ekleyin. Anahtarlar standart/hücre bazlı hata kodlarıyla eşleşsin (örn. `ERR_MONTHLY_TOKENS_EXCEEDED`, `ERR_FILE_TOO_LARGE`, `ERR_INVALID_REQUEST_BODY`, `ERR_ONLY_URL_REFRESH`, `ERR_REFRESH_COOLDOWN_ACTIVE`, `CHAT_TIMEOUT_OR_INCOMPLETE`, `HANDOFF_EMAIL_SUBJECT`, `HANDOFF_NOT_ENABLED` vb.).
   - TR/EN config’lerde bu `Errors` sözlüğünü doldurun. Mevcut alanlar (`DefaultSystemPrompt`, `ErrorMessage`, `WelcomeMessage`, `DefaultPersonaPrompt`) korunur.

2. Standart hata kodları
   - `internal/api/errors.go` içinde alanı genişletin veya `internal/api/codes.go` ekleyin: domain odaklı kodlar (örn. `ERR_MONTHLY_TOKENS_EXCEEDED`, `ERR_URL_LIMIT_REACHED`, `ERR_DUPLICATE_URL`). HTTP tabanlı mevcut kodlar (`BAD_REQUEST` vb.) kalır; `ErrorResponse.Code` alanına domain kodu yazılacak.

3. Dil seçimi yardımcıları
   - `internal/api` içine yardımcı ekleyin: `ResolveLangConfig(r *http.Request, chatbotID string) langconfig.LanguageConfig`.
     - Chatbot ID varsa DB’den `LanguageCode` alıp `tr`/`en` tabanına indirger.
     - Yoksa sırasıyla `?lang=` sorgu parametresi → `Accept-Language` başlığı → varsayılan `tr`.
   - Orta vadede `middleware` ile `Accept-Language` çözümleyip `context`’a `lang` ekleyin.

4. Yerelleştirilmiş hata yazımı
   - `internal/api/errors.go` içine `WriteLocalizedError(w, status, code, cfg)` ekleyin: mesajı `cfg.ResponseTemplates.Errors[code]`’dan seçer; yoksa `cfg.ResponseTemplates.ErrorMessage` fallback.
   - Ek olarak `WriteErrorWithDetails` karşılığı `WriteLocalizedErrorWithDetails(...)` sağlayın.

5. Handler refaktörü
   - Yukarıda listelenen handler’larda `http.Error(...)` ve doğrudan `WriteHeader(...)`+sabit metin kullanımını `WriteLocalizedError(...)` ile değiştirin.
   - Chatbot ilişkili isteklerde `ResolveLangConfig(..., chatbotID)` ile botun dilinden seçim yapın; ilişkisiz isteklerde `ResolveLangConfig(r, "")`.

6. Servis katmanı
   - `internal/services/chat_service.go:312` sabit TR fallback’ı `cfg.ResponseTemplates.Errors["CHAT_TIMEOUT_OR_INCOMPLETE"]` veya genel `ErrorMessage` ile değiştirin.
   - `internal/services/handoff_service.go` e‑posta konu/gövde ve durum metinlerini `cfg.ResponseTemplates.Errors[...]` ile üreten küçük bir yardımcıya taşıyın; dil kaynağı: ilgili chatbot’un dili, yoksa `ResolveLangConfig` fallback.

7. Testler
   - `pkg/langconfig`: `Errors` sözlüğünde TR/EN anahtarların dolu olduğu ve `Get("")` ile `tr` varsayılanı doğrulansın.
   - API handler testi: TR botla istek → Türkçe hata mesajı; EN botla istek → İngilizce.
   - Servis testi: ChatService timeout/incomplete → ilgili dilde fallback; Handoff e‑posta konu/gövde yerelleştirmesi.

8. Dokümantasyon
   - `docs/i18n.md`: dil belirleme stratejisi, hata kodları, `ResponseTemplates.Errors` anahtar sözleşmesi.

## Beklenen Sonuçlar
- API ve servis çıktıları tamamen `langconfig` üzerinden yerelleşir; varsayılan Türkçe.
- Hata yanıtları standart `code` ile tutarlı ve dil‑bağımsız olur.
- Chatbot diline veya istekteki `lang`/`Accept-Language`’a göre TR/EN değişimi yapılır.

## Geriye Uyum
- `ErrorResponse` şeması değişmez; sadece `error` alanı yerelleşir ve `code` alanı daha anlamlı domain kodları kullanır.
- Varsayılan davranış değişmeden Türkçe kalır; yeni diller eklemek için yalnızca `pkg/langconfig` genişletmek yeterlidir.