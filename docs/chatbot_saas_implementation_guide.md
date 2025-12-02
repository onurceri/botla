# Türkiye Pazarı için AI Chatbot SaaS Platformu - Kapsamlı Uygulama Rehberi

## 📋 İçindekiler
1. [Hazırlık ve Ortam Kurulumu](#1-hazırlık-ve-ortam-kurulumu)
2. [Proje Yapısı ve Mimari Planı](#2-proje-yapısı-ve-mimari-planı)
3. [Veritabanı Tasarımı](#3-veritabanı-tasarımı)
4. [Backend Geliştirme Aşamaları](#4-backend-geliştirme-aşamaları)
5. [Web Scraping Sistemi](#5-web-scraping-sistemi)
6. [PDF İşleme](#6-pdf-işleme)
7. [RAG Pipeline](#7-rag-pipeline)
8. [Frontend Dashboard](#8-frontend-dashboard)
9. [Chat Widget](#9-chat-widget)
10. [Ödeme Sistemi](#10-ödeme-sistemi)
11. [Deployment ve DevOps](#11-deployment-ve-devops)
12. [Testler ve Kalite Kontrol](#12-testler-ve-kalite-kontrol)
13. [Launch Checklist](#13-launch-checklist)

---

## 1. Hazırlık ve Ortam Kurulumu

### 1.1 Sistem Gereksinimleri

Geliştirme makinenizde kurulması gereken araçlar:

**Go (Backend)**
- İndir: https://golang.org/dl
- Sürüm: 1.21 veya daha yeni
- Kontrol: `go version`
- GOPATH ayarı: `echo $GOPATH` (genellikle ~/go)

**PostgreSQL (Veritabanı)**
- İndir: https://www.postgresql.org/download/
- Veya Docker: `docker run -d -e POSTGRES_PASSWORD=postgres -p 5432:5432 postgres:15-alpine`
- Kontrol: `psql --version`

**Node.js (Frontend)**
- İndir: https://nodejs.org/
- Sürüm: 18 LTS veya daha yeni
- Kontrol: `node --version` ve `npm --version`

**Docker & Docker Compose (Deployment)**
- İndir: https://www.docker.com/products/docker-desktop
- Kontrol: `docker --version` ve `docker-compose --version`

**Git (Versiyon Kontrolü)**
- İndir: https://git-scm.com/
- Kontrol: `git --version`

**Postman veya Insomnia (API Testleri)**
- https://www.postman.com/downloads/ veya https://insomnia.rest/

**Metin Editörü / IDE**
- VS Code (Önerilir) + Go extension (golang.go)
- GoLand (JetBrains)

### 1.2 Go Ortamı İlk Kurulum

```bash
# GOPATH kontrolü
echo $GOPATH

# Eğer boş çıktıysa, shell konfigürasyonuna ekle:
# ~/.bashrc veya ~/.zshrc dosyasına:
# export GOPATH=$HOME/go
# export PATH=$PATH:$GOPATH/bin
# Sonra: source ~/.bashrc
```

### 1.3 Proje Dizini ve Git Yapısı

```bash
# Proje dizini oluştur
mkdir chatbot-saas
cd chatbot-saas

# Git repo başlat
git init
git config user.name "Adınız"
git config user.email "email@example.com"

# Go modülü başlat
go mod init github.com/yourusername/chatbot-saas

# Temel dizin yapısını oluştur
mkdir -p cmd/server cmd/worker cmd/migrate
mkdir -p internal/{api,auth,db,scraper,pdf,rag,models,payment}
mkdir -p pkg/{logger,config,utils,middleware}
mkdir -p migrations
mkdir -p frontend dashboard widget
mkdir -p docs tests
mkdir -p scripts
```

### 1.4 .gitignore Dosyası

Proje köküne `.gitignore` oluştur:

```
# Go
*.o
*.a
*.so
*.exe
*.exe~
*.dll
*.dylib
dist/
bin/
vendor/

# Environment
.env
.env.local
.env.*.local

# IDE
.vscode/
.idea/
*.swp
*.swo
*.sublime-project
*.sublime-workspace

# OS
.DS_Store
Thumbs.db

# Database
*.db
*.sql
database.backup

# Logs
*.log
logs/

# Node
node_modules/
dist/
build/
.next/
.cache/
```

### 1.5 İlk Commit

```bash
touch README.md
git add .
git commit -m "Initial project structure"
```

---

## 2. Proje Yapısı ve Mimari Planı

### 2.1 Sistem Bileşenleri ve İletişimi

```
┌─────────────────────────────────────────────────────────┐
│                  MÜŞTERİNİN WEB SİTESİ                │
│                   (HTML/CSS/JS)                         │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Chat Widget (Preact 3KB + Tailwind CSS)         │  │
│  │  - Shadow DOM'da izole                           │  │
│  │  - Siteyi etkilemez                              │  │
│  │  - Her mesaj API'ye gönder                       │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
                           ↕ (HTTPS)
┌──────────────────────────────────────────────────────┐
│              GO BACKEND API (Main)                  │
│  ┌───────────────────────────────────────────────┐  │
│  │ REST Endpoints:                               │  │
│  │ - /api/v1/auth/* (Login, Register)            │  │
│  │ - /api/v1/chatbots/* (CRUD)                   │  │
│  │ - /api/v1/sources/* (Veri kaynakları)         │  │
│  │ - /api/v1/chat (Sohbet)                       │  │
│  │ - /api/v1/analytics/* (Metrikler)             │  │
│  │ - /api/v1/subscription/* (Ödeme)              │  │
│  └───────────────────────────────────────────────┘  │
│  ┌───────────────────────────────────────────────┐  │
│  │ Background Workers (Goroutines):              │  │
│  │ - Scraper (URL tarama)                        │  │
│  │ - PDF Processor (MuPDF)                       │  │
│  │ - Embedding Generator (OpenAI API)            │  │
│  │ - Subscription Checker (Billing)              │  │
│  └───────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────┘
          ↕                    ↕                   ↕
       ┌─────────┐    ┌──────────────┐    ┌───────────┐
       │PostgreSQL   │  Qdrant       │    │ External  │
       │(Relational)│ (Vector DB)   │    │  APIs     │
       │ Users       │ Embeddings    │    │ OpenAI    │
       │ Chatbots    │ (Multitenancy)│   │ Iyzico    │
       │ Logs        │              │    │           │
       └─────────────┘    └──────────────┘    └───────────┘
```

### 2.2 Veri Akışı (Örnek: Chatbot Oluşturma ve PDF İşleme)

```
1. KULLANICI:
   - Dashboard'da "Yeni Chatbot" tıkla
   - Adı gir, dil seç
   - PDF dosyasını yükle

2. FRONTEND:
   - POST /api/v1/chatbots → Chatbot oluştur
   - POST /api/v1/sources (file upload) → Kaynağı kaydet (status: pending)
   - Response: source_id al
   - Polling: GET /api/v1/sources/{source_id} (status kontrolü)

3. BACKEND:
   - İstek al, source status = "processing" olarak güncelle
   - Goroutine başlat (non-blocking)
   - İçeride:
     a) PDF dosyasını oku (gen2brain/go-fitz)
     b) Metin çıkar
     c) Türçe karakterleri normalize et
     d) Chunks'a böl (neurosnap/sentences kullan)
     e) Embedding oluştur (OpenAI API)
     f) Qdrant'ta kaydet
     g) Status = "completed" → DB güncelle

4. FRONTEND (Polling):
   - Status "completed" olunca: "Başarıyla yüklendi" mesajı göster
   - Widget'ı test etmeye çağır

5. CHAT (Widget):
   - "Merhaba" mesajı gönder
   - Backend: Embedding oluştur → Qdrant ara → Context al → OpenAI sor
   - Cevap widget'a gelir → Göster
```

### 2.3 Geliştirme İçin Sabit Hedefler

**MVP (Minimum Viable Product) Sürümü İçin Zorunlu:**
- [ ] PostgreSQL şeması (Users, Chatbots, Sources, Messages, Analytics)
- [ ] Go HTTP server (Gin veya Echo)
- [ ] Temel auth (Register, Login, JWT)
- [ ] Chatbot CRUD
- [ ] PDF upload ve metin çıkarma
- [ ] Basic chunking (Türçe uyumlu)
- [ ] OpenAI embedding
- [ ] Qdrant bağlantısı
- [ ] Chat endpoint
- [ ] React dashboard (basic)
- [ ] Preact widget (basic)
- [ ] Docker Compose
- [ ] Temel hata loglama

**İlk Sonra Eklenecek Özellikler:**
- [ ] Web scraping (Colly + Headless)
- [ ] Türçe NLP (Advanced chunking)
- [ ] Iyzico payment
- [ ] Gelişmiş analytics
- [ ] Whitelabel
- [ ] Theme customization

---

## 3. Veritabanı Tasarımı

### 3.1 PostgreSQL Şeması (Detalılı)

**Tablo 1: USERS**
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    avatar_url TEXT,
    is_email_verified BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Subscription
    subscription_plan VARCHAR(50) DEFAULT 'free', -- free, starter, pro, enterprise
    subscription_started_at TIMESTAMP,
    subscription_expires_at TIMESTAMP,
    
    -- Payment
    payment_customer_id VARCHAR(255), -- Iyzico customer ref
    
    -- KVKK
    kvkk_accepted BOOLEAN DEFAULT false,
    kvkk_accepted_at TIMESTAMP,
    
    -- Soft delete
    deleted_at TIMESTAMP DEFAULT NULL
);

CREATE INDEX idx_users_email ON users(email);
```

**Tablo 2: CHATBOTS**
```sql
CREATE TABLE chatbots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Model settings
    system_prompt TEXT DEFAULT 'Sen yararlı, kibar ve bilgili bir yapay zeka asistanısın.',
    model VARCHAR(100) DEFAULT 'gpt-3.5-turbo',
    temperature FLOAT DEFAULT 0.7,
    max_tokens INTEGER DEFAULT 512,
    
    -- Styling
    theme_color VARCHAR(7) DEFAULT '#3b82f6', -- Hex color
    welcome_message TEXT DEFAULT 'Merhaba! Size nasıl yardımcı olabilirim?',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL
);

CREATE INDEX idx_chatbots_user_id ON chatbots(user_id);
```

**Tablo 3: DATA_SOURCES**
```sql
CREATE TABLE data_sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chatbot_id UUID NOT NULL REFERENCES chatbots(id) ON DELETE CASCADE,
    
    source_type VARCHAR(50) NOT NULL, -- 'pdf', 'url', 'text'
    source_url VARCHAR(2048),
    file_path VARCHAR(1024),
    original_filename VARCHAR(255),
    text_content TEXT,
    
    -- Processing status
    status VARCHAR(50) DEFAULT 'pending', -- pending, processing, completed, failed
    error_message TEXT,
    
    processed_at TIMESTAMP,
    chunk_count INTEGER DEFAULT 0,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL
);

CREATE INDEX idx_data_sources_chatbot_id ON data_sources(chatbot_id);
CREATE INDEX idx_data_sources_status ON data_sources(status);
```

**Tablo 4: CONVERSATIONS**
```sql
CREATE TABLE conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chatbot_id UUID NOT NULL REFERENCES chatbots(id) ON DELETE CASCADE,
    
    session_id VARCHAR(255), -- Widget'ın session ID'si
    visitor_name VARCHAR(255),
    visitor_email VARCHAR(255),
    
    -- Visitor tracking (KVKK dikkate alınarak)
    visitor_ip_hash VARCHAR(64), -- Hash'lenmiş IP (original tutma)
    user_agent_hash VARCHAR(64),
    
    message_count INTEGER DEFAULT 0,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_conversations_chatbot_id ON conversations(chatbot_id);
```

**Tablo 5: MESSAGES**
```sql
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    
    role VARCHAR(20) NOT NULL, -- 'user', 'assistant'
    content TEXT NOT NULL,
    
    -- Token tracking
    tokens_used INTEGER,
    
    -- User feedback
    thumbs_up BOOLEAN,
    feedback_text TEXT,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_messages_conversation_id ON messages(conversation_id);
```

**Tablo 6: ANALYTICS (Materialized View veya Scheduled Aggregation)**
```sql
CREATE TABLE analytics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chatbot_id UUID NOT NULL REFERENCES chatbots(id) ON DELETE CASCADE,
    
    analytics_date DATE NOT NULL,
    
    total_conversations INTEGER DEFAULT 0,
    total_messages INTEGER DEFAULT 0,
    unanswered_messages INTEGER DEFAULT 0,
    thumbs_up_count INTEGER DEFAULT 0,
    thumbs_down_count INTEGER DEFAULT 0,
    average_tokens_per_message FLOAT,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(chatbot_id, analytics_date)
);

CREATE INDEX idx_analytics_chatbot_date ON analytics(chatbot_id, analytics_date);
```

**Tablo 7: PAYMENTS (Billing)**
```sql
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'TRY',
    
    status VARCHAR(50) DEFAULT 'pending', -- pending, success, failed, refunded
    payment_method VARCHAR(50),
    
    iyzico_payment_id VARCHAR(255),
    iyzico_conversation_id VARCHAR(255),
    
    plan_type VARCHAR(50),
    billing_period_start DATE,
    billing_period_end DATE,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_payments_user_id ON payments(user_id);
CREATE INDEX idx_payments_status ON payments(status);
```

### 3.2 Qdrant Collection Tasarımı

**Collection: "embeddings"**

```json
{
  "name": "embeddings",
  "config": {
    "params": {
      "vectors": {
        "size": 1536,
        "distance": "Cosine"
      },
      "shard_number": 2,
      "replication_factor": 1,
      "write_consistency_factor": 1
    }
  }
}
```

**Her Vektör Kaydı:**
```json
{
  "id": 12345,
  "vector": [0.123, 0.456, ..., 384 boyutlu array],
  "payload": {
    "chatbot_id": "uuid-123e4567",
    "source_id": "uuid-789a0bcd",
    "chunk_index": 0,
    "original_text": "Bu, PDF'den çıkarılan gerçek metin...",
    "source_type": "pdf",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

**Neden Multitenancy (Tek koleksiyon)?**
- 1000 chatbot = 1000 koleksiyon = Qdrant RAM tüketimi ↑↑↑
- 1000 chatbot = 1 koleksiyon + payload filtering = Verimli
- Filtreleme: `chatbot_id == "uuid-123"` → Hızlı (Payload Index sayesinde)

### 3.3 Migration Stratejisi

**golang-migrate Kurulumu:**
```bash
# Mac
brew install golang-migrate

# Linux
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/migrate

# Go ile
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

**İlk Migration Oluştur:**
```bash
migrate create -ext sql -dir migrations -seq init_schema
```

Bu 2 dosya oluşturur:
- `migrations/000001_init_schema.up.sql` (Oluştur)
- `migrations/000001_init_schema.down.sql` (Geri al)

**up.sql'e tüm CREATE TABLE'ları koy**
**down.sql'e DROP TABLE'ları koy**

**Migration Çalıştırma (Go kodu içinden):**
```go
// cmd/migrate/main.go içinde
m, err := migrate.New(
    "file://migrations",
    "postgres://user:pass@localhost/chatbot_saas",
)
m.Up()
```

---

## 4. Backend Geliştirme Aşamaları

### 4.1 Aşama 1: Temel Server ve Config

**Yapılacaklar:**

1. **pkg/config/config.go** - Çevre değişkenlerini oku
   - DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD
   - QDRANT_URL
   - OPENAI_API_KEY
   - IYZICO_API_KEY
   - JWT_SECRET
   - PORT (Server portu)

2. **cmd/server/main.go** - Server başlatma
   - Config yükle
   - Database pool oluştur
   - HTTP server başlat
   - Graceful shutdown (SIGINT, SIGTERM)

3. **internal/db/db.go** - Database bağlantısı
   - `*sql.DB` connection pool
   - Ping testi
   - Connection limits (MaxOpenConns: 25, MaxIdleConns: 5)

4. **pkg/logger/logger.go** - Basit logging
   - Stdout'a JSON formatında log
   - Levels: DEBUG, INFO, WARN, ERROR

5. **Health Check Endpoint**
   - GET /health → `{ "status": "ok" }`
   - Bağımlılıkları test et (DB, Qdrant)

**Go Öğrenme Noktaları:**
- `os.Getenv()`: Çevre değişkenleri
- `sql.Open()`: Database connection
- `defer` statement: Resource cleanup
- `goroutine` ve `channel`: Background tasks için temel

---

### 4.2 Aşama 2: Kullanıcı Yönetimi (Auth)

**Yapılacaklar:**

1. **internal/auth/password.go** - Şifre işlemleri
   - `HashPassword(password string) (string, error)` - bcrypt
   - `VerifyPassword(hash, password string) bool` - Doğrulama

2. **internal/auth/jwt.go** - Token yönetimi
   - `GenerateToken(userID string) (string, error)` - JWT oluştur
   - `VerifyToken(tokenString string) (userID string, error)` - Doğrula
   - `Claims` struct: userID, expiresAt

3. **internal/api/handlers/auth.go** - HTTP handlers
   - `RegisterHandler(w http.ResponseWriter, r *http.Request)`
     - POST /api/v1/auth/register
     - Body: `{ "email": "...", "password": "...", "full_name": "..." }`
     - Email zaten varsa: 409 Conflict
     - Başarılı: 201 Created + token
   
   - `LoginHandler(w http.ResponseWriter, r *http.Request)`
     - POST /api/v1/auth/login
     - Body: `{ "email": "...", "password": "..." }`
     - Şifre yanlışsa: 401 Unauthorized
     - Başarılı: 200 OK + token

4. **pkg/middleware/auth.go** - Auth middleware
   - `AuthMiddleware(next http.Handler) http.Handler`
   - Header'dan token al (Authorization: Bearer ...)
   - Doğrula
   - Başarısız: 401 dönüş
   - Başarılı: userID'yi context'e koy

5. **Postman Test:**
   - POST /auth/register
   - POST /auth/login (Dönen token'ı kopyala)
   - GET /protected-endpoint (Token ile header'a ekle)

**Go Öğrenme Noktaları:**
- `crypto/sha256`, `crypto/md5`: Hashing
- `github.com/golang-jwt/jwt`: JWT kütüphanesi
- `bcrypt`: Güvenli password hashing
- `context.Context`: Request-scoped values

---

### 4.3 Aşama 3: Chatbot CRUD

**Yapılacaklar:**

1. **internal/models/chatbot.go** - Data model
   ```go
   type Chatbot struct {
       ID              string
       UserID          string
       Name            string
       SystemPrompt    string
       Model           string
       Temperature     float32
       ThemeColor      string
       CreatedAt       time.Time
   }
   ```

2. **internal/db/chatbot.go** - Database işlemleri
   - `CreateChatbot(ctx context.Context, bot *Chatbot) error`
     - INSERT chatbots table'a
   - `GetChatbotsByUserID(ctx context.Context, userID string) ([]Chatbot, error)`
     - SELECT WHERE user_id = :userID
   - `GetChatbotByID(ctx context.Context, id string) (*Chatbot, error)`
     - SELECT WHERE id = :id
   - `UpdateChatbot(ctx context.Context, bot *Chatbot) error`
     - UPDATE chatbots SET ...
   - `DeleteChatbot(ctx context.Context, id string) error`
     - UPDATE chatbots SET deleted_at = NOW() (Soft delete)

3. **internal/api/handlers/chatbot.go** - HTTP handlers
   - `CreateChatbot(w http.ResponseWriter, r *http.Request)`
     - POST /api/v1/chatbots
     - userID'yi context'ten al
     - JSON parse + validate
     - DB'ye kaydet
     - 201 döndür
   
   - `ListChatbots(w http.ResponseWriter, r *http.Request)`
     - GET /api/v1/chatbots
     - userID'ye göre listele
     - 200 döndür
   
   - `GetChatbot(w http.ResponseWriter, r *http.Request)`
     - GET /api/v1/chatbots/:id
     - URL parametreden id çıkar
     - Ownership kontrol (bu chatbot bu user'a mı ait?)
     - Döndür
   
   - `UpdateChatbot(w http.ResponseWriter, r *http.Request)`
     - PUT /api/v1/chatbots/:id
     - Ownership kontrol
     - Update + 200 döndür
   
   - `DeleteChatbot(w http.ResponseWriter, r *http.Request)`
     - DELETE /api/v1/chatbots/:id
     - Ownership kontrol
     - Soft delete + 204 döndür

4. **Routing**
   - Gin/Echo router'da routes tanımla:
     ```
     POST   /api/v1/chatbots
     GET    /api/v1/chatbots
     GET    /api/v1/chatbots/:id
     PUT    /api/v1/chatbots/:id
     DELETE /api/v1/chatbots/:id
     ```

5. **Postman Test:**
   - Auth token'ı al
   - Chatbot oluştur
   - Listele
   - Detay al
   - Güncelle
   - Sil

**Go Öğrenme Noktaları:**
- SQL parametrization: `SELECT * FROM chatbots WHERE id = $1` (SQL injection prevention)
- `json.Marshal()` / `json.Unmarshal()`: JSON serialization
- `mux.Vars(r)`: URL parametreler (chi) veya `r.FormValue()` (http)
- Error wrapping: `fmt.Errorf("failed to create chatbot: %w", err)`

---

### 4.4 Aşama 4: Veri Kaynağı Yönetimi

**Yapılacaklar:**

1. **internal/models/source.go** - Data model
   ```go
   type DataSource struct {
       ID               string
       ChatbotID        string
       SourceType       string // "pdf", "url", "text"
       SourceURL        string
       FilePath         string
       OriginalFilename string
       Status           string // "pending", "processing", "completed", "failed"
       ErrorMessage     string
       ChunkCount       int
       ProcessedAt      time.Time
       CreatedAt        time.Time
   }
   ```

2. **internal/api/handlers/source.go** - HTTP handlers
   - `AddDataSource(w http.ResponseWriter, r *http.Request)`
     - POST /api/v1/chatbots/:chatbot_id/sources
     - Ownership kontrol
     - Form parse (multipart/form-data)
     - Status = "pending" ile kaydet
     - Queue'ya ekle (Background processing)
     - Response: source_id
   
   - `ListSources(w http.ResponseWriter, r *http.Request)`
     - GET /api/v1/chatbots/:chatbot_id/sources
     - Listing + status ile
   
   - `GetSourceStatus(w http.ResponseWriter, r *http.Request)`
     - GET /api/v1/sources/:source_id
     - Status bilgisi dön (Frontend polling'i için)
   
   - `DeleteSource(w http.ResponseWriter, r *http.Request)`
     - DELETE /api/v1/sources/:source_id
     - Qdrant'ta ilişkili vektörleri sil
     - DB'den kaynak sil

3. **File Upload Handling**
   - `r.FormFile("file")` ile dosyayı al
   - Validate: Type kontrol (PDF gibi), Size limit (10MB)
   - Temp dizinine kaydet: `/tmp/uploads/source-{uuid}.pdf`
   - DB'ye dosya yolunu kaydet

4. **Background Processing Queue**
   - Go channel veya Redis queue
   - Worker goroutine: Kuyruğu monitorle
   - PDF varsa → PDF processor başlat
   - URL varsa → Scraper başlat
   - Status update et

**Go Öğrenme Noktaları:**
- `multipart.Form`: File upload parsing
- `os.MkdirAll()`: Dizin oluşturma
- `io.Copy()`: File kopyalama
- Channel'lar: `make(chan Task, 100)` - Buffered channels
- `select` statement: Channel'ları dinleme

---

### 4.5 Aşama 5: Chat Endpoint (Sohbet Motoru)

**Yapılacaklar:**

1. **internal/models/message.go** - Message model
   ```go
   type Message struct {
       ID               string
       ConversationID   string
       Role             string // "user", "assistant"
       Content          string
       TokensUsed       int
       ThumbsUp         *bool
       CreatedAt        time.Time
   }
   ```

2. **internal/api/handlers/chat.go** - Chat handler
   - `Chat(w http.ResponseWriter, r *http.Request)`
     - POST /api/v1/chatbots/:chatbot_id/chat
     - Body: `{ "message": "Merhaba", "session_id": "..." }`
     - Response: `{ "response": "Merhaba! Ben yapay zeka asistanıyım.", "tokens": 50 }`

3. **Chat Processing Pipeline:**
   ```
   a) session_id'dan conversation al (veya oluştur)
   b) User message'ı DB'ye kaydet
   c) User message'ı embedding'e çevir (OpenAI API)
   d) Qdrant'ta ara (top_k=5, filter: chatbot_id)
   e) Context oluştur (benzer chunk'ları birleştir)
   f) Prompt'u oluştur: System + Context + User question
   g) OpenAI API'ye gönder
   h) Response'u al
   i) Assistant message'ı DB'ye kaydet
   j) Token kullanımını hesapla ve kaydet
   k) JSON response dön
   ```

4. **OpenAI API Integration** (internal/rag/openai.go)
   - `CreateEmbedding(text string) ([]float32, error)`
     - Model: text-embedding-3-small
     - API Key: env'den al
     - HTTP POST + error handling
   
   - `CreateCompletion(systemPrompt, context, userMessage string) (string, int, error)`
     - Model: gpt-3.5-turbo (başlangıç için)
     - Temperature, max_tokens: chatbot settings'den al
     - Prompt formatting: OpenAI best practices'e göre
     - Token count dön

5. **Qdrant Search** (internal/rag/qdrant.go)
   - `SearchSimilar(embedding []float32, chatbotID string, topK int) ([]ChunkResult, error)`
     - Qdrant client'ı oluştur
     - Search query: embedding + payload filter (chatbot_id)
     - Results'ı ChunkResult'a çevir
     - original_text'leri birleştir (context'e katıl)

6. **Error Scenarios:**
   - OpenAI rate limit: Retry logic (exponential backoff)
   - Qdrant bağlantı hatası: Fallback response ("Şu an hata oluştu")
   - Token limit: Cevap truncate et

7. **Response Format:**
   ```json
   {
     "response": "Yapay zeka'nın cevabı...",
     "tokens_used": 120,
     "sources_used": [
       {"chunk_index": 0, "source_type": "pdf"}
     ]
   }
   ```

**Go Öğrenme Noktaları:**
- HTTP client: `net/http.Client` + timeout
- JSON encoding: `json.NewEncoder()`
- Error handling ve retry: `time.Sleep()` + loop
- Context timeout: `context.WithTimeout()`

---

## 5. Web Scraping Sistemi

### 5.1 Statik Tarama (Colly + Türçe Kodlama)

**Yapılacaklar:**

1. **internal/scraper/colly.go** - Colly kurulumu
   - Collector oluştur
   - AllowedDomains ayarla
   - UserAgent setini kur (5-10 farklı UA)
   - Timeout: 30 saniye
   - RateLimit: 2 request/saniye
   - Callback'ler:
     - `OnHTML`: Body tag'ini bul ve text'i çıkar
     - `OnError`: Log error
     - `OnScraped`: Request tamamlandı

2. **Türçe Kodlama Çözümü** (internal/scraper/encoding.go)
   - `NormalizeText(rawHTML string) (string, error)`
     - golang.org/x/net/html/charset ile otomatik detect + convert
     - UTF-8 validation
     - BOM (Byte Order Mark) temizleme
     - Invalid UTF-8 karakterleri replace et
   
   - `IsValidUTF8(data []byte) bool` - Basit kontrol

3. **Tarama Worker** (internal/scraper/worker.go)
   - Task struct: URL, ChatbotID, SourceID
   - `ScrapeURL(task ScrapingTask) (string, error)`
     - Colly'yi çalıştır
     - HTML parse (goquery kütüphanesi)
     - Body text'ini çıkar (sadece görünür text, script/style kaldır)
     - Normalize
     - Tekrar tarama kontrolü (cache)
     - Result döndür

4. **Cache Mekanizması**
   - Redis veya simple in-memory cache
   - Key: `scraped:{md5(url)}`
   - Value: { content, timestamp }
   - Validity: 7 gün

5. **Background Job Queue Integration**
   - Worker goroutine'ler
   - Channel'dan task al
   - Scrape + Process
   - Status güncellemesi

**Go Öğrenme Noktaları:**
- `net/http` custom client
- `golang.org/x/net` package'lar
- Goroutine pools (WaitGroup)
- sync.Mutex: Thread-safe cache

---

### 5.2 Dinamik Tarama (Headless Browser)

**Yapılacaklar:**

1. **internal/scraper/browser.go** - Headless setup
   - Chrome/Chromium yüklü olmalı
   - go-rod veya similar kütüphane
   - Browser instance pool
   - Connection timeout

2. **Tarama Fonksiyonu**
   - `ScrapeDynamicURL(url string) (string, error)`
     - Browser tab aç
     - URL'ye git (navigation timeout: 10 saniye)
     - JavaScript execute (Wait for page load)
     - DOM'dan HTML al
     - Tab kapat
     - Content döndür

3. **Resource Management** (KRITIK)
   - Browser tab sayısı sınırlandırması (max 2 concurrent)
   - Memory leak prevention:
     - Her tab işleminden sonra `page.Close()`
     - Idle time'ı aşan browser'ları kapat
   
4. **Fallback Stratejisi**
   - İlk: Static tarama (Colly)
   - Hata: Dynamic tarama dene (Headless)
   - Hata: Boş response + hata loglama

---

## 6. PDF İşleme

### 6.1 Metin Çıkarma (gen2brain/go-fitz + MuPDF)

**Yapılacaklar:**

1. **internal/pdf/extractor.go** - PDF parser
   - `ExtractPDFText(filePath string) (string, error)`
     - go-fitz ile PDF aç
     - Sayfa sayısını al
     - Her sayfadan text blocks çıkar
     - Blok'ların koordinatlarına göre satır birleştir
     - Sayfaları bir string'de topla
     - Return

2. **Türçe Karakter Validasyonu**
   - Metin UTF-8 mi kontrol et
   - Invalid bytes replace et (? karakteri ile)
   - Encoding normalization

3. **CGO Bağımlılıkları** (Önemli!)
   - macOS: `brew install mupdf`
   - Linux: `apt-get install libmupdf-dev`
   - Windows: Precompiled binary veya WSL
   - go-fitz derleme sırasında errors olabilir

4. **Error Handling**
   - Dosya açılamaz
   - PDF corrupted
   - Sayfa < 1

**Derleme ve Test Notları**
- Üretici derlemeler için MuPDF kurulumu gerekir ve gerçek çıkarıcı yalnızca `fitz` build etiketi ile etkinleşir: `go build -tags fitz ./...`
- Varsayılan derlemede (etiketsiz) `internal/pdf/extractor_stub.go` çalışır ve uygun hata döndürür.
- Linux: `apt-get install -y libmupdf-dev pkg-config`
- macOS: `brew install mupdf`
- CI/yerel testlerde örnek PDF yolu sağlanırsa fitz testi çalışır: `BOTLA_PDF_PATH=/path/to/sample.pdf go test -tags fitz ./internal/pdf -v`

---

### 6.2 OCR (Taranmış PDF'ler)

**Yapılacaklar:**

1. **internal/pdf/ocr.go** - Tesseract integration
   - `ExtractPDFWithOCR(filePath string) (string, error)`
     - go-fitz ile PDF'i render et (300 DPI)
     - Her sayfa için image oluştur
     - gosseract'a gönder
     - OCR + Türçe dil paketi
     - Metin al
     - Sayfaları birleştir

2. **Tesseract Kurulumu**
   - Server'da yüklü olmalı
   - Türçe dil paketi: tesseract-ocr-tur
   - Kontrol: `tesseract --list-langs | grep tur`

3. **Fallback Logic**
   - Eğer metin çok az (< 100 chars) → OCR dene
   - OCR da başarısızsa → Error log + user notification

---

### 6.3 PDF Processing Pipeline

**Adım Adım:**

1. Upload sırasında:
   - File validation (PDF mi, boyut < 50MB)
   - Temp dizine kaydet
   - DB'ye kaydet (status: pending)
   - Task kuyruğa ekle

2. Worker (Goroutine):
   - Task al
   - Status: processing
   - go-fitz ile metin çıkar (try)
   - Başarısız: OCR dene (fallback)
   - Başarısızsa: Status = failed + error message
   - Başarılı: Text normalize + chunk'la

3. Chunking sonrası:
   - Her chunk'ı Qdrant'a kaydet (embedding ile)
   - Status: completed
   - chunk_count güncelleşti

---

## 7. RAG Pipeline

### 7.1 Metin Parçalama (Türçe Uyumlu)

**Yapılacaklar:**

1. **internal/rag/chunker.go** - Chunking algoritması
   - `ChunkText(text string, targetTokens int) ([]Chunk, error)`
     - Input: Raw text
     - Output: []Chunk { Text: "...", TokenCount: 512 }

2. **Chunking Algoritması (Aşamalar):**

   **Aşama 1: Paragraph Bölmesi**
   - Split by: `\n\n` (2+ newline)
   - Her paragraph bir logical unit

   **Aşama 2: Cümle Bölmesi (Turkish NLP)**
   - neurosnap/sentences kullan
   - Türçe exception'ları: "Dr.", "Prof.", "vb."
   - Output: []Sentence

   **Aşama 3: Recursive Chunking**
   - Target: 512 token
   - Sentences'ı ekle, token sayarken toplam token takip et
   - Chunk full olunca başa dön
   - Cümle sınırını korumayı ön planda tut

   **Aşama 4: Overlap Ekleme**
   - Her chunk'ın sonundan %15'ini al
   - Sonraki chunk'ın başına ekle
   - Bağlam kopmasını prevent et

3. **Token Counting** (internal/rag/tokens.go)
   - `CountTokens(text string) int`
   - OpenAI token counter veya estimation
   - Türçe için ~1.3x multiplier (karakter sayısı / 4)

---

### 7.2 Embedding Pipeline

**Yapılacaklar:**

1. **internal/rag/embedding.go** - Embedding orchestration
   - `GenerateEmbeddings(chunks []Chunk, chatbotID string) error`
     - Chunks batch'le (25 chunk per request)
     - OpenAI API çağır
     - Response parse
     - Qdrant'a kaydet
     - Ingest kuyruğuna entegre edildi

2. **Batch Processing**
   - 1 request = max 25 chunk
   - Loop: Batch send + result collect
   - Rate limiting: 3500 req/min
   - Exponential backoff on rate limit

3. **Error Recovery**
   - 1 chunk failed: Skip + log (diğerleri continue)
   - Tüm request failed: Retry logic (1 retry, total 2 attempts)
   - Final fail: Mark status = "failed"

4. **Cost Tracking** (Optional ama İyi)
   - Her embedding'in token sayısını kaydet
   - Toplam cost = tokens * $0.02 / 1M
   - Kontrol panelinde göster

---

### 7.3 Qdrant Arama ve Context Oluşturma

**Yapılacaklar:**

1. **internal/rag/search.go** - Vector search
   - `SearchContext(queryEmbedding []float32, chatbotID string) (string, []ChunkMetadata, error)`
     - Qdrant client init
     - Search query: embedding + filter (chatbot_id)
     - Top k=5 results
     - Confidence score check
     - Context concatenate
     - Metadata (sources) döndür

2. **Context Formulation**
   - Retrieved chunks'ı sırala (relevance'a göre)
   - Concatenate: "\n---\n" ile ayır
   - Max context length: 2000 token (OpenAI limit'ine uygun)
   - Format: "Aşağıdaki belgeler sorgularına cevap vermek için kullanılmıştır:\n\n{context}"

3. **No Context Scenario**
   - Eğer 0 result veya score < threshold
   - System prompt'ta: "Bilmiyorum cevapla"
   - User'a: "Yeterli bilgi bulamadım" döndür

---

## 8. Frontend Dashboard

### 8.1 React Project Setup

**Yapılacaklar:**

```bash
cd frontend
npm create vite@latest . -- --template react
npm install
npm install -D @types/react @types/react-dom typescript
npm install shadcn-ui @radix-ui/react-* lucide-react
npm install axios react-query react-router-dom
npm install tailwindcss postcss autoprefixer
npx tailwindcss init -p
```

### 8.2 Folder Structure

```
frontend/
├── src/
│   ├── components/          (Reusable UI)
│   │   ├── layout/
│   │   │   ├── Header.tsx
│   │   │   ├── Sidebar.tsx
│   │   │   └── MainLayout.tsx
│   │   ├── chatbot/
│   │   │   ├── ChatbotCard.tsx
│   │   │   ├── ChatbotForm.tsx
│   │   │   └── SourceUploader.tsx
│   │   └── shared/
│   │       ├── Button.tsx
│   │       ├── Card.tsx
│   │       └── Modal.tsx
│   ├── pages/               (Route pages)
│   │   ├── LoginPage.tsx
│   │   ├── DashboardPage.tsx
│   │   ├── ChatbotsPage.tsx
│   │   ├── ChatbotDetailPage.tsx
│   │   ├── AnalyticsPage.tsx
│   │   └── SettingsPage.tsx
│   ├── api/                 (API client)
│   │   ├── client.ts        (Axios instance)
│   │   ├── auth.ts
│   │   ├── chatbot.ts
│   │   ├── source.ts
│   │   └── analytics.ts
│   ├── hooks/               (Custom React hooks)
│   │   ├── useAuth.ts
│   │   ├── useChatbots.ts
│   │   └── usePagination.ts
│   ├── types/               (TypeScript types)
│   │   ├── auth.ts
│   │   ├── chatbot.ts
│   │   ├── api.ts
│   │   └── index.ts
│   ├── utils/
│   │   ├── localStorage.ts
│   │   ├── format.ts
│   │   └── validators.ts
│   ├── App.tsx
│   └── main.tsx
├── public/
├── vite.config.ts
├── tailwind.config.js
└── package.json
```

### 8.3 Ana Sayfa Tasarımları

**Sayfalar ve Amaçları:**

1. **LoginPage** (/login)
   - E-posta ve şifre girişi
   - "Kaydol" linki
   - Şifremi unuttum
   - Error messaging

2. **DashboardPage** (/)
   - Welcome message
   - İstatistikler (toplam chatbot, toplam mesaj)
   - Son aktiviteler
   - "Yeni Chatbot" CTA

3. **ChatbotsPage** (/chatbots)
   - Chatbot listesi (table veya card grid)
   - Status, mesaj sayısı, created date
   - Actions: Edit, Delete, View Analytics, Copy Widget Code
   - Pagination ve search

4. **ChatbotDetailPage** (/chatbots/:id)
   - Chatbot adı, açıklaması
   - Settings: Name, System Prompt, Model, Temperature
   - Data Sources panel (Add PDF, Add URL, Add Text)
   - Widget embed code (copy to clipboard)
   - Test chat (widget'ı embed et)

5. **AnalyticsPage** (/chatbots/:id/analytics)
   - Grafik: Günlük mesaj sayısı (recharts)
   - Grafik: Konuşma sayısı
   - Metrik: Toplam mesaj, yanıtlanamayan %, avg rating
   - Tarih range picker

6. **SettingsPage** (/settings)
   - Profil: E-posta, ad, avatar
   - Şifre değişim
   - Plan ve faturalandırma
   - KVKK onayları

### 8.4 API Client (axios + React Query)

**Setup (src/api/client.ts):**
```
- Base URL: process.env.REACT_APP_API_URL
- Interceptor: Authorization header (Bearer token)
- Error handler: 401 → Redirect to login
- Timeout: 30 saniye
```

**Hook Examples:**
```
useQuery(['chatbots'], () => api.getChatbots())
useMutation((data) => api.createChatbot(data), {
  onSuccess: () => queryClient.invalidateQueries(['chatbots'])
})
```

### 8.5 Authentication State

**Yöntem:**
- JWT token'ı localStorage'da sakla
- Custom hook: `useAuth()`
- Token expiration kontrolü
- Auto-logout on 401

---

## 9. Chat Widget

### 9.1 Widget Project Setup

```bash
cd widget
npm create vite@latest . -- --template react
npm install
npm install preact preact/compat
npm install -D preact-cli tailwindcss
```

### 9.2 Widget Teknik Mimarı

**İşlem Sırası:**

1. **Müşteri's sitesinde:**
   ```html
   <script src="https://chatbot.example.com/widget.js?chatbot-id=uuid-123"></script>
   ```

2. **Widget.js yüklenir:**
   - Query param'dan chatbot-id çıkar
   - DOM'da host element ara (`#chatbot-widget-host`)
   - Yoksa oluştur
   - Shadow DOM attach et
   - Tailwind CSS enjekte et
   - Preact App render et

3. **Preact App (Lightweight):**
   - Chat bubble UI
   - Message input
   - Conversation history
   - API calls

### 9.3 Shadow DOM Kurulumu (Detaylı)

**Adım 1: Widget.js Bootstrap**
```javascript
// widget.js entry point
(function() {
  const chatbotId = getChatbotIdFromScript();
  const host = document.getElementById('chatbot-widget-host') || createHost();
  
  fetch(host, chatbotId)
    .then(() => initializeWidget(host, chatbotId))
    .catch(err => console.error('Widget init failed', err));
})();
```

**Adım 2: Shadow DOM Creation**
```javascript
function initializeWidget(host, chatbotId) {
  const shadowRoot = host.attachShadow({ mode: 'open' });
  
  // Inject Tailwind CSS (minified string)
  const style = document.createElement('style');
  style.textContent = TAILWIND_CSS_STRING; // ~10KB
  shadowRoot.appendChild(style);
  
  // Render Preact App
  render(<App chatbotId={chatbotId} />, shadowRoot);
}
```

**Adım 3: Stil İzolasyonu**
- Müşteri sitesinin `h1 { color: red }` → Widget h1'i etkilemez
- Widget'ın `body { margin: 100px }` → Site body'si etkilemez
- Sadece CSS Inheritance'ler (font-family, color gibi) traverse eder

### 9.4 Widget UI Bileşenleri

**Ana Component'ler:**

1. **ChatBubble**
   - Animated bubble (sağ alt köşe)
   - Badge: unread message count
   - Click → Drawer aç

2. **ChatDrawer**
   - Header: Chatbot adı + close button
   - Messages list (scrollable)
   - Input field + send button
   - Powered by link (biz için branding)

3. **Message**
   - Avatar (user circle, bot icon)
   - Timestamp
   - Message text
   - Typing indicator (bot yazıyor)

### 9.5 Session Management

**Akış:**

1. Widget ilk yüklenir:
   - localStorage'ı kontrol et: `chatbot_session_{chatbotId}`
   - Varsa: conversation_id al + messages yükle
   - Yoksa: Yeni session oluştur + UUID al

2. Her mesaj:
   - POST /api/v1/chatbots/{chatbotId}/chat
   - Body: `{ message: "...", session_id: "..." }`
   - Response: `{ response: "...", tokens: X }`

3. Sayfa refresh:
   - localStorage intact
   - Eski mesajler gösterilir
   - Devam etme imkanı

---

## 10. Ödeme Sistemi

### 10.1 Iyzico Entegrasyonu

**Adımlar:**

1. **Account Oluşturma**
   - https://merchant.iyzipay.com → Register
   - API Key al
   - Secret Key al
   - Test kartı: 5890040000050009 (Sandbox)

2. **PKI String (İmza) Oluşturma** (Go)
   - `{apiKey}{requestString}{secretKey}` → SHA1 → base64
   - Her istek'e bu imza header'ında gitmeli

3. **Payment Flow**

   **A) Kart Tokenization (İlk Ödeme)**
   ```
   1. Frontend: Kart bilgisi al (Iyzico form embed)
   2. Iyzico JS: Card token oluştur
   3. Backend: Token al + amount + plan_id
   4. Iyzico API: CreatePayment çağır (token kullan)
   5. Success/Fail response
   6. Success: DB'ye payment kaydet + subscription active + user plan upgrade
   7. Widget'a gönder: "Ödeme başarılı"
   ```

   **B) Recurring Payment (Abonelik Yenilemesi)**
   ```
   1. Her ay bir adet job çalışır (Cron job)
   2. Expiring subscriptions bulur
   3. Iyzico API: CreateRecurringPayment (customer_id + cardToken)
   4. Başarı: subscription_expires_at += 1 month
   5. Başarısız: Dunning e-mail gönder
   6. 3x başarısız: Subscription suspend
   ```

