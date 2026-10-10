# Multi-language plan

Botla is built for the Turkish market. The dashboard is Turkish-only, sentence
splitting for chunking uses a Punkt model trained on Turkish, and privacy
flows follow KVKK. Each chatbot already has its own language (`tr` or `en`),
and the widget and LLM answers follow it.

This document estimates what full multi-language support (Turkish and English
dashboard, extensible to more languages) would take, so it can be planned
deliberately rather than bolted on. The estimates come from a code inventory
taken in October 2026 and are rough.

## Summary

| Area | Effort | Biggest item |
|---|---|---|
| Dashboard (`frontend/`) | ~75–115 h | Extracting ~900–1,100 unique strings from 95 files |
| Backend (`internal/`, `pkg/`) | ~30–45 h | One language-resolution path instead of scattered `"tr"` defaults |
| Sentence tokenizer data | 1–3 days per new language | Training a Punkt model on a representative corpus |
| Widget | done | UI strings follow the bot's `language_code` (`widget/src/i18n/index.ts`) |
| **Total** | **~105–160 h (3–4 engineer-weeks)** | plus translation and legal review |

## What is already language-aware

- **Per-chatbot language.** `chatbots.language_id` references the `languages`
  table (seeded `tr-TR`, `en-US`). `ChatContextBuilder` resolves the bot's
  language and loads `langconfig.Get(code)`.
- **LLM prompts.** `internal/services/chat_prompts.go` is English and appends a
  language directive ("respond ONLY in <language>"). Topic extraction and
  suggested questions take the language name as a parameter.
- **API errors.** 247 of 248 error responses return a code
  (`WriteErrorCode`), and the frontend maps codes to text.
  `frontend/src/i18n/errors.ts` already has `tr` and `en`.
- **Language config.** `pkg/langconfig/config.go` holds per-language
  abbreviations, token multiplier, tokenizer data and user-facing messages.
- **Widget.** Follows the bot's language, with an optional `data-lang` override.

## Turkish-only today

### Dashboard
- About 391 unique Turkish literals in app code, plus English-only labels.
  Estimated 900–1,100 unique user-facing strings in total. About 45% are in
  `features/chatbot` and 20% in `pages/admin`.
- No i18n library. Landing copy (`src/i18n/landing.ts`) and privacy text
  (`src/i18n/privacy.ts`) are centralised but Turkish-only.
- 15+ hardcoded `tr-TR` date and number formats. `Intl.NumberFormat` with a
  `TRY` fallback in `PlanPage`. `date-fns` `tr` locale in admin pages.
- Strings built by concatenation (e.g. `SourceUploader` character counters),
  which need ICU-style interpolation and English plurals.
- 419 text-based test queries in 59 unit test files, and Turkish selectors in
  13 Playwright files.
- `<html lang="en">` in `index.html` regardless of the UI language.

### Backend
- 15 `"tr"` literals across 9 files, with about 7 duplicated
  normalize/default-language helpers that default to `tr` or `tr-TR`.
- `langconfig.Get` falls back to `tr` for unknown codes.
- Default organization and workspace names are always Turkish
  (`auth.go` uses `langconfig.Get("tr")`).
- `User.PreferredLanguageID` is stored but never used.
- The restricted-fallback prompt template has Turkish greeting examples. The DB
  default `chatbots.system_prompt` (migration 000001) is Turkish.
- Plans are priced in TRY (migration defaults and seed rows). The
  `plan_translations` table exists but is unused.
- `NormalizeTR` is language-neutral despite its name.

### Sentence tokenizer
- Punkt training data is loaded per language from R2 (`tokenizer/<code>.json`)
  in production, or from `cfg.TokenizerData` on disk in development.
- When it's missing, the chunker falls back to the English model. The server
  now logs `tokenizer_fallback_english` at startup when that happens.
- `CountTokens` is a chars/4 heuristic times a per-language multiplier
  (tr 1.3, en 1.0).

## Recommended approach

### Phase 0: decisions (2–4 h)
- Default language policy: Turkish stays the default, detect from the browser,
  and persist per user.
- The meaning of `PreferredLanguageID`. Store a locale code instead of a
  foreign-key ID, or map ID to code in the API.
- Currency: keep TRY pricing and only localize display formatting, or
  introduce multi-currency plans (a business decision).
- Legal: Turkish KVKK text stays authoritative. An English version needs
  legal review, not just translation.

### Phase 1: backend plumbing (8–12 h)
- One `ResolveLanguage(bot, user, request)` path. Remove the `"tr"` literals
  and duplicate helpers.
- Default org and workspace names from the user's locale.
- Reject or explicitly handle unknown locales instead of silently mapping
  them to `tr`.
- Rename `NormalizeTR` to `NormalizeText`.

### Phase 2: dashboard infrastructure (10–16 h)
- `react-i18next` + `i18next-browser-languagedetector` +
  `i18next-resources-to-backend`. The existing `src/i18n/*.ts` dictionaries map
  directly onto i18next resources.
- Typed keys through `CustomTypeOptions`. Turkish is the source locale.
- A language switcher, persisted in localStorage for guests and in the
  profile for logged-in users. Keep `document.documentElement.lang` in sync.
- A `useFormatters()` hook wrapping `Intl` and the `date-fns` locale, to
  replace the hardcoded `tr-TR` calls.
- `eslint-plugin-i18next` (no literal strings) and `i18next-parser`.

### Phase 3: extraction (35–55 h)
- Feature by feature, starting with chatbot (~45%), then admin.
- A codemod (ast-grep or jscodeshift) for plain JSX text and attributes. Do
  interpolation, plurals and toasts by hand.

### Phase 4: translation (10–16 h)
- LLM first draft with a glossary, then native review. Landing copy needs
  marketing review, privacy text needs legal review.

### Phase 5: tests and QA (18–28 h)
- Force `tr` in `setupTests.ts` so the existing assertions keep working, then
  add English tests and a key-parity test (every `en` key exists in `tr` and
  vice versa).
- Playwright: the same forced locale, plus screenshot checks per locale for
  layout overflow, since English strings are often longer.

### Adding a third language
- Train Punkt data, upload it to R2, and add a `langconfig` entry
  (abbreviations, token multiplier, messages) and a `languages` row. The code
  is about 1 hour; the data is the bulk of the work.
- Add the dashboard and widget dictionaries (about 1,100 + 30 strings).

## Risks

- **Silent fallbacks.** Unknown codes map to `tr`, and missing tokenizer data
  maps to English splitting. Add startup checks per language (the tokenizer
  check exists now).
- **Untranslated keys in the UI.** Mitigate with the key-parity test and the
  lint rule.
- **Layout regressions** from longer English text in tables and tabs.
- **Merge conflicts** with ongoing chatbot UI work, since that feature holds
  most of the strings.
