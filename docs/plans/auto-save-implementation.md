# Auto-Save Implementation Plan

> **Tarih:** 2024-12-13  
> **Durum:** Onaylandı  
> **Kapsam:** Chatbot ayar tab'larında auto-save implementasyonu

---

## Özet

Mevcut "Değişiklikleri Kaydet" butonunu kaldırıp tüm chatbot ayar tab'larını **auto-save** yapısına geçirmek.

| Karar | Değer |
|-------|-------|
| Debounce süresi | 800ms |
| Retry sayısı | 2 |
| Retry bekleme süresi | 3 saniye |

---

## Değişiklik Listesi

| Dosya | İşlem | Açıklama |
|-------|-------|----------|
| `useAutoSave.ts` | YENİ | Auto-save hook |
| `SaveIndicator.tsx` | YENİ | Kaydetme durumu göstergesi |
| `useChatbotForm.ts` | GÜNCELLE | 7 payload builder ekle |
| `OverviewTab.tsx` | GÜNCELLE | Auto-save entegrasyonu |
| `GuardrailsTab.tsx` | GÜNCELLE | Auto-save entegrasyonu |
| `HandoffTab.tsx` | GÜNCELLE | Auto-save entegrasyonu |
| `SuggestionsTab.tsx` | GÜNCELLE | Auto-save entegrasyonu |
| `PlaygroundTab.tsx` | GÜNCELLE | Auto-save entegrasyonu |
| `ConnectTab.tsx` | GÜNCELLE | Auto-save entegrasyonu |
| `SourcesTab.tsx` | GÜNCELLE | URL ayarları için auto-save |
| `HeaderActions.tsx` | GÜNCELLE | Save butonunu kaldır |
| `ChatbotDetailPage.tsx` | GÜNCELLE | handleSave → handleCreate |
| `save-success.test.tsx` | SİL | Artık geçerli değil |
| `save-error.test.tsx` | GÜNCELLE | Sadece POST testi |
| `save-delete.test.tsx` | GÜNCELLE | Update testini kaldır |
| `HeaderActions.test.tsx` | GÜNCELLE | Save testlerini kaldır |
| `useAutoSave.test.tsx` | YENİ | Hook testleri |
| `auto-save.test.tsx` | YENİ | Entegrasyon testleri |

---

## 1. Yeni Dosya: useAutoSave.ts

**Yol:** `frontend/src/features/chatbot/hooks/useAutoSave.ts`

```typescript
import { useState, useEffect, useRef, useCallback } from 'react'
import { useParams } from 'react-router-dom'
import { api } from '@/api/client'

// === TYPES ===

type AutoSaveState = {
  isSaving: boolean
  lastSavedAt: Date | null
  error: string | null
}

type UseAutoSaveOptions = {
  /** Gönderilecek veri */
  payload: Record<string, unknown>
  /** false ise save tetiklenmez (validation için) */
  enabled?: boolean
  /** Debounce süresi (ms) */
  debounceMs?: number
  /** Başarılı save sonrası callback */
  onSuccess?: () => void
  /** Hata sonrası callback */
  onError?: (error: string) => void
}

// === CONSTANTS ===

const MAX_RETRIES = 2
const RETRY_DELAY_MS = 3000

// === HOOK ===

export function useAutoSave({
  payload,
  enabled = true,
  debounceMs = 800,
  onSuccess,
  onError
}: UseAutoSaveOptions): AutoSaveState {
  const { id } = useParams()
  
  // State
  const [state, setState] = useState<AutoSaveState>({
    isSaving: false,
    lastSavedAt: null,
    error: null
  })
  
  // Refs
  const timeoutRef = useRef<NodeJS.Timeout | null>(null)
  const abortRef = useRef<AbortController | null>(null)
  const prevPayloadRef = useRef<string>('')
  const retryCountRef = useRef<number>(0)
  const retryTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  
  // === EDGE CASE 1: Cleanup on unmount (race condition önleme) ===
  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current)
        timeoutRef.current = null
      }
      if (retryTimeoutRef.current) {
        clearTimeout(retryTimeoutRef.current)
        retryTimeoutRef.current = null
      }
      if (abortRef.current) {
        abortRef.current.abort()
        abortRef.current = null
      }
    }
  }, [])
  
  // === EDGE CASE 6: Unsaved changes warning ===
  useEffect(() => {
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      // Pending debounce timeout veya saving durumu varsa uyar
      if (timeoutRef.current || state.isSaving) {
        e.preventDefault()
        e.returnValue = '' // Chrome için gerekli
      }
    }
    
    window.addEventListener('beforeunload', handleBeforeUnload)
    return () => window.removeEventListener('beforeunload', handleBeforeUnload)
  }, [state.isSaving])
  
  // === SAVE FUNCTION ===
  const save = useCallback(async (isRetry = false) => {
    // Yeni chatbot için save yapma
    if (!id || id === 'new') return
    
    // === EDGE CASE 2: Önceki request'i iptal et ===
    if (abortRef.current) {
      abortRef.current.abort()
    }
    abortRef.current = new AbortController()
    
    setState(s => ({ ...s, isSaving: true, error: null }))
    
    try {
      await api.put(`/api/v1/chatbots/${id}`, payload, {
        signal: abortRef.current.signal
      })
      
      // Başarılı - reset retry count
      retryCountRef.current = 0
      setState(s => ({ ...s, isSaving: false, lastSavedAt: new Date(), error: null }))
      onSuccess?.()
      
    } catch (e: any) {
      // İptal edilen request'leri ignore et
      if (e.name === 'CanceledError' || e.name === 'AbortError') {
        return
      }
      
      const errorMsg = e.response?.data?.error || 'Kaydetme başarısız'
      
      // === EDGE CASE 4: Retry mekanizması ===
      if (!isRetry && retryCountRef.current < MAX_RETRIES) {
        retryCountRef.current++
        setState(s => ({ ...s, isSaving: false, error: `${errorMsg} - Tekrar deneniyor...` }))
        
        retryTimeoutRef.current = setTimeout(() => {
          save(true) // isRetry = true
        }, RETRY_DELAY_MS)
        return
      }
      
      // Max retry'a ulaşıldı veya retry'dan geldik
      retryCountRef.current = 0
      setState(s => ({ ...s, isSaving: false, error: errorMsg }))
      onError?.(errorMsg)
    }
  }, [id, payload, onSuccess, onError])
  
  // === DEBOUNCED SAVE TRIGGER ===
  useEffect(() => {
    // === EDGE CASE 5: Validation kontrolü ===
    if (!enabled) return
    
    // Yeni chatbot için save yapma
    if (!id || id === 'new') return
    
    // === EDGE CASE 7: Payload değişmeden save yapma ===
    const payloadStr = JSON.stringify(payload)
    if (payloadStr === prevPayloadRef.current) return
    prevPayloadRef.current = payloadStr
    
    // Önceki timeout'u temizle (debounce)
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current)
    }
    
    // Yeni timeout başlat
    timeoutRef.current = setTimeout(() => {
      timeoutRef.current = null
      save()
    }, debounceMs)
    
  }, [payload, enabled, debounceMs, save, id])
  
  return state
}
```