### 10.2 Plan Yapısı

**Planlar:**

| Plan | Aylık Ücret | Chatbot # | Mesaj/Ay | Features |
|------|-------------|-----------|----------|----------|
| Ücretsiz | 0 | 1 | 100 | PDF + URL |
| Başlangıç | 399₺ | 3 | 5000 | + Analytics |
| Profesyonel | 999₺ | 10 | 20000 | + Custom branding |
| Ajans | 2999₺ | Unlimited | Unlimited | + Whitelabel |

### 10.3 Subscription State Machine

```
Free User
    ↓ (Upgrade click)
Pending Payment
    ↓ (Payment success)
Active Subscriber ←→ (Auto-renew) ←→ Active Subscriber
    ↓ (Renewal failed x3)
Suspended
    ↓ (Manual payment)
Active Subscriber
```

### 10.4 Implementation Details (Go)

Geçici durum: API key temin edilene kadar ödeme entegrasyonu devre dışı. Aşağıdaki örnekler tüm gerçek çağrıları yorum satırına alır ve sadece log üretir.

```go
// internal/payment/iyzico.go
package payment

import "log"

type User struct{ ID string }

type Payment struct {
    Amount int64
    PlanID string
    Token  string
}

// CreateCustomer(user *User) (customerID string, error)
func CreateCustomer(user *User) (string, error) {
    log.Printf("[PAYMENT] CreateCustomer skipped (awaiting API keys) user=%s", user.ID)
    // İyzico entegrasyonu aktif olduğunda açılacak:
    // iyzicoAPI.CreateCustomer(...)
    return "stub-customer-id", nil
}

// CreatePayment(payment *Payment) (transactionID string, error)
func CreatePayment(payment *Payment) (string, error) {
    log.Printf("[PAYMENT] CreatePayment skipped amount=%d plan=%s", payment.Amount, payment.PlanID)
    // Gerçek ödeme çağrısı entegrasyon aktif olana kadar kapalı:
    // iyzicoAPI.CreatePayment(...)
    return "stub-transaction-id", nil
}

// GetPaymentStatus(transactionID string) (status string, error)
func GetPaymentStatus(transactionID string) (string, error) {
    log.Printf("[PAYMENT] GetPaymentStatus stub transactionID=%s", transactionID)
    // iyzicoAPI.GetPaymentStatus(...)
    return "PENDING", nil
}

// CreateRecurringPayment(customerID string) error
func CreateRecurringPayment(customerID string) error {
    log.Printf("[PAYMENT] CreateRecurringPayment skipped customerID=%s", customerID)
    // iyzicoAPI.CreateRecurringPayment(...)
    return nil
}
```

