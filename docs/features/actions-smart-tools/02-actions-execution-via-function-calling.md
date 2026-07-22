> Extracted from: actions-smart-tools.md — Section: 1.2 Actions — Execution via Function Calling

## 1.2 Actions — Execution via Function Calling

- [ ] Enabled actions are included in the LLM function-calling definition
- [ ] LLM can choose to invoke an action during chat generation
- [ ] Action execution is preformed server-side with configurable timeout
- [ ] HTTP actions make outbound requests with configured method, headers, and body
- [ ] Parameters from the LLM's function call arguments are substituted into the request
- [ ] Action response is injected back into the LLM context for final response generation
- [ ] Action execution errors (timeout, non-2xx, connection failure) are surfaced to the LLM
- [ ] Disabled actions are excluded from function-calling definitions
- [ ] Action execution is logged in `actions_logs`
- [ ] SSRF protection is applied to action endpoints (private IP blocks)
- [ ] Rate limiting applies to action executions
