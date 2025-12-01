## Kullanıcı Karmaşıklığı
- Basic embed: Müşteri yalnızca `<script src="...">` ekler. Ekstra adım yok.
- Secure embed: Müşteri backend’i olan projelerde tek küçük endpoint ile kısa ömürlü `embed_jwt` üretir (≈5–10 satır) ve bir ortam değişkeni ekler. Backend’i olmayan projelerde Turnstile/recaptcha anahtarı eklenir; biz sunucuda doğrularız.

## Güvenlik Seçenekleri
- Allowed domains (dashboard): Müşteri alan adlarını girer; biz `Origin` kontrolü ve CORS ile kısıtlarız. Düşük karmaşıklık, orta güvenlik.
- Embed secret + token (önerilen): Her bot için `embed_secret` veririz. Müşteri kendi endpoint’inde bu secret ile `embed_jwt` üretir. Public chat ucu token’ı doğrular. Düşük ek yük, yüksek güvenlik.
- Captcha + Rate Limit: Backend yoksa captcha (Turnstile/recaptcha) gerektiririz ve güçlü rate limit uygularız.

## Teknik Uygulama
1) Backend Public Uçları
- `GET /api/v1/public/chatbots/:id`: Tema, pozisyon, ikon, ad, karşılama.
- `POST /api/v1/public/chatbots/:id/chat`: `embed_jwt` doğrulama (varsa), captcha doğrulama (yoksa), rate limit, yanıt üretimi.

2) Dashboard Ayarları
- `allowed_domains` listesi.
- Bot’a özel `embed_secret` oluşturma/gösterme.

3) Widget Güncellemeleri
- Parametre: `embed-token-url` veya `data-embed-token`. Varsa önce tokenı alıp `POST`’lara ekler.
- Captcha entegrasyonu için `data-captcha-site-key` desteği.

4) Testler
- Playwright E2E: Basic ve Secure embed akışlarını, Shadow DOM render’ını, mesaj gönderimi ve localStorage’ı doğrular.
- CORS/Origin testleri: İzinli/izinsiz origin davranışları.
- Go integration: Public uçlar için `200/401/403/429` durumları ve token/captcha doğrulaması.

5) Dokümantasyon
- Quick Start (Basic): Sadece `<script>` ekleyerek çalışma.
- Secure Mode: Env değişkeni + minimal endpoint örnekleri (Node/Next, PHP vb.) ve widget parametresi.

## Sonuç
- Basic modla sıfır ek yük.
- Secure modla tipik müşteri için tek küçük endpoint ve bir env ile yüksek güvenlik.
- Her iki mod için otomatik testler ve net dokümantasyon sağlanır.