```go
// internal/subscription/manager.go
package subscription

import "log"

// UpgradeSubscription(userID, planType string) error
func UpgradeSubscription(userID, planType string) error {
    log.Printf("[SUBSCRIPTION] UpgradeSubscription user=%s plan=%s (payment disabled)", userID, planType)
    return nil
}

// CheckExpiredSubscriptions() error (Scheduled daily)
func CheckExpiredSubscriptions() error {
    log.Printf("[SUBSCRIPTION] CheckExpiredSubscriptions (payment disabled)")
    return nil
}

// ProcessRenewal(userID string) error
func ProcessRenewal(userID string) error {
    log.Printf("[SUBSCRIPTION] ProcessRenewal user=%s (payment disabled)", userID)
    return nil
}

// SuspendSubscription(userID string) error
func SuspendSubscription(userID string) error {
    log.Printf("[SUBSCRIPTION] SuspendSubscription user=%s", userID)
    return nil
}
```

---

## 11. Deployment ve DevOps

### 11.1 Docker Setup

**docker-compose.yml** (Production-ready)

```yaml
version: '3.8'
services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: chatbot_saas
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  qdrant:
    image: qdrant/qdrant:latest
    environment:
      QDRANT_API_KEY: ${QDRANT_API_KEY}
    volumes:
      - qdrant_data:/qdrant/storage
    ports:
      - "6333:6333"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:6333/health"]
      interval: 10s
      timeout: 5s
      retries: 5

  backend:
    build:
      context: .
      dockerfile: Dockerfile
    depends_on:
      postgres:
        condition: service_healthy
      qdrant:
        condition: service_healthy
    environment:
      DB_HOST: postgres
      DB_NAME: chatbot_saas
      DB_USER: postgres
      DB_PASSWORD: ${DB_PASSWORD}
      QDRANT_URL: http://qdrant:6333
      OPENAI_API_KEY: ${OPENAI_API_KEY}
      IYZICO_API_KEY: ${IYZICO_API_KEY}
      JWT_SECRET: ${JWT_SECRET}
      PORT: 8080
    ports:
      - "8080:8080"
    restart: unless-stopped

  caddy:
    image: caddy:latest
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile
      - caddy_data:/data
      - caddy_config:/config
    ports:
      - "80:80"
      - "443:443"
    depends_on:
      - backend
    restart: unless-stopped

volumes:
  postgres_data:
  qdrant_data:
  caddy_data:
  caddy_config:
```

