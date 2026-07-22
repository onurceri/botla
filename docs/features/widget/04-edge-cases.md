> Extracted from: widget.md — Section: 4. Edge Cases

# 4. Edge Cases

## 4.1 Widget Edge Cases

| # | Edge Case | Expected Behavior |
|---|---|---|
| 4.1.1 | Host page uses CSP that blocks inline scripts | Widget loads successfully; CSP is for host page, not shadow root |
| 4.1.2 | Host page has `document.domain` set | postMessage origin validation still works correctly |
| 4.1.3 | Browser localStorage is full | Session save fails silently; widget continues without persistence |
| 4.1.4 | Browser localStorage is disabled (private mode in some browsers) | `getSession`/`saveSession` catch exceptions; widget works in-memory only |
| 4.1.5 | `crypto.randomUUID()` is not available (HTTP, older browser) | Falls back to alternate UUID generation |
| 4.1.6 | Multiple rapid open/close clicks | Debounced or state-checked; no duplicate renders |
| 4.1.7 | Very long message (> 10,000 chars) | Input is truncated or handled gracefully |
| 4.1.8 | Very long bot response (> 10,000 chars) | Message is scrollable within the drawer |
| 4.1.9 | Widget is embedded in an iframe | postMessage integration works; widget is fully functional |
| 4.1.10 | `data-custom-branding` contains invalid JSON | Silently ignored; falls back to default branding |
| 4.1.11 | `data-suggestions` contains invalid JSON | Silently ignored; no suggestions shown |
| 4.1.12 | Widget CSS variable contains invalid value (e.g., `not-a-color`) | CSS default values are used; no rendering crash |
| 4.1.13 | Host page removes the target element after widget mount | Widget handles gracefully (target element not found errors) |
| 4.1.14 | 100+ messages in a single session | Virtual scrolling or truncation; performance is acceptable |
| 4.1.15 | Simultaneous sends (user double-clicks send) | Loading state prevents duplicate sends |
| 4.1.16 | API returns malformed response (missing `reply` field) | Widget shows fallback error message; does not crash |
| 4.1.17 | API returns HTML instead of JSON | Widget handles parse error; shows generic error |
| 4.1.18 | Widget loaded on a page with existing React/Vue/Preact | Shadow DOM prevents framework conflicts |
| 4.1.19 | Mobile device with very small screen (320px width) | Widget adapts; all functionality is usable |
| 4.1.20 | High-DPI / Retina displays | Bot icons and branding images are crisp |
| 4.1.21 | Screen reader / VoiceOver / TalkBack | All interactive elements have proper ARIA labels and roles |
| 4.1.22 | Host page uses `transform: translateZ(0)` or 3D transforms | Widget z-index is high enough to render above transformed elements |
