> Extracted from: training-jobs.md — Section: 2.2 Training Jobs — Retry Job

# 2.2 Training Jobs — Retry Job

- [ ] `POST /api/v1/sources/{id}/job/retry` retries a failed training job
- [ ] Only failed jobs can be retried (returns 400 Bad Request with `ERR_JOB_NOT_FAILED` for non-failed jobs)
- [ ] Retry count is reset before re-enqueueing
- [ ] Job is re-enqueued for processing
- [ ] Returns 202 Accepted with `job_id` and `message`
- [ ] Non-existent job returns 404
