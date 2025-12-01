## Genel Yol Haritası
1. pkg/logger/logger.go: JSON stdout logger ve seviye fonksiyonları
2. internal/db/db.go: PostgreSQL `*sql.DB` havuzu, ping ve limitler
3. cmd/server/main.go: config yükleme, logger+DB init, HTTP router, `/health`
4. Qdrant sağlık kontrolü: `GET {QDRANT_URL}/healthz` ile doğrulama
5. Graceful shutdown: SIGINT/SIGTERM yakala, `server.Shutdown` ve `db.Close`

## Logger (pkg/logger/logger.go)
- Amaç: Stdout'a tek satır JSON yazan, seviyeli bir logger.
- API:
  - `type Logger struct{ level string }`
  - `func New(level string) *Logger`
  - `func (l *Logger) Debug(msg string, fields map[string]any)`
  - `Info`, `Warn`, `Error` aynı imza
- Çıktı şeması: `{ "ts": RFC3339, "level": "INFO", "msg": "...", "fields": { ... } }`
- Uygulama: `encoding/json` ve `os.Stdout` kullan; seviye filtrelemesi basit string önceliği ile.

## Database (internal/db/db.go)
- Amaç: Tek bir `New` fonksiyonu ile `*sql.DB` havuzu oluşturmak.
- Sürücü: Postgres için `github.com/jackc/pgx/v5/stdlib` öneriyorum (veya `lib/pq`).
- API:
  - `func New(cfg *config.Config) (*sql.DB, error)`
- Adımlar:
  - DSN: `postgres://user:pass@host:port/dbname?sslmode=disable`
  - `sql.Open("pgx", dsn)` (pgx stdlib kullanıyorsak)
  - `db.SetMaxOpenConns(25)`, `db.SetMaxIdleConns(5)`
  - Ping testi: `ctx, 2s timeout` ile `db.PingContext(ctx)`

## Sunucu (cmd/server/main.go)
- Amaç: Config+Logger+DB init, HTTP server ve `/health` endpoint.
- Adımlar:
  - `cfg := config.LoadConfig()` (`pkg/config/config.go:23`)
  - `log := logger.New("INFO")`
  - `db := db.New(cfg)` ve hata yönetimi
  - Router: stdlib `http` ile `http.NewServeMux()`
  - `/health` handler:
    - DB ping: `PingContext(1s)`
    - Qdrant check: `GET {QDRANT_URL}/healthz` (`api-key` varsa header)
    - Yanıt: `200` `{ "status": "ok" }` ya da `503` ile detaylar
  - `srv := &http.Server{ Addr: ":"+cfg.PORT, Handler: mux }`
  - `go srv.ListenAndServe()`; hata `http.ErrServerClosed` dışında logla

## Qdrant Sağlık Kontrolü
- İstek: `GET {QDRANT_URL}/healthz`
- Header: `api-key` gerekiyorsa ekle (`QDRANT_API_KEY` opsiyonel).
- Timeout: `http.Client{Timeout: 2 * time.Second}`.
- Başarılı: `200`; aksi halde hata döndür.

## Graceful Shutdown
- `signal.Notify` ile `SIGINT`, `SIGTERM` yakala.
- `ctx, 5s timeout` ile `srv.Shutdown(ctx)` çağır.
- `db.Close()` ve kapanış logla.

## Örnek Kod Parçaları
- Logger kullanımı:
  - `log.Info("server_start", map[string]any{"port": cfg.PORT})`
- Health handler yanıtı:
  - Başarılı: `w.WriteHeader(http.StatusOK); json.NewEncoder(w).Encode(map[string]string{"status": "ok"})`
  - Başarısız: `w.WriteHeader(http.StatusServiceUnavailable)` ve hangi bağımlılık çöktüğünü döndür.

## Bağımlılıklar
- Postgres sürücüsü: `github.com/jackc/pgx/v5/stdlib` eklenmeli.
- Ek bir router paketi gerekmez; stdlib yeterli.

## Teslim Sonrası Doğrulama
- `GET /health` çağrısı: DB ve Qdrant down/up durumlarını doğru raporluyor mu?
- Shutdown sırasında istekler düzgün sonlanıyor mu?

Hazırsanız bu plana göre dosyaları uygulayalım ve kodu birlikte yazalım.