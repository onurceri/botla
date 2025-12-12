# 3.5 Chatbot Advanced Configuration Test Plan

## Overview
This test plan covers guardrails, threshold config, fallback messages, topic restrictions, and handoff.

---

## 3.5.1 Confidence Thresholds

### Test Cases

| Test | Action | Expected Result |
|------|--------|-----------------|
| Default config | Create chatbot | high=0.50, medium=0.30, fallback="smart" |
| Update high_threshold | Set to 0.70 | 200 OK |
| Update medium_threshold | Set to 0.40 | 200 OK |
| Validation: high >= medium | Set high=0.30, medium=0.50 | 400 Bad Request |
| Invalid range | Set threshold to 1.5 | 400 Bad Request |
| Negative threshold | Set to -0.1 | 400 Bad Request |

---

## 3.5.2 Fallback Modes

### Test Cases

| Test | Action | Expected Result |
|------|--------|-----------------|
| Static mode | fallback_mode = "static" | Returns static fallback_messages |
| Smart mode | fallback_mode = "smart" | AI generates contextual response |
| Escalate mode | fallback_mode = "escalate" | Triggers handoff flow |
| Plan restriction | Free user uses "smart" | 403 Forbidden |

---

## 3.5.3 Fallback Messages

### Test Cases

| Test | Action | Expected Result |
|------|--------|-----------------|
| Update no_info_found | Custom message | Displayed when no sources |
| Update error_message | Custom message | Displayed on error |
| Update handoff_message | Custom message | Displayed on handoff |
| Empty message | Use default | System default used |
| Localization | Turkish chatbot | Turkish messages displayed |

---

## 3.5.4 Topic Restrictions

### Test Cases

| Test | Action | Expected Result |
|------|--------|-----------------|
| Set allowed_topics | ["support", "pricing"] | Only these topics allowed |
| Set blocked_topics | ["politics", "religion"] | These topics blocked |
| Blocked topic query | Ask about politics | blocked_message returned |
| Allowed topic query | Ask about pricing | Normal response |
| Plan restriction | Free user sets topics | 403 Forbidden |

---

## 3.5.5 Handoff Configuration

### Test Cases

| Test | Action | Expected Result |
|------|--------|-----------------|
| Enable handoff | handoff_enabled = true | 200 OK |
| Set handoff type | handoff_type = "email" | 200 OK |
| Configure email | handoff_config.email | 200 OK |
| Trigger handoff | Low confidence + escalate mode | Handoff triggered |
| Collect email | User prompted | Email collected |
| Handoff tracking | Request created | In handoff_requests table |

---

## How to Run Tests

```bash
go test -v ./internal/integration/... -run "Guardrails|Threshold|Fallback|Topic|Handoff"
```
