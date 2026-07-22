> Extracted from: background-processing.md — Section: 6. Expected Behavior

# 6. Expected Behavior

### Success Conditions

| Condition | Indicator |
|-----------|-----------|
| Processing completed | Source `status: "completed"`, `chunk_count > 0`, `processed_at` set |

### Failure Conditions

| Condition | Indicator |
|-----------|-----------|
| Processing: network error | `ERR_SCRAPE_NETWORK` at `parse_content` |
| Processing: timeout | `ERR_SCRAPE_TIMEOUT` at `parse_content` |
| Processing: forbidden | `ERR_SCRAPE_FORBIDDEN` at `parse_content` |
| Processing: empty content | `ERR_EMPTY_CONTENT` at `parse_content` |
| Processing: dynamic required | `ERR_DYNAMIC_REQUIRED` at `parse_content` |
| Processing: PDF download failed | `ERR_PDF_DOWNLOAD_FAILED` at `fetch_source` |
| Processing: PDF parse failed | `ERR_PDF_PARSE_FAILED` at `parse_content` |
| Processing: chunking failed | `ERR_CHUNKING_FAILED` at `chunk_text` |
| Processing: embedding failed | `ERR_EMBEDDING_FAILED` at `embed_chunks` |
| Processing: LLM not supported | `ERR_LLM_NOT_SUPPORTED` at `embed_chunks` |
| Processing: unknown source type | `ERR_UNKNOWN_SOURCE_TYPE` at `fetch_source` |
| Processing: queue full | `QUEUE_FULL` at `fetch_source` |
| Processing: storage required | `ERR_STORAGE_REQUIRED` at `fetch_source` |
| Processing: empty file path | `ERR_EMPTY_FILE_PATH` at `fetch_source` |
| Processing: empty URL | `ERR_EMPTY_URL` at `fetch_source` |
