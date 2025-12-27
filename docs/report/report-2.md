PHASE 2 — Bugs, Correctness & Reliability
1) Training job idempotency belirsiz

Kanıt: Aynı kaynağın tekrar eklenmesi durumunda job deduplikasyonu açıkça görülmüyor.

Senaryo: Kullanıcı aynı PDF’i iki kez ekler → çifte embedding/storage.

Öneri: Kaynak hash’i ile idempotency.

Güven: Medium

2) Uzun süren işler için timeout/geri dönüş stratejisi net değil

Kanıt: Scraping/embedding adımları ardışık.

Senaryo: LLM provider yavaş → request lifecycle kilitlenir.

Öneri: Async job + polling / webhook.

Güven: High