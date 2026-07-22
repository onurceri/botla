> Extracted from: chatbot-management.md — Section 2.3

# 2.3 Suggestions

## Regenerate Suggestions — `POST /api/v1/chatbots/{id}/suggestions/regenerate`
- [ ] Creates a suggestion job in the database with status `pending`
- [ ] Submits job to worker pool for async processing
- [ ] Returns `202 Accepted` with `job_id` in response body
- [ ] `404 Not Found` for non-existent chatbot
- [ ] `403 Forbidden` for unauthorized access

## Check Suggestion Status — `GET /api/v1/chatbots/{id}/suggestions/status`
- [ ] Returns latest suggestion job for the chatbot
- [ ] Response includes: `job_id`, `status`, optional `suggested_questions`, optional `error_message`
- [ ] `status` values: `"pending"`, `"running"`, `"completed"`, `"failed"`
- [ ] `200 OK` when job exists
- [ ] `404 Not Found` when no suggestion job exists for this chatbot
- [ ] `suggested_questions` only included when there are actual questions (non-empty)