---

## 2. Yeni Dosya: SaveIndicator.tsx

**Yol:** `frontend/src/features/chatbot/components/SaveIndicator.tsx`

```tsx
import { Check, Loader2, AlertCircle, RefreshCw } from 'lucide-react'

type Props = {
  isSaving: boolean
  lastSavedAt: Date | null
  error: string | null
}

export function SaveIndicator({ isSaving, lastSavedAt, error }: Props) {
  // Hata durumu
  if (error) {
    const isRetrying = error.includes('Tekrar deneniyor')
    return (
      <span className="flex items-center gap-1.5 text-sm text-destructive animate-in fade-in duration-200">
        {isRetrying ? (
          <RefreshCw className="w-4 h-4 animate-spin" />
        ) : (
          <AlertCircle className="w-4 h-4" />
        )}
        <span className="max-w-[200px] truncate">{error}</span>
      </span>
    )
  }
  
  // Kaydetme durumu
  if (isSaving) {
    return (
      <span className="flex items-center gap-1.5 text-sm text-muted-foreground animate-pulse">
        <Loader2 className="w-4 h-4 animate-spin" />
        Kaydediliyor...
      </span>
    )
  }
  
  // Başarılı kayıt
  if (lastSavedAt) {
    return (
      <span className="flex items-center gap-1.5 text-sm text-green-600 animate-in fade-in duration-200">
        <Check className="w-4 h-4" />
        Kaydedildi
      </span>
    )
  }
  
  // Henüz değişiklik yok
  return null
}
```

---

## 3. Güncelleme: useChatbotForm.ts

**Yol:** `frontend/src/features/chatbot/hooks/useChatbotForm.ts`

**Değişiklik:** Return bloğuna aşağıdaki fonksiyonları ekle:

```typescript
// === MEVCUT return bloğunun ÜSTÜNE bu fonksiyonları ekle ===

function buildOverviewPayload() {
  return {
    name,
    custom_instruction: customInstruction,
    model,
    temperature,
    max_tokens: maxTokens,
  }
}

function buildGuardrailsPayload() {
  return {
    threshold_config: thresholdConfig,
    fallback_messages: fallbackMessages,
    topic_restrictions: topicRestrictions,
  }
}

function buildHandoffPayload() {
  return {
    handoff_enabled: handoffEnabled,
    handoff_type: handoffType,
    handoff_config: handoffEnabled ? handoffConfig : null,
  }
}

function buildSuggestionsPayload() {
  return {
    suggestions_enabled: suggestionsEnabled,
    suggested_questions: suggestedQuestions,
  }
}

function buildAppearancePayload() {
  return {
    bot_display_name: botDisplayName,
    bot_icon: botIcon,
    welcome_message: welcomeMessage,
    position,
    chat_font_family: chatFontFamily,
    theme_color: themeColor,
    chat_background_color: chatBackgroundColor,
    chat_header_color: chatHeaderColor,
    chat_header_text_color: chatHeaderTextColor,
    bot_message_color: botMessageColor,
    bot_message_text_color: botMessageTextColor,
    user_message_color: userMessageColor,
    user_message_text_color: userMessageTextColor,
    hide_branding: hideBranding,
    custom_branding: hideBranding ? customBranding : null,
  }
}

function buildConnectPayload() {
  return {
    secure_embed_enabled: secureEmbedEnabled,
    allowed_domains: secureEmbedEnabled ? allowedDomains : undefined,
    embed_secret: secureEmbedEnabled ? embedSecret : undefined,
  }
}

function buildSourceSettingsPayload() {
  return {
    discovery_mode: discoveryMode,
    refresh_policy: refreshPolicy,
    refresh_frequency: refreshFrequency,
    include_paths: includePaths,
    exclude_paths: excludePaths,
    selector_whitelist: selectorWhitelist,
  }
}

// === MEVCUT return bloğuna bu fonksiyonları ekle ===

return {
  // ... mevcut tüm değerler aynen kalacak ...
  
  // Yeni eklenenler:
  buildOverviewPayload,
  buildGuardrailsPayload,
  buildHandoffPayload,
  buildSuggestionsPayload,
  buildAppearancePayload,
  buildConnectPayload,
  buildSourceSettingsPayload,
}
```

