# Widget Kritik Güvenlik İyileştirmeleri

> **Öncelik:** 🔴 Kritik  
> **Tahmini Süre:** 4-6 saat  
> **Etki:** Production'da güvenlik açıkları

---

## 1. PostMessage Origin Validation

### Problem

Widget, `postMessage` ile gelen konfigürasyon güncellemelerinde origin kontrolü yapmıyor.

**Dosya:** [widget.tsx](file:///Users/onur/Documents/workspace/botla-co/widget/src/widget.tsx#L120-L173)

```typescript
// Mevcut kod - GÜVENSİZ
window.addEventListener('message', (event) => {
  if (event.data?.type === 'WIDGET_CONFIG') {
    // ⚠️ Herhangi bir origin'den mesaj kabul edilir!
    const newConfig = event.data.config
    // ...
  }
})
```

### Risk

Kötü niyetli bir sayfa, widget'ı embed eden siteye iframe ile erişip widget konfigürasyonunu değiştirebilir:
- `api-base` değiştirilerek mesajlar farklı sunucuya yönlendirilebilir
- Kullanıcı verisi çalınabilir

### Çözüm

```typescript
// src/widget.tsx - GÜVENLİ
const ALLOWED_ORIGINS = [
  import.meta.env.VITE_DASHBOARD_URL,
  import.meta.env.VITE_API_BASE_URL,
].filter(Boolean)

window.addEventListener('message', (event) => {
  // Origin kontrolü
  if (!ALLOWED_ORIGINS.some(origin => event.origin.startsWith(origin))) {
    console.warn('[Widget] Unauthorized postMessage origin:', event.origin)
    return
  }
  
  if (event.data?.type === 'WIDGET_CONFIG') {
    const newConfig = event.data.config
    // ...
  }
})
```

### Environment Variables

`.env.production` dosyasına ekle:
```env
VITE_DASHBOARD_URL=https://app.botla.co
```

---

## 2. XSS Prevention - Markdown Rendering

### Problem

`markdown-to-jsx` kütüphanesi varsayılan olarak raw HTML'i render edebilir.

**Dosya:** [Message.tsx](file:///Users/onur/Documents/workspace/botla-co/widget/src/components/Message.tsx#L139)

```typescript
// Mevcut kod - XSS RİSKİ
<Markdown options={{ createElement }}>{m.content}</Markdown>
```

### Risk

Backend'den gelen veya manipüle edilmiş mesaj içeriğinde:
```html
<script>document.cookie</script>
<img src=x onerror="alert('XSS')">
```

### Çözüm

```typescript
// src/components/Message.tsx - GÜVENLİ
<Markdown 
  options={{ 
    createElement,
    disableParsingRawHTML: true,  // HTML parsing'i kapat
    forceBlock: true,
    overrides: {
      // Sadece güvenli elementlere izin ver
      a: {
        component: ({ children, href, ...props }) => (
          <a 
            {...props} 
            href={sanitizeUrl(href)} 
            target="_blank" 
            rel="noopener noreferrer"
          >
            {children}
          </a>
        )
      },
      script: () => null,  // Script taglerini engelle
      iframe: () => null,  // Iframe'leri engelle
    }
  }}
>
  {m.content}
</Markdown>
```

---

## 3. URL Sanitization Güçlendirme

### Problem

Mevcut sanitization sadece quote karakterlerini temizliyor.

**Dosya:** [widgetApp.tsx](file:///Users/onur/Documents/workspace/botla-co/widget/src/widgetApp.tsx#L64-L67)

```typescript
// Mevcut kod - YETERSİZ
function sanitizeUrl(u?: string) {
  if (!u) return undefined
  return u.replace(/[`'\"]/g, '').trim()
}
```

### Risk

```
javascript:alert('XSS')
data:text/html,<script>...</script>
```

### Çözüm

```typescript
// src/utils/sanitize.ts
const ALLOWED_PROTOCOLS = ['http:', 'https:', 'data:'];
const DATA_MIME_WHITELIST = ['image/png', 'image/jpeg', 'image/gif', 'image/svg+xml', 'image/webp'];

export function sanitizeUrl(u?: string): string | undefined {
  if (!u) return undefined;
  
  const trimmed = u.replace(/[`'\"<>]/g, '').trim();
  
  try {
    const url = new URL(trimmed);
    
    // Protocol kontrolü
    if (!ALLOWED_PROTOCOLS.includes(url.protocol)) {
      console.warn('[Widget] Blocked unsafe URL protocol:', url.protocol);
      return undefined;
    }
    
    // Data URL için MIME type kontrolü
    if (url.protocol === 'data:') {
      const mimeMatch = trimmed.match(/^data:([^;,]+)/);
      if (mimeMatch && !DATA_MIME_WHITELIST.includes(mimeMatch[1])) {
        console.warn('[Widget] Blocked unsafe data URL MIME:', mimeMatch[1]);
        return undefined;
      }
    }
    
    return url.toString();
  } catch {
    // Relative URL'ler için
    if (trimmed.startsWith('/') || trimmed.startsWith('./')) {
      return trimmed;
    }
    return undefined;
  }
}
```

### Kullanım

```typescript
// widgetApp.tsx
import { sanitizeUrl } from './utils/sanitize'

const botIcon = sanitizeUrl(
  (useOverrides && typeof botIconOverride !== 'undefined') 
    ? botIconOverride 
    : config?.bot_icon
)
```

---

## 4. TypeScript Strict Mode

### Problem

`strict: false` ile type safety devre dışı.

**Dosya:** [tsconfig.json](file:///Users/onur/Documents/workspace/botla-co/widget/tsconfig.json#L11)

### Çözüm

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "module": "ESNext",
    "moduleResolution": "Bundler",
    "skipLibCheck": true,
    "jsx": "react-jsx",
    "jsxImportSource": "preact",
    "allowJs": true,
    "checkJs": false,
    "noEmit": true,
    "strict": true,
    "noImplicitAny": true,
    "strictNullChecks": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "types": ["vite/client"],
    "lib": ["ES2020", "DOM"]
  },
  "include": ["src/**/*"]
}
```

### Beklenen Type Hataları

Strict mode açıldığında düzeltilmesi gereken alanlar:

| Dosya | Satır | Sorun |
|-------|-------|-------|
| widgetApp.tsx | 23 | `useState<any>` → proper interface |
| ChatDrawer.tsx | 52, 59 | `(e: any)` → `(e: Event)` |
| widget.tsx | 8 | `(window as any)` → global.d.ts |

---

## Doğrulama Adımları

```bash
# 1. Type check
cd widget && npx tsc --noEmit

# 2. Lint
npm run lint

# 3. Build test
npm run build

# 4. Security audit
npm audit
```

---

## Checklist

- [ ] Origin validation eklendi
- [ ] Markdown XSS koruması eklendi  
- [ ] URL sanitization güçlendirildi
- [ ] TypeScript strict mode açıldı
- [ ] Type hatalar düzeltildi
- [ ] Testler geçti
