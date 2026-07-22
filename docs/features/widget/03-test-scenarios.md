> Extracted from: widget.md — Section: 3. Test Scenarios

# 3. Test Scenarios

## 3.1 Widget Rendering Scenarios

### 3.1.1 Happy Path — Standard Widget Load

1. Embed `<div data-chatbot-id="abc123">` on a clean HTML page
2. Load `widget.js` script
3. **Expected:** Shadow DOM is created, bubble appears at bottom-right, config API is fetched
4. Click the bubble
5. **Expected:** Drawer slides open, welcome message is displayed, suggestions are loaded

### 3.1.2 Invalid Chatbot ID

1. Embed with `data-chatbot-id="nonexistent-id"`
2. Load `widget.js`
3. **Expected:** Config API returns 404
4. **Expected:** Widget displays "Chatbot not found" or "Yapılandırma bulunamadı" error state
5. **Expected:** No API calls to chat endpoint are made

### 3.1.3 Empty Chatbot ID

1. Embed with `data-chatbot-id=""` or missing attribute
2. Load `widget.js`
3. **Expected:** Widget refuses to render
4. **Expected:** Console error: "chatbot-id is required"
5. **Expected:** No DOM manipulation occurs

### 3.1.4 Network Error on Config Fetch

1. Embed with valid `data-chatbot-id`
2. Simulate network failure (offline, DNS failure, CORS error)
3. **Expected:** Config fetch fails
4. **Expected:** Widget falls back to `data-*` attributes for configuration
5. **Expected:** Widget still renders and is functional
6. **Expected:** Error is logged to console

### 3.1.5 Network Error on Chat Send

1. Widget is loaded and drawer is open
2. Type a message and send
3. Simulate network failure during `sendMessage`
4. **Expected:** Message is shown in the chat as "sending" or queued
5. **Expected:** After timeout, error message is shown: "Message could not be sent"
6. **Expected:** Retry button or option is available

### 3.1.6 Embed Token Flow

1. Set `data-embed-token-url="https://tokens.example.com/widget-token"`
2. Load widget
3. **Expected:** Widget fetches token from the URL
4. **Expected:** Token is added as `X-Embed-Token` header to all API requests
5. **Expected:** If token fetch returns 401/403, widget shows auth error
6. **Expected:** If token fetch returns 5xx, widget retries with backoff

### 3.1.7 Embed Token Expiry

1. Widget has a valid embed token
2. Token expires mid-session
3. Next API call returns 401
4. **Expected:** Widget attempts to re-fetch token from `embed-token-url`
5. **Expected:** If re-fetch succeeds, the original request is retried
6. **Expected:** If re-fetch fails, widget shows "Session expired" message

### 3.1.8 Suggestions Loading

1. Widget is loaded with `data-suggestions='["What is X?", "How do I Y?"]'`
2. Open drawer
3. **Expected:** Suggestions are rendered as chips in a horizontal carousel
4. Tap a suggestion
5. **Expected:** Suggestion is sent as a message, response is received

### 3.1.9 Suggestions Overflow

1. Widget is loaded with 15+ suggestions
2. Open drawer
3. **Expected:** Prev/next arrows are visible
4. **Expected:** Counter shows "1/3", "2/3", etc.

### 3.1.10 Custom Styling — All CSS Variables

1. Set all `data-*` color/border/radius attributes to distinct values
2. Load widget
3. **Expected:** Each visual element reflects the custom style
4. **Expected:** No style leaks from host page
5. **Expected:** No style leaks into host page

### 3.1.11 Shadow DOM Style Isolation

1. Host page has aggressive global styles (e.g., `* { font-family: Comic Sans !important; }`)
2. Load widget
3. **Expected:** Widget styles are unaffected by host page styles
4. **Expected:** Widget fonts, colors, spacing are as configured

### 3.1.12 RTL / Non-Turkish Content

1. Configure widget with English welcome message and suggestions
2. **Expected:** Widget displays English text correctly
3. **Expected:** No layout issues with non-ASCII characters

### 3.1.13 Accessibility — Keyboard Navigation

1. Open widget
2. Tab through all interactive elements
3. **Expected:** Focus order is logical (input → send → suggestions → close)
4. **Expected:** Enter/Space activates buttons
5. **Expected:** Escape closes the drawer
6. **Expected:** Escape closes the confirm modal
7. **Expected:** Focus is trapped in modal