### 11.2 Dockerfile (Backend)

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder
RUN apk add --no-cache git ca-certificates make libmupdf-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o server cmd/server/main.go

# Runtime stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates libmupdf
WORKDIR /root/
COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations
EXPOSE 8080
CMD ["./server"]
```

### 11.3 .env.example

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=chatbot_saas
DB_USER=postgres
DB_PASSWORD=your_secure_password

# Qdrant
QDRANT_URL=http://localhost:6333
QDRANT_API_KEY=your_api_key

# OpenAI
OPENAI_API_KEY=sk-...

# Iyzico (Sandbox)
IYZICO_API_KEY=...
IYZICO_SECRET_KEY=...

# JWT
JWT_SECRET=your_jwt_secret_key_min_32_chars

# Server
PORT=8080
ENVIRONMENT=production

# Frontend
REACT_APP_API_URL=https://api.chatbot.local
```

### 11.4 CI/CD (GitHub Actions)

**.github/workflows/deploy.yml:**

```yaml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.21
      - run: go test ./...

  deploy:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v2
      
      # Build and push Docker image
      - uses: docker/setup-buildx-action@v1
      - uses: docker/login-action@v1
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      
      - uses: docker/build-push-action@v2
        with:
          context: .
          push: true
          tags: ghcr.io/${{ github.repository }}:latest
      
      # SSH deploy
      - name: Deploy to VPS
        uses: appleboy/ssh-action@master
        with:
          host: ${{ secrets.VPS_HOST }}
          username: ${{ secrets.VPS_USER }}
          key: ${{ secrets.VPS_SSH_KEY }}
          script: |
            cd /app/chatbot-saas
            docker-compose pull
            docker-compose up -d
            docker-compose exec -T backend ./server -migrate
```