---

## 4. Güncelleme: OverviewTab.tsx

**Yol:** `frontend/src/features/chatbot/pages/tabs/OverviewTab.tsx`

**TAM YENİ DOSYA İÇERİĞİ:**

```tsx
import { Bot, Cpu, Sparkles, Gauge } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { useChatbotContext } from '../../context/ChatbotContext'
import { useAutoSave } from '../../hooks/useAutoSave'
import { SaveIndicator } from '../../components/SaveIndicator'

export default function OverviewTab() {
  const {
    name, setName,
    customInstruction, setCustomInstruction,
    model, setModel,
    temperature, setTemperature,
    maxTokens, setMaxTokens,
    buildOverviewPayload,
  } = useChatbotContext()

  // Auto-save hook - sadece isim doluysa kaydet (validation)
  const { isSaving, lastSavedAt, error } = useAutoSave({
    payload: buildOverviewPayload(),
    enabled: !!name.trim(),
  })

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <div className="flex flex-col gap-2">
        <div className="flex items-center justify-between">
          <h2 className="text-2xl font-bold tracking-tight">Genel Bakış</h2>
          <SaveIndicator isSaving={isSaving} lastSavedAt={lastSavedAt} error={error} />
        </div>
        <p className="text-muted-foreground">
          Chatbotunuzun kimliğini ve temel yapay zeka davranışlarını yapılandırın.
        </p>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <Card className="h-full border-muted-foreground/20 shadow-sm hover:shadow-md transition-shadow">
          <CardHeader>
            <div className="flex items-center gap-2 mb-2">
              <div className="p-2 rounded-lg bg-primary/10 text-primary">
                <Bot className="w-5 h-5" />
              </div>
              <CardTitle>Kimlik</CardTitle>
            </div>
            <CardDescription>
              Botunuzun ismi ve özel talimatları.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="space-y-2">
              <Label htmlFor="name">Bot İsmi</Label>
              <Input
                id="name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Örn: Müşteri Temsilcisi"
                className="bg-background/50"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="customInstruction" className="flex justify-between">
                <span>Özel Talimatlar</span>
                <span className="text-xs text-muted-foreground font-normal">{customInstruction.length} karakter</span>
              </Label>
              <Textarea
                id="customInstruction"
                className="min-h-[300px] resize-none bg-background/50 leading-relaxed font-mono text-sm"
                value={customInstruction}
                onChange={(e) => setCustomInstruction(e.target.value)}
                placeholder="Botunuza özel davranış kuralları ekleyin... Örn: Müşterilere resmi bir dil kullan, fiyat bilgisi verme..."
              />
              <p className="text-xs text-muted-foreground">
                Botunuzun nasıl davranması gerektiğini, tonunu ve özel kurallarını buraya yazın. Dil ve kapsam kuralları otomatik eklenir.
              </p>
            </div>
          </CardContent>
        </Card>

        <Card className="h-full border-muted-foreground/20 shadow-sm hover:shadow-md transition-shadow">
          <CardHeader>
            <div className="flex items-center gap-2 mb-2">
              <div className="p-2 rounded-lg bg-blue-500/10 text-blue-500">
                <Cpu className="w-5 h-5" />
              </div>
              <CardTitle>Model Ayarları</CardTitle>
            </div>
            <CardDescription>
              Yapay zeka modelini ve teknik parametreleri seçin.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-8">
            <div className="space-y-2">
              <Label>Yapay Zeka Modeli</Label>
              <Select value={model} onValueChange={setModel}>
                <SelectTrigger className="bg-background/50 h-11">
                  <SelectValue placeholder="Model seçin" />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    <SelectLabel>OpenAI</SelectLabel>
                    <SelectItem value="openai:gpt-4o">GPT-4o (Önerilen)</SelectItem>
                    <SelectItem value="openai:gpt-4o-mini">GPT-4o Mini (Hızlı)</SelectItem>
                  </SelectGroup>
                  <SelectGroup>
                    <SelectLabel>Anthropic</SelectLabel>
                    <SelectItem value="anthropic:claude-3-5-sonnet-latest">Claude 3.5 Sonnet</SelectItem>
                    <SelectItem value="anthropic:claude-3-5-haiku-latest">Claude 3.5 Haiku</SelectItem>
                  </SelectGroup>
                  <SelectGroup>
                    <SelectLabel>Google</SelectLabel>
                    <SelectItem value="google:gemini-1.5-pro">Gemini 1.5 Pro</SelectItem>
                    <SelectItem value="google:gemini-1.5-flash">Gemini 1.5 Flash</SelectItem>
                  </SelectGroup>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-6 pt-6 border-t border-border">
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <Label className="flex items-center gap-2">
                    <Sparkles className="w-4 h-4 text-amber-500" />
                    Yaratıcılık (Temperature)
                  </Label>
                  <span className="text-sm font-bold text-primary bg-primary/10 px-3 py-1 rounded-full min-w-[3rem] text-center">
                    {temperature}
                  </span>
                </div>
                <input
                  type="range"
                  min="0"
                  max="1"
                  step="0.1"
                  className="w-full h-2 bg-secondary rounded-lg appearance-none cursor-pointer accent-primary hover:accent-primary/80 transition-all"
                  value={temperature}
                  onChange={(e) => setTemperature(parseFloat(e.target.value))}
                />
                <div className="flex justify-between text-xs text-muted-foreground font-medium">
                  <span>Daha Tutarlı (0.0)</span>
                  <span>Daha Yaratıcı (1.0)</span>
                </div>
              </div>

              <div className="space-y-4 pt-2">
                 <div className="flex items-center justify-between">
                  <Label className="flex items-center gap-2">
                    <Gauge className="w-4 h-4 text-green-500" />
                    Maksimum Token
                  </Label>
                </div>
                 <div className="flex items-center gap-4">
                    <Input
                      type="number"
                      min="1"
                      max="8192"
                      value={maxTokens}
                      onChange={(e) => setMaxTokens(parseInt(e.target.value) || 512)}
                      className="bg-background/50 h-11"
                    />
                 </div>
                 <p className="text-xs text-muted-foreground">
                   Her cevap için üretilecek maksimum kelime/token sayısı.
                 </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
```

