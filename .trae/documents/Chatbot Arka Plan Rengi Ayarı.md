## Amaç
- `/chatbots/:id` sayfasından sohbet penceresinin arka plan rengini ayarlamak.
- Varsayılan renk: `#FFF5E6` (krem).
- Ayar, widget’a otomatik yansısın ve public config API’si ile taşınsın.

## Backend Değişiklikleri
- Model güncellemesi:
  - `internal/models/chatbot.go` içine `ChatBackgroundColor string \\`json:"chat_background_color"\\`` alanı eklenir.
- DB/migrasyon:
  - Yeni kolon `chat_background_color TEXT NOT NULL DEFAULT '#FFF5E6'` `chatbots` tablosuna eklenir.
  - Dosyalar: `db/migrations/000006_add_chat_background_color.up.sql` ve `...down.sql` (kolonu geri almak için `ALTER TABLE chatbots DROP COLUMN chat_background_color`).
- Repository fonksiyonları:
  - `internal/db/chatbot.go`:
    - `CreateChatbot(...)` INSERT listesine `chat_background_color` eklenir ve değeri yazılır.
    - `GetChatbotsByUserID(...)` ve `GetChatbotByID(...)` SELECT listesine `chat_background_color` eklenir ve scan edilir.
    - `UpdateChatbot(...)` UPDATE set listesine `chat_background_color=$N` eklenir.
- Handler’lar:
  - `internal/api/handlers/chatbot.go`:
    - `createChatbotRequest` içine `ChatBackgroundColor *string \\`json:"chat_background_color"\\`` alanı eklenir.
    - `ListOrCreate` (POST) içinde varsayılan olarak `#FFF5E6` kullanılır.
    - `ByID` (PUT) içinde gönderilirse mevcut değeri güncellenir.
  - Public config:
    - `internal/api/handlers/public.go`:
      - `publicChatbot` yapısına `ChatBackgroundColor string \\`json:"chat_background_color"\\`` eklenir.
      - `PublicChatbotConfig` dönen JSON’a bu alan eklenir (ör. `public.go:34`).
- Validasyon:
  - Renk formatını basit bir regex ile `#RGB` veya `#RRGGBB` hex olarak doğrula; geçersizse 400.

## Frontend Değişiklikleri (Dashboard)
- Sayfa: `frontend/src/pages/ChatbotDetailPage.tsx` (rota `/chatbots/:id`, bileşen `ChatbotDetailPage`).
- State ve veri yükleme:
  - Yeni state: `chatBackgroundColor`, başlangıç `#FFF5E6`.
  - GET ile gelen `data.chat_background_color` varsa state’e yaz.
- Form alanı:
  - "Renkler" bölümüne "Chat Arka Plan" için color picker + text input ekle.
  - Kaydet payload’ına `chat_background_color: chatBackgroundColor` ekle (`handleSave`).
- Playground önizleme:
  - Mesajlar konteynerinin arka planını `style={{ background: chatBackgroundColor }}` ile uygula (mevcut `bg-gray-50` yerine).

## Widget Entegrasyonu
- Mevcut destek:
  - Widget, `config.chat_background_color` varsa `--cbw-chat-bg` CSS değişkenini set ediyor (`widget/src/widgetApp.tsx:50`).
  - URL ile `chat-bg-color` parametresi override edebiliyor (`widget/src/widget.tsx:57`).
- Yapılacak:
  - Public config çıktısına `chat_background_color` eklendiğinde widget otomatik kullanır; ekstra kod değişikliği gerekmez.

## Test ve Doğrulama
- Backend:
  - POST `/api/v1/chatbots` ile oluşturulan botta `chat_background_color` default `#FFF5E6` olduğunu doğrula.
  - PUT ile farklı renk gönderildiğinde kalıcılığını ve public endpoint’te görünmesini doğrula.
- Frontend:
  - Renk picker ile seçilen değer kaydedilir ve sayfa yenilendiğinde geri yüklenir.
  - Playground’da arka plan değişimini görsel olarak doğrula.
- Widget:
  - `GET /api/v1/public/chatbots/:id` yanıtında alanı gördüğünde `cbw-messages` arka planının değiştiğini kontrol et.

## Güvenlik ve Varsayılanlar
- Varsayılan krem `#FFF5E6` hem DB default hem de frontend başlangıç state’i olarak kullanılır.
- Renk input’u için basit hex doğrulaması ve trimming yapılır.
- Public endpoint sadece görsel alanları döner; business verisi sızdırılmaz (`internal/api/handlers/public.go:34`).

## Yayın ve Geriye Dönüş
- Migrasyonu çalıştır, sunucuyu yeniden başlat.
- Eski botlar default krem rengi alır; istenirse dashboard’tan güncellenir.
- Down migrasyon ile gerektiğinde kolonu kaldırıp önceki sürüme dönebilirsin.
