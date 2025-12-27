PHASE 6 — Testing, Observability & Dev Experience
1) Test kapsamı orkestrasyon seviyesinde zayıf

Kanıt: Unit test izleri sınırlı, e2e yok.

Etki: Regresyon riski.

Öneri: Use-case bazlı integration testleri.

2) Observability minimal

Kanıt: Structured logging ve trace id izleri sınırlı.

Etki: Prod debug zor.

Öneri: Request-id propagation, basic metrics.