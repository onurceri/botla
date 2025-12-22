import { useEffect, useRef, useState } from 'react'
import { ChatBubble } from './components/ChatBubble'
import { ChatDrawer } from './components/ChatDrawer'

type Message = { id?: string; role: 'user' | 'assistant'; content: string; ts?: number; feedback?: boolean; type?: 'welcome' | 'handoff' | 'normal'; handoffRequestId?: string; emailSubmitted?: boolean }

export function WidgetApp({ chatbotId, apiBase, themeColor, headerColor, headerTextColor, botMessageColor, botMessageTextColor, userMessageColor, userMessageTextColor, fontFamily, position, botNameOverride, botIconOverride, panelHeight, panelWidth, panelBg, inputBg, inputText, chatBg, bubbleRadius, sendButtonColor, welcome, embedTokenUrl, captchaSiteKey, autoOpen, useOverrides, resetSession, sessionIdOverride, suggestions: suggestionsOverride, hideBrandingOverride, customBrandingOverride, positionStrategy = 'fixed', previewMode = false }: { chatbotId: string; apiBase?: string; themeColor?: string; headerColor?: string; headerTextColor?: string; botMessageColor?: string; botMessageTextColor?: string; userMessageColor?: string; userMessageTextColor?: string; fontFamily?: string; position?: 'bottom-right' | 'bottom-left'; botNameOverride?: string; botIconOverride?: string; panelHeight?: string; panelWidth?: string; panelBg?: string; inputBg?: string; inputText?: string; chatBg?: string; bubbleRadius?: string; sendButtonColor?: string; welcome?: string; embedTokenUrl?: string; captchaSiteKey?: string; autoOpen?: boolean; useOverrides?: boolean; resetSession?: boolean; sessionIdOverride?: string; suggestions?: string[]; hideBrandingOverride?: boolean; customBrandingOverride?: { logo_url?: string; text?: string; link?: string }; positionStrategy?: 'fixed' | 'absolute'; previewMode?: boolean }) {
  const [open, setOpen] = useState(!!autoOpen)
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const panelRef = useRef<HTMLDivElement | null>(null)
  const [config, setConfig] = useState<any>(null)
  const [sid, setSid] = useState<string>('')
  const [embedToken, setEmbedToken] = useState<string>('')
  const [unread, setUnread] = useState(0)
  const [suggestions, setSuggestions] = useState<string[]>([])

  const emitEvent = (type: string, payload?: any) => {
    if (window.parent !== window) {
      window.parent.postMessage({ type: `WIDGET_EVENT_${type.toUpperCase()}`, payload }, '*')
    }
  }

  useEffect(() => {
    const base = apiBase || ''
    const url = `${base}/api/v1/public/chatbots/${encodeURIComponent(chatbotId)}`
    fetch(url).then(r => r.json()).then(data => {
      setConfig(data)
      emitEvent('CONFIG_LOADED', data)
      if (Array.isArray(data.suggested_questions)) {
        if (!useOverrides) setSuggestions(data.suggested_questions as string[])
      }
      if (data.welcome_message && !welcome) {
        // If welcome message is not overridden by prop, use from config
        // Logic to handle welcome message if not already present
      }
    }).catch(() => {})
  }, [chatbotId, apiBase, useOverrides])

  useEffect(() => {
    if (useOverrides && suggestionsOverride) {
      setSuggestions(suggestionsOverride)
    }
  }, [useOverrides, suggestionsOverride])

  const color = (useOverrides && themeColor) || config?.theme_color || '#3b82f6'
  const pos = (useOverrides && position) || config?.position || 'bottom-right'
  function sanitizeUrl(u?: string) {
    if (!u) return undefined
    return u.replace(/[`'\"]/g, '').trim()
  }
  const botName = (useOverrides && typeof botNameOverride !== 'undefined') ? botNameOverride : config?.bot_display_name
  const botIcon = sanitizeUrl((useOverrides && typeof botIconOverride !== 'undefined') ? botIconOverride : config?.bot_icon)
  const hideBrand = (useOverrides && typeof hideBrandingOverride !== 'undefined') ? hideBrandingOverride : config?.hide_branding
  const customBrand = (useOverrides && customBrandingOverride) ? { ...customBrandingOverride, link: sanitizeUrl(customBrandingOverride.link) } : (config?.custom_branding ? { ...config?.custom_branding, link: sanitizeUrl(config?.custom_branding?.link) } : undefined)
  
  useEffect(() => {
    if (panelRef.current) {
      panelRef.current.style.setProperty('--cbw-color', color)
      panelRef.current.style.setProperty('--cbw-bot-msg-color', botMessageColor || config?.bot_message_color || color)
      panelRef.current.style.setProperty('--cbw-user-msg-color', userMessageColor || config?.user_message_color || color)
      panelRef.current.style.setProperty('--cbw-bot-msg-text-color', botMessageTextColor || config?.bot_message_text_color || '#ffffff')
      panelRef.current.style.setProperty('--cbw-user-msg-text-color', userMessageTextColor || config?.user_message_text_color || '#ffffff')
      panelRef.current.style.setProperty('--cbw-header-color', headerColor || config?.chat_header_color || color)
      panelRef.current.style.setProperty('--cbw-header-text-color', headerTextColor || config?.chat_header_text_color || '#ffffff')
      panelRef.current.style.setProperty('--cbw-font-family', fontFamily || config?.chat_font_family || 'inherit')
      if (panelBg || (config as any)?.chat_panel_bg_color) panelRef.current.style.setProperty('--cbw-panel-bg', panelBg || (config as any)?.chat_panel_bg_color)
      if (chatBg || (config as any)?.chat_background_color) panelRef.current.style.setProperty('--cbw-chat-bg', chatBg || (config as any)?.chat_background_color)
      if (inputBg || (config as any)?.input_background_color) panelRef.current.style.setProperty('--cbw-input-bg', inputBg || (config as any)?.input_background_color)
      if (inputText || (config as any)?.input_text_color) panelRef.current.style.setProperty('--cbw-input-text', inputText || (config as any)?.input_text_color)
      if (bubbleRadius || (config as any)?.bubble_radius) panelRef.current.style.setProperty('--cbw-bubble-radius', bubbleRadius || (config as any)?.bubble_radius)
      if (sendButtonColor || (config as any)?.send_button_color) panelRef.current.style.setProperty('--cbw-send-bg', sendButtonColor || (config as any)?.send_button_color)
      if (panelHeight || (config as any)?.chat_panel_height) {
        panelRef.current.style.setProperty('--cbw-panel-height', panelHeight || (config as any)?.chat_panel_height)
      }
      if (panelWidth || (config as any)?.chat_panel_width) {
        panelRef.current.style.setProperty('--cbw-panel-width', panelWidth || (config as any)?.chat_panel_width)
      }
      
      // Position - skip in preview mode for full container fill
      if (!previewMode) {
        panelRef.current.style.bottom = '20px'
        if (pos === 'bottom-left') {
          panelRef.current.style.left = '20px'
          panelRef.current.style.right = 'auto'
        } else {
          panelRef.current.style.right = '20px'
          panelRef.current.style.left = 'auto'
        }
      }
    }
  }, [color, config, headerColor, headerTextColor, botMessageColor, botMessageTextColor, userMessageColor, userMessageTextColor, fontFamily, pos, panelHeight, panelWidth, panelBg, inputBg, inputText, chatBg, bubbleRadius, sendButtonColor, previewMode])

  useEffect(() => {
    if (resetSession) {
      clearSession(chatbotId)
    }
    if (sessionIdOverride && sessionIdOverride.length > 0) {
      setSessionId(chatbotId, sessionIdOverride)
    }
    const s = getSession(chatbotId)
    setSid(s.sessionId)
    if (s.messages && s.messages.length > 0) {
      // In playground/overrides mode, if the only message is the welcome message, update it
      if (useOverrides && s.messages.length === 1 && s.messages[0].type === 'welcome') {
        const msg = welcome || config?.welcome_message
        if (msg && s.messages[0].content !== msg) {
          const wm = { role: 'assistant', content: msg, ts: Date.now(), type: 'welcome' } as Message
          setMessages([wm])
          saveSession(chatbotId, { sessionId: s.sessionId, messages: [wm] })
        } else {
          setMessages(s.messages)
        }
      } else {
        setMessages(s.messages)
      }
    } else if (welcome || config?.welcome_message) {
      const msg = welcome || config?.welcome_message
      const wm = { role: 'assistant', content: msg, ts: Date.now(), type: 'welcome' } as Message
      setMessages([wm])
      saveSession(chatbotId, { sessionId: s.sessionId, messages: [wm] })
    }
  }, [chatbotId, config, resetSession, sessionIdOverride, welcome, useOverrides])

  const send = async () => {
    if (loading) return
    const text = input.trim()
    if (!text) return
    setInput('')
    const um = { role: 'user', content: text, ts: Date.now() } as Message
    emitEvent('MESSAGE_SENT', { content: text })
    setMessages((m) => {
      const nm = [...m, um]
      saveSession(chatbotId, { sessionId: ensureSession(chatbotId, sid, setSid), messages: nm })
      return nm
    })
    setLoading(true)
    try {
      const base = apiBase || ''
      const url = `${base}/api/v1/public/chatbots/${encodeURIComponent(chatbotId)}/chat`
      let token = embedToken
      if (!token && embedTokenUrl) {
        try { const tr = await fetch(embedTokenUrl); if (tr.ok) { const t = await tr.text(); token = t; setEmbedToken(t) } } catch {}
      }
      let captchaToken = ''
      if (captchaSiteKey && (window as any).getCaptchaToken) {
        try { captchaToken = await (window as any).getCaptchaToken(captchaSiteKey) } catch {}
      }
      const headers: Record<string, string> = { 'Content-Type': 'application/json' }
      if (token) {
        headers['X-Embed-Token'] = token
      }
      const res = await fetch(url, {
        method: 'POST',
        headers,
        body: JSON.stringify({ message: text, session_id: ensureSession(chatbotId, sid, setSid), captcha_token: captchaToken }),
      })
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      const data = await res.json()
      const ans: string = data.response || 'Merhaba!'
      
      emitEvent('RESPONSE_RECEIVED', { content: ans, message_id: data.message_id })

      // Check if this is a handoff response
      const isHandoff = !!data.handoff_request_id
      const am: Message = { 
        id: data.message_id, 
        role: 'assistant', 
        content: ans, 
        ts: Date.now(),
        type: isHandoff ? 'handoff' : 'normal',
        handoffRequestId: data.handoff_request_id || undefined
      }
      
      setMessages((m) => {
        const nm = [...m, am]
        saveSession(chatbotId, { sessionId: ensureSession(chatbotId, sid, setSid), messages: nm })
        return nm
      })
      if (!open) setUnread((u) => u + 1)
    } catch (e: any) {
      emitEvent('ERROR', { message: e.message })
      console.error('Chat error:', e)
      await new Promise((r) => setTimeout(r, 300))
      const em = { role: 'assistant', content: 'Şu an bir hata oluştu, lütfen tekrar deneyin.', ts: Date.now() } as Message
      setMessages((m) => {
        const nm = [...m, em]
        saveSession(chatbotId, { sessionId: ensureSession(chatbotId, sid, setSid), messages: nm })
        return nm
      })
      if (!open) setUnread((u) => u + 1)
    } finally {
      setLoading(false)
    }
  }

  const pickSuggestion = (q: string) => {
    if (loading) return
    setInput(q)
    setTimeout(() => {
      if (!loading) send()
    }, 0)
  }

  const handleFeedback = async (id: string, isPositive: boolean) => {
    // Optimistic update
    setMessages((prev) => prev.map(m => m.id === id ? { ...m, feedback: isPositive } : m))
    
    try {
      const base = apiBase || ''
      const url = `${base}/api/v1/public/chatbots/${encodeURIComponent(chatbotId)}/feedback`
      await fetch(url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ message_id: id, thumbs_up: isPositive }),
      })
    } catch {
      // Silent fail for feedback
    }
  }

  const submitHandoffEmail = async (requestId: string, email: string): Promise<void> => {
    const base = apiBase || ''
    const url = `${base}/api/v1/public/chatbots/${encodeURIComponent(chatbotId)}/handoff/${encodeURIComponent(requestId)}/contact`
    const res = await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email }),
    })
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    
    // Mark the message as email submitted in state
    setMessages((prev) => {
      const updated = prev.map(m => m.handoffRequestId === requestId ? { ...m, emailSubmitted: true } : m)
      saveSession(chatbotId, { sessionId: ensureSession(chatbotId, sid, setSid), messages: updated })
      return updated
    })
  }

  const toggle = () => setOpen((v) => {
    const nv = !v
    if (nv) setUnread(0)
    return nv
  })

  // Preview mode container styles
  const previewContainerStyle = previewMode ? {
    position: 'absolute' as const,
    inset: 0,
    width: '100%',
    height: '100%',
    maxWidth: '100%',
    maxHeight: '100%',
  } : {
    position: positionStrategy as 'fixed' | 'absolute',
  }

  // Position class for preview mode
  const posClass = pos === 'bottom-left' ? 'cbw-pos-left' : 'cbw-pos-right'

  return (
    <div 
      className={`cbw-container ${previewMode ? 'cbw-preview-mode' : ''} ${posClass}`} 
      ref={panelRef} 
      style={previewContainerStyle}
    >
      {open ? (
        <ChatDrawer 
          color={color} 
          messages={messages} 
          loading={loading} 
          input={input} 
          setInput={setInput} 
          onSend={send} 
          onClose={toggle}
          botName={botName}
          botIcon={botIcon}
          suggestions={suggestions}
          onPickSuggestion={pickSuggestion}
          maxChars={config?.max_chars}
          hideBranding={hideBrand}
          customBranding={customBrand}
          onFeedback={handleFeedback}
          onSubmitEmail={submitHandoffEmail}
          isPreviewMode={previewMode}
        />
      ) : (
        <ChatBubble color={color} unread={unread} onClick={toggle} icon={botIcon} />
      )}
    </div>
  )
}

type SessionData = { sessionId: string; messages: Message[] }
function storageKey(chatbotId: string) { return `chatbot_session_${chatbotId}` }
function getSession(chatbotId: string): SessionData {
  try {
    const raw = localStorage.getItem(storageKey(chatbotId))
    if (raw) {
      const parsed = JSON.parse(raw) as SessionData
      if (parsed.sessionId && Array.isArray(parsed.messages)) return parsed
    }
  } catch {}
  const sid = crypto.randomUUID()
  const init: SessionData = { sessionId: sid, messages: [] }
  saveSession(chatbotId, init)
  return init
}
function saveSession(chatbotId: string, data: SessionData) {
  try { localStorage.setItem(storageKey(chatbotId), JSON.stringify(data)) } catch {}
}
function clearSession(chatbotId: string) {
  try { localStorage.removeItem(storageKey(chatbotId)) } catch {}
}
function setSessionId(chatbotId: string, sessionId: string) {
  saveSession(chatbotId, { sessionId, messages: [] })
}
function ensureSession(chatbotId: string, sid: string, setSid: (v: string) => void): string {
  if (sid && sid.length > 0) return sid
  const s = getSession(chatbotId)
  setSid(s.sessionId)
  return s.sessionId
}
