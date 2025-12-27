# Widget Düşük Öncelikli İyileştirmeler - Performance & Testing

> **Öncelik:** 🟡 Düşük  
> **Tahmini Süre:** 6-10 saat  
> **Etki:** Bundle size, performance, reliability

---

## 1. Bundle Size Optimizasyonu

### Problem

`react` ve `react-dom` bağımlılıkları gereksiz yere package.json'da var.

**Dosya:** [package.json](file:///Users/onur/Documents/workspace/botla-co/widget/package.json#L17-18)

```json
"dependencies": {
  "react": "^19.2.0",        // Gereksiz
  "react-dom": "^19.2.0"     // Gereksiz
}
```

Widget zaten `preact/compat` alias kullanıyor.

### Çözüm

```json
{
  "dependencies": {
    "@botla/ui-shared": "file:../packages/ui-shared",
    "markdown-to-jsx": "^9.3.5",
    "preact": "^10.27.2"
  },
  "devDependencies": {
    "@types/react": "^19.2.5",
    "@types/react-dom": "^19.2.3"
  }
}
```

### Beklenen Kazanım

| Metric | Önce | Sonra |
|--------|------|-------|
| Bundle Size | ~85KB | ~55KB |
| Gzip Size | ~28KB | ~18KB |

---

## 2. CSS Minification

### Problem

`styles.css` - 891 satır, ~19KB raw

### Çözüm

#### postcss.config.js Güncelleme

```javascript
import tailwindcss from 'tailwindcss'
import autoprefixer from 'autoprefixer'
import cssnano from 'cssnano'

export default {
  plugins: [
    tailwindcss(),
    autoprefixer(),
    ...(process.env.NODE_ENV === 'production' 
      ? [cssnano({ preset: 'default' })] 
      : []
    )
  ]
}
```

#### Kurulum

```bash
npm install -D cssnano
```

---

## 3. Environment Variable Validation

### Problem

Production build'de env değişkenleri validate edilmiyor.

### Çözüm

#### Yeni Dosya: `src/utils/env.ts`

```typescript
interface WidgetEnv {
  apiBaseUrl?: string
  dashboardUrl?: string
  marketingUrl: string
  isDev: boolean
  isProd: boolean
}

function validateEnv(): WidgetEnv {
  const env: WidgetEnv = {
    apiBaseUrl: import.meta.env.VITE_API_BASE_URL,
    dashboardUrl: import.meta.env.VITE_DASHBOARD_URL,
    marketingUrl: import.meta.env.VITE_MARKETING_URL || 'https://botla.app',
    isDev: import.meta.env.DEV,
    isProd: import.meta.env.PROD,
  }

  if (env.isProd) {
    if (!env.apiBaseUrl) {
      console.warn('[Widget] VITE_API_BASE_URL not set for production')
    }
    if (!env.dashboardUrl) {
      console.warn('[Widget] VITE_DASHBOARD_URL not set for production')
    }
  }

  return env
}

export const env = validateEnv()
```

---

## 4. Test Coverage Genişletme

### Mevcut Durum

```
src/components/__tests__/
├── ChatDrawer.test.tsx ✅
├── Message.test.tsx ✅
└── Suggestions.test.tsx ✅
```

### Eksik Testler

#### 4.1 Widget Entry Point Tests

**Yeni Dosya:** `src/__tests__/widget.test.tsx`

```typescript
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

// Mock DOM
beforeEach(() => {
  document.body.innerHTML = ''
})

afterEach(() => {
  vi.restoreAllMocks()
})

describe('Widget Mount', () => {
  it('should create host element if not exists', async () => {
    const { mount } = await import('../widget')
    
    // Mock script with data-bot
    const script = document.createElement('script')
    script.dataset.bot = 'test-bot-id'
    document.body.appendChild(script)
    
    mount()
    
    const host = document.getElementById('chatbot-widget-host')
    expect(host).toBeDefined()
    expect(host?.shadowRoot).toBeDefined()
  })

  it('should log error when chatbot-id is missing', async () => {
    const consoleSpy = vi.spyOn(console, 'error')
    const { mount } = await import('../widget')
    
    mount()
    
    expect(consoleSpy).toHaveBeenCalledWith(
      expect.stringContaining('chatbot-id is required')
    )
  })

  it('should unmount and clear shadow root', async () => {
    const { mount, unmount } = await import('../widget')
    
    const script = document.createElement('script')
    script.dataset.bot = 'test-bot-id'
    document.body.appendChild(script)
    
    mount()
    unmount()
    
    const host = document.getElementById('chatbot-widget-host')
    expect(host?.shadowRoot?.innerHTML).toBe('')
  })
})

describe('PostMessage Handler', () => {
  it('should update config on WIDGET_CONFIG message', async () => {
    // Test postMessage config updates
  })

  it('should ignore messages from unauthorized origins', async () => {
    // Test origin validation (after implementing)
  })
})
```

#### 4.2 Session Management Tests

**Yeni Dosya:** `src/utils/__tests__/session.test.ts`

```typescript
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { getSession, saveSession, clearSession, ensureSession } from '../session'

describe('Session Management', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('should create new session if none exists', () => {
    const session = getSession('test-bot')
    
    expect(session.sessionId).toBeDefined()
    expect(session.sessionId.length).toBeGreaterThan(0)
    expect(session.messages).toEqual([])
  })

  it('should return existing session', () => {
    const testSession = {
      sessionId: 'existing-id',
      messages: [{ role: 'user', content: 'hello' }]
    }
    localStorage.setItem('chatbot_session_test-bot', JSON.stringify(testSession))
    
    const session = getSession('test-bot')
    
    expect(session.sessionId).toBe('existing-id')
    expect(session.messages).toHaveLength(1)
  })

  it('should save session to localStorage', () => {
    const session = {
      sessionId: 'save-test',
      messages: [{ role: 'assistant', content: 'hi' }]
    }
    
    saveSession('test-bot', session)
    
    const stored = localStorage.getItem('chatbot_session_test-bot')
    expect(stored).toBeDefined()
    expect(JSON.parse(stored!).sessionId).toBe('save-test')
  })

  it('should clear session', () => {
    saveSession('test-bot', { sessionId: 'clear-test', messages: [] })
    
    clearSession('test-bot')
    
    expect(localStorage.getItem('chatbot_session_test-bot')).toBeNull()
  })

  it('should handle localStorage errors gracefully', () => {
    const mockSetItem = vi.spyOn(Storage.prototype, 'setItem')
    mockSetItem.mockImplementation(() => {
      throw new Error('QuotaExceeded')
    })
    
    // Should not throw
    expect(() => saveSession('test-bot', { sessionId: 'x', messages: [] })).not.toThrow()
    
    mockSetItem.mockRestore()
  })
})
```

#### 4.3 API Integration Tests

**Yeni Dosya:** `src/__tests__/api.test.ts`

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest'

describe('Chat API', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn())
  })

  it('should send message with correct payload', async () => {
    const mockFetch = vi.mocked(fetch)
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ response: 'Hello!', message_id: 'msg-1' })
    } as Response)

    // Test chat send logic
  })

  it('should handle rate limit error', async () => {
    const mockFetch = vi.mocked(fetch)
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 429,
    } as Response)

    // Test rate limit handling
  })

  it('should handle network error', async () => {
    const mockFetch = vi.mocked(fetch)
    mockFetch.mockRejectedValueOnce(new Error('Network error'))

    // Test network error handling
  })
})
```

#### 4.4 Sanitization Tests

**Yeni Dosya:** `src/utils/__tests__/sanitize.test.ts`

```typescript
import { describe, it, expect } from 'vitest'
import { sanitizeUrl } from '../sanitize'

