> Extracted from: widget.md — Section: 2.11 Logger (`utils/logger.ts`)

## 2.11 Logger (`utils/logger.ts`)

- [ ] Console logger with levels: debug, info, warn, error
- [ ] Debug mode is enabled by `botla_debug` localStorage key
- [ ] Logger prefix includes `[botla]` for easy filtering
- [ ] Sensitive data (tokens, full messages) is redacted in production mode
- [ ] Logger is disabled in production unless debug mode is active
