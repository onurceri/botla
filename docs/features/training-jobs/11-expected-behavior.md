> Extracted from: training-jobs.md — Section: 6. Expected Behavior

# 6. Expected Behavior

### Success Conditions

| Condition | Indicator |
|-----------|-----------|
| Job status retrieved | 200 OK with JobStatusResponse |
| Job retried | 202 Accepted with `{"job_id": "...", "message": "Job queued for retry"}` |
| Pending URLs listed | 200 OK with `{"urls": [...], "total": N, "page": P, "per_page": PP}` |
| Pending URLs approved | 200 OK with `{"approved_count": N, "sources_created": N}` |
| Pending URLs rejected | 200 OK with `{"rejected_count": N}` |
| Pending URLs cleared | 200 OK with `{"cleared_count": N}` |

### Failure Conditions

| Condition | Indicator |
|-----------|-----------|
| Job not failed (retry) | 400 Bad Request with `ERR_JOB_NOT_FAILED` |
| Job not found (retry) | 404 Not Found |