---

## 5. Güncelleme: GuardrailsTab.tsx

**Yol:** `frontend/src/features/chatbot/pages/tabs/GuardrailsTab.tsx`

**TAM YENİ DOSYA İÇERİĞİ:**

```tsx
import GuardrailsSettings from '../../components/GuardrailsSettings'
import { useChatbotContext } from '../../context/ChatbotContext'
import { useAutoSave } from '../../hooks/useAutoSave'
import { SaveIndicator } from '../../components/SaveIndicator'

export default function GuardrailsTab() {
  const {
    fallbackMessages, setFallbackMessages,
    topicRestrictions, setTopicRestrictions,
    thresholdConfig, setThresholdConfig,
    planConfig,
    buildGuardrailsPayload,
  } = useChatbotContext()

  const { isSaving, lastSavedAt, error } = useAutoSave({
    payload: buildGuardrailsPayload(),
  })

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <div className="flex flex-col gap-2">
        <div className="flex items-center justify-between">
          <h2 className="text-2xl font-bold tracking-tight">Güvenlik ve Sınırlar</h2>
          <SaveIndicator isSaving={isSaving} lastSavedAt={lastSavedAt} error={error} />
        </div>
        <p className="text-muted-foreground">
          Botunuzun hangi konularda cevap vereceğini ve güven eşiklerini yapılandırın.
        </p>
      </div>

      <GuardrailsSettings
        thresholdConfig={thresholdConfig}
        setThresholdConfig={setThresholdConfig}
        fallbackMessages={fallbackMessages}
        setFallbackMessages={setFallbackMessages}
        topicRestrictions={topicRestrictions}
        setTopicRestrictions={setTopicRestrictions}
        
        canCustomizeThresholds={planConfig?.guardrails?.can_customize_thresholds}
        canUseSmartFallback={planConfig?.guardrails?.can_use_smart_fallback}
        canUseEscalateFallback={planConfig?.guardrails?.can_use_escalate_fallback}
        canManageTopics={planConfig?.guardrails?.can_manage_topics}
        canCustomizeMessages={planConfig?.guardrails?.can_customize_messages}
      />
    </div>
  )
}
```

---

## 6. Güncelleme: HandoffTab.tsx

**Yol:** `frontend/src/features/chatbot/pages/tabs/HandoffTab.tsx`

**TAM YENİ DOSYA İÇERİĞİ:**

```tsx
import HandoffSettings from '../../components/HandoffSettings'
import { useChatbotContext } from '../../context/ChatbotContext'
import { useAutoSave } from '../../hooks/useAutoSave'
import { SaveIndicator } from '../../components/SaveIndicator'

export default function HandoffTab() {
  const {
    handoffEnabled, setHandoffEnabled,
    handoffType, setHandoffType,
    handoffConfig, setHandoffConfig,
    planConfig,
    buildHandoffPayload,
  } = useChatbotContext()

  const { isSaving, lastSavedAt, error } = useAutoSave({
    payload: buildHandoffPayload(),
  })

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <div className="flex flex-col gap-2">
        <div className="flex items-center justify-between">
          <h2 className="text-2xl font-bold tracking-tight">İnsan Devri</h2>
          <SaveIndicator isSaving={isSaving} lastSavedAt={lastSavedAt} error={error} />
        </div>
        <p className="text-muted-foreground">
          Botun cevap veremediği veya kullanıcının talep ettiği durumlarda konuşmayı insana yönlendirin.
        </p>
      </div>

      <HandoffSettings
        handoffEnabled={handoffEnabled}
        setHandoffEnabled={setHandoffEnabled}
        handoffType={handoffType}
        setHandoffType={setHandoffType}
        handoffConfig={handoffConfig}
        setHandoffConfig={setHandoffConfig}
        canUseHandoff={planConfig?.guardrails?.can_use_escalate_fallback}
      />
    </div>
  )
}
```

---

## 7. Güncelleme: SuggestionsTab.tsx

**Yol:** `frontend/src/features/chatbot/pages/tabs/SuggestionsTab.tsx`

**TAM YENİ DOSYA İÇERİĞİ:**

