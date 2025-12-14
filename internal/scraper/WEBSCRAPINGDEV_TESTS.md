# Web Scraping Test Suite

Bu dizinde [web-scraping.dev](https://web-scraping.dev/) sitesindeki challenge'ları test etmek için 3 ayrı test dosyası bulunmaktadır.

> **⚠️ ÖNEMLİ:** Bu testler **normal test döngüsünde çalışmaz**. Sadece `webscraping` build tag'i ile manuel olarak çalıştırılabilir.

## Test Dosyaları

### 1. 🟢 Beginner Tests (`webscrapingdev_beginner_test.go`)
**Güvenli** - Bu testler IP bloğuna neden olmaz.

- ✅ Static Paging - Sunucu taraflı sayfalama
- ✅ Forced New Tab Links - Yeni sekmede açılan linkler
- ✅ Product HTML Markup - Temel HTML yapısı
- ✅ Cookie Popup - Cookie popup'ı
- ✅ Example Block Page - Block sayfası örneği

### 2. 🟡 Intermediate Tests (`webscrapingdev_intermediate_test.go`)
**Dikkatli** - Çoğu güvenli, bazıları dynamic scraping gerektirir.

- ⚙️ Endless Scroll Paging - Dinamik scroll sayfalama (JS gerekli)
- ⚙️ Secret API Token - Gizli API token'ı
- ⚙️ Endless Button Paging - "Load More" butonu (JS gerekli)
- ✅ Hidden Web Data - HTML içinde gizli JSON
- ⚙️ Local Storage - localStorage kullanımı (JS gerekli)
- ✅ Cookies Based Login - Cookie tabanlı login
- ✅ PDF Downloads - PDF indirme linkleri
- ✅ Form File Attachment Download - Form ile dosya indirme
- ✅ AI Content Obfuscation - Unicode ile gizlenmiş içerik

### 3. 🔴 Advanced Tests (`webscrapingdev_advanced_test.go`)
**⚠️ TEHLİKELİ** - Bu testler IP bloğuna neden olabilir!

- 🔥 GraphQL Background Requests - GraphQL API istekleri
- 🔥 CSRF Token Locks - CSRF token koruması
- 🔥 Blocking Redirect for Invalid Referer - Referer kontrolü
- ☢️ Persistent Cookie-Based Blocking - **Kalıcı cookie bloğu** (EN TEHLİKELİ!)

---

## Testleri Çalıştırma

### ⚠️ ÖNEMLİ: Build Tag Gereklidir

Bu testler normal `go test` komutunda **ÇALIŞMAZ**. Manuel olarak çalıştırmak için `-tags=webscraping` flag'i eklemeniz gerekir:

```bash
# ❌ ÇALIŞMAZ (testler görünmez)
go test ./internal/scraper/

# ✅ ÇALIŞIR
go test -tags=webscraping ./internal/scraper/
```

### Beginner Testleri

```bash
# Tüm beginner testleri
go test -v -tags=webscraping -run TestBeginner ./internal/scraper/

# Tek bir test
go test -v -tags=webscraping -run TestBeginner_StaticPaging ./internal/scraper/

# Hepsi bir arada
go test -v -tags=webscraping -run TestBeginner_AllChallenges ./internal/scraper/
```

### Intermediate Testleri

```bash
# Sadece statik testler (dynamic scraping olmadan)
go test -v -tags=webscraping -run "TestIntermediate_(SecretAPIToken|HiddenWebData|CookiesBasedLogin|PDFDownloads|FormFileAttachmentDownload|AIContentObfuscation)" ./internal/scraper/

# Sadece dynamic testler (headless browser gerekli)
go test -v -tags=webscraping -run "TestIntermediate_(EndlessScrollPaging|EndlessButtonPaging|LocalStorage)" ./internal/scraper/

# Tüm intermediate testleri
go test -v -tags=webscraping -run TestIntermediate_AllChallenges ./internal/scraper/

# Tek bir test
go test -v -tags=webscraping -run TestIntermediate_HiddenWebData ./internal/scraper/
```

### Advanced Testleri

```bash
# Sadece güvenli advanced testleri (ÖNERİLEN)
go test -v -tags=webscraping -run TestAdvanced_SafeOnly ./internal/scraper/

# Tüm advanced testleri (RİSKLİ!)
go test -v -tags=webscraping -run TestAdvanced_AllChallenges ./internal/scraper/

# Persistent blocking testini ATLA (önerilen)
go test -v -tags=webscraping -short -run TestAdvanced ./internal/scraper/

# Tek bir test
go test -v -tags=webscraping -run TestAdvanced_CSRFTokenLocks ./internal/scraper/
```

---

## Önemli Notlar

### ⚠️ Güvenlik Uyarıları

1. **Advanced testleri çalıştırmadan önce:**
   - VPN veya proxy kullanmayı düşünün
   - Rate limiting ayarlarını kontrol edin
   - Persistent blocking testini `-short` flag ile atlayın

2. **Persistent Blocking Testi:**
   ```bash
   # Bu testi ATLAMAK için -short kullanın
   go test -v -tags=webscraping -short -run TestAdvanced_PersistentCookieBasedBlocking ./internal/scraper/
   ```
   - Bu test kalıcı blocking cookie set edebilir
   - Sadece test ediyorsanız, cookies'i temizlemeye hazır olun
   - Production ortamında ÇALIŞTIRMAYIN

3. **Rate Limiting:**
   - Testler 1-2 saniye rate limit ile çalışır
   - Hepsini birden çalıştırmak yerine gruplar halinde çalıştırın

4. **CI/CD Entegrasyonu:**
   - Bu testler `webscraping` tag'i olmadan **asla çalışmaz**
   - Normal CI/CD pipeline'ınıza **etki etmez**
   - Manuel olarak veya ayrı bir job'da çalıştırabilirsiniz

---

## Test Sonuçları (Son Çalıştırma)

**Tarih:** 2025-12-14  
**Sonuç:** ✅ 17/17 PASSED (100%)

| Kategori | Testler | Durum | Süre |
|----------|---------|-------|------|
| Beginner | 5 | ✅ PASSED | 6.27s |
| Intermediate (Static) | 6 | ✅ PASSED | 9.86s |
| Intermediate (Dynamic) | 3 | ✅ PASSED | 19.94s |
| Advanced (Safe) | 3 | ✅ PASSED | 8.93s |
| Advanced (Risky) | - | ⚠️ SKIPPED | - |

**Toplam:** 45 saniye

---

## Gereksinimler

- **Static Scraping:** Colly (default)
- **Dynamic Scraping:** Rod + Chromium (bazı intermediate ve advanced testler için)

---

## Makefile Entegrasyonu (Opsiyonel)

Makefile'ınıza ekleyebilirsiniz:

```makefile
# Web scraping testleri (manuel)
.PHONY: test-webscraping
test-webscraping:
	@echo "Running web scraping tests (beginner + intermediate)..."
	go test -v -tags=webscraping -run "Test(Beginner|Intermediate)" ./internal/scraper/

# Sadece güvenli testler
.PHONY: test-webscraping-safe
test-webscraping-safe:
	@echo "Running safe web scraping tests..."
	go test -v -tags=webscraping -run "TestBeginner|TestIntermediate_(SecretAPIToken|HiddenWebData|CookiesBasedLogin|PDFDownloads|FormFileAttachmentDownload|AIContentObfuscation)" ./internal/scraper/

# Advanced testler (dikkatli!)
.PHONY: test-webscraping-advanced
test-webscraping-advanced:
	@echo "⚠️  WARNING: Running advanced tests..."
	go test -v -tags=webscraping -run TestAdvanced_SafeOnly ./internal/scraper/
```

Kullanım:
```bash
make test-webscraping
make test-webscraping-safe
make test-webscraping-advanced
```

---

## 🔧 Troubleshooting

**"no tests to run" hatası:**
```bash
# ❌ Yanlış
go test -run TestBeginner ./internal/scraper/

# ✅ Doğru
go test -tags=webscraping -run TestBeginner ./internal/scraper/
```

**"domain_not_allowed" hatası:**
```bash
# .env dosyasına ekleyin:
export SCRAPER_ALLOWED_DOMAINS=web-scraping.dev
```

**"context deadline exceeded" hatası:**
- Timeout süresini artırın
- İnternet bağlantınızı kontrol edin
- VPN kullanıyorsanız kapatmayı deneyin

**Dynamic scraping çalışmıyor:**
- Rod ve Chromium yüklü olduğundan emin olun
- Headless browser testlerini deneyin: `go test -tags=webscraping -v -run TestBrowser`

**Block edildiniz mi?**
- Cookies'i temizleyin
- IP adresinizi değiştirin (VPN/proxy)
- 24 saat bekleyin

---

## 📚 Kaynaklar

- [web-scraping.dev](https://web-scraping.dev/) - Test platformu
- [Colly Documentation](http://go-colly.org/)
- [Rod Documentation](https://go-rod.github.io/)
- [Go Build Tags](https://pkg.go.dev/cmd/go#hdr-Build_constraints)

---

## 📝 Notlar

- Bu testler **eğitim amaçlıdır** - web scraper'ımızın farklı senaryolara nasıl yanıt verdiğini anlamak için
- Production'da gerçek siteleri scrape ederken **her zaman site kurallarına uyun** (robots.txt, ToS)
- Rate limiting ve User-Agent rotasyonu kullanın
- Hata yönetimi ve retry logic ekleyin
- **Normal test döngünüzde otomatik çalışmazlar** (build tag sayesinde)
