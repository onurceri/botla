> Extracted from: widget.md — Section: 2.12 Sanitization (`utils/sanitize.ts`)

## 2.12 Sanitization (`utils/sanitize.ts`)

- [ ] `sanitizeUrl()` blocks `javascript:` protocol
- [ ] `sanitizeUrl()` blocks `data:` protocol (except images)
- [ ] `sanitizeUrl()` blocks `vbscript:` protocol
- [ ] `sanitizeUrl()` allows `http:`, `https:`, `mailto:`
- [ ] `sanitizeMessage()` removes `<script>` tags
- [ ] `sanitizeMessage()` removes `onerror`, `onclick`, etc. event handlers
- [ ] `sanitizeMessage()` removes `data:` URIs in attributes
- [ ] `sanitizeMarkdown()` passes through safe HTML (bold, italic, links)
- [ ] `sanitizeMarkdown()` strips dangerous HTML before markdown rendering
