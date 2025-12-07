# Plan 3.2: Custom Domain Routing

## Özet

Müşterilerin kendi alan adlarında (chat.example.com) chatbot widget'ı sunabilmesi.

---

## Hedef Mimari

```
┌────────────────────────────────────────────────────────────────┐
│                     Custom Domain Flow                          │
├────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. Müşteri: chat.example.com → Botla'ya CNAME                 │
│                                                                 │
│  2. DNS Doğrulama:                                              │
│     chat.example.com CNAME → widget.botla.co                   │
│     _botla.example.com TXT → "verify=abc123"                   │
│                                                                 │
│  3. SSL Otomasyonu (Let's Encrypt / Caddy)                     │
│                                                                 │
│  4. Reverse Proxy:                                              │
│     Request → Caddy → Host header → Org lookup → Serve         │
│                                                                 │
└────────────────────────────────────────────────────────────────┘
```

---

## Uygulama Adımları

### Adım 1: Veritabanı Migration

**Dosya:** `db/migrations/000018_custom_domains.up.sql`

```sql
CREATE TABLE IF NOT EXISTS custom_domains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    chatbot_id UUID REFERENCES chatbots(id) ON DELETE CASCADE, -- NULL = org-level
    domain TEXT UNIQUE NOT NULL,
    verification_token TEXT NOT NULL,
    verified BOOLEAN DEFAULT FALSE,
    verified_at TIMESTAMP,
    ssl_issued BOOLEAN DEFAULT FALSE,
    ssl_expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    
    CONSTRAINT domain_format CHECK (domain ~ '^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)+$')
);

CREATE INDEX idx_custom_domains_domain ON custom_domains(domain) WHERE verified = true;
```

### Adım 2: Models

**Dosya:** `internal/models/custom_domain.go` (YENİ)

```go
type CustomDomain struct {
    ID                string     `json:"id"`
    OrganizationID    string     `json:"organization_id"`
    ChatbotID         *string    `json:"chatbot_id,omitempty"`
    Domain            string     `json:"domain"`
    VerificationToken string     `json:"verification_token"`
    Verified          bool       `json:"verified"`
    VerifiedAt        *time.Time `json:"verified_at,omitempty"`
    SSLIssued         bool       `json:"ssl_issued"`
    SSLExpiresAt      *time.Time `json:"ssl_expires_at,omitempty"`
    CreatedAt         time.Time  `json:"created_at"`
}

type DomainVerificationStatus struct {
    Domain   string `json:"domain"`
    CNAME    bool   `json:"cname_configured"`
    TXT      bool   `json:"txt_configured"`
    Verified bool   `json:"verified"`
}
```

### Adım 3: Domain Service

**Dosya:** `internal/services/domain_service.go` (YENİ)

```go
type DomainService struct {
    DB  *sql.DB
    Log *logger.Logger
}

// AddDomain registers a new custom domain
func (s *DomainService) AddDomain(ctx context.Context, orgID string, domain string, chatbotID *string) (*CustomDomain, error) {
    // Validate domain format
    // Generate verification token
    // Insert to DB
}

// VerifyDomain checks DNS configuration
func (s *DomainService) VerifyDomain(ctx context.Context, domainID string) (*DomainVerificationStatus, error) {
    domain, _ := db.GetDomain(ctx, s.DB, domainID)
    
    // Check CNAME
    cnameOK := s.checkCNAME(domain.Domain)
    
    // Check TXT
    txtOK := s.checkTXT(domain.Domain, domain.VerificationToken)
    
    if cnameOK && txtOK {
        db.MarkDomainVerified(ctx, s.DB, domainID)
        // Trigger SSL issuance
        go s.requestSSL(domain.Domain)
    }
    
    return &DomainVerificationStatus{
        Domain:   domain.Domain,
        CNAME:    cnameOK,
        TXT:      txtOK,
        Verified: cnameOK && txtOK,
    }, nil
}

func (s *DomainService) checkCNAME(domain string) bool {
    cname, err := net.LookupCNAME(domain)
    return err == nil && strings.HasSuffix(cname, "botla.co.")
}

func (s *DomainService) checkTXT(domain string, token string) bool {
    records, err := net.LookupTXT("_botla." + domain)
    if err != nil {
        return false
    }
    for _, r := range records {
        if strings.Contains(r, "verify="+token) {
            return true
        }
    }
    return false
}
```

### Adım 4: Caddy Konfigürasyonu

**Dosya:** `Caddyfile` (GÜNCELLE)

```
# Dynamic domain handler
{
    on_demand_tls {
        ask http://localhost:8080/api/internal/verify-domain
        interval 5m
        burst 5
    }
}

:443 {
    tls {
        on_demand
    }
    
    reverse_proxy localhost:8080 {
        header_up X-Forwarded-Host {host}
    }
}
```

