# Plan 1.7: White-Label ve Branding Kaldırma

## Özet

Plan bazlı "Powered by Botla" ibaresini kaldırma ve özelleştirilebilir branding seçenekleri.

---

## Mevcut Durum

### İlgili Dosyalar

| Dosya | Mevcut Durum |
|-------|--------------|
| `widget/src/components/ChatDrawer.tsx` | Sabit branding gösterimi |
| `internal/models/chatbot.go` | hide_branding alanı **yok** |
| Plan config | Branding özelliği **yok** |

---

## Hedef Mimari

```
┌────────────────────────────────────────┐
│           Plan Kontrolü                │
├────────────────────────────────────────┤
│                                        │
│  Free Plan:                            │
│    - hide_branding: false (zorunlu)    │
│    - "Powered by Botla.co" gösterilir  │
│                                        │
│  Pro Plan:                             │
│    - hide_branding: true/false         │
│    - Kullanıcı seçebilir               │
│                                        │
│  Enterprise Plan:                      │
│    - hide_branding: true (varsayılan)  │
│    - custom_branding: { logo, text }   │
│                                        │
└────────────────────────────────────────┘
```

---

## Uygulama Adımları

### Adım 1: Veritabanı Migration

**Dosya:** `db/migrations/000013_branding_options.up.sql`

```sql
ALTER TABLE chatbots
ADD COLUMN IF NOT EXISTS hide_branding BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS custom_branding JSONB DEFAULT NULL;

COMMENT ON COLUMN chatbots.hide_branding IS 'Hide Powered by Botla branding';
COMMENT ON COLUMN chatbots.custom_branding IS 'Custom branding config: {logo_url, text, link}';

-- Plan config'e branding izni ekle
-- (Bu zaten plan JSON'ında yönetiliyor, migration gereksiz olabilir)
```

### Adım 2: Model Güncelleme

**Dosya:** `internal/models/chatbot.go`

```go
HideBranding    bool            `json:"hide_branding"`
CustomBranding  *CustomBranding `json:"custom_branding,omitempty"`

type CustomBranding struct {
    LogoURL string `json:"logo_url,omitempty"`
    Text    string `json:"text,omitempty"`
    Link    string `json:"link,omitempty"`
}
```

### Adım 3: Plan Config Güncelleme

**Dosya:** `pkg/plans/plans.go` (veya ilgili config)

```go
type PlanConfig struct {
    // ... mevcut alanlar
    Branding BrandingConfig `json:"branding"`
}

type BrandingConfig struct {
    CanHideBranding   bool `json:"can_hide_branding"`
    CanCustomBranding bool `json:"can_custom_branding"`
}
```

**Plan değerleri:**
```json
{
  "free": {
    "branding": { "can_hide_branding": false, "can_custom_branding": false }
  },
  "pro": {
    "branding": { "can_hide_branding": true, "can_custom_branding": false }
  },
  "enterprise": {
    "branding": { "can_hide_branding": true, "can_custom_branding": true }
  }
}
```

### Adım 4: API Endpoint Güncellemesi

**Dosya:** `internal/api/handlers/chatbot.go`

**Değişiklikler:**

1. `UpdateChatbot` handler'ında branding kontrolü:

```go
func (h *ChatbotHandlers) UpdateChatbot(c *gin.Context) {
    // ... mevcut kod
    
    // Branding değişikliği kontrolü
    if req.HideBranding != nil && *req.HideBranding {
        if !plan.Config.Branding.CanHideBranding {
            c.JSON(http.StatusForbidden, gin.H{
                "error": "Your plan does not allow hiding branding",
                "upgrade_required": true,
            })
            return
        }
    }
    
    if req.CustomBranding != nil {
        if !plan.Config.Branding.CanCustomBranding {
            c.JSON(http.StatusForbidden, gin.H{
                "error": "Custom branding requires Enterprise plan",
                "upgrade_required": true,
            })
            return
        }
    }
}
```

### Adım 5: Public Chatbot Config API'sı

**Dosya:** `internal/api/handlers/public.go`

**Güncelleme:** `PublicChatbotConfig` response'una branding bilgisi ekle:

```go
type publicChatbot struct {
    // ... mevcut alanlar
    HideBranding   bool            `json:"hide_branding"`
    CustomBranding *CustomBranding `json:"custom_branding,omitempty"`
}
```

### Adım 6: Widget Güncellemesi

**Dosya:** `widget/src/components/ChatDrawer.tsx`

**Mevcut branding kodu:**
```tsx
// Muhtemelen sabit bir "Powered by Botla" var
```

**Yeni mantık:**
```tsx
interface ChatDrawerProps {
    // ... mevcut props
    hideBranding?: boolean;
    customBranding?: {
        logoUrl?: string;
        text?: string;
        link?: string;
    };
}

// Footer bileşeni
const BrandingFooter = ({ hideBranding, customBranding }) => {
    if (hideBranding && !customBranding) {
        return null;
    }
    
    if (customBranding) {
        return (
            <div className="branding-footer">
                {customBranding.logoUrl && <img src={customBranding.logoUrl} />}
                <a href={customBranding.link}>{customBranding.text}</a>
            </div>
        );
    }
    
    // Default branding
    return (
        <div className="branding-footer">
            <a href="https://botla.co" target="_blank">
                Powered by Botla.co
            </a>
        </div>
    );
};
```