```tsx
import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { api } from '@/api/client'
import SuggestionsPanel from '../../components/SuggestionsPanel'
import { useChatbotContext } from '../../context/ChatbotContext'
import { useAutoSave } from '../../hooks/useAutoSave'
import { SaveIndicator } from '../../components/SaveIndicator'

export default function SuggestionsTab() {
  const { id: chatbotId } = useParams()
  const {
    suggestionsEnabled, setSuggestionsEnabled,
    suggestedQuestions, setSuggestedQuestions,
    allSuggestedQuestions, setAllSuggestedQuestions,
    buildSuggestionsPayload,
  } = useChatbotContext()

  const [isRegenerating, setIsRegenerating] = useState(false)

  const { isSaving, lastSavedAt, error } = useAutoSave({
    payload: buildSuggestionsPayload(),
  })

  const handleRegenerate = async () => {
    if (!chatbotId) return
    setIsRegenerating(true)
    try {
      await api.post(`/api/v1/chatbots/${chatbotId}/suggestions/regenerate`)
      // Refetch chatbot to get updated suggestions after a short delay
      setTimeout(async () => {
        const { data } = await api.get(`/api/v1/chatbots/${chatbotId}`)
        if (data.suggested_questions) {
          setSuggestedQuestions(() => data.suggested_questions)
        }
        if (data.all_suggested_questions) {
          setAllSuggestedQuestions(() => data.all_suggested_questions)
        }
        setIsRegenerating(false)
      }, 2000) // Wait 2 seconds for background processing
    } catch (err) {
      console.error('Failed to regenerate suggestions:', err)
      setIsRegenerating(false)
    }
  }

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div className="flex flex-col gap-2">
          <div className="flex items-center gap-4">
            <h2 className="text-2xl font-bold tracking-tight">Konuşma Başlatıcılar</h2>
            <SaveIndicator isSaving={isSaving} lastSavedAt={lastSavedAt} error={error} />
          </div>
          <p className="text-muted-foreground">
            Kullanıcıların sohbeti başlatmasını kolaylaştıracak hazır sorular ekleyin.
          </p>
        </div>
        <Button 
          variant="outline" 
          size="sm"
          onClick={handleRegenerate}
          disabled={isRegenerating}
          className="self-start sm:self-auto"
        >
          <RefreshCw className={`w-4 h-4 mr-2 ${isRegenerating ? 'animate-spin' : ''}`} />
          {isRegenerating ? 'Yeniden Üretiliyor...' : 'Yeniden Üret'}
        </Button>
      </div>

      <SuggestionsPanel 
        suggestionsEnabled={suggestionsEnabled}
        setSuggestionsEnabled={setSuggestionsEnabled}
        suggestedQuestions={suggestedQuestions}
        setSuggestedQuestions={setSuggestedQuestions}
        allSuggestedQuestions={allSuggestedQuestions}
      />
    </div>
  )
}
```

---

## 8. Güncelleme: PlaygroundTab.tsx

**Yol:** `frontend/src/features/chatbot/pages/tabs/PlaygroundTab.tsx`

**DEĞİŞİKLİKLER:** İlk satırlara import ekle, sonra hook çağır ve header'a SaveIndicator ekle.

```tsx
// === EN ÜSTE EKLE ===
import { useAutoSave } from '../../hooks/useAutoSave'
import { SaveIndicator } from '../../components/SaveIndicator'

// === useChatbotContext() çağrısına buildAppearancePayload ekle ===
const {
  // ... mevcut değerler ...
  buildAppearancePayload, // YENİ
} = useChatbotContext()

// === useChatbotContext sonrasına ekle ===
const { isSaving, lastSavedAt, error } = useAutoSave({
  payload: buildAppearancePayload(),
})

// === Header alanını güncelle ===
<div className="flex flex-col gap-2">
  <div className="flex items-center justify-between">
    <h2 className="text-2xl font-bold tracking-tight">Görünüm ve Test</h2>
    <SaveIndicator isSaving={isSaving} lastSavedAt={lastSavedAt} error={error} />
  </div>
  <p className="text-muted-foreground">
    Chatbotunuzun görünümünü özelleştirin ve anlık olarak test edin.
  </p>
</div>
```

---

## 9. Güncelleme: ConnectTab.tsx

**Yol:** `frontend/src/features/chatbot/pages/tabs/ConnectTab.tsx`

**TAM YENİ DOSYA İÇERİĞİ:**

```tsx
import { useParams } from 'react-router-dom'
import EmbeddingCodePanel from '../../components/EmbeddingCodePanel'
import { useChatbotContext } from '../../context/ChatbotContext'
import { useAutoSave } from '../../hooks/useAutoSave'
import { SaveIndicator } from '../../components/SaveIndicator'

export default function ConnectTab() {
  const { id = '' } = useParams()
  const {
    secureEmbedEnabled, setSecureEmbedEnabled,
    allowedDomains, setAllowedDomains,
    embedSecret, setEmbedSecret,
    planConfig,
    buildConnectPayload,
  } = useChatbotContext()

  const { isSaving, lastSavedAt, error } = useAutoSave({
    payload: buildConnectPayload(),
  })

  return (
    <div className="space-y-6 animate-in fade-in duration-500">
      <div className="flex flex-col gap-2">
        <div className="flex items-center justify-between">
          <h2 className="text-2xl font-bold tracking-tight">Bağlantı & Entegrasyon</h2>
          <SaveIndicator isSaving={isSaving} lastSavedAt={lastSavedAt} error={error} />
        </div>
        <p className="text-muted-foreground">
          Botunuzu web sitenize eklemek için gerekli kodları ve güvenlik ayarlarını yönetin.
        </p>
      </div>

      <EmbeddingCodePanel
        id={id}
        secureEmbedPlanEnabled={!!planConfig?.security?.secure_embed_enabled}
        secureEmbedEnabled={secureEmbedEnabled}
        allowedDomains={allowedDomains}
        embedSecret={embedSecret}
        onToggleSecure={setSecureEmbedEnabled}
        onDomainsChange={setAllowedDomains}
        onSecretChange={setEmbedSecret}
        onSecretRefresh={() => setEmbedSecret(Math.random().toString(36).slice(2)+Math.random().toString(36).slice(2))}
      />
    </div>
  )
}
```

