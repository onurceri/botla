## Amaç
- Statik taramada görünmeyen içerikleri (JS ile sonradan yüklenenler) headless browser ile almak.
- Kaynak tüketimini kontrol etmek (pool ve tab sınırlaması), timeouts ve fallback stratejisi.

## Bağımlılıklar
- `github.com/go-rod/rod` ve opsiyonel `github.com/go-rod/stealth` (anti-bot önlemleri için).
- Chrome/Chromium sistemde kurulu olmalı (varsayılan headless).

## Dosyalar
- `internal/scraper/browser.go`:
  - Browser pool yöneticisi (`type BrowserPool`) ve `NewBrowserPool(size int, idleTTL time.Duration)`.
  - `ScrapeDynamicURL(url string) (string, error)` – tek işlevsel API.
  - Yardımcılar: `normalizeHTMLToText(html string) (string)`, `isAllowedDomain(url string, allowed []string) bool`.

## Browser Pool Tasarımı
- `poolSize` (env: `SCRAPER_BROWSER_POOL_SIZE`, varsayılan: 2).
- Her iş için: bir tarayıcı seçilir (round-robin) ve yeni tab açılır (`browser.Page`).
- Tab sayısı kontrolü bir semaphore ile (max 2 concurrent toplam); job bittiğinde `page.Close()`.
- Idle tarayıcı reaper: `idleTTL` (env: `SCRAPER_DYNAMIC_IDLE_SECS`, varsayılan: 60s) aşıldığında tarayıcıyı kapat.

## ScrapeDynamicURL
- Domain whitelist kontrolü (env: `SCRAPER_ALLOWED_DOMAINS`).
- Navigation timeout: 10s (env: `SCRAPER_NAV_TIMEOUT_MS`).
- Adımlar:
  1. Pool’dan browser, yeni tab aç.
  2. `page.Navigate(url)` ve timeout.
  3. `page.WaitLoad()` + kısa `WaitIdle()` veya `Evaluate` ile `document.readyState` kontrolü.
  4. `page.HTML()` ile rendered DOM HTML al.
  5. `page.Close()` ve sonucu normalize edip döndür.

## Resource Management
- Her tab için `defer page.Close()`; panic safe.
- Network idle bekleme süresi üst sınır: 1–2s.
- Büyük içeriklere karşı `len(html)` üst limit kontrolü (örn. 2MB) ve truncate.

## Fallback Stratejisi
- Yeni fonksiyon: `ScrapeURLWithFallback(task ScrapingTask, cfg CollectorConfig) (string, error)`.
  - Adım 1: Statik `ScrapeURL` (Colly) dene.
  - Adım 2: Eğer boş/hata: `ScrapeDynamicURL(task.URL)` dene.
  - Adım 3: Hala boşsa: boş içerik ve hata logla.
- Cache ve TTL mantığı statik ile aynı key altında (dinamik sonuç da cache’lenir).

## Konfigürasyon
- `SCRAPER_BROWSER_POOL_SIZE` (int, default 2)
- `SCRAPER_DYNAMIC_IDLE_SECS` (int, default 60)
- `SCRAPER_NAV_TIMEOUT_MS` (int, default 10000)
- `SCRAPER_ALLOWED_DOMAINS` (statikle ortak)
- `SCRAPER_DYNAMIC_ENABLED` (bool, default true)

## Güvenlik
- Domain whitelist zorunlu; liste dışı URL reddedilir.
- Kullanıcı-agent ve ayak izi azaltma için `stealth` opsiyonel.
- Yüksek kaynak kullanımı ve sınırsız tab açımını engelleyen semaphore/pool.

## Testler
- Dinamik içerik testi (httptest server): JS ile body’yi değiştiren sayfa; statik boş, dinamik dolu.
- Timeout testi: 10s timeout ile başarısız navigasyon.
- Pool ve concurrency testi: 3 eşzamanlı istek → yalnızca 2 paralel tab açılır.

## Entegrasyon
- `internal/processing/sources_queue.go`: `url` tipinde kaynaklar için `ScrapeURLWithFallback` kullanımı.
- Loglama `pkg/logger` ile; cache `scraper/cache.go` ile.

## Adımlar
1. Bağımlılık ekle: `rod` (+ opsiyonel `stealth`).
2. `internal/scraper/browser.go` – pool ve `ScrapeDynamicURL`.
3. `ScrapeURLWithFallback` ekle ve `sources_queue.go` içinde kullan.
4. Konfigürasyon env değişkenleri.
5. Testler ve çalıştırma.

## Sonraki Geliştirmeler
- Dinamik bekleme stratejilerinde `wait for selector` kullanımı.
- CAPTCHA tespiti ve bypass politikaları (etik sınırlar dahilinde).
- Render sonrası görünür metnin daha iyi temizlenmesi (CSS computed style değerlendirmek).