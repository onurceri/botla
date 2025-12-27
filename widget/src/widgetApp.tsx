import { useEffect, useRef, useState } from 'react'
import { ChatBubble } from './components/ChatBubble'
import { ChatDrawer } from './components/ChatDrawer'
import { sanitizeUrl } from './utils/sanitize'
import { logger } from './utils/logger'
import type { ChatMessage, ChatbotConfig } from './types'
import type { WidgetAppProps } from './types/props'
import { getSession, saveSession, clearSession, ensureSession, setSessionId } from './utils/session'
import {
  DEFAULT_THEME_COLOR,
  DEFAULT_POSITION,
  DEFAULT_MAX_CHARS,
  DEFAULT_ERROR_MESSAGE,
  ERROR_DISPLAY_DELAY_MS
} from './constants'

export function WidgetApp(props: WidgetAppProps) {
  const { 
    chatbotId, apiBase, themeColor, headerColor, headerTextColor, botMessageColor, 
    botMessageTextColor, userMessageColor, userMessageTextColor, fontFamily, 
    position, botNameOverride, botIconOverride, panelHeight, panelWidth, 
    panelBg, inputBg, inputText, chatBg, bubbleRadius, sendButtonColor, 
    welcome, embedTokenUrl, captchaSiteKey, autoOpen, useOverrides, 
    resetSession, sessionIdOverride, suggestions: suggestionsOverride, 
    hideBrandingOverride, customBrandingOverride, 
    positionStrategy = 'fixed', previewMode = false, onOpenChange 
  } = props

  const [open, setOpen] = useState(!!autoOpen)
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const panelRef = useRef<HTMLDivElement | null>(null)
  const [config, setConfig] = useState<ChatbotConfig | null>(null)
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const [sid, setSid] = useState<string>('')
  const [embedToken, setEmbedToken] = useState<string>('')
  const [unread, setUnread] = useState(0)
  const [suggestions, setSuggestions] = useState<string[]>([])

  const onOpenChangeRef = useRef(onOpenChange)
  onOpenChangeRef.current = onOpenChange

  const emitEvent = (type: string, payload?: unknown) => {
    if (window.parent !== window) {
      window.parent.postMessage({ type: `WIDGET_EVENT_${type.toUpperCase()}`, payload }, '*')
    }
  }

  // Report open state to parent on change (initial and updates)
  useEffect(() => {
    onOpenChangeRef.current?.(open)
  }, [open])

  useEffect(() => {
    const base = apiBase || ''
    const url = `${base}/api/v1/public/chatbots/${encodeURIComponent(chatbotId)}`
    fetch(url)
      .then(r => {
        if (!r.ok) throw new Error(`HTTP ${r.status}`)
        return r.json()
      })
      .then((data: ChatbotConfig) => {
        setConfig(data)
        logger.debug('Config loaded', data)
        emitEvent('CONFIG_LOADED', data)
        if (Array.isArray(data.suggested_questions)) {
          if (!useOverrides) setSuggestions(data.suggested_questions as string[])
        }
      })
      .catch((error) => {
        logger.error('Failed to load config', error)
        emitEvent('ERROR', { type: 'config_load_error', message: error.message })
      })
  }, [chatbotId, apiBase, useOverrides])

  useEffect(() => {
    if (useOverrides && suggestionsOverride) {
      setSuggestions(suggestionsOverride)
    }
  }, [useOverrides, suggestionsOverride])

  const color = (useOverrides && themeColor) || (config?.theme_color as string | undefined) || DEFAULT_THEME_COLOR
  const pos = (useOverrides && position) || (config?.position as string | undefined) || DEFAULT_POSITION
  const botName = (useOverrides && typeof botNameOverride !== 'undefined') ? botNameOverride : (config?.bot_display_name as string | undefined)
  const botIcon = sanitizeUrl((useOverrides && typeof botIconOverride !== 'undefined') ? botIconOverride : (config?.bot_icon as string | undefined))
  const hideBrand = (useOverrides && typeof hideBrandingOverride !== 'undefined') ? hideBrandingOverride : (config?.hide_branding ?? false)
  const customBrand = (useOverrides && customBrandingOverride) 
    ? { ...customBrandingOverride, link: sanitizeUrl(customBrandingOverride.link) } 
    : config?.custom_branding 
      ? { ...config.custom_branding, link: sanitizeUrl(config.custom_branding.link) } 
      : undefined
  
  // Determine effective values for styling with proper override logic
  const effectiveBotMsgColor = (useOverrides && botMessageColor) ? botMessageColor : (config?.bot_message_color ?? color)
  const effectiveUserMsgColor = (useOverrides && userMessageColor) ? userMessageColor : (config?.user_message_color ?? color)
  const effectiveBotMsgTextColor = (useOverrides && botMessageTextColor) ? botMessageTextColor : (config?.bot_message_text_color ?? '#ffffff')
  const effectiveUserMsgTextColor = (useOverrides && userMessageTextColor) ? userMessageTextColor : (config?.user_message_text_color ?? '#ffffff')
  const effectiveHeaderColor = (useOverrides && headerColor) ? headerColor : (config?.chat_header_color ?? color)
  const effectiveHeaderTextColor = (useOverrides && headerTextColor) ? headerTextColor : (config?.chat_header_text_color ?? '#ffffff')
  const effectiveFontFamily = (useOverrides && fontFamily) ? fontFamily : (config?.chat_font_family ?? 'inherit')
  const effectivePanelBg = (useOverrides && panelBg) ? panelBg : config?.chat_panel_bg_color
  const effectiveChatBg = (useOverrides && chatBg) ? chatBg : config?.chat_background_color
  const effectiveInputBg = (useOverrides && inputBg) ? inputBg : config?.input_background_color
  const effectiveInputText = (useOverrides && inputText) ? inputText : config?.input_text_color
  const effectiveBubbleRadius = (useOverrides && bubbleRadius) ? bubbleRadius : config?.bubble_radius
  const effectiveSendBtnColor = (useOverrides && sendButtonColor) ? sendButtonColor : config?.send_button_color
  const effectivePanelHeight = (useOverrides && panelHeight) ? panelHeight : config?.chat_panel_height
  const effectivePanelWidth = (useOverrides && panelWidth) ? panelWidth : config?.chat_panel_width

  // Load Google Fonts dynamically
  useEffect(() => {
    if (!effectiveFontFamily || effectiveFontFamily === 'inherit') return
    
    // Extract the font family name (before any fallbacks like ", sans-serif")
    const fontName = effectiveFontFamily.split(',')[0].trim().replace(/['"]/g, '')
    if (!fontName) return
    
    // Skip system fonts
    const systemFonts = ['inherit', 'system-ui', '-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'Helvetica', 'Arial', 'sans-serif', 'serif', 'monospace']
    if (systemFonts.some(sf => fontName.toLowerCase() === sf.toLowerCase())) return
    
    // Check if font is already loaded in the document
    const fontId = `cbw-font-${fontName.replace(/\s+/g, '-').toLowerCase()}`
    
    // For shadow DOM, we need to load the font in the parent document
    const targetDocument = document
    if (targetDocument.getElementById(fontId)) return
    
    // Create and inject Google Fonts link
    const link = targetDocument.createElement('link')
    link.id = fontId
    link.rel = 'stylesheet'
    link.href = `https://fonts.googleapis.com/css2?family=${encodeURIComponent(fontName)}:wght@300;400;500;600;700&display=swap`
    targetDocument.head.appendChild(link)
    
    logger.debug('Google Font loaded:', fontName)
  }, [effectiveFontFamily])

  useEffect(() => {
    if (panelRef.current) {
      panelRef.current.style.setProperty('--cbw-color', color)
      panelRef.current.style.setProperty('--cbw-bot-msg-color', effectiveBotMsgColor)
      panelRef.current.style.setProperty('--cbw-user-msg-color', effectiveUserMsgColor)
      panelRef.current.style.setProperty('--cbw-bot-msg-text-color', effectiveBotMsgTextColor)
      panelRef.current.style.setProperty('--cbw-user-msg-text-color', effectiveUserMsgTextColor)
      panelRef.current.style.setProperty('--cbw-header-color', effectiveHeaderColor)
      panelRef.current.style.setProperty('--cbw-header-text-color', effectiveHeaderTextColor)
      panelRef.current.style.setProperty('--cbw-font-family', effectiveFontFamily)
      if (effectivePanelBg) panelRef.current.style.setProperty('--cbw-panel-bg', effectivePanelBg)
      if (effectiveChatBg) panelRef.current.style.setProperty('--cbw-chat-bg', effectiveChatBg)
      if (effectiveInputBg) panelRef.current.style.setProperty('--cbw-input-bg', effectiveInputBg)
      if (effectiveInputText) panelRef.current.style.setProperty('--cbw-input-text', effectiveInputText)
      if (effectiveBubbleRadius) panelRef.current.style.setProperty('--cbw-bubble-radius', effectiveBubbleRadius)
      if (effectiveSendBtnColor) panelRef.current.style.setProperty('--cbw-send-bg', effectiveSendBtnColor)
      if (effectivePanelHeight) panelRef.current.style.setProperty('--cbw-panel-height', effectivePanelHeight)
      if (effectivePanelWidth) panelRef.current.style.setProperty('--cbw-panel-width', effectivePanelWidth)
      
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
  }, [color, effectiveBotMsgColor, effectiveUserMsgColor, effectiveBotMsgTextColor, effectiveUserMsgTextColor, effectiveHeaderColor, effectiveHeaderTextColor, effectiveFontFamily, pos, effectivePanelHeight, effectivePanelWidth, effectivePanelBg, effectiveInputBg, effectiveInputText, effectiveChatBg, effectiveBubbleRadius, effectiveSendBtnColor, previewMode])

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
          const wm = { role: 'assistant', content: msg, ts: Date.now(), type: 'welcome' } as ChatMessage
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
      const wm = { role: 'assistant', content: msg, ts: Date.now(), type: 'welcome' } as ChatMessage
      setMessages([wm])
      saveSession(chatbotId, { sessionId: s.sessionId, messages: [wm] })
    }
  }, [chatbotId, config, resetSession, sessionIdOverride, welcome, useOverrides])

  const send = async () => {
    if (loading) return
    const text = input.trim()
    if (!text) return
    setInput('')
    const um = { role: 'user', content: text, ts: Date.now() } as ChatMessage
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
        try { const tr = await fetch(embedTokenUrl); if (tr.ok) { const t = await tr.text(); token = t; setEmbedToken(t) } } catch (e) { logger.warn('Embed token fetch failed', e) }
      }
      let captchaToken = ''
      if (captchaSiteKey && (window as any).getCaptchaToken) {
        try { captchaToken = await (window as any).getCaptchaToken(captchaSiteKey) } catch (e) { logger.warn('Captcha token fetch failed', e) }
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
      const am: ChatMessage = { 
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
      const error = e instanceof Error ? e : new Error(String(e))
      logger.error('Chat request failed', error)
      emitEvent('ERROR', { type: 'chat_error', message: error.message })
      
      // Use standard delay constant
      await new Promise((r) => setTimeout(r, ERROR_DISPLAY_DELAY_MS))
      
      const em = { role: 'assistant', content: DEFAULT_ERROR_MESSAGE.tr, ts: Date.now() } as ChatMessage
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
    } catch (e) {
      logger.warn('Feedback failed', e)
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
    // onOpenChange handled by effect
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
          maxChars={config?.max_chars ?? DEFAULT_MAX_CHARS}
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
