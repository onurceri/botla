# 03. Turkish Language & Character Tests

> **Priority**: Critical  
> **Test Count**: 20  
> **Source Files**: `pkg/langconfig/config.go`, `internal/rag/chunker.go`

---

## 3.1 Character Encoding

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| TRK-001 | Turkish special chars in user message: `şŞğĞıİöÖüÜçÇ` | Chars preserved in DB and response | ✅ |
| TRK-002 | Turkish special chars in chatbot name | Chars preserved | ✅ |
| TRK-003 | Turkish special chars in source content | Chars preserved through embedding | ✅ |
| TRK-004 | Turkish chars in system prompt | Chars preserved in LLM context | ✅ |
| TRK-005 | Turkish chars in response from LLM | Chars properly decoded | ✅ |
| TRK-006 | URL encoding of Turkish chars | Correct UTF-8 encoding | ✅ |
| TRK-007 | JSON encoding of Turkish chars | No escaped unicode (e.g., `\u015f`) | ✅ |

### Test Data

```
Input: "Türkiye'de şeker üretimi çok önemlidir. Iğdır şehri güzel."
Expected: Same string preserved in:
- Database columns
- API responses
- Widget display
- LLM context
```

---

## 3.2 Localized Error Messages

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| TRK-010 | Error `ERR_MONTHLY_TOKENS_EXCEEDED` | "Aylık token sınırı aşıldı" | ✅ |
| TRK-011 | Error `ERR_DUPLICATE_URL` | "Yinelenen URL" | ✅ |
| TRK-012 | Error `CHAT_TIMEOUT_OR_INCOMPLETE` | "İşlem tamamlanamadı veya çok uzun sürdü." | ✅ |
| TRK-013 | All 21 error codes have Turkish translations | No `[object Object]` or English fallbacks | ✅ |
| TRK-014 | Default `NoInfoFound` response | "Yeterli bilgi bulamadım." | ✅ |
| TRK-015 | Default system prompt | "Her zaman Türkçe yanıt ver..." | ✅ |

### Error Codes to Test

```go
// From pkg/langconfig/config.go:
"ERR_MONTHLY_TOKENS_EXCEEDED"       // "Aylık token sınırı aşıldı"
"ERR_NAME_AND_ACTION_TYPE_REQUIRED" // "'name' ve 'action_type' alanları zorunludur"
"ERR_PDF_LIMIT_REACHED"             // "Sınır aşıldı: Chatbot başına en fazla PDF dosyası"
"ERR_FILE_TOO_LARGE"                // "Dosya çok büyük"
"ERR_READD_COOLDOWN_ACTIVE"         // "Yeniden ekleme bekleme süresi aktif"
"ERR_DUPLICATE_URL"                 // "Yinelenen URL"
"ERR_ONLY_URL_REFRESH"              // "Yalnızca URL kaynakları yenilenebilir"
"ERR_SOURCE_ALREADY_PROCESSING"     // "Kaynak zaten işleniyor"
"ERR_PLAN_REFRESH_UNAVAILABLE"      // "Planınızda yenileme özelliği mevcut değil"
"ERR_MONTHLY_REFRESH_EXCEEDED"      // "Aylık yenileme sınırı aşıldı"
"ERR_REFRESH_COOLDOWN_ACTIVE"       // "Yenileme bekleme süresi aktif"
"ERR_INVALID_REQUEST_BODY"          // "Geçersiz istek gövdesi"
"ERR_NO_URLS_PROVIDED"              // "Herhangi bir URL sağlanmadı"
"ERR_URL_LIMIT_REACHED"             // "Bu chatbot için URL sınırı aşıldı"
"ERR_MONTHLY_INGESTION_EXCEEDED"    // "Aylık içe‑alma sınırı aşıldı"
"ERR_SITEMAP_PARSE_FAILED"          // "Site haritası ayrıştırılamadı"
"CHAT_TIMEOUT_OR_INCOMPLETE"        // "İşlem tamamlanamadı veya çok uzun sürdü."
"HANDOFF_NOT_ENABLED"               // "Bu chatbot için devretme etkin değil"
"HANDOFF_CREATE_FAILED"             // "Devretme talebi oluşturulamadı"
"HANDOFF_EMAIL_NOT_CONFIGURED"      // "Devretme için e‑posta adresi yapılandırılmamış"
"ERR_INVALID_STATUS"                // "Geçersiz durum: %s"
```

---

## 3.3 Sentence Tokenization

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| TRK-020 | Turkish abbreviation "Dr." not split | "Sayın Dr. Ahmet" → 1 sentence | ✅ |
| TRK-021 | Turkish abbreviation "vb." not split | "araba, bisiklet vb. araçlar" → 1 sentence | ✅ |
| TRK-022 | Turkish sentence end "." | Properly split | ✅ |
| TRK-023 | Turkish sentence end "?" | Properly split | ✅ |
| TRK-024 | Turkish sentence end "!" | Properly split | ✅ |
| TRK-025 | Mixed Turkish/English text | Both tokenizers work | ✅ |
| TRK-026 | Trained tokenizer file exists | `data/sentences/turkish.json` loaded | ✅ |

### Turkish Abbreviations

```go
// From pkg/langconfig/config.go:
[]string{"Dr.", "Prof.", "vb.", "Av.", "Ecz.", "Doç.", "Yrd.", "Cad.", "Sok.", "Mah."}
```

### Test Sentences

```
Input: "Sayın Dr. Ahmet Bey araba, bisiklet vb. araçlar konusunda uzman."
Expected chunks: 1 (single sentence, abbreviations preserved)

Input: "Merhaba. Nasılsınız?"
Expected chunks: 2 (two sentences)
```

---

## Existing Test Coverage

| File | Coverage |
|------|----------|
| `internal/integration/chat_prompt_default_tr_test.go` | Default Turkish prompts |
| `pkg/langconfig/config_test.go` | Config loading, Error messages |
| `internal/rag/chunker_test.go` | Tokenization, Turkish characters |
