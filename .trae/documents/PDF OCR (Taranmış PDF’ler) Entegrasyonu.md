## Mevcut Durum
- PDF metin çıkarımı `go-fitz` ile hazır: `internal/pdf/extractor_fitz.go:18` → `ExtractPDFText(filePath string)` sayfa bazlı çıkarım yapıyor.
- İşleme hattı `pdf.ExtractPDFText` çağırıyor: `internal/processing/sources_queue.go:86`.
- OCR/Tesseract entegrasyonu yok; dokümandaki hedef dosya `internal/pdf/ocr.go` mevcut değil.

## Hedefler
1. `internal/pdf/ocr.go` ekle ve `ExtractPDFWithOCR(filePath string) (string, error)` implemente et.
2. `go-fitz` ile her sayfayı 300 DPI image’a render et ve `gosseract` ile OCR yap.
3. Türkçe dil paketi (`tur`) kullan, gerektiğinde `psm/oem` ve `user_defined_dpi=300` ayarları.
4. Fallback: `ExtractPDFText` sonucu kısa (<100 karakter) ise OCR’a geç ve sayfa metinlerini birleştir.
5. Hata/operasyon: Tesseract yoksa veya dil paketi eksikse anlamlı hata ve log; kullanıcı bildirimi.

## Teknik Tasarım
- Bağımlılıklar: `github.com/otiai10/gosseract/v2` (Go), mevcut `github.com/gen2brain/go-fitz`.
- Sayfa render:
  - `img, err := doc.ImageDPI(page, 300)` (bkz. go-fitz API). Gerekirse `ImagePNG(page, 300)` ile PNG baytları.
- OCR istemci:
  - `c := gosseract.NewClient(); defer c.Close()`
  - `c.SetLanguage("tur")`
  - `c.SetVariable("user_defined_dpi", "300")` (DPI uyarılarını önlemek için)
  - Tercihen: `c.SetPageSegMode(gosseract.PSM_AUTO)` veya `PSM_SPARSE_TEXT` taranmış formlar için.
  - Görseli aktar: `SetImageFromBytes(pngBytes)` veya geçici dosya ile `SetImage(path)`.
- Metin birleştirme:
  - Her sayfa için OCR sonucu `strings.TrimSpace` ile normalize edilip `\n\n` ile birleştirilecek.
- UTF-8 ve normalizasyon:
  - Mevcut `scraper.NormalizeText` ve `IsValidUTF8` yardımcıları yeniden kullanılacak.

## Fallback Entegrasyonu
- `internal/pdf/extractor_fitz.go:18` içindeki çıkarım sonrası toplam uzunluk kontrolü.
- Kısa metinde `ExtractPDFWithOCR` çağrısı yap; başarılıysa OCR sonucunu döndür, aksi halde özgün metinle devam.
- `internal/processing/sources_queue.go:86` akışı değişmeden kalır; `pdf.ExtractPDFText` içindeki fallback tüm PDF’ler için otomatikleşir.

## Hata Yönetimi ve Kontroller
- Tesseract kurulumu kontrolü: `tesseract --list-langs` çıktı doğrulaması (opsiyonel, sadece bilgilendirici).
- Dil paketi yoksa: açık hata mesajı ve log; `Source` kaydı `failed` olarak güncellenir.
- OCR başarısızsa: hata yakalanır, loglanır; dokümandaki yönergeye uygun kullanıcı bildirimi yapılır.

## Testler
- Birim testi: taranmış örnek PDF ile `ExtractPDFWithOCR` sonucu boş değil.
- Ortam değişkeni destekli test: mevcut `BOTLA_PDF_PATH` yaklaşımının benzeri ile OCR testi (`//go:build fitz`).
- Kenar durumları: bozuk PDF, çok sayfalı, düşük kaliteli taramalar (PSM varyasyonları).

## Operasyonel Gereksinimler
- Sunucuda Tesseract kurulu olmalı; Türkçe dil paketi: `tesseract-ocr-tur`.
- Build: `go build -tags fitz ./...` (MuPDF kurulumuna bağlı).
- Performans: sayfa başına OCR pahalıdır; ileride işçi sayısı ve zaman aşımı ayarları eklenebilir.

## Sonuç
- Dokümandaki 6.2 maddesini bire bir karşılayan `ocr.go` ve fallback akışı eklenecek.
- Uygulama, metni olmayan/taranmış PDF’lerde otomatik OCR yaparak kaliteyi artıracak.

Onaylarsanız uygulamayı hemen ekleyip testleri çalıştıracağım.