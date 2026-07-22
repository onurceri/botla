> Extracted from: chatbot-management.md — Section 6

# 6. Expected Behavior — Success/Failure Conditions

## Success Conditions
- `200 OK`: Successful read or update; JSON body contains chatbot object
- `201 Created`: Successful creation; JSON body contains new chatbot object
- `202 Accepted`: Async job submitted (suggestions, sitemap); JSON body contains `job_id`
- `204 No Content`: Successful deletion; no response body

## Failure Conditions
- `400 Bad Request`:
  - Missing required field (`name`)
  - Empty/whitespace-only `name`
  - Invalid JSON body
  - Non-UUID path parameter
  - Reserved path value `"new"`
- `401 Unauthorized`:
  - Missing or invalid auth cookie
  - Expired JWT
  - No user ID in context
- `403 Forbidden`:
  - Feature not available on plan (body: `{"error": "...", "upgrade_required": true, "feature": "..."}`)
  - Color validation failed (body: `{"error": "Invalid color format...", "upgrade_required": false, "feature": "color"}`)
  - Workspace non-membership
  - Organization non-membership
  - Chatbot limit exceeded (`ERR_MAX_CHATBOTS_EXCEEDED`)
- `404 Not Found`:
  - Non-existent chatbot ID
  - Soft-deleted chatbot
  - No suggestion job found
- `405 Method Not Allowed`:
  - Wrong HTTP method on dispatch handler
- `500 Internal Server Error`:
  - Database query failure
  - Repository is nil
  - Plan fetch failure
  - Unexpected errors
