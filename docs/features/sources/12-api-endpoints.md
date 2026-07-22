> Extracted from: sources.md — Section: 3. API Endpoints Table

# 3. API Endpoints Table

| Method | Path | Handler | Success Codes | Error Codes |
|--------|------|---------|---------------|-------------|
| GET | `/api/v1/chatbots/{id}/sources` | `ChatbotSources` → `listSources` | 200 | 401, 404 |
| POST | `/api/v1/chatbots/{id}/sources` | `ChatbotSources` → `createSource` | 201 | 400, 401, 403, 404, 409, 413, 422, 402, 503 |
| POST | `/api/v1/chatbots/{id}/sources/bulk` | `BulkCreateSources` | 200, 201 | 400, 401, 403, 404, 422 |
| GET | `/api/v1/sources/{id}` | `GetSourceStatusOrDelete` → `getSourceStatus` | 200, 304 | 401, 404 |
| DELETE | `/api/v1/sources/{id}` | `GetSourceStatusOrDelete` → `deleteSource` | 204 | 401, 404 |
| POST | `/api/v1/sources/{id}/refresh` | `RefreshSource` | 202 | 400, 401, 403, 404, 409, 429, 402 |
| GET | `/api/v1/sources/{id}/chunks` | `GetSourceChunks` | 200 | 401, 404, 500 |
| POST | `/api/v1/chatbots/{id}/sitemap/discover` | `DiscoverSitemap` | 200 | 400, 401, 404 |
