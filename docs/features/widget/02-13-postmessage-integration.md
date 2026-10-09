> Extracted from: widget.md — Section: 2.13 postMessage Integration

## 2.13 postMessage Integration

- [ ] Widget listens for `message` events on `window`
- [x] Origin is validated against allowed origins (exact match, fail closed; configurable via `VITE_DASHBOARD_URL`)
- [x] Outbound events are posted only to the dashboard origins and an optional `parent-origin`, never `'*'` (see `docs/project/features.md` §4.5)
- [ ] Supported commands: `open`, `close`, `toggle`, `sendMessage`, `resetSession`
- [ ] Widget responds with `botla:{eventType}` messages for parent frame
- [ ] Events: `ready`, `opened`, `closed`, `messageSent`, `error`
- [ ] Playground integration can set/get all configuration via postMessage
