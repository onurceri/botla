> Extracted from: widget.md — Section: 2.6 Message Component

## 2.6 Message Component

- [ ] User messages are right-aligned (blue bubbles)
- [ ] Bot messages are left-aligned (gray bubbles)
- [ ] Markdown is rendered: bold, italic, links, lists, code blocks, tables
- [ ] GFM (GitHub Flavored Markdown) is supported (strikethrough, task lists)
- [ ] Code blocks have syntax highlighting (or at least monospace + background)
- [ ] Links open in new tab (`target="_blank"`, `rel="noopener noreferrer"`)
- [ ] HTML is sanitized — `<script>`, `<style>`, `<iframe>` are removed (`rehype-sanitize`)
- [ ] `sanitizeUrl()` prevents `javascript:` URLs
- [ ] `sanitizeMessage()` strips dangerous HTML from raw message text
- [ ] Thumbs up/down feedback buttons appear on bot messages
- [ ] Feedback state is persisted (cannot re-vote after selection)
- [ ] Human handoff email form appears after handoff request
- [ ] Email validation on the handoff form
- [ ] Timestamp is shown below each message (or relative time)
- [ ] Citations/sources are rendered if present in response
