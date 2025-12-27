# Widget Orta Öncelikli İyileştirmeler - Code Quality

> **Öncelik:** 🟠 Orta  
> **Tahmini Süre:** 6-8 saat  
> **Etki:** Maintainability ve developer experience

---

## 1. Type Definitions Merkezileştirme

### Problem

Aynı `Message` type tanımı 3 farklı dosyada tekrarlanıyor.

**Lokasyonlar:**
- `widgetApp.tsx:6-15`
- `ChatDrawer.tsx:5`
- `Message.tsx:4-13`

### Çözüm

#### Yeni Dosya: `src/types/index.ts`

```typescript
/**
 * Widget type definitions
 */

export interface ChatMessage {
  id?: string
  role: 'user' | 'assistant'
  content: string
  ts?: number
  feedback?: boolean
  type?: 'welcome' | 'handoff' | 'normal'
  handoffRequestId?: string
  emailSubmitted?: boolean
}

export interface ChatbotConfig {
  theme_color?: string
  position?: 'bottom-right' | 'bottom-left'
  welcome_message?: string
  suggested_questions?: string[]
  bot_display_name?: string
  bot_icon?: string
  hide_branding?: boolean
  custom_branding?: CustomBranding
  max_chars?: number
  // Styling
  bot_message_color?: string
  bot_message_text_color?: string
  user_message_color?: string
  user_message_text_color?: string
  chat_header_color?: string
  chat_header_text_color?: string
  chat_font_family?: string
  chat_panel_bg_color?: string
  chat_background_color?: string
  input_background_color?: string
  input_text_color?: string
  bubble_radius?: string
  send_button_color?: string
  chat_panel_height?: string
  chat_panel_width?: string
}

export interface CustomBranding {
  logo_url?: string
  text?: string
  link?: string
}

export interface SessionData {
  sessionId: string
  messages: ChatMessage[]
}

export type WidgetPosition = 'bottom-right' | 'bottom-left'
export type PositionStrategy = 'fixed' | 'absolute'
```

#### Kullanım

```typescript
// widgetApp.tsx
import type { ChatMessage, ChatbotConfig, SessionData } from './types'

const [messages, setMessages] = useState<ChatMessage[]>([])
const [config, setConfig] = useState<ChatbotConfig | null>(null)
```

---

## 2. Props Interface Refactoring

### Problem

`WidgetApp` bileşeninin props tanımı tek satırda 30+ parametre içeriyor.

