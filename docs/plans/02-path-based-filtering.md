# Plan 1.2: Path Tabanlı Include/Exclude Filtreleme

## Özet

Web sitesi tarama sırasında kullanıcının belirli URL path'lerini dahil etmesi veya hariç tutması için filtreleme sistemi.

---

## Mevcut Durum

### İlgili Dosyalar

| Dosya | Mevcut Durum |
|-------|--------------|
| `internal/models/chatbot.go` | Path alanları yok |
| `internal/scraper/worker.go` | `ExtractLinks` filtresiz |
| `internal/processing/url_processor.go` | `discoverSubPages` tüm linkleri ekler |

### Mevcut ExtractLinks Fonksiyonu

```go
// internal/scraper/worker.go:87-138
func ExtractLinks(htmlContent string, baseURL string) ([]string, error) {
    // Sadece aynı domain kontrolü var
    // Path filtreleme YOK
}
```

---

## Hedef Mimari

```
Kullanıcı Tanımlar:
┌────────────────────────────────────────┐
│ include_paths: ["/blog/*", "/docs/*"]  │
│ exclude_paths: ["/admin/*", "/tag/*"]  │
└────────────────────────────────────────┘
                    │
                    ▼
┌────────────────────────────────────────┐
│         Path Filter Engine             │
│  ─────────────────────────────────────│
│  1. Glob pattern matching              │
│  2. Priority: exclude > include        │
│  3. Boş include = hepsini al           │
└────────────────────────────────────────┘
                    │
                    ▼
┌────────────────────────────────────────┐
│      Filtrelenmiş URL Listesi          │
└────────────────────────────────────────┘
```

---

## Uygulama Adımları

### Adım 1: Veritabanı Migration

**Dosya:** `db/migrations/000009_path_filters.up.sql`

```sql
ALTER TABLE chatbots 
ADD COLUMN IF NOT EXISTS include_paths TEXT[] DEFAULT '{}',
ADD COLUMN IF NOT EXISTS exclude_paths TEXT[] DEFAULT '{}';

COMMENT ON COLUMN chatbots.include_paths IS 'Glob patterns for paths to include (empty = all)';
COMMENT ON COLUMN chatbots.exclude_paths IS 'Glob patterns for paths to exclude';
```

**Dosya:** `db/migrations/000009_path_filters.down.sql`

```sql
ALTER TABLE chatbots 
DROP COLUMN IF EXISTS include_paths,
DROP COLUMN IF EXISTS exclude_paths;
```

### Adım 2: Model Güncelleme

**Dosya:** `internal/models/chatbot.go`

**Eklenecek alanlar:**
```go
IncludePaths []string `json:"include_paths,omitempty"`
ExcludePaths []string `json:"exclude_paths,omitempty"`
```

### Adım 3: Path Filter Utility Oluştur

**Dosya:** `internal/scraper/path_filter.go` (YENİ)

**Fonksiyonlar:**

```
PathFilter struct:
  - includePaths: []string
  - excludePaths: []string
  - compiledInclude: []*regexp.Regexp
  - compiledExclude: []*regexp.Regexp

Methodlar:
  - NewPathFilter(include, exclude []string) (*PathFilter, error)
  - (f *PathFilter) Match(urlPath string) bool
  - (f *PathFilter) FilterURLs(urls []string) []string
```

**Pattern Kuralları:**
- `*` → herhangi bir karakter dizisi (regex: `.*`)
- `/blog/*` → `/blog/foo`, `/blog/bar/baz` eşleşir
- `/docs/v1` → tam eşleşme
- Büyük/küçük harf duyarsız

### Adım 4: ExtractLinks Fonksiyonunu Güncelle

**Dosya:** `internal/scraper/worker.go`

**Değişiklik:** `ExtractLinks` signature'ını güncelle

```go
// Eski
func ExtractLinks(htmlContent string, baseURL string) ([]string, error)

// Yeni
func ExtractLinks(htmlContent string, baseURL string, filter *PathFilter) ([]string, error)
```

**İç mantık:**
1. Mevcut link çıkarma mantığını koru
2. Her link için `filter.Match(parsedURL.Path)` kontrolü ekle
3. `filter == nil` ise tüm linkleri döndür (geriye uyumluluk)

### Adım 5: URLProcessor'ı Güncelle

**Dosya:** `internal/processing/url_processor.go`

**Değişiklik:** `discoverSubPages` fonksiyonunda filter kullan

```go
func (p *URLProcessor) discoverSubPages(ctx context.Context, s *models.DataSource, bot *models.Chatbot, plan *models.Plan, content string) {
    // PathFilter oluştur
    filter, _ := scraper.NewPathFilter(bot.IncludePaths, bot.ExcludePaths)
    
    // Links'i filtrele
    links, lerr := scraper.ExtractLinks(content, *s.SourceURL, filter)
    // ... geri kalan mantık aynı
}
```

### Adım 6: API Endpoint Güncelleme

**Dosya:** `internal/api/handlers/chatbot.go`

**Değişiklikler:**

1. `CreateChatbot` handler'ında `include_paths` ve `exclude_paths` kabul et
2. `UpdateChatbot` handler'ında güncelleme

**Request body:**
```json
{
  "name": "My Bot",
  "include_paths": ["/blog/*", "/docs/*"],
  "exclude_paths": ["/admin/*", "/tag/*", "/author/*"]
}
```

### Adım 7: Frontend Değişiklikleri

**Dosya:** `frontend/src/features/chatbot/ChatbotSettings.tsx`

**Eklenecek UI:**