describe('sanitizeUrl', () => {
  it('should allow https URLs', () => {
    expect(sanitizeUrl('https://example.com/image.png')).toBe('https://example.com/image.png')
  })

  it('should allow http URLs', () => {
    expect(sanitizeUrl('http://example.com/image.png')).toBe('http://example.com/image.png')
  })

  it('should block javascript: protocol', () => {
    expect(sanitizeUrl('javascript:alert(1)')).toBeUndefined()
  })

  it('should block data: with non-image MIME', () => {
    expect(sanitizeUrl('data:text/html,<script></script>')).toBeUndefined()
  })

  it('should allow data: with image MIME', () => {
    expect(sanitizeUrl('data:image/png;base64,abc')).toBeDefined()
  })

  it('should handle undefined input', () => {
    expect(sanitizeUrl(undefined)).toBeUndefined()
  })

  it('should handle empty string', () => {
    expect(sanitizeUrl('')).toBeUndefined()
  })

  it('should strip quotes', () => {
    const result = sanitizeUrl('"https://example.com"')
    expect(result).not.toContain('"')
  })

  it('should allow relative URLs', () => {
    expect(sanitizeUrl('/images/bot.png')).toBe('/images/bot.png')
    expect(sanitizeUrl('./images/bot.png')).toBe('./images/bot.png')
  })
})
```

---

## 5. E2E Test Setup

### Playwright Kurulumu

```bash
npm install -D @playwright/test
npx playwright install
```

#### playwright.config.ts

```typescript
import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: './e2e',
  use: {
    baseURL: 'http://localhost:5173',
  },
  webServer: {
    command: 'npm run dev',
    port: 5173,
    reuseExistingServer: !process.env.CI,
  },
})
```

#### e2e/widget.spec.ts

```typescript
import { test, expect } from '@playwright/test'

