PHASE 1 — Architecture & System Design Review
1) Güçlü ama yer yer bulanık sınırlar

Açıklama: Katmanlı mimari (Controller → Service → Domain → Infra) net; ancak bazı iş kararları servislerde yoğunlaşıyor.

Kanıt: Training orkestrasyonu, provider seçimi ve guardrail uygulamaları aynı servis akışında.

Etki: Değişiklikler zincirleme etki yaratabilir; test izolasyonu zorlaşır.

Öneri: Orkestrasyon ile domain kararlarını ayıran “use-case” katmanı eklenebilir.

2) Frontend–Backend sözleşmesi iyi ama tiplenmemiş

Açıklama: Error code’lar güçlü bir sözleşme oluşturuyor.

Kanıt: Frontend error mapping’lerinin backend code’larla birebir eşleşmesi.

Etki: Versiyonlama ve evrim zorlaşır.

Öneri: Paylaşılan şema (OpenAPI + enum’lar) veya contract testleri.

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

PHASE 3 — Code Quality & Technical Debt
1) Service layer şişkinliği

Problem: Tek servis çok fazla sorumluluk üstleniyor.

Etki: Okunabilirlik ve test edilebilirlik düşer.

Kanıt: Training, validation, provider seçimi aynı dosyalarda.

Öneri: Küçük, amaç odaklı servisler (Single Responsibility).

2) Frontend’de ürün semantiği UI’ya gömülü

Problem: Domain kavramları component’ler içinde.

Etki: UI refactor’ları iş kurallarını etkiler.

Öneri: Frontend domain/config katmanı.

PHASE 4 — Performance & Scalability
1) Ardışık pipeline darboğazı

Risk: Fetch → parse → chunk → embed sıralı.

Etki: Kaynak sayısı arttıkça eğitim süresi lineer artar.

Öneri: Paralel chunk embedding, batch API’ler.

Varsayım: Orta–yüksek veri hacmi.

2) Vector store erişim paterni

Risk: Chat sırasında fazla context çekimi.

Etki: Latency artışı.

Öneri: Top-k sınırlama, cache.

PHASE 5 — Security & Safety Review
1) Input validation kapsamı sınırlı

Kanıt: URL/PDF kaynaklarında temel doğrulama.

Risk: SSRF / büyük dosya saldırıları.

Öneri: Allowlist, boyut sınırı, content-type doğrulama.

Risk seviyesi: Medium

2) Prompt/guardrail enforcement backend’de doğru yerde

Pozitif bulgu: Guardrail’lerin servis katmanında uygulanması.

Etki: Widget/Frontend atlatmaları engellenir.

PHASE 6 — Testing, Observability & Dev Experience
1) Test kapsamı orkestrasyon seviyesinde zayıf

Kanıt: Unit test izleri sınırlı, e2e yok.

Etki: Regresyon riski.

Öneri: Use-case bazlı integration testleri.

2) Observability minimal

Kanıt: Structured logging ve trace id izleri sınırlı.

Etki: Prod debug zor.

Öneri: Request-id propagation, basic metrics.

PHASE 7 — Improvements & Evolution Ideas
1) Async training mimarisi

Gerekçe: Uzun süren işler.

Değer: Ölçeklenebilirlik + UX iyileşmesi.

Yön: Job queue + status API.

2) Widget-first API düşüncesi

Gerekçe: Widget ayrı ama core’a bağlı.

Değer: Daha net public API.

Yön: Read-only chat endpoints, token-scoped auth.

3) Contract-driven development

Gerekçe: Frontend–backend bağı güçlü.

Değer: Güvenli evrim.

Yön: OpenAPI + contract tests.

✅ SONUÇ (Koç Yorumu)

Bu sistem:

Ciddi ve doğru kurgulanmış

MVP’yi geçmiş

Ölçeklenmeye yakın ama henüz hazır değil

En yüksek kaldıraç:

Async training

Service parçalama

Contract + test