### 3.1.14 Mobile Viewport

1. Set viewport to 375×812 (iPhone X)
2. Load widget
3. **Expected:** Bubble is smaller and positioned correctly
4. Open drawer
5. **Expected:** Drawer is full width, near full height
6. **Expected:** Safe area insets are respected (no notch overlap)

### 3.1.15 Multiple Chatbots on One Page

1. Embed two widgets with different chatbot IDs
2. **Expected:** Both widgets render independently
3. **Expected:** Each has its own session in localStorage (different keys)
4. **Expected:** Messages to one widget don't affect the other

## 3.2 Session & State Scenarios

### 3.2.1 Session Persistence Across Reload

1. Send a few messages in the widget
2. Reload the page
3. **Expected:** Conversation history is restored
4. **Expected:** Session ID is the same (from localStorage)

### 3.2.2 New Session (reset-session)

1. Set `data-reset-session="true"`
2. Load widget
3. **Expected:** Previous session (if any) is cleared
4. **Expected:** New session ID is generated
5. **Expected:** Conversation starts fresh

### 3.2.3 Explicit Session ID

1. Set `data-session-id="my-custom-session"`
2. Load widget
3. **Expected:** Session ID is "my-custom-session"
4. **Expected:** API calls include this session ID

### 3.2.4 Session Expiry

1. Configure session TTL (e.g., 30 minutes)
2. Send messages, wait for TTL to expire
3. **Expected:** Session is considered expired
4. **Expected:** New session is automatically created

## 3.3 Markdown & Sanitization Scenarios

### 3.3.1 Markdown Rendering

- Bot response contains: `**bold**, *italic*, [link](https://example.com), \`code\``
- **Expected:** Bold, italic, link, and inline code are rendered correctly

### 3.3.2 Code Block Rendering

- Bot response contains a fenced code block
- **Expected:** Code block is rendered with monospace font and background
- **Expected:** Syntax is preserved (no HTML escaping issues)

### 3.3.3 Table Rendering (GFM)

- Bot response contains a GFM table
- **Expected:** Table is rendered with proper columns and rows

### 3.3.4 XSS via Markdown

- Bot response contains: `<script>alert('xss')</script>`
- **Expected:** Script tag is removed, not executed
- **Expected:** Safe rendering of any remaining text

### 3.3.5 XSS via Link URL

- Bot response contains: `[click](javascript:alert('xss'))`
- **Expected:** Link is rendered but URL is sanitized
- **Expected:** Clicking the link does nothing or navigates safely

### 3.3.6 XSS via Image Tag

- Bot response contains: `<img src=x onerror="alert('xss')">`
- **Expected:** `onerror` handler is stripped
- **Expected:** Image tag is rendered safely (broken image or removed)

### 3.3.7 XSS via iframe

- Bot response contains: `<iframe src="https://malicious.com"></iframe>`
- **Expected:** iframe is removed entirely
- **Expected:** No embedded content from external sources

## 3.4 Feedback & Handoff Scenarios

### 3.4.1 Feedback — Thumbs Up

1. Receive a bot message
2. Click thumbs up
3. **Expected:** Feedback is submitted to API
4. **Expected:** Thumbs up is highlighted (active state)
5. **Expected:** Thumbs down remains inactive
6. **Expected:** Clicking thumbs up again does nothing (idempotent)

### 3.4.2 Feedback — Thumbs Down

1. Receive a bot message
2. Click thumbs down
3. **Expected:** Same as 3.4.1 but for thumbs down

### 3.4.3 Feedback — API Failure

1. Click thumbs up/down
2. API returns 500
3. **Expected:** Visual state does not change (shows error silently)
4. **Expected:** User can retry

### 3.4.4 Human Handoff — Success

1. Click "Talk to human" (or similar handoff trigger)
2. Email form appears
3. Enter valid email and optional message
4. Click submit
5. **Expected:** POST to handoff endpoint succeeds
6. **Expected:** Success message is shown: "We'll get back to you soon"

### 3.4.5 Human Handoff — Invalid Email

1. Click "Talk to human"
2. Enter invalid email ("notanemail")
3. Click submit
4. **Expected:** Form validation error is shown
5. **Expected:** API call is NOT made

### 3.4.6 Human Handoff — API Failure

1. Submit valid handoff form
2. API returns 500
3. **Expected:** Error message is shown: "Could not submit request"
4. **Expected:** User can retry
