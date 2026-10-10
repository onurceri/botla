# Botla chat widget

The embeddable chat widget: a Preact + TypeScript bundle that mounts into a Shadow DOM, so the host page's CSS can't affect it. It loads the bot's public configuration, then sends messages to the public chat API. Domain allowlists and optional signed embed tokens are enforced by the API.

```html
<script src="https://<your-widget-host>/widget.js" data-bot="<chatbot-id>"></script>
```

The UI language follows the bot's language (`tr` or `en`, defaulting to `tr`). Add `data-lang="en"` to the script tag to force one.

## Embedding

### Script attributes

| Attribute | Meaning |
|---|---|
| `data-bot` | Chatbot ID (required) |
| `data-lang` | Force the UI language (`tr` or `en`) instead of the bot's language |
| `data-parent-origin` | Origin of a page that frames your site and should receive widget events (see below) |

The same options can be passed as query parameters on the script URL (`widget.js?chatbot-id=...&lang=en`).

### Query parameters

| Parameter | Meaning |
|---|---|
| `auto-open=1` | Open the chat panel on load |
| `welcome` | Welcome message shown in an empty conversation |
| `suggestions` | JSON array of suggested questions |
| `position`, `position-strategy` | Bubble placement on the page |
| `reset-session=1`, `session-id` | Start a fresh conversation, or continue a known session |
| `embed-token-url` | Your endpoint that returns a signed embed token (for bots with secure embed enabled) |
| `captcha-site-key` | Captcha site key, if the bot requires one |
| `skip-auto-mount=1` | Load the script without mounting; call `ChatbotWidget.mount()` yourself |
| `debug=1` | Show a diagnostic banner if mounting fails |

Appearance is configured in the dashboard. To override it from the embed code, add `use-url-overrides=1` together with any of: `color`, `header-color`, `header-text-color`, `bot-name`, `bot-icon`, `bot-message-color`, `bot-message-text-color`, `user-message-color`, `user-message-text-color`, `chat-bg-color`, `panel-bg-color`, `input-bg-color`, `input-text-color`, `send-button-color`, `bubble-radius`, `font-family`, `panel-width`, `panel-height`.

### JavaScript API

```js
window.ChatbotWidget.mount()    // mount (after skip-auto-mount=1)
window.ChatbotWidget.unmount()  // remove the widget from the page
```

### Events

When your site is framed by another page (for example, a site builder preview), the widget can tell the parent window what happens in the chat. Set `data-parent-origin` to that page's origin. Events are sent only to that origin, never to `*`:

| `type` | When |
|---|---|
| `WIDGET_EVENT_MESSAGE_SENT` | The visitor sent a message |
| `WIDGET_EVENT_RESPONSE_RECEIVED` | The bot replied |
| `WIDGET_EVENT_SESSION_RESET` | The visitor started a new conversation |
| `WIDGET_EVENT_ERROR` | A request failed |

### Isolation

The widget renders inside a Shadow DOM attached to `<div id="chatbot-widget-host">`, so your page's CSS can't affect it and its CSS can't leak out. The chat API only answers for domains on the bot's allowlist.

## Develop

Install dependencies once from the repository root (`npm ci`), then:

```bash
cp .env.example .env     # VITE_API_BASE_URL, VITE_MARKETING_URL
npm run dev              # Vite dev server; public/widget-test.html embeds the widget
```

## Scripts

| Script | What it does |
|---|---|
| `npm run ci` | `tsc --noEmit`, ESLint, Vitest with coverage |
| `npm run build` | Builds `dist/widget.js` |
| `npm run test:e2e` | Playwright against `vite preview`; run `npm run build` first |
| `npm run build:analyze` | Build and open a bundle-size report |

Deploy to Cloudflare Pages with `make widget-deploy` from the repository root.