**Dosya:** [widgetApp.tsx](file:///Users/onur/Documents/workspace/botla-co/widget/src/widgetApp.tsx#L17)

### Çözüm

#### Yeni Dosya: `src/types/props.ts`

```typescript
import type { CustomBranding, WidgetPosition, PositionStrategy } from './index'

export interface WidgetThemeProps {
  themeColor?: string
  headerColor?: string
  headerTextColor?: string
  botMessageColor?: string
  botMessageTextColor?: string
  userMessageColor?: string
  userMessageTextColor?: string
  fontFamily?: string
  panelBg?: string
  chatBg?: string
  inputBg?: string
  inputText?: string
  bubbleRadius?: string
  sendButtonColor?: string
}

export interface WidgetLayoutProps {
  position?: WidgetPosition
  positionStrategy?: PositionStrategy
  panelHeight?: string
  panelWidth?: string
  previewMode?: boolean
}

export interface WidgetBrandingProps {
  hideBrandingOverride?: boolean
  customBrandingOverride?: CustomBranding
}

export interface WidgetAppProps extends 
  WidgetThemeProps, 
  WidgetLayoutProps, 
  WidgetBrandingProps {
  // Required
  chatbotId: string
  
  // API
  apiBase?: string
  embedTokenUrl?: string
  captchaSiteKey?: string
  
  // Bot customization
  botNameOverride?: string
  botIconOverride?: string
  welcome?: string
  suggestions?: string[]
  
  // Session
  resetSession?: boolean
  sessionIdOverride?: string
  
  // Behavior
  autoOpen?: boolean
  useOverrides?: boolean
  
  // Callbacks
  onOpenChange?: (isOpen: boolean) => void
}
```

#### Güncellenmiş Component

```typescript
// widgetApp.tsx
import type { WidgetAppProps } from './types/props'

export function WidgetApp(props: WidgetAppProps) {
  const {
    chatbotId,
    apiBase,
    themeColor,
    headerColor,
    // ... destructure needed props
  } = props
  
  // ...
}
```

---

## 3. Session Logic Extraction

### Problem

Session yönetim fonksiyonları component dosyası içinde tanımlı.

**Dosya:** [widgetApp.tsx](file:///Users/onur/Documents/workspace/botla-co/widget/src/widgetApp.tsx#L311-L340)

### Çözüm

#### Yeni Dosya: `src/utils/session.ts`

```typescript
import type { ChatMessage, SessionData } from '../types'

const STORAGE_PREFIX = 'chatbot_session_'

function storageKey(chatbotId: string): string {
  return `${STORAGE_PREFIX}${chatbotId}`
}

export function getSession(chatbotId: string): SessionData {
  try {
    const raw = localStorage.getItem(storageKey(chatbotId))
    if (raw) {
      const parsed = JSON.parse(raw) as SessionData
      if (parsed.sessionId && Array.isArray(parsed.messages)) {
        return parsed
      }
    }
  } catch (error) {
    console.warn('[Widget] Failed to parse session:', error)
  }
  
  const newSession: SessionData = {
    sessionId: crypto.randomUUID(),
    messages: []
  }
  saveSession(chatbotId, newSession)
  return newSession
}

export function saveSession(chatbotId: string, data: SessionData): void {
  try {
    localStorage.setItem(storageKey(chatbotId), JSON.stringify(data))
  } catch (error) {
    console.warn('[Widget] Failed to save session:', error)
  }
}

export function clearSession(chatbotId: string): void {
  try {
    localStorage.removeItem(storageKey(chatbotId))
  } catch (error) {
    console.warn('[Widget] Failed to clear session:', error)
  }
}

export function updateSessionMessages(
  chatbotId: string, 
  sessionId: string, 
  messages: ChatMessage[]
): void {
  saveSession(chatbotId, { sessionId, messages })
}

export function ensureSession(
  chatbotId: string, 
  currentSid: string, 
  setSid: (v: string) => void
): string {
  if (currentSid && currentSid.length > 0) return currentSid
  const session = getSession(chatbotId)
  setSid(session.sessionId)
  return session.sessionId
}
```

---

## 4. Error Handling İyileştirmesi

### Problem

Hatalar sessizce yutulur, debugging zorlaşır.

**Örnekler:**
```typescript
// widgetApp.tsx:43
.catch(() => {})  // Silent fail

// widgetApp.tsx:159
try { ... } catch {}  // Silent fail
```

### Çözüm

#### Yeni Dosya: `src/utils/logger.ts`

```typescript
type LogLevel = 'debug' | 'info' | 'warn' | 'error'

const LOG_PREFIX = '[Botla Widget]'

function shouldLog(): boolean {
  return import.meta.env.DEV || 
         localStorage.getItem('botla_debug') === '1'
}

export const logger = {
  debug: (message: string, ...args: unknown[]) => {
    if (shouldLog()) {
      console.debug(`${LOG_PREFIX} ${message}`, ...args)
    }
  },
  
  info: (message: string, ...args: unknown[]) => {
    if (shouldLog()) {
      console.info(`${LOG_PREFIX} ${message}`, ...args)
    }
  },
  
  warn: (message: string, ...args: unknown[]) => {
    console.warn(`${LOG_PREFIX} ${message}`, ...args)
  },
  
  error: (message: string, error?: unknown, ...args: unknown[]) => {
    console.error(`${LOG_PREFIX} ${message}`, error, ...args)
  }
}
```

#### Güncellenmiş Error Handling

```typescript
// widgetApp.tsx
import { logger } from './utils/logger'

// Config fetch
fetch(url)
  .then(r => {
    if (!r.ok) throw new Error(`HTTP ${r.status}`)
    return r.json()
  })
  .then(data => {
    setConfig(data)
    logger.debug('Config loaded', data)
  })
  .catch((error) => {
    logger.error('Failed to load config', error)
    emitEvent('ERROR', { type: 'config_load_error', message: error.message })
  })

// Chat send
} catch (e: unknown) {
  const error = e instanceof Error ? e : new Error(String(e))
  logger.error('Chat request failed', error)
  emitEvent('ERROR', { type: 'chat_error', message: error.message })
  // Show user-friendly error
  const em: ChatMessage = { 
    role: 'assistant', 
    content: getWidgetErrorMessage('INTERNAL_ERROR'), 
    ts: Date.now() 
  }
  // ...
}
```

---

## 5. Magic Values Elimination

### Problem

Hardcoded değerler kod boyunca dağınık.

### Çözüm

#### Yeni Dosya: `src/constants.ts`

```typescript
// Z-index for maximum layer priority
export const WIDGET_Z_INDEX = 2147483647

// Default values
export const DEFAULT_MAX_CHARS = 1000
export const DEFAULT_THEME_COLOR = '#3b82f6'
export const DEFAULT_POSITION: 'bottom-right' = 'bottom-right'

// Timeouts
export const SCROLL_DELAY_MS = 10
export const ERROR_DISPLAY_DELAY_MS = 300

// Storage
export const STORAGE_PREFIX = 'chatbot_session_'
export const DEBUG_STORAGE_KEY = 'botla_debug'

// API
export const DEFAULT_API_ENDPOINTS = {
  config: (chatbotId: string) => `/api/v1/public/chatbots/${encodeURIComponent(chatbotId)}`,
  chat: (chatbotId: string) => `/api/v1/public/chatbots/${encodeURIComponent(chatbotId)}/chat`,
  feedback: (chatbotId: string) => `/api/v1/public/chatbots/${encodeURIComponent(chatbotId)}/feedback`,
  handoff: (chatbotId: string, requestId: string) => 
    `/api/v1/public/chatbots/${encodeURIComponent(chatbotId)}/handoff/${encodeURIComponent(requestId)}/contact`,
} as const

// i18n defaults
export const DEFAULT_ERROR_MESSAGE = {
  tr: 'Şu an bir hata oluştu, lütfen tekrar deneyin.',
  en: 'An error occurred, please try again.'
}
```

---

## 6. ESLint Rules Düzeltme

### Problem

ESLint kuralları suppress edilmiş.

```typescript
// widgetApp.tsx:38
}, []) // eslint-disable-line react-hooks/exhaustive-deps
```

### Çözüm

```typescript
// Correct dependency handling
const onOpenChangeRef = useRef(onOpenChange)
onOpenChangeRef.current = onOpenChange

useEffect(() => {
  onOpenChangeRef.current?.(open)
}, [open])
```

---

## Dosya Yapısı (Sonrası)

```
widget/src/
├── components/
│   ├── ChatBubble.tsx
│   ├── ChatDrawer.tsx
│   ├── Message.tsx
│   └── Suggestions.tsx
├── types/
│   ├── index.ts          # Core types
│   └── props.ts          # Component props
├── utils/
│   ├── session.ts        # Session management
│   ├── sanitize.ts       # URL/content sanitization
│   └── logger.ts         # Logging utility
├── i18n/
│   └── errors.ts         # Error translations
├── constants.ts          # Magic values
├── widget.tsx            # Entry point
├── widgetApp.tsx         # Main component
└── styles.css            # Styles
```

---

## Checklist

- [ ] Type definitions merkezileştirildi
- [ ] Props interface refactor edildi
- [ ] Session logic ayrı modüle çıkarıldı
- [ ] Error handling iyileştirildi
- [ ] Magic values constants'a taşındı
- [ ] ESLint uyarıları düzeltildi
- [ ] Import'lar güncellendi
- [ ] Testler geçti
