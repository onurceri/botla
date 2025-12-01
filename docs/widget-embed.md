# Widget Embed

## Quick Start
- Sitenize şu kodu ekleyin: `<script src="https://cdn.botla.co/widget.js" data-bot="<CHATBOT_ID>"></script>`
- Parametreler:
  - `chatbot-id` veya `data-bot`: Bot kimliği
  - `api-base`: API tabanı (örn. `https://api.botla.co`)
  - `color`: Tema rengi (hex)
  - `welcome`: Karşılama mesajı

## Secure Mode
- Allowed domains: Dashboard’da bot ayarlarından izinli alan adlarını girin.
- Embed secret: Dashboard’da gizli anahtar tanımlayın.
- Token üretimi: Müşteri backend’inde kısa ömürlü `embed_jwt` üretip bir URL üzerinden sağlayın.
  - Widget parametresi: `embed-token-url` ile token URL’sini geçin.
- Captcha: Gerekirse `captcha-site-key` parametresi ile captcha entegrasyonu yapın.

## Örnek
- `<script src="https://cdn.botla.co/widget.js?chatbot-id=abc123&api-base=https://api.botla.co&color=#3b82f6"></script>`
