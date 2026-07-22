> Extracted from: widget.md — Section: 2.13 postMessage Integration

## 2.13 postMessage Integration

- [ ] Widget listens for `message` events on `window`
- [ ] Origin is validated against allowed origins (configurable)
- [ ] Supported commands: `open`, `close`, `toggle`, `sendMessage`, `resetSession`
- [ ] Widget responds with `botla:{eventType}` messages for parent frame
- [ ] Events: `ready`, `opened`, `closed`, `messageSent`, `error`
- [ ] Playground integration can set/get all configuration via postMessage
