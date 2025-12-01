## Amaç ve Kapsam
- `register`, `login` akışları ve JWT tabanlı kimlik doğrulama middleware’i.
- Standart `net/http` ile mevcut sunucu mimarisi korunur.

## Mimari Tercihler
- HTTP: `net/http` ve `http.ServeMux` (bkz. `cmd/server/main.go:26-50`).
- DB: `database/sql` + pgx stdlib sürücüsü (bkz. `internal/db/db.go:13-27`).
- Şifre: `golang.org/x/crypto/bcrypt`.
- JWT: `github.com/golang-jwt/jwt/v5` (HS256, `cfg.JWT_SECRET`, bkz. `pkg/config/config.go:45-63`).

## Dosya/Dizinler
- `internal/auth/password.go`: `HashPassword`, `VerifyPassword`.
- `internal/auth/jwt.go`: `GenerateToken`, `VerifyToken`, `Claims`.
- `pkg/middleware/auth.go`: `AuthMiddleware(next http.Handler)`.
- `internal/api/handlers/auth.go`: `RegisterHandler`, `LoginHandler`.
- Router güncellemeleri: `cmd/server/main.go` içine `/api/v1/auth/register` ve `/api/v1/auth/login`.
- (Opsiyonel) SQLC: `db/queries/auth/*.sql` ile sorgu dosyaları, ileride `internal/db/auth` jenerasyonu.

## Uygulama Detayı
- Şifre Yardımcıları
  - `HashPassword(password string) (string, error)`: `bcrypt.GenerateFromPassword([]byte(password), cost)`; default cost: `bcrypt.DefaultCost`.
  - `VerifyPassword(hash, password string) bool`: `bcrypt.CompareHashAndPassword` sonucu doğrulama.
- JWT Yardımcıları
  - `Claims`: `userID string`, `expiresAt time.Time` (veya `RegisteredClaims` içinde `Subject=userID`, `ExpiresAt`).
  - `GenerateToken(userID string) (string, error)`: HS256; expire 24 saat; secret: `cfg.JWT_SECRET`.
  - `VerifyToken(tokenString string) (userID string, error)`: imza ve expire kontrolü; subject’den `userID` döndür.
- Handlers
  - `RegisterHandler`
    - Input: `{ "email": "...", "password": "...", "full_name": "..." }`.
    - Adımlar: email format/tight validation → email var mı kontrol (`SELECT id FROM users WHERE email=$1`) → varsa `409 Conflict` → yoksa şifreyi hashle, kullanıcıyı ekle (`INSERT INTO users (email, password_hash, full_name) VALUES ($1,$2,$3) RETURNING id`) → token üret → `201 Created` + `{ token }`.
  - `LoginHandler`
    - Input: `{ "email": "...", "password": "..." }`.
    - Adımlar: kullanıcıyı getir (`SELECT id, password_hash FROM users WHERE email=$1`) → yoksa/şifre yanlışsa `401 Unauthorized` → doğruysa token üret → `200 OK` + `{ token }`.
  - Content-Type: `application/json`; hatalarda JSON mesajı.
- Middleware
  - `AuthMiddleware(next http.Handler) http.Handler`:
    - `Authorization: Bearer <jwt>`; yok/format yanlış: `401`.
    - `VerifyToken` ile doğrula; başarısız: `401`.
    - Başarılı: `context.WithValue(r.Context(), contextKeyUserID, userID)` ve `next.ServeHTTP`.
- Router
  - `cmd/server/main.go` `mux.HandleFunc("/api/v1/auth/register", RegisterHandler)`.
  - `mux.HandleFunc("/api/v1/auth/login", LoginHandler)`.
  - Örnek korumalı endpoint: `mux.Handle("/api/v1/protected", AuthMiddleware(http.HandlerFunc(ProtectedHandler)))`.

## Veri Erişimi
- MVP’de doğrudan `*sql.DB` üzerinden `QueryRowContext`, `ExecContext` kullanılacak.
- (Opsiyonel) Sonraki iterasyonda `db/queries/auth/*.sql` + `sqlc` ile tipli erişime geçiş; mevcut `sqlc.yaml` pgx/v5 hedefli olduğundan `pgxpool` eklenmesi veya `sql_package`’ın `stdlib`’a uyarlanması değerlendirilecek.

## Güvenlik
- Bcrypt cost→ `DefaultCost` ile başla; prod’da `12` hedeflenebilir.
- Sabit zamanlı kıyaslama: `bcrypt.CompareHashAndPassword`.
- Gizli verileri loglama yok; hatalarda detay sızdırma yok.
- JWT secret rotasyonu için `kid` alanı ileride eklenebilir.
- Token süreleri ve clock skew toleransı (ör. ±2 dk) ayarlanabilir.

## Test
- Postman
  - `POST /api/v1/auth/register` → `201` ve token.
  - `POST /api/v1/auth/login` → `200` ve token.
  - `GET /api/v1/protected` → Header `Authorization: Bearer <token>`; `401`/`200` doğrulama.
- Unit
  - `password.go`: hash/verify.
  - `jwt.go`: generate/verify ve expire.
- Integration
  - Handler’lar için DB ile başarılı/başarısız senaryolar.

## Kabul Kriterleri
- Dokümandaki HTTP statü kodları ve gövdeler aynen karşılanır.
- `JWT_SECRET` zorunlu ve kullanılıyor (bkz. `pkg/config/config.go:45-63`).
- Korumasız endpoint’ler çalışır; korumalılar geçerli token olmadan `401` döner.

## Bağımlılıklar ve Etki
- Yeni bağımlılık: `github.com/golang-jwt/jwt/v5` (go.mod’a eklenecek).
- Kod referansları: `cmd/server/main.go:26-50`, `internal/db/db.go:13-27`, `pkg/config/config.go:45-63`.

Onayınız sonrası implementasyona başlayıp doğrulama testleriyle birlikte teslim edeceğim.