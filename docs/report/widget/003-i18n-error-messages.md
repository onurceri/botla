# Widget Task 003: Localize Error Messages

## Background
The widget currently uses a hardcoded Turkish error message (`DEFAULT_ERROR_MESSAGE.tr`). It should respect the chatbot's configured language.

**File:** `widget/src/widgetApp.tsx`
**Location:** Line 268

## Integration Plan
1.  **Use Configured Language**
    - Access `config.language` (if available in config).
    - Or detect browser language if config is missing.

2.  **Select Message**
    - Use `DEFAULT_ERROR_MESSAGE[lang] || DEFAULT_ERROR_MESSAGE.en`.

3.  **Verify**
    - Test with a chatbot configured for English. Verify error message is in English.

## Checklist
- [ ] Identify `config.language` field
- [ ] Implement fallback logic for language selection
- [ ] Update error message selection code
- [ ] Verify with different languages
