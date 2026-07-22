> Extracted from: widget.md — Section: 2.2 Configuration (`data-*` attributes)

## 2.2 Configuration (`data-*` attributes)

- [ ] `data-chatbot-id` — Required; widget refuses to render if missing
- [ ] `data-api-base` — Custom API endpoint override (defaults to `/api/v1/public/chatbots`)
- [ ] `data-color` — Primary accent color for UI elements
- [ ] `data-welcome` — Welcome message shown on open
- [ ] `data-position` — `bottom-left` or `bottom-right` (default: `bottom-right`)
- [ ] `data-auto-open` — Automatically opens the drawer on load (milliseconds delay supported)
- [ ] `data-panel-height` — Max height of the chat drawer (CSS value, default: `520px`)
- [ ] `data-panel-width` — Width of the chat drawer (CSS value, default: `380px`)
- [ ] `data-header-color` — Header background color
- [ ] `data-header-text-color` — Header text color
- [ ] `data-bot-message-color` — Bot message bubble background
- [ ] `data-user-message-color` — User message bubble background
- [ ] `data-bot-message-text-color` — Bot message text color
- [ ] `data-user-message-text-color` — User message text color
- [ ] `data-chat-bg-color` — Chat area background color
- [ ] `data-input-bg-color` — Input field background color
- [ ] `data-input-text-color` — Input field text color
- [ ] `data-send-button-color` — Send button color
- [ ] `data-bubble-radius` — Border radius for message bubbles
- [ ] `data-bot-name` — Display name for the bot
- [ ] `data-bot-icon` — URL or emoji for bot avatar
- [ ] `data-hide-branding` — Hides "Powered by botla" branding
- [ ] `data-custom-branding` — JSON object with custom branding (logo URL, text, link)
- [ ] `data-embed-token-url` — URL to fetch a one-time embed token for secure embedding
- [ ] `data-captcha-site-key` — reCAPTCHA site key for bot protection
- [ ] `data-reset-session` — Forces a new session on mount
- [ ] `data-session-id` — Explicit session ID override
- [ ] `data-preview-mode` — Enables preview/debug mode
- [ ] `data-position-strategy` — `fixed` (default) or `absolute` positioning
- [ ] `data-use-url-overrides` — Allows URL query parameters to override configuration
- [ ] `data-suggestions` — JSON array of suggested questions
