# Plan 3.1: Çok Kiracılı Organizasyon Yapısı (Multi-Tenant)

## Özet

Ajansların birden fazla müşteri için chatbot yönetebilmesi için organizasyon ve workspace yapısı.

---

## Hedef Mimari

```
┌────────────────────────────────────────────────────────────────┐
│                        Organization                             │
│                     (Ajans: XYZ Digital)                       │
├────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │ Workspace 1 │  │ Workspace 2 │  │ Workspace 3 │             │
│  │ Client: ABC │  │ Client: DEF │  │ Client: GHI │             │
│  ├─────────────┤  ├─────────────┤  ├─────────────┤             │
│  │ - Chatbot A │  │ - Chatbot C │  │ - Chatbot E │             │
│  │ - Chatbot B │  │ - Chatbot D │  │             │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
│                                                                 │
│  Members:                                                       │
│  - owner@xyzdigital.com (Owner)                                │
│  - admin@xyzdigital.com (Admin)                                │
│  - dev@xyzdigital.com (Member)                                 │
│                                                                 │
└────────────────────────────────────────────────────────────────┘
```

---

## Uygulama Adımları

### Adım 1: Veritabanı Migration

**Dosya:** `db/migrations/000017_multi_tenant.up.sql`

```sql
-- Organizations
CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    slug TEXT UNIQUE NOT NULL,
    owner_id UUID NOT NULL REFERENCES users(id),
    plan_id TEXT DEFAULT 'agency_starter',
    branding JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Organization memberships
CREATE TABLE IF NOT EXISTS memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'member', -- owner, admin, member
    created_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(organization_id, user_id)
);

-- Workspaces (client groupings)
CREATE TABLE IF NOT EXISTS workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    client_name TEXT,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(organization_id, slug)
);

-- Update chatbots to belong to workspace (optional)
ALTER TABLE chatbots
ADD COLUMN IF NOT EXISTS workspace_id UUID REFERENCES workspaces(id),
ADD COLUMN IF NOT EXISTS organization_id UUID REFERENCES organizations(id);

-- Index for fast lookups
CREATE INDEX idx_memberships_user ON memberships(user_id);
CREATE INDEX idx_workspaces_org ON workspaces(organization_id);
CREATE INDEX idx_chatbots_workspace ON chatbots(workspace_id);
```

### Adım 2: Models

**Dosya:** `internal/models/organization.go` (YENİ)

```go
type Organization struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Slug      string    `json:"slug"`
    OwnerID   string    `json:"owner_id"`
    PlanID    string    `json:"plan_id"`
    Branding  *Branding `json:"branding,omitempty"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type Membership struct {
    ID             string    `json:"id"`
    OrganizationID string    `json:"organization_id"`
    UserID         string    `json:"user_id"`
    Role           string    `json:"role"` // owner, admin, member
    CreatedAt      time.Time `json:"created_at"`
}

type Workspace struct {
    ID             string    `json:"id"`
    OrganizationID string    `json:"organization_id"`
    Name           string    `json:"name"`
    Slug           string    `json:"slug"`
    ClientName     *string   `json:"client_name,omitempty"`
    CreatedAt      time.Time `json:"created_at"`
}
```

### Adım 3: Organization Service

**Dosya:** `internal/services/organization_service.go` (YENİ)

```go
type OrganizationService struct {
    DB  *sql.DB
    Log *logger.Logger
}

// CreateOrganization creates a new organization with the owner
func (s *OrganizationService) CreateOrganization(ctx context.Context, name, slug string, ownerID string) (*Organization, error)

// AddMember adds a user to the organization
func (s *OrganizationService) AddMember(ctx context.Context, orgID, userID, role string) error

// CreateWorkspace creates a new workspace
func (s *OrganizationService) CreateWorkspace(ctx context.Context, orgID, name, slug string) (*Workspace, error)

// GetUserOrganizations returns all organizations a user belongs to
func (s *OrganizationService) GetUserOrganizations(ctx context.Context, userID string) ([]*Organization, error)