```
┌─────────────────────────────────────────┐
│ URL Filtreleme Ayarları                 │
├─────────────────────────────────────────┤
│ ✅ Dahil Edilecek Yollar (Include)      │
│ ┌─────────────────────────────────────┐ │
│ │ /blog/*                         [x] │ │
│ │ /docs/*                         [x] │ │
│ │ + Yeni ekle...                      │ │
│ └─────────────────────────────────────┘ │
│                                         │
│ ❌ Hariç Tutulacak Yollar (Exclude)     │
│ ┌─────────────────────────────────────┐ │
│ │ /admin/*                        [x] │ │
│ │ /tag/*                          [x] │ │
│ │ + Yeni ekle...                      │ │
│ └─────────────────────────────────────┘ │
│                                         │
│ ℹ️ Örnek: /blog/* tüm blog yazılarını   │
│    dahil eder                           │
└─────────────────────────────────────────┘
```

**Dosyalar:**
- `frontend/src/features/chatbot/PathFilterInput.tsx` (YENİ)
- `frontend/src/api/chatbot.ts` (GÜNCELLE - tipler)

---

## Dosya Değişiklikleri Özeti

| Dosya | İşlem | Açıklama |
|-------|-------|----------|
| `db/migrations/000009_*.sql` | YENİ | Migration |
| `internal/models/chatbot.go` | GÜNCELLE | Yeni alanlar |
| `internal/scraper/path_filter.go` | YENİ | Filter engine |
| `internal/scraper/path_filter_test.go` | YENİ | Unit testler |
| `internal/scraper/worker.go` | GÜNCELLE | ExtractLinks |
| `internal/processing/url_processor.go` | GÜNCELLE | Filter kullan |
| `internal/api/handlers/chatbot.go` | GÜNCELLE | API |
| `frontend/src/features/chatbot/*` | GÜNCELLE | UI |

---

## Test Planı

### Unit Testler

**Dosya:** `internal/scraper/path_filter_test.go` (YENİ)

**Test Senaryoları:**

```go
func TestPathFilter_Match(t *testing.T) {
    testCases := []struct {
        name     string
        include  []string
        exclude  []string
        path     string
        expected bool
    }{
        // Include testleri
        {"include all when empty", nil, nil, "/any/path", true},
        {"include glob match", []string{"/blog/*"}, nil, "/blog/post-1", true},
        {"include no match", []string{"/blog/*"}, nil, "/docs/intro", false},
        
        // Exclude testleri
        {"exclude takes priority", []string{"/*"}, []string{"/admin/*"}, "/admin/users", false},
        {"exclude partial match", nil, []string{"/tag/*"}, "/tag/golang", false},
        
        // Edge cases
        {"exact match", []string{"/about"}, nil, "/about", true},
        {"trailing slash", []string{"/blog/*"}, nil, "/blog/", true},
    }
}
```

**Komut:** `go test ./internal/scraper/... -v -run TestPathFilter`

### Integration Testler

**Dosya:** `internal/scraper/worker_test.go`

**Test:**
```go
func TestExtractLinks_WithFilter(t *testing.T) {
    // HTML içeriği hazırla
    // Filter ile ExtractLinks çağır
    // Filtrelenmiş sonuçları doğrula
}
```

**Komut:** `go test ./internal/scraper/... -v -run TestExtractLinks`

### API Testleri

**Dosya:** `internal/api/handlers/chatbot_test.go`

**Test:**
```go
func TestCreateChatbot_WithPathFilters(t *testing.T) {
    // POST /chatbots with include_paths and exclude_paths
    // Verify saved correctly
}
```

### Manuel Test Prosedürü

1. **Setup:**
   - Dashboard'a giriş yap
   - Yeni bir chatbot oluştur

2. **Path Filter Test:**
   - Ayarlar → URL Filtreleme
   - Include: `/blog/*` ekle
   - Exclude: `/tag/*` ekle
   - Kaydet

3. **Crawl Test:**
   - Yeni URL kaynağı ekle: `https://example.com`
   - Tarama tamamlanınca kaynakları kontrol et
   - Sadece `/blog/` altındaki sayfaların eklendiğini doğrula
   - `/tag/` sayfalarının eklenmediğini doğrula

---

## Doğrulama Kriterleri

| Kriter | Doğrulama Yöntemi |
|--------|-------------------|
| ✅ Migration başarılı | `make migrate-up` |
| ✅ Model alanları çalışıyor | Unit test |
| ✅ Filter engine doğru | Path filter testleri |
| ✅ ExtractLinks filtreli | Integration test |
| ✅ API kabul ediyor | API test |
| ✅ Frontend UI çalışıyor | Manuel test |
| ✅ %90 coverage | `make cover-gate` |
| ✅ Lint temiz | `make lint` |

---

## Edge Cases

| Durum | Beklenen Davranış |
|-------|-------------------|
| Boş include, boş exclude | Tüm URL'leri al |
| Include var, exclude yok | Sadece include'a uyanları al |
| Include yok, exclude var | Exclude dışındakileri al |
| Hem include hem exclude | Önce include filtrele, sonra exclude uygula |
| Geçersiz pattern | Hata döndür, logla |

---

## Tahmini Süre

| Görev | Süre |
|-------|------|
| Migration | 30 dk |
| Model update | 30 dk |
| Path filter engine | 3-4 saat |
| ExtractLinks update | 1-2 saat |
| URLProcessor update | 1 saat |
| API update | 1-2 saat |
| Frontend UI | 3-4 saat |
| Testler | 3-4 saat |
| **TOPLAM** | **~1 hafta** |

---

## Sonraki Adımlar

Bu plan tamamlandığında:
- ✅ [1.3 CSS Selector](./03-css-selector-scraping.md) için temel hazır
- ✅ [1.4 Sitemap Parser](./04-sitemap-parser.md) için temel hazır
- ✅ [1.5 URL Checkbox UI](./05-url-checkbox-ui.md) için temel hazır