---

## 10. Güncelleme: SourcesTab.tsx

**Yol:** `frontend/src/features/chatbot/pages/tabs/SourcesTab.tsx`

**DEĞİŞİKLİKLER:** Import ekle, hook çağır, SaveIndicator'ı URLAdvancedSettings'in yanına ekle.

```tsx
// === EN ÜSTE EKLE ===
import { useAutoSave } from '../../hooks/useAutoSave'
import { SaveIndicator } from '../../components/SaveIndicator'

// === useChatbotContext() çağrısına buildSourceSettingsPayload ekle ===
const {
  // ... mevcut değerler ...
  buildSourceSettingsPayload, // YENİ
} = useChatbotContext()

// === useChatbotContext sonrasına ekle ===
const { isSaving, lastSavedAt, error } = useAutoSave({
  payload: buildSourceSettingsPayload(),
})

// === Header alanını güncelle ===
<div className="flex flex-col gap-2">
  <div className="flex items-center justify-between">
    <h2 className="text-2xl font-bold tracking-tight">Bilgi Bankası</h2>
    <SaveIndicator isSaving={isSaving} lastSavedAt={lastSavedAt} error={error} />
  </div>
  <p className="text-muted-foreground">
    Botunuzun soruları cevaplarken kullanacağı kaynakları yönetin.
  </p>
</div>
```

---

## 11. Güncelleme: HeaderActions.tsx

**Yol:** `frontend/src/features/chatbot/components/HeaderActions.tsx`

**TAM YENİ DOSYA İÇERİĞİ:**

```tsx
import { Save, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'

type HeaderActionsProps = {
  isNew: boolean
  name: string
  isDeleting: boolean
  isCreating?: boolean     // YENİ: sadece isNew için
  disabled?: boolean
  onDelete: () => void
  onCreate?: () => void    // YENİ: sadece isNew için
}

export default function HeaderActions({
  isNew,
  name,
  isDeleting,
  isCreating = false,
  disabled,
  onDelete,
  onCreate,
}: HeaderActionsProps) {
  return (
    <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-border pb-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight text-foreground">
          {isNew ? 'Yeni Chatbot' : name}
        </h1>
        <p className="text-muted-foreground">
          {isNew ? 'Asistanınızı yapılandırın' : 'Bot ayarlarını ve kaynaklarını yönetin'}
        </p>
      </div>
      <div className="flex items-center gap-2">
        {/* Silme butonu - sadece mevcut chatbot'lar için */}
        {!isNew && (
          <Button
            variant="destructive"
            size="icon"
            className="mr-2"
            onClick={onDelete}
            isLoading={isDeleting}
            aria-label="Sil"
          >
            <Trash2 className="w-4 h-4" />
          </Button>
        )}
        
        {/* Oluştur butonu - sadece yeni chatbot için */}
        {isNew && onCreate && (
          <Button 
            onClick={onCreate} 
            className="gap-2" 
            isLoading={isCreating} 
            disabled={disabled} 
            aria-label="Oluştur"
          >
            <Save className="w-4 h-4" />
            Oluştur
          </Button>
        )}
        
        {/* NOT: Mevcut chatbot'lar için "Değişiklikleri Kaydet" butonu KALDIRILDI */}
        {/* Auto-save her tab'da ayrı ayrı çalışıyor */}
      </div>
    </div>
  )
}
```

---

## 12. Güncelleme: ChatbotDetailPage.tsx

**Yol:** `frontend/src/pages/ChatbotDetailPage.tsx`

**DEĞİŞİKLİKLER:**

```tsx
// === STATE DEĞİŞİKLİĞİ ===
// ESKİ:
const [isSaving, setIsSaving] = useState(false)
// YENİ:
const [isCreating, setIsCreating] = useState(false)

// === FONKSİYON DEĞİŞİKLİĞİ ===
// handleSave fonksiyonunu handleCreate olarak değiştir ve sadece POST kısmını tut:

const handleCreate = async () => {
  if (!validate()) {
    toasts.error('Lütfen bir bot ismi girin.')
    return
  }

  if (!currentWorkspace) {
    toasts.error('Lütfen önce bir çalışma alanı seçin.')
    return
  }

  setIsCreating(true)
  const payload = buildPayload()

  try {
    const { data } = await api.post('/api/v1/chatbots', payload)
    toast('Chatbot başarıyla oluşturuldu.', 'success')
    navigate(`/dashboard/chatbots/${data.id}`)
  } catch (error: any) {
    console.error(error)
    const msg = error.response?.data?.error || 'Bir hata oluştu. Lütfen tekrar deneyin.'
    toasts.error(msg)
  } finally {
    setIsCreating(false)
  }
}

// === HEADER ACTIONS DEĞİŞİKLİĞİ ===
// ESKİ:
<HeaderActions
  isNew={isNew}
  name={name}
  isDeleting={isDeleting}
  isSaving={isSaving}
  disabled={isOrgLoading}
  onDelete={handleDelete}
  onSave={handleSave}
/>

// YENİ:
<HeaderActions
  isNew={isNew}
  name={name}
  isDeleting={isDeleting}
  isCreating={isCreating}
  disabled={isOrgLoading}
  onDelete={handleDelete}
  onCreate={isNew ? handleCreate : undefined}
/>
```

