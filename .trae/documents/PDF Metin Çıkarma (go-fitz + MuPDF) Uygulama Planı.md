## Amaç
- PDF dosyalarından yüksek doğrulukta metin çıkarmak ve UTF-8/Türkçe karakter korunumunu sağlamak.
- `internal/pdf/extractor.go` içinde `ExtractPDFText(filePath string) (string, error)` fonksiyonunu hayata geçirmek.

## Dosya/Dizin Yapısı
- `internal/pdf/extractor.go`: Metin çıkarma ana fonksiyonu ve yardımcılar.
- Mevcut `internal/scraper/encoding.go`: UTF-8 validasyon/normalizasyon yardımcılarını yeniden kullan.
- (İleride) `internal/pdf/ocr.go`: Görsel tabanlı fallback (Tesseract) için yer.

## Bağımlılıklar ve Derleme
- Go modülüne `github.com/gen2brain/go-fitz` eklenir.
- macOS: `brew install mupdf`.
- Linux (Docker dahil): `apt-get install -y libmupdf-dev pkg-config`.
- Gerekirse build tag: `extlib` ve `pkgconfig` kullanımı; CJK/Türkçe fontlar için sistem kütüphanesi tercih edilir.
- Not: Aynı dokümanda `Text()`/`Image()` çağrılarının eşzamanlı kullanımı desteklenmez; sayfalar seri işlenecek.

## Uygulama Adımları (extractor)
1. Dosya açılışı
   - `doc, err := fitz.New(filePath)`; `defer doc.Close()`.
   - Hata türlerini ayrıştır: `fitz.ErrOpenDocument`, `fitz.ErrNeedsPassword` vb.
2. Sayfa doğrulama
   - `pages := doc.NumPage()`; `pages < 1` ise anlamlı bir hata döndür.
3. Sayfa bazlı metin çıkarımı
   - Hızlı yol: `text, err := doc.Text(n)`; sayfa sonuna satır sonu ekle.
   - Düzen-koruma (isteğe bağlı): `html, err := doc.HTML(n, false)` ile stil/koordinatları parse edip `top` (Y) değerine göre satır gruplama, `left` (X) ile sıralama; blokları birleştirme. Büyük tablolar/çok sütunlu düzenlerde daha tutarlı sonuç verir.
   - Her sayfa için `strings.Builder` ile biriktir.
4. UTF-8/Türkçe karakter işleme
   - `bytes.ToValidUTF8([]byte(s), []byte("?"))` ile geçersiz baytları düzelt.
   - Mevcut `internal/scraper/encoding.go` yardımcılarını çağırarak UTF-8’e dönüştürme ve normalizasyon (NFC) uygula.
5. Çıktı
   - Sayfalar arası ayraç olarak `\n\n` kullan; toplam metni döndür.

## Hata Yönetimi
- Dosya açılamadı/eksik: kaynak yolu ve sarmalanmış hata ile döndür.
- Şifreli PDF: `fitz.ErrNeedsPassword`; kullanıcıya uygun mesaj.
- Bozuk PDF/sayfa yüklenemedi: ilgili `fitz` hata türlerini aynen yüzeye çıkar.
- Sayfa < 1: özel hata ile.
- Kaynak serbest bırakma: `defer doc.Close()` garanti edilir.

## Test ve Doğrulama
- Unit testler (`internal/pdf/extractor_test.go`):
  - Küçük örnek PDF ile `ExtractPDFText` doğrulaması.
  - Türkçe karakterler (`ş,ğ,İ,ı,Ö,Ç`) kontrolü; UTF-8 geçerlilik ve NFC.
  - Bozuk PDF ve şifreli PDF negatif senaryoları.
- Performans testleri: 100+ sayfalı PDF’de süre ve bellek izleme.

## Entegrasyon Noktaları
- `internal/processing/sources_queue.go` içindeki `"pdf"` dalında:
  - Yüklenen PDF yolunu `ExtractPDFText` ile işle.
  - Dönüş metnini mevcut içerik indeksleme/embedding pipeline’ına aktar.
  - İşlem durumlarını (started, completed, failed) güncelle.
- API katmanı (`internal/api/handlers/source.go`): yükleme sonrası kuyruğa PDF işleme işi ekli.

## Performans ve Sınırlar
- Büyük PDF’lerde sayfa sayfa işleme ve builder kullanımı ile düşük bellek ayak izi.
- Aynı dokümanda eşzamanlı `Text()` çağrıları yapılmaz.
- Çok sütunlu mizanpajlar için HTML tabanlı pozisyon-parsing seçeneği.

## Fallback (Opsiyonel)
- go-fitz başarısız olursa veya çıktı çok zayıfsa OCR fallback (`internal/pdf/ocr.go`, Tesseract/gosseract) devreye alınır.

## Dev/CI Ortamı
- macOS: `brew install mupdf` dokümante edilir.
- Docker: `libmupdf-dev` ve `pkg-config` paketi eklenir; derleme aşaması öncesi cache optimize edilir.
- CI’de örnek PDF ile smoke test koşulur.