### 11.5 VPS Hazırlama (Ubuntu 22.04)

**İlk Kurulum:**

```bash
# SSH ile bağlan
ssh root@your_vps_ip

# Sistem güncelle
apt update && apt upgrade -y

# Docker yükle
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Docker Compose yükle
sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Git clone
git clone https://github.com/yourusername/chatbot-saas.git /app/chatbot-saas
cd /app/chatbot-saas

# .env dosya oluştur
cp .env.example .env
# Düzenle: nano .env

# Başlat
docker-compose up -d

# Kontrol
docker-compose ps
docker-compose logs backend
```

### 11.6 SSL/TLS (Let's Encrypt + Caddy)

**Caddyfile:**

```
api.chatbot-saas.com {
  reverse_proxy backend:8080
  encode gzip
  
  # CORS headers
  header {
    Access-Control-Allow-Origin *
    Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS"
  }
}

chatbot-saas.com {
  root * /var/www/html
  file_server
  
  # SPA routing
  try_files {path} {path}/ /index.html
}

# Widget endpoint
widget.chatbot-saas.com {
  reverse_proxy backend:8080
}
```

Caddy otomatik Let's Encrypt sertifikası alır ve yeniler.

---

## 12. Testler ve Kalite Kontrol

### 12.1 Unit Tests (Go)

