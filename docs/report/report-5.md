PHASE 5 — Security & Safety Review
1) Input validation kapsamı sınırlı

Kanıt: URL/PDF kaynaklarında temel doğrulama.

Risk: SSRF / büyük dosya saldırıları.

Öneri: Allowlist, boyut sınırı, content-type doğrulama.

Risk seviyesi: Medium

2) Prompt/guardrail enforcement backend’de doğru yerde

Pozitif bulgu: Guardrail’lerin servis katmanında uygulanması.

Etki: Widget/Frontend atlatmaları engellenir.