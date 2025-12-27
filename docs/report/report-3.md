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