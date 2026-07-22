> Extracted from: training-jobs.md — Section: 1. Feature Overview

# 1. Feature Overview

The Training Jobs & Pending URLs subsystem handles async job tracking for source ingestion and the pending URL approval workflow. Training jobs track step-by-step progress (0–100%) through the processing pipeline with retry support for failed jobs. Pending URLs are discovered sub-page URLs awaiting user approval before being created as sources.

**Key components:**
- **Training Jobs** — Async job tracking with step-by-step progress (0–100%), retry for failed jobs
- **Pending URLs** — Discovered sub-page URLs awaiting user approval (approve/reject/clear)