---

## 13. Test Güncellemeleri

### 13.1 Silinecek Dosya

**Yol:** `frontend/src/pages/__tests__/ChatbotDetailPage.save-success.test.tsx`

Bu dosyayı **SİL** - Manuel save butonunu test ediyor, artık geçerli değil.

---

### 13.2 Güncellenecek: save-error.test.tsx

**Yol:** `frontend/src/pages/__tests__/ChatbotDetailPage.save-error.test.tsx`

**"existing chatbot: PUT fails" testini SİL, sadece ilk testi tut:**

```tsx
import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { ToastProvider } from '@/components/ui/toast'
import ChatbotDetailPage from '../ChatbotDetailPage'
import { api } from '@/api/client'

describe('ChatbotDetailPage save error branches', () => {
  it('new chatbot: validate ok but POST fails shows error toast', async () => {
    const user = userEvent.setup()
    vi.spyOn(api, 'post').mockRejectedValueOnce(new Error('fail'))
    render(
      <ToastProvider>
        <MemoryRouter initialEntries={["/chatbots/new"]}>
          <Routes>
            <Route path="/chatbots/:id" element={<ChatbotDetailPage />} />
          </Routes>
        </MemoryRouter>
      </ToastProvider>
    )
    const nameInput = await screen.findByPlaceholderText('Örn: Müşteri Temsilcisi')
    await user.type(nameInput, 'Yeni Bot')
    const createBtn = await screen.findByRole('button', { name: 'Oluştur' })
    await user.click(createBtn)
    const errs1 = await screen.findAllByText('Bir hata oluştu. Lütfen tekrar deneyin.')
    expect(errs1.length).toBeGreaterThan(0)
  })
  
  // "existing chatbot: PUT fails" testi SİLİNDİ - auto-save kullanıyor artık
})
```

---

### 13.3 Güncellenecek: save-delete.test.tsx

**Yol:** `frontend/src/pages/__tests__/ChatbotDetailPage.save-delete.test.tsx`

**"updates existing chatbot" testini SİL, diğerlerini tut:**

```tsx
// "updates existing chatbot and shows success toast" testini SİL
// Sadece şu testler kalsın:
// - "creates new chatbot on valid form and shows success toast"
// - "deletes existing chatbot and shows success toast"
```

---

### 13.4 Güncellenecek: HeaderActions.test.tsx

**Yol:** `frontend/src/features/chatbot/components/__tests__/HeaderActions.test.tsx`

**TAM YENİ DOSYA İÇERİĞİ:**

```tsx
import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import HeaderActions from '../HeaderActions'

describe('HeaderActions', () => {
  it('renders title and create button for new chatbot', () => {
    const onCreate = vi.fn()
    render(
      <HeaderActions
        isNew={true}
        name=""
        isDeleting={false}
        isCreating={false}
        onDelete={() => {}}
        onCreate={onCreate}
      />
    )
    expect(screen.getByText('Yeni Chatbot')).toBeInTheDocument()
    const createButton = screen.getByRole('button', { name: /Oluştur/i })
    fireEvent.click(createButton)
    expect(onCreate).toHaveBeenCalledTimes(1)
  })

  it('renders name and delete button (no save button) for existing chatbot', () => {
    const onDelete = vi.fn()
    render(
      <HeaderActions
        isNew={false}
        name="Destek Botu"
        isDeleting={false}
        onDelete={onDelete}
      />
    )
    expect(screen.getByText('Destek Botu')).toBeInTheDocument()
    
    // Silme butonu olmalı
    const deleteButton = screen.getByLabelText('Sil')
    fireEvent.click(deleteButton)
    expect(onDelete).toHaveBeenCalledTimes(1)
    
    // Kaydet butonu OLMAMALI
    expect(screen.queryByRole('button', { name: /Değişiklikleri Kaydet/i })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /Oluştur/i })).not.toBeInTheDocument()
  })
})
```

---

### 13.5 Yeni Test: useAutoSave.test.tsx

**Yol:** `frontend/src/features/chatbot/hooks/__tests__/useAutoSave.test.tsx`