**Konvensyon:** Her `.go` dosyası için `*_test.go` dosyası

```
internal/auth/password.go
internal/auth/password_test.go ← Test burada
```

**Test Örnekleri:**

1. **Password Hashing Test** (internal/auth/password_test.go)
   - `TestHashPassword` - Hash'in unique olup olmadığını kontrol
   - `TestVerifyPassword` - Doğru şifre doğrulanıyor mu?
   - `TestVerifyPasswordInvalid` - Yanlış şifre reject ediliyor mu?

2. **Chunker Test** (internal/rag/chunker_test.go)
   - `TestChunkTurkishText` - Türçe metin doğru parçalanıyor mu?
   - `TestChunkingPreservesSemantics` - Cümle sınırları korunuyor mu?
   - `TestOverlapCorrect` - Chunk'lar arası overlap var mı?

3. **Embedding Test** (internal/rag/embedding_test.go)
   - `TestBatchEmbedding` - Batch işleme çalışıyor mu?
   - `TestEmbeddingVectorSize` - Vektör boyutu 1536 mı?

**Test Yazma İlkeleri:**
- Table-driven tests kullan
- Mock external APIs (OpenAI, Iyzico)
- Cleanup (defer statements)

### 12.2 Integration Tests

**Yapılacaklar:**
- Database test (fixtures)
- Qdrant test
- API endpoint tests