// GetOrganizationMembers returns all members of an organization
func (s *OrganizationService) GetOrganizationMembers(ctx context.Context, orgID string) ([]*Membership, error)
```

### Adım 4: Authorization Middleware

**Dosya:** `pkg/middleware/organization.go` (YENİ)

```go
// RequireOrganizationAccess checks if user has access to organization
func RequireOrganizationAccess(minRole string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("user_id")
        orgID := c.Param("orgId")
        
        membership, err := getMembership(c.Request.Context(), userID, orgID)
        if err != nil || membership == nil {
            c.AbortWithStatusJSON(403, gin.H{"error": "Access denied"})
            return
        }
        
        if !hasMinRole(membership.Role, minRole) {
            c.AbortWithStatusJSON(403, gin.H{"error": "Insufficient permissions"})
            return
        }
        
        c.Set("organization_id", orgID)
        c.Set("membership", membership)
        c.Next()
    }
}
```

### Adım 5: API Endpoints

**Dosya:** `internal/api/handlers/organization.go` (YENİ)

```
# Organization CRUD
POST   /api/organizations
GET    /api/organizations
GET    /api/organizations/:orgId
PATCH  /api/organizations/:orgId
DELETE /api/organizations/:orgId

# Members
GET    /api/organizations/:orgId/members
POST   /api/organizations/:orgId/members
DELETE /api/organizations/:orgId/members/:memberId

# Workspaces
GET    /api/organizations/:orgId/workspaces
POST   /api/organizations/:orgId/workspaces
GET    /api/organizations/:orgId/workspaces/:wsId
PATCH  /api/organizations/:orgId/workspaces/:wsId
DELETE /api/organizations/:orgId/workspaces/:wsId

# Workspace chatbots
GET    /api/organizations/:orgId/workspaces/:wsId/chatbots
POST   /api/organizations/:orgId/workspaces/:wsId/chatbots
```

### Adım 6: Frontend - Organization Switcher

**Dosya:** `frontend/src/components/OrganizationSwitcher.tsx` (YENİ)

**UI:**

```
┌─────────────────────────────────────────────────────────────┐
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ 🏢 XYZ Digital                                      ▼  │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                             │
│ ├── 📁 ABC Corporation                                      │
│ │   ├── 🤖 Support Bot                                     │
│ │   └── 🤖 Sales Bot                                       │
│ │                                                           │
│ ├── 📁 DEF Industries                                       │
│ │   └── 🤖 FAQ Bot                                         │
│ │                                                           │
│ └── + Yeni Workspace Ekle                                   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Adım 7: Billing Değişiklikleri

**Organizasyon bazlı plan:**
- Tek fatura organizasyon seviyesinde
- Kaynak limitleri organizasyon bazında
- Workspace'ler arasında paylaşımlı

---

## Dosya Değişiklikleri Özeti

| Dosya | İşlem | Açıklama |
|-------|-------|----------|
| `db/migrations/000017_*.sql` | YENİ | Tablolar |
| `internal/models/organization.go` | YENİ | Models |
| `internal/services/organization_service.go` | YENİ | Business logic |
| `pkg/middleware/organization.go` | YENİ | Auth middleware |
| `internal/api/handlers/organization.go` | YENİ | API handlers |
| `internal/api/router.go` | GÜNCELLE | Routes |
| `frontend/src/components/OrganizationSwitcher.tsx` | YENİ | UI |
| `frontend/src/pages/*` | GÜNCELLE | Org context |

---

## Test Planı

### Unit Testler

```go
func TestCreateOrganization(t *testing.T) {
    // Create org with owner
    // Verify owner membership created
}

func TestAddMember(t *testing.T) {
    // Add member with role
    // Verify membership
}

func TestOrganizationAccess(t *testing.T) {
    // User without membership → 403
    // User with member role, admin required → 403
    // User with admin role → 200
}
```

### Manuel Test

1. Yeni organizasyon oluştur
2. Başka kullanıcı davet et
3. Workspace oluştur
4. Workspace'e chatbot taşı
5. Farklı kullanıcı ile giriş yap
6. Organizasyon chatbot'larını gör

---

## Tahmini Süre

| Görev | Süre |
|-------|------|
| Migration | 2-3 saat |
| Models + Service | 4-6 saat |
| Middleware | 2-3 saat |
| API endpoints | 6-8 saat |
| Frontend | 8-12 saat |
| Billing integration | 4-6 saat |
| Testler | 4-6 saat |
| **TOPLAM** | **~2-3 hafta** |

---

## Bağımlılıklar

**Önceki:** Plan 1.7 (White-Label) - branding altyapısı

**Sonraki:** 
- Plan 3.2 (Custom Domain)
- Plan 3.3 (Advanced Analytics)