```tsx
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { renderHook, act, waitFor } from '@testing-library/react'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { useAutoSave } from '../useAutoSave'
import { api } from '@/api/client'

// Wrapper for hooks that need router
const createWrapper = (id: string) => {
  return ({ children }: { children: React.ReactNode }) => (
    <MemoryRouter initialEntries={[`/chatbots/${id}`]}>
      <Routes>
        <Route path="/chatbots/:id" element={children} />
      </Routes>
    </MemoryRouter>
  )
}

describe('useAutoSave', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
    vi.restoreAllMocks()
  })

  it('should not save when id is "new"', async () => {
    const putSpy = vi.spyOn(api, 'put')
    
    renderHook(
      () => useAutoSave({ payload: { name: 'Test' } }),
      { wrapper: createWrapper('new') }
    )

    await act(async () => {
      vi.advanceTimersByTime(1000)
    })

    expect(putSpy).not.toHaveBeenCalled()
  })

  it('should debounce multiple rapid changes', async () => {
    const putSpy = vi.spyOn(api, 'put').mockResolvedValue({ data: {} })
    
    const { rerender } = renderHook(
      ({ payload }) => useAutoSave({ payload }),
      { 
        wrapper: createWrapper('123'),
        initialProps: { payload: { name: 'Test1' } }
      }
    )

    // Birden fazla hızlı değişiklik
    rerender({ payload: { name: 'Test2' } })
    rerender({ payload: { name: 'Test3' } })
    rerender({ payload: { name: 'Test4' } })

    await act(async () => {
      vi.advanceTimersByTime(800)
    })

    // Sadece 1 kez çağrılmalı (son değer ile)
    expect(putSpy).toHaveBeenCalledTimes(1)
    expect(putSpy).toHaveBeenCalledWith(
      '/api/v1/chatbots/123',
      { name: 'Test4' },
      expect.any(Object)
    )
  })

  it('should set isSaving to true during save', async () => {
    vi.spyOn(api, 'put').mockImplementation(() => new Promise(resolve => setTimeout(() => resolve({ data: {} }), 100)))
    
    const { result } = renderHook(
      () => useAutoSave({ payload: { name: 'Test' } }),
      { wrapper: createWrapper('123') }
    )

    await act(async () => {
      vi.advanceTimersByTime(800) // Debounce
    })

    expect(result.current.isSaving).toBe(true)

    await act(async () => {
      vi.advanceTimersByTime(100) // API response
    })

    expect(result.current.isSaving).toBe(false)
  })

  it('should set lastSavedAt on success', async () => {
    vi.spyOn(api, 'put').mockResolvedValue({ data: {} })
    
    const { result } = renderHook(
      () => useAutoSave({ payload: { name: 'Test' } }),
      { wrapper: createWrapper('123') }
    )

    expect(result.current.lastSavedAt).toBeNull()

    await act(async () => {
      vi.advanceTimersByTime(800)
    })

    await waitFor(() => {
      expect(result.current.lastSavedAt).toBeInstanceOf(Date)
    })
  })

  it('should set error on failure', async () => {
    vi.spyOn(api, 'put').mockRejectedValue({ 
      response: { data: { error: 'Test error' } } 
    })
    
    const { result } = renderHook(
      () => useAutoSave({ payload: { name: 'Test' } }),
      { wrapper: createWrapper('123') }
    )

    await act(async () => {
      vi.advanceTimersByTime(800) // Debounce
    })

    // İlk deneme + 2 retry için bekle
    await act(async () => {
      vi.advanceTimersByTime(3000) // 1st retry
    })
    await act(async () => {
      vi.advanceTimersByTime(3000) // 2nd retry
    })

    await waitFor(() => {
      expect(result.current.error).toBe('Test error')
    })
  })

  it('should not save when payload is unchanged', async () => {
    const putSpy = vi.spyOn(api, 'put').mockResolvedValue({ data: {} })
    const payload = { name: 'Same' }
    
    const { rerender } = renderHook(
      () => useAutoSave({ payload }),
      { wrapper: createWrapper('123') }
    )

    await act(async () => {
      vi.advanceTimersByTime(800)
    })

    expect(putSpy).toHaveBeenCalledTimes(1)

    // Aynı payload ile rerender
    rerender()

    await act(async () => {
      vi.advanceTimersByTime(800)
    })

    // Hâlâ 1 kez çağrılmış olmalı
    expect(putSpy).toHaveBeenCalledTimes(1)
  })

  it('should not save when enabled is false', async () => {
    const putSpy = vi.spyOn(api, 'put')
    
    renderHook(
      () => useAutoSave({ payload: { name: 'Test' }, enabled: false }),
      { wrapper: createWrapper('123') }
    )

    await act(async () => {
      vi.advanceTimersByTime(1000)
    })

    expect(putSpy).not.toHaveBeenCalled()
  })
})
```

---

## 14. Verification Checklist

### Otomatik Testler

```bash
cd frontend && npm run test
```

### Manuel Testler

| Test | Adımlar | Beklenen Sonuç |
|------|---------|----------------|
| Auto-save çalışıyor | Overview'da ismi değiştir, 1sn bekle | "Kaydedildi" görünür |
| Veri kalıcı | Kayıt sonrası sayfayı yenile | Değişiklik duruyor |
| Error gösterimi | Backend'i durdur, değişiklik yap | Hata mesajı görünür |
| Retry çalışıyor | Backend'i durdur, değişiklik yap, backend'i başlat | Auto retry başarılı |
| Yeni chatbot | /chatbots/new'da isim gir, Oluştur'a tıkla | POST çalışır, yönlendirir |
| Tab değişimi | Tab A'da değiştir, hemen Tab B'ye geç | Race condition yok |
| Sayfa kapatma | Değişiklik yap, hemen tab'ı kapatmayı dene | Uyarı çıkar |

---

## Summary

| Kategori | Dosya Sayısı |
|----------|--------------|
| Backend | 0 |
| Yeni Frontend Dosyaları | 2 |
| Güncellenecek Tab Dosyaları | 7 |
| Güncellenecek Core Dosyalar | 3 |
| Silinecek Test Dosyaları | 1 |
| Güncellenecek Test Dosyaları | 3 |
| Yeni Test Dosyaları | 1 |
| **Toplam** | 17 dosya |