test.describe('Widget E2E', () => {
  test('should open and close chat panel', async ({ page }) => {
    await page.goto('/preview.html?chatbot-id=test')
    
    // Click bubble to open
    await page.click('.cbw-bubble')
    await expect(page.locator('.cbw-panel')).toBeVisible()
    
    // Click close to minimize
    await page.click('.cbw-close-btn')
    await expect(page.locator('.cbw-panel')).not.toBeVisible()
    await expect(page.locator('.cbw-bubble')).toBeVisible()
  })

  test('should send message and receive response', async ({ page }) => {
    await page.goto('/preview.html?chatbot-id=test')
    await page.click('.cbw-bubble')
    
    await page.fill('.cbw-input-field', 'Hello')
    await page.click('.cbw-send-btn')
    
    // Wait for response
    await expect(page.locator('.cbw-msg.assistant')).toBeVisible({ timeout: 10000 })
  })
})
```

---

## 6. Build Performance

### Vite Analyze Plugin

```bash
npm install -D rollup-plugin-visualizer
```

#### vite.config.js Güncelleme

```javascript
import { visualizer } from 'rollup-plugin-visualizer'

export default defineConfig({
  plugins: [
    preact({ jsxImportSource: 'preact' }),
    previewHtmlPlugin(),
    visualizer({
      filename: 'dist/stats.html',
      open: false,
      gzipSize: true,
    }),
  ],
  // ...
})
```

---

## 7. npm Scripts Güncelleme

```json
{
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "build:analyze": "vite build && open dist/stats.html",
    "lint": "eslint .",
    "typecheck": "tsc --noEmit",
    "preview": "vite preview",
    "test": "vitest run",
    "test:watch": "vitest",
    "test:coverage": "vitest run --coverage",
    "test:e2e": "playwright test",
    "ci": "npm run typecheck && npm run lint && npm run test:coverage"
  }
}
```

---

## Coverage Hedefleri

| Metric | Mevcut | Hedef |
|--------|--------|-------|
| Statements | ~40% | 80% |
| Branches | ~35% | 75% |
| Functions | ~50% | 80% |
| Lines | ~40% | 80% |

---

## Checklist

- [ ] React dependencies kaldırıldı
- [ ] CSS minification eklendi
- [ ] Environment validation eklendi
- [ ] Widget entry point testleri yazıldı
- [ ] Session management testleri yazıldı
- [ ] Sanitization testleri yazıldı
- [ ] E2E test setup yapıldı
- [ ] Build analyze eklendi
- [ ] CI scripts güncellendi
- [ ] Coverage hedefleri karşılandı
