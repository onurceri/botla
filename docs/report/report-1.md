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
