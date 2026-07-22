> Extracted from: actions-smart-tools.md — Section: 3. Test Scenarios

# 3. Test Scenarios

## Scenario 17: Action CRUD — Create

- [ ] Send POST to `/api/v1/chatbots/{id}/actions` with valid body:
  ```json
  {
    "name": "Get Weather",
    "description": "Fetches current weather for a location",
    "type": "http",
    "endpoint": "https://api.weather.com/current",
    "method": "GET",
    "headers": {"Authorization": "Bearer test"},
    "parameters": {"location": {"type": "string", "required": true}}
  }
  ```
- [ ] Verify 201 response with created action object
- [ ] Verify `id` is returned in the response
- [ ] Verify `isEnabled` defaults to `true`
- [ ] Verify action is persisted in the database

## Scenario 18: Action CRUD — Read

- [ ] GET `/api/v1/chatbots/{id}/actions` — verify array of actions
- [ ] GET `/api/v1/chatbots/{id}/actions/{actionId}` — verify single action details
- [ ] GET with invalid actionId → 404
- [ ] GET with action from different chatbot → 404

## Scenario 19: Action CRUD — Update

- [ ] PUT `/api/v1/chatbots/{id}/actions/{actionId}` with updated name and description
- [ ] Verify 200 and updated fields
- [ ] Verify non-updated fields remain unchanged
- [ ] PUT with invalid body → 400
- [ ] PUT non-existent action → 404

## Scenario 20: Action CRUD — Delete

- [ ] DELETE `/api/v1/chatbots/{id}/actions/{actionId}` → 204
- [ ] Verify action no longer appears in list
- [ ] Verify GET for deleted action returns 404
- [ ] DELETE non-existent action → 404

## Scenario 21: Action Execution via Function Calling

**Preconditions:** Chatbot has at least one enabled HTTP action.

- [ ] Send a chat message that should trigger the action (e.g., "What's the weather in London?" for a weather action)
- [ ] Verify the LLM function-calling logic invokes the action
- [ ] Verify the action executes server-side (check action logs)
- [ ] Verify the action response is incorporated into the final LLM response
- [ ] Verify the final response references the action result
- [ ] Verify `ToolCallsJSON` on the assistant message contains the invocation details

## Scenario 22: Action Execution — Error Handling

- [ ] Configure an action with an unreachable endpoint
- [ ] Send a trigger message
- [ ] Verify the LLM receives the error and generates a graceful response (e.g., "I'm sorry, I couldn't fetch the weather right now")
- [ ] Verify the error is logged in action execution logs
- [ ] Configure an action that times out → verify graceful error handling

## Scenario 23: Action Execution — Disabled Actions

- [ ] Disable an action via PUT
- [ ] Send a trigger message
- [ ] Verify the action is NOT invoked
- [ ] Verify the LLM responds without the action data (or notes it's unavailable)

## Scenario 24: Action Logs

- [ ] Execute several actions (some successful, some failing)
- [ ] GET `/api/v1/chatbots/{id}/actions/logs`
- [ ] Verify logs include: action ID, status, request/response summary, timestamp, duration
- [ ] Verify pagination works (if applicable)
- [ ] Verify filtering by action ID or status works (if applicable)

## Scenario 25: SSRF Protection for Actions

- [ ] Create an action with endpoint pointing to `http://localhost:8080/admin`
- [ ] Create an action with endpoint pointing to `http://10.0.0.1/secret`
- [ ] Create an action with endpoint pointing to `http://169.254.169.254/latest/meta-data/`
- [ ] Verify these are rejected at creation time or blocked at execution time
- [ ] Create an action with a legitimate public endpoint → allowed
