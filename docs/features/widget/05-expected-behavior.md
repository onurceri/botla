> Extracted from: widget.md — Section: 5. Expected Behavior

# 5. Expected Behavior — Success/Failure Conditions

## 5.1 Widget

| Condition | Success Indicator | Failure Indicator |
|---|---|---|
| Widget mounts successfully | Shadow DOM created, bubble visible, config fetched | Console error, no DOM elements, "chatbot-id is required" |
| Chat message sent | Message appears in chat, bot responds | Network error message, message stuck as "sending" |
| Session restored | Previous messages visible on reload | Empty chat, new session created |
| Feedback submitted | Thumb icon highlighted, API called 200 | Icon not highlighted, error logged silently |
| Handoff submitted | Success message shown, email sent | Error message, retry option |
| Custom styling applied | All elements reflect custom colors | Default styling used, CSS variables not applied |
| Markdown rendered | Bold, code, tables rendered correctly | Raw markdown visible, HTML entities shown |
| XSS attempted | Script tags removed, no JS execution | Script executes, alert shows |
