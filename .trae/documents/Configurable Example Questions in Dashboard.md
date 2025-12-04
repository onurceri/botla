**Model & Behavior**

* Single list: `chatbots.suggested_questions` holds the current suggestions; ingestion populates when empty, admin edits directly.

* Toggle: `chatbots.suggestions_enabled BOOLEAN` controls visibility.

* Public output: if enabled → return `suggested_questions`, else return `[]`.

**Caching**

* Cache computed public suggestions per chatbot: key `public:chatbot:<id>:suggestions:v1`, TTL 5–10 minutes.

* Implementation: use existing cache interface (`internal/scraper/cache.go`) with Redis when `REDIS_URL` is set, otherwise memory.

* Invalidation triggers:

  * Admin updates via `PUT /api/v1/chatbots/:id` when `suggested_questions` or `suggestions_enabled` changes.

  * Ingestion aggregation writes `chatbots.suggested_questions` (only when empty) → invalidate.

  * Source deletion (`DELETE /api/v1/sources/:id`) → invalidate chatbot suggestions cache (best‑effort), since content may be re‑ingested later.

**API Changes**

* Auth API (`GET/PUT /api/v1/chatbots/:id`): include `suggestions_enabled` and `suggested_questions` with server‑side validation (≤6 items, ≤120 chars, trim/dedupe).

* Public API (`GET /api/v1/public/chatbots/:id`):

  * Check cache for final suggestions array; if miss, compute `(enabled ? list : [])`, set cache, and return.

**Dashboard UI**

* Add “Örnek Sorular” in Chatbot settings page:

  * Toggle “Örnek soruları göster”.

  * Editable list bound to `suggested_questions` with add/remove/edit; limits enforced.

  * Button “Kaynaklardan Yenile” repopulates from ingestion aggregation only if list is currently empty or admin confirms overwrite.

**Ingestion**

* After extracting suggestions per source and aggregating, write to `chatbots.suggested_questions` only if empty; do not overwrite admin edits.

* Post‑write, invalidate the public suggestions cache.

**Validation & Tests (100% for modified parts)**

* Backend unit/integration:

  * Chatbot PUT/GET round‑trips new fields; enforcement of limits; invalidation called.

  * Public handler caches result; returns from cache; invalidates on changes.

  * Ingestion aggregation respects “only-if-empty” rule and invalidates cache.

* Frontend unit:

  * Settings section renders toggle and editable chips; save sends the correct payload; loads binds from GET.

