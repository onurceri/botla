> Extracted from: training-jobs.md — Section: 2.1 Training Jobs — Get Job Status

# 2.1 Training Jobs — Get Job Status

- [ ] `GET /api/v1/sources/{id}/job` returns the latest training job for a source
- [ ] Response includes: job_id, source_id, status, current_step, progress_percent, error_code, error_message, failed_step, started_at, completed_at
- [ ] If no job exists, source status is mapped to job status (pending→pending, processing→running, completed→completed, failed→failed)
- [ ] Progress is 0% for pending, 50% for processing, 100% for completed
- [ ] Step progress mapping: fetch_source=10%, parse_content=30%, chunk_text=50%, embed_chunks=80%, store_vectors=100%
- [ ] Cross-tenant access returns 404