**Örnek: Chat endpoint test**

```go
func TestChatEndpoint(t *testing.T) {
  // Setup: DB + Qdrant + fixtures
  defer cleanup()
  
  // Test steps
  req := &ChatRequest{Message: "Merhaba"}
  resp, err := client.Chat(req)
  
  // Assert
  if err != nil { t.Fatal(err) }
  if resp.Response == "" { t.Error("No response") }
}
```

### 12.3 Load Testing

**Araçlar:**
- Apache Bench: `ab -n 1000 -c 10 http://localhost:8080/health`
- k6: https://k6.io/
- Locust: https://locust.io/

**Test Senaryoları:**
1. 100 concurrent user
2. 10.000 requests
3. Monitor: Response time, error rate, CPU/Memory

### 12.4 Linting ve Code Quality

**Tools:**
- `golangci-lint` - Birden fazla linter'ı bir arada çalıştırır
- `gofmt` - Format kontrol
- `go vet` - Şüpheli patterns bulur

**Komutlar:**
```bash
gofmt -l .                    # Formatı kontrol et
go vet ./...                  # Vet analizi
golangci-lint run ./...       # Tüm linter'ları çalıştır
```

### 12.5 Frontend Testing (React)

**Setup:** Vitest + React Testing Library

```bash
npm install -D vitest @testing-library/react @testing-library/jest-dom
```

**Test Örnekleri:**

1. **Login Component Test**
   - Form render'lanıyor mu?
   - Email input'ı var mı?
   - Submit button'u çalışıyor mu?

2. **API Integration Test**
   - Mock API response
   - Component data gösteriyor mu?
   - Error handling var mı?

### 12.6 Manual Testing Checklist

**Sürüm öncesi kontrol:**

- [ ] Kayıt et + oturum aç
- [ ] Chatbot oluştur
- [ ] PDF yükle (başarı + başarısızlık)
- [ ] Widget'ı test et
- [ ] Mesaj gönder (response var mı?)
- [ ] Analytics sayfası yükleniyor mu?
- [ ] Ödeme akışını test et (Sandbox)
- [ ] Abonelik yenilemesi tetikleniyor mu?
- [ ] Mobil responsive (tablet + phone)
- [ ] Türçe karakterler (ş, ü, ö, ğ, ç, ı) doğru görünüyor mu?

---

## 13. Launch Checklist

### 13.1 Teknik Hazırlık

- [ ] Tüm environment variables set edildi
- [ ] Database migrations ran successfully
- [ ] Qdrant koleksiyonu oluşturuldu
- [ ] Iyzico Sandbox test edildi
- [ ] OpenAI API key aktif
- [ ] SSL sertifikası yüklü
- [ ] Backup stratejisi hazırlandı
- [ ] Logging aktif (ELK stack veya basit file logging)
- [ ] Error monitoring setup (Sentry, Rollbar)
- [ ] Rate limiting configure edildi

### 13.2 Güvenlik Checklist

- [ ] JWT secret strong (64+ char, random)
- [ ] Password hashing: bcrypt kullanılıyor
- [ ] SQL injection: Parametrized queries
- [ ] CORS: Doğru domain'ler whitelisted
- [ ] HTTPS: Tüm endpoints HTTPS
- [ ] API Keys: Environment variables (hardcoded değil)
- [ ] KVKK compliance: Privacy policy hazır
- [ ] PII redaction: Loglardan sensitive data çıkarıldı
- [ ] Rate limiting: API endpoints korunuyor
- [ ] CSRF protection: State management var (if needed)

### 13.3 Performance Checklist

- [ ] Widget boyutu < 50KB (gzipped)
- [ ] Dashboard page load < 2 saniye
- [ ] API response time < 500ms (median)
- [ ] Database queries optimized (indexes)
- [ ] Vector search < 200ms (Qdrant)
- [ ] CDN configured (Static assets)

### 13.4 Monitoring ve Alerting

**Kurulacaklar:**

1. **Application Monitoring**
   - Error rate threshold (> 1% alert)
   - Response time p99 (> 1s alert)
   - Database connection pool
   - Memory usage (> 80% alert)

2. **Infrastructure Monitoring**
   - CPU usage
   - Disk space
   - Network I/O
   - Docker container health

3. **Business Metrics**
   - New user signups (daily)
   - Subscription conversions
   - Revenue tracking
   - Active users

**Tools:** Prometheus + Grafana, New Relic, DataDog

### 13.5 Backup ve Disaster Recovery

**Backup Plan:**

1. **PostgreSQL**
   - Daily full backup
   - Weekly incremental
   - S3 veya cloud storage'a yükle
   - Test: Monthly restore test

2. **Qdrant**
   - Daily snapshot
   - Cloud storage'a yükle
   - Recovery time < 1 hour

3. **Restoration Procedure:**
   - DB restore script
   - Qdrant recovery steps
   - RTO (Recovery Time Objective): 4 saat
   - RPO (Recovery Point Objective): 1 gün

### 13.6 Müşteri Hazırlığı

- [ ] Rehberler yazıldı (Kurulum, kullanım, sık sorular)
- [ ] Widget embed kodu jeneratörü
- [ ] API documentation (OpenAPI/Swagger)
- [ ] Support email setup
- [ ] Billing page tamamlandı
- [ ] Email templates (Welcome, Payment confirmation, Subscription expiry)

### 13.7 Beta Launch

**Aşama 1: Closed Beta (10-20 müşteri)**
- Yakın arkadaşlar, kollegler
- Feedback toplayın
- Bug fix'leri yapın
- Performance optimize edin

**Aşama 2: Open Beta**
- Herkesin erişimine aç
- Feature complete olmasa da çalışan sistem
- Müşteri feedback devam ediyor

**Aşama 3: Production (v1.0)**
- Tüm critical bugs fixed
- Performance tested
- Security audit completed
- Full documentation ready

---

## 14. Sonrası: Büyüme ve Optimizasyon

### 14.1 Müşteri Feedback Döngüsü