### Adım 7: Widget App'e Props Aktarımı

**Dosya:** `widget/src/widgetApp.tsx`

```tsx
// Config'ten branding bilgisini al
const { hideBranding, customBranding } = botConfig;

<ChatDrawer
    // ... mevcut props
    hideBranding={hideBranding}
    customBranding={customBranding}
/>
```

### Adım 8: Frontend Ayarları

**Dosya:** `frontend/src/features/chatbot/BrandingSettings.tsx` (YENİ)

**UI Tasarımı:**

```
┌─────────────────────────────────────────────────────────────┐
│ 🏷️ Branding Ayarları                                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ ☑️ "Powered by Botla.co" ibaresini gizle                    │
│                                                             │
│ ─────────────────────────────────────────────────────────── │
│                                                             │
│ 🎨 Özel Branding (Enterprise)                    [🔒 Upgrade]│
│                                                             │
│ Logo URL:                                                   │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ https://mycompany.com/logo.png                          │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                             │
│ Metin:                                                      │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ Powered by MyCompany                                    │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                             │
│ Link:                                                       │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ https://mycompany.com                                   │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Dosya Değişiklikleri Özeti

| Dosya | İşlem | Açıklama |
|-------|-------|----------|
| `db/migrations/000013_*.sql` | YENİ | Migration |
| `internal/models/chatbot.go` | GÜNCELLE | Branding alanları |
| `pkg/plans/plans.go` | GÜNCELLE | Plan config |
| `internal/api/handlers/chatbot.go` | GÜNCELLE | Plan kontrolü |
| `internal/api/handlers/public.go` | GÜNCELLE | Public API |
| `widget/src/components/ChatDrawer.tsx` | GÜNCELLE | Branding render |
| `widget/src/widgetApp.tsx` | GÜNCELLE | Props aktarımı |
| `frontend/src/features/chatbot/BrandingSettings.tsx` | YENİ | UI |

---

## Test Planı

### Unit Testler

**Dosya:** `internal/api/handlers/chatbot_test.go`

```go
func TestUpdateChatbot_HideBranding_FreePlan(t *testing.T) {
    // Free plan'da hide_branding = true yapılmaya çalışılırsa
    // 403 Forbidden döndürmeli
}

func TestUpdateChatbot_HideBranding_ProPlan(t *testing.T) {
    // Pro plan'da hide_branding = true başarılı olmalı
}

func TestUpdateChatbot_CustomBranding_ProPlan(t *testing.T) {
    // Pro plan'da custom_branding 403 döndürmeli
}

func TestUpdateChatbot_CustomBranding_EnterprisePlan(t *testing.T) {
    // Enterprise'da custom_branding başarılı olmalı
}
```

### Widget Testi

**Dosya:** `widget/src/components/ChatDrawer.test.tsx`

```tsx
test('shows default branding when hideBranding is false', () => {
    render(<ChatDrawer hideBranding={false} />);
    expect(screen.getByText('Powered by Botla.co')).toBeInTheDocument();
});

test('hides branding when hideBranding is true', () => {
    render(<ChatDrawer hideBranding={true} />);
    expect(screen.queryByText('Powered by Botla.co')).not.toBeInTheDocument();
});

test('shows custom branding when provided', () => {
    const custom = { text: 'Powered by Test', link: 'https://test.com' };
    render(<ChatDrawer customBranding={custom} />);
    expect(screen.getByText('Powered by Test')).toBeInTheDocument();
});
```

### Manuel Test Prosedürü

1. **Free Plan Test:**
   - Free kullanıcı ile giriş
   - Chatbot → Ayarlar → Branding
   - "Gizle" checkbox'ının disabled olduğunu doğrula
   - Widget'ta "Powered by Botla" görünür

2. **Pro Plan Test:**
   - Pro kullanıcı ile giriş
   - Branding → Gizle seç, kaydet
   - Widget'ta branding görünmez

3. **Enterprise Test:**
   - Enterprise kullanıcı
   - Custom branding ayarla
   - Widget'ta custom logo/text görünür

---

## Doğrulama Kriterleri

| Kriter | Doğrulama Yöntemi |
|--------|-------------------|
| ✅ Migration başarılı | `make migrate-up` |
| ✅ Plan kontrolü çalışıyor | Unit test |
| ✅ Widget branding koşullu | Widget test |
| ✅ API doğru çalışıyor | API test |
| ✅ Frontend UI | Manuel test |
| ✅ %90 coverage | `make cover-gate` |

---

## Tahmini Süre

| Görev | Süre |
|-------|------|
| Migration | 30 dk |
| Model + Plan config | 1-2 saat |
| API kontrolü | 2-3 saat |
| Widget update | 2-3 saat |
| Frontend UI | 2-3 saat |
| Testler | 2-3 saat |
| **TOPLAM** | **~3-4 gün** |

---

## Bağımlılıklar

**Önceki:** Yok (bağımsız)

**Sonraki:** Plan 3.1 (Multi-Tenant) için temel