### Adım 5: Internal API for Caddy

**Dosya:** `internal/api/handlers/internal.go` (YENİ)

```go
// GET /api/internal/verify-domain?domain=chat.example.com
// Called by Caddy to check if domain is allowed
func (h *InternalHandlers) VerifyDomainForSSL(c *gin.Context) {
    domain := c.Query("domain")
    
    // Check if domain exists and is verified
    d, err := db.GetVerifiedDomain(c.Request.Context(), h.DB, domain)
    if err != nil || d == nil {
        c.Status(http.StatusNotFound)
        return
    }
    
    c.Status(http.StatusOK)
}
```

### Adım 6: Widget Hostname Resolution

**Dosya:** `internal/api/handlers/public.go`

```go
func (h *PublicHandlers) PublicChatbotConfig(c *gin.Context) {
    botID := c.Param("botId")
    
    // Check if request is from custom domain
    host := c.GetHeader("X-Forwarded-Host")
    if host != "" && host != "botla.co" {
        // Lookup chatbot by custom domain
        domain, _ := db.GetDomainByHost(c.Request.Context(), h.DB, host)
        if domain != nil && domain.ChatbotID != nil {
            botID = *domain.ChatbotID
        }
    }
    
    // Continue with normal flow...
}
```

### Adım 7: API Endpoints

```
POST   /api/organizations/:orgId/domains
GET    /api/organizations/:orgId/domains
DELETE /api/organizations/:orgId/domains/:domainId
POST   /api/organizations/:orgId/domains/:domainId/verify
GET    /api/organizations/:orgId/domains/:domainId/status
```

### Adım 8: Frontend UI

**UI:**

```
┌─────────────────────────────────────────────────────────────┐
│ 🌐 Özel Alan Adları                                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ chat.example.com                                        │ │
│ │ ✅ Doğrulandı  │  🔒 SSL Aktif                          │ │
│ │ Chatbot: Support Bot                                    │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                             │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ help.mysite.com                           [🔄 Doğrula]  │ │
│ │ ⏳ Doğrulama Bekleniyor                                 │ │
│ │                                                         │ │
│ │ DNS Ayarları:                                           │ │
│ │ ┌───────────────────────────────────────────────────┐   │ │
│ │ │ Type: CNAME                                        │   │ │
│ │ │ Name: help                                         │   │ │
│ │ │ Value: widget.botla.co                            │   │ │
│ │ └───────────────────────────────────────────────────┘   │ │
│ │ ┌───────────────────────────────────────────────────┐   │ │
│ │ │ Type: TXT                                          │   │ │
│ │ │ Name: _botla                                       │   │ │
│ │ │ Value: verify=abc123def456                        │   │ │
│ │ └───────────────────────────────────────────────────┘   │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                             │
│                              [+ Yeni Alan Adı Ekle]         │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Dosya Değişiklikleri Özeti

| Dosya | İşlem | Açıklama |
|-------|-------|----------|
| `db/migrations/000018_*.sql` | YENİ | Tablolar |
| `internal/models/custom_domain.go` | YENİ | Model |
| `internal/services/domain_service.go` | YENİ | DNS + SSL logic |
| `internal/api/handlers/domain.go` | YENİ | API |
| `internal/api/handlers/internal.go` | YENİ | Caddy callback |
| `internal/api/handlers/public.go` | GÜNCELLE | Host resolution |
| `Caddyfile` | GÜNCELLE | On-demand TLS |
| `frontend/src/features/domains/*` | YENİ | UI |

---

## Test Planı

### Unit Testler

```go
func TestCheckCNAME(t *testing.T) {
    // Mock DNS resolver
}

func TestCheckTXT(t *testing.T) {
    // Mock DNS resolver
}

func TestPublicChatbotConfig_CustomDomain(t *testing.T) {
    // Request with X-Forwarded-Host
    // Verify correct chatbot returned
}
```

### Manuel Test

1. Custom domain ekle
2. DNS'te CNAME ve TXT ekle
3. Doğrula butonuna bas
4. SSL sertifikası oluştuğunu bekle (1-2 dk)
5. https://chat.example.com adresine git
6. Widget'ın çalıştığını doğrula

---

## Tahmini Süre

| Görev | Süre |
|-------|------|
| Migration + Models | 1-2 saat |
| Domain service | 4-6 saat |
| Caddy config | 2-4 saat |
| API endpoints | 3-4 saat |
| Host resolution | 2-3 saat |
| Frontend UI | 4-6 saat |
| Testler | 3-4 saat |
| **TOPLAM** | **~1.5 hafta** |

---

## Bağımlılıklar

**Önceki:** Plan 3.1 (Multi-Tenant)

**Sonraki:** Bağımsız
