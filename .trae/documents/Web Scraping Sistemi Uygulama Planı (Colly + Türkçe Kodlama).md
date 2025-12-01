## Genel Yaklaşım
- Amaç: Belirtilen URL’lerden görünür metni güvenli ve doğru kodlamayla çekmek, normalize etmek, cache’lemek ve arka plan iş kuyruğu ile ölçekli şekilde işlemek.
- Dil/Stack: Go (mevcut), kütüphaneler: `colly`, `goquery`, `x/net/html/charset`, opsiyonel `redis`.
- Konum: `internal/scraper/*` altında modüler paketler; mevcut `pkg/logger`, `pkg/config` kalıplarını kullanma.

## Bağımlılıklar
- `github.com/gocolly/colly` – HTTP tarayıcı ve crawling.
- `github.com/PuerkitoBio/goquery` – HTML parse/görünür metin çıkarımı.
- `golang.org/x/net/html/charset` – otomatik kodlama tespiti ve UTF-8’e dönüşüm.
- `github.com/redis/go-redis/v9` – Cache için Redis (opsiyonel, yoksa in-memory).

## Dosya Yapısı
- `internal/scraper/colly.go` – Colly collector kurulumu ve callback’ler.
- `internal/scraper/encoding.go` – Türkçe kodlama normalizasyon yardımcıları.
- `internal/scraper/worker.go` – Task tipi, `ScrapeURL` ve iş akışı.
- `internal/scraper/cache.go` – Redis + in-memory cache adaptörü.
- Entegrasyon: `internal/processing/sources_queue.go` içinde job üretimi/tüketimi; `cmd/server/main.go` içinde opsiyonel HTTP endpoint’leri.

## Colly Kurulumu (`internal/scraper/colly.go`)
- Collector oluştur: `colly.NewCollector()`
- `AllowedDomains` ayarla: sadece hedef domain’ler.
- UserAgent seti: 5–10 farklı UA; her istek için rastgele seçim.
- `Timeout`: 30 saniye.
- `RateLimit`: 2 req/sn (domain-based `LimitRule`).
- Callback’ler:
  - `OnHTML("body", ...)`: görünür metni çıkar (script/style hariç, `goquery` ile).
  - `OnError`: hataları `pkg/logger` ile kaydet.
  - `OnScraped`: istek tamamlandı işaretle.

## Türkçe Kodlama (`internal/scraper/encoding.go`)
- `NormalizeText(rawHTML string) (string, error)`:
  - `charset.NewReader` ile otomatik tespit+dönüştür; daima UTF-8.
  - UTF-8 validation; invalid karakterleri güvenli değişimle (� yerine boşluk/silme) ele al.
  - BOM temizleme.
- `IsValidUTF8(data []byte) bool` – hızlı kontrol.

## Worker ve Akış (`internal/scraper/worker.go`)
- `ScrapingTask`: `URL`, `ChatbotID`, `SourceID` alanları.
- `ScrapeURL(task ScrapingTask) (string, error)`:
  - Cache kontrol: `scraped:{md5(url)}`; 7 gün geçerlilik.
  - Colly ile fetch; `goquery` ile görünür metni çıkar.
  - `NormalizeText` ile temizle.
  - Sonucu döndür ve cache’e yaz.
- Görünür metin çıkarımı:
  - `script`, `style`, `noscript` düğümlerini hariç tut.
  - `display:none` ve `hidden` sınıflarını opsiyonel filtrele.
  - Beyaz boşlukları normalize et; paragraf sınırlarını koru.

## Cache (`internal/scraper/cache.go`)
- Arayüz: `Get(key) (string, bool)`, `Set(key, val string, ttl time.Duration) error`.
- Uygulama:
  - Redis varsa `REDIS_URL` ile bağlan; yoksa in-memory LRU/TTL.
  - Key: `scraped:{md5(url)}`; TTL: 7 gün.

## Background Job Queue Entegrasyonu
- Worker goroutine’ler: sabit boyutlu worker pool.
- Kanal: `chan ScrapingTask` üzerinden üretici/tüketici.
- `internal/processing/sources_queue.go` ile köprü:
  - Kaynak eklendiğinde veya güncellendiğinde scraping task üret.
  - İş tamamlandığında durum/istatistik güncelle.

## Konfigürasyon
- `pkg/config`: `AllowedDomains`, `RateLimit`, `Timeout`, `RedisURL`, `UAList` parametreleri.
- Varsayılanlar ve environment override.

## Gözlemlenebilirlik
- `pkg/logger`: istek, hata, sonuç uzunluğu ve süre ölçümü.
- Metrikler: istek sayısı, hata oranı, cache hit/miss.

## Güvenlik ve Uyumluluk
- Robots.txt saygısı (isteğe bağlı; `colly` ile `RespectRobotsTxt`).
- Domain whitelist; harici sitelere yayılmayı engelle.
- Büyük sayfa ve binary içeriklere karşı koruma (boyut sınırları ve `Content-Type` kontrolü).

## Testler
- Unit test: `NormalizeText`, `IsValidUTF8`, görünür metin çıkarımı.
- Integration test: örnek HTML’ler (Türkçe ISO-8859-9/Windows-1254), cache davranışı, rate-limit.

## Endpoint/İç API (opsiyonel)
- `GET /scrape?url=...` yalnızca authorized kullanıcılar için (rate-limit ve domain kontrolü ile).
- İç servisler için `internal/scraper` paket fonksiyonları kullanımı.

## Adım Adım Uygulama
1. Bağımlılıkları `go.mod`’a ekle.
2. `internal/scraper/colly.go` içinde collector ve callback’leri yaz.
3. `internal/scraper/encoding.go` fonksiyonlarını uygula.
4. `internal/scraper/cache.go`’yu Redis + in-memory ile yaz.
5. `internal/scraper/worker.go`’da `ScrapingTask` ve `ScrapeURL` akışını bitir.
6. `internal/processing/sources_queue.go` ile entegrasyonu ekle.
7. Konfigürasyon ve logger entegrasyonunu tamamla.
8. Testleri yaz ve çalıştır; örnek sayfalarla doğrula.

## Sonraki Geliştirmeler
- Dinamik siteler için `chromedp` alternatifi.
- Dil tespiti ve dil-özel temizleme.
- Ekstra deduplication: içerik hash’i ve değişim algılama.
- Failover/Retry stratejileri ve backoff.