1. **Feedback Collection**
   - In-app feedback widget
   - Email surveys
   - Support tickets analysis

2. **Prioritization**
   - Feature requests scoring
   - Bug severity levels
   - User impact assessment

3. **Implementation**
   - Sprint planning
   - Development
   - Beta testing with customer

### 14.2 Scaling Stratejisi

**Şu anki yapı (2GB VPS) handle edebilir:**
- ~50 active chatbots
- ~1000 conversations/day
- ~5000 messages/day

**Scaling Points:**

**Level 1: Database Optimization (0 cost)**
- Query optimization
- Index optimization
- Connection pooling tuning

**Level 2: Caching Layer ($20/month)**
- Redis cache (Embedding results)
- API response cache
- Session cache

**Level 3: Multi-server ($50+/month)**
- Load balancer
- Multiple API instances
- Dedicated database server
- Dedicated Qdrant server

**Level 4: Kubernetes (Enterprise)**
- Auto-scaling
- Global distribution
- Advanced monitoring

### 14.3 Maliyet Modeli

**Aylık Operating Costs (İlk aşama):**

| Item | Cost |
|------|------|
| VPS (2GB, 2 vCPU) | $10 |
| Domain | $1 |
| SSL Certificate | $0 (Let's Encrypt) |
| OpenAI Embeddings | ~$50 (100K requests/month) |
| OpenAI Completions | ~$100 (100K tokens/month) |
| **Total** | ~$161 |

**Revenue Model (Turkish Pricing):**
- Free: 0₺ (Ad support veya API limited)
- Starter: 399₺/month → 70₺ net profit/user
- Professional: 999₺/month → 150₺ net profit/user
- Enterprise: 2999₺/month → 500₺+ net profit/user

**Break-even point:**
- 3 Starter user = 210₺ revenue
- Operating cost: ~500₺
- Need: ~7-10 Starter users

### 14.4 Feature Roadmap (İlk 6 ay)

**Month 1-2: MVP Launch**
- Basic chat functionality
- Simple PDF upload
- Simple dashboard

**Month 3: Enhancements**
- Web scraping (Colly)
- Better analytics
- Whitelabel basic support

**Month 4: Monetization**
- Payment integration
- Tiered pricing
- Usage tracking

**Month 5-6: Scale**
- Performance optimization
- Multi-language support (EN, DE)
- API for integrations

---

## 15. Go Learning Resources (While Building)

### 15.1 Kavramlar (Priority Order)

1. **Week 1-2: Basics**
   - goroutines ve channels
   - Error handling (if err != nil)
   - Struct ve interface'ler
   - Documentation: https://go.dev/tour

2. **Week 3-4: Web**
   - net/http package
   - JSON marshaling
   - Database/sql
   - Book: "Web Development with Go" (https://www.usegolang.com/)

3. **Week 5+: Advanced**
   - Context ve cancellation
   - Middleware patterns
   - Testing (table-driven tests)
   - Documentation: Effective Go (https://go.dev/doc/effective_go)

### 15.2 Kütüphane Seçim Rehberi

**HTTP Framework seçimi:**
- **chi**: Minimal, standart library friendly (Önerilir MVP için)
- **Gin**: Full-featured, fast (Scaling için)
- **Echo**: Lightweight, good middleware

**Database ORM:**
- **sqlc**: Type-safe SQL (En iyisi)
- **gorm**: Full-featured ORM (Hızlı başlangıç)
- **sql.DB directly**: Ultimate control (Ama verbose)

**Logging:**
- **logrus**: Structured logging
- **zap**: High performance
- **slog**: Standart library (Go 1.21+) - Yeni başlayanlar için iyi

### 15.3 Sık Hatalar ve Çözümleri

| Hata | Sebep | Çözüm |
|------|-------|-------|
| `database/sql: no rows in result set` | Query sonuç vermedi | `sql.ErrNoRows` kontrol et |
| `context deadline exceeded` | İstek çok uzun sürdü | Timeout'u arttır veya query optimize et |
| `goroutine leak` | Goroutine'ler kapatılmadı | Channel'ları yazar, context.Done() kontrol et |
| `Deadlock` | Goroutine'ler birbirini bekliyor | Channel sıralaması review et |
| `Out of Memory` | Vektör DB çok RAM tüketiyor | Qdrant mmap aktif et, RAM limit set et |

---

## 16. Devam Eden Destek Kaynakları

### 16.1 Community

- **Go Discord**: https://discord.gg/golang
- **Reddit r/golang**: https://reddit.com/r/golang
- **Stack Overflow**: Tag: [go] veya [golang]

### 16.2 Documentation

- **Official Docs**: https://go.dev/doc
- **Package Reference**: https://pkg.go.dev
- **Ebook: The Go Programming Language**: Donovan & Kernighan

### 16.3 DevOps

- **Docker Best Practices**: https://docs.docker.com/develop/dev-best-practices
- **Docker Compose**: https://docs.docker.com/compose/
- **GitHub Actions**: https://github.com/features/actions

### 16.4 Türçe Kaynaklar

- **Medium (Go yazıları)**: Türkçe yazarları takip et
- **GitHub (Türk projeler)**: Kod incelemesi
- **Meetup.com**: Yerel Go meetup'larına katıl

---

## 17. Zaman Tahmini

Baştan sona geliştirme süresi (**solo geliştirici**):

| Aşama | Task | Saat | Haftalar |
|-------|------|------|----------|
| Setup | Ortam + İlk repo | 8 | 0.2 |
| Backend Auth | Login + Register | 16 | 0.4 |
| CRUD | Chatbot operations | 20 | 0.5 |
| Scraping | Colly + Encoding | 24 | 0.6 |
| PDF | MuPDF + OCR | 20 | 0.5 |
| RAG | Chunking + Embedding | 32 | 0.8 |
| Chat | Full pipeline | 24 | 0.6 |
| Frontend | Dashboard React | 48 | 1.2 |
| Widget | Preact + Shadow DOM | 32 | 0.8 |
| Payment | Iyzico integration | 16 | 0.4 |
| Testing | Tests + QA | 40 | 1 |
| Deployment | Docker + CI/CD | 20 | 0.5 |
| Docs + Launch | Documentation | 24 | 0.6 |
| **TOTAL** | | **384** | **8 hafta** |

**Hızlandırma stratejileri:**
- MVP: Sadece PDF + metin kaynakları (ilk 4 hafta)
- Scraping sonra ekle (hafta 5-6)
- Payment başta Stripe (Iyzico daha sonra)
- Analytics basit başlasın

---

## 18. Troubleshooting Rehberi

### 18.1 Ortak Sorunlar

**Problem: "Türçe karakterler bozuk gözüküyor"**
- Database character set: utf8mb4 veya utf8
- PDF extraction: Encoding detection başarısız mı?
- Widget: CSS font-family türçeyi support ediyor mu?

**Problem: "Widget yavaş yükleniyor"**
- Preact bundle boyutu > 50KB mı? → Tree-shake
- CSS file inline mi? → Evet olmalı
- API latency > 500ms? → Backend optimize

**Problem: "Qdrant high memory"**
- mmap enabled mi? `storage:` config
- Koleksiyon sayısı çok mu? → Multitenancy'e geç
- Vector size ne? (384 vs 1536)

**Problem: "Belli user'lar için chatbot çalışmıyor"**
- Ownership check: Chatbot bu user'a mı ait?
- Token expired? → Refresh token implement et
- Source process status? → pending'de takılmış mı?

### 18.2 Debug Techniques

**Go Debugging:**
```bash
# Verbose logging
export LOG_LEVEL=DEBUG

# Profiling
go tool pprof http://localhost:6060/debug/pprof/profile

# Race detection
go run -race cmd/server/main.go
```

**Database Query Debug:**
```go
// Query ile log
if err := db.QueryRowContext(ctx, query, args...).Scan(&result); err != nil {
  log.Printf("Query failed: %q Args: %v Error: %v", query, args, err)
}
```

---

## Son Notlar

Bu rehber, **MVP**'den **Production**'a giden tüm adımları kapsar. 

**Yapılması Gerekenler:**
1. **Bu dokümanı çıktı al** veya bookmark'le
2. **Adım adım takip et** - Jump around yapma
3. **Errors'ı Google'da ara** - Go community çok yardımcı
4. **İlk müşteri bul** - Tamamlandıktan hemen sonra
5. **Feedback al ve iterate et**

**En Önemli Tavsiye:**
> "Done is better than perfect. MVP ile başla, sonra iterate et."

---

**Yazarı:** Chatbot SaaS Development Guide
**Son Güncelleme:** 2024
**Sürüm:** 1.0
