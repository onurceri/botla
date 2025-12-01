import { useEffect, useRef, useState } from 'react'
import { ChatBubble } from './components/ChatBubble'
import { ChatDrawer } from './components/ChatDrawer'

type Message = { role: 'user' | 'assistant'; content: string; ts?: number }

export function WidgetApp({ chatbotId, apiBase, themeColor, welcome }: { chatbotId: string; apiBase?: string; themeColor?: string; welcome?: string }) {
  const [open, setOpen] = useState(false)
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const panelRef = useRef<HTMLDivElement | null>(null)
  const [config, setConfig] = useState<any>(null)
  const [sid, setSid] = useState<string>('')

  useEffect(() => {
    const base = apiBase || ''
    const url = `${base}/api/v1/public/chatbots/${encodeURIComponent(chatbotId)}`
    fetch(url).then(r => r.json()).then(data => {
      setConfig(data)
      if (data.welcome_message && !welcome) {
        // If welcome message is not overridden by prop, use from config
        // Logic to handle welcome message if not already present
      }
    }).catch(() => {})
  }, [chatbotId, apiBase])

  const color = themeColor || config?.theme_color || '#3b82f6'
  const position = config?.position || 'bottom-right'
  
  useEffect(() => {
    if (panelRef.current) {
      panelRef.current.style.setProperty('--cbw-color', color)
      panelRef.current.style.setProperty('--cbw-bot-msg-color', config?.bot_message_color || color)
      panelRef.current.style.setProperty('--cbw-user-msg-color', config?.user_message_color || color)
      panelRef.current.style.setProperty('--cbw-bot-msg-text-color', config?.bot_message_text_color || '#ffffff')
      panelRef.current.style.setProperty('--cbw-user-msg-text-color', config?.user_message_text_color || '#ffffff')
      panelRef.current.style.setProperty('--cbw-header-color', config?.chat_header_color || color)
      panelRef.current.style.setProperty('--cbw-header-text-color', config?.chat_header_text_color || '#ffffff')
      panelRef.current.style.setProperty('--cbw-font-family', config?.chat_font_family || 'inherit')
      
      // Position
      panelRef.current.style.bottom = '20px'
      if (position === 'bottom-left') {
        panelRef.current.style.left = '20px'
        panelRef.current.style.right = 'auto'
      } else {
        panelRef.current.style.right = '20px'
        panelRef.current.style.left = 'auto'
      }
    }
  }, [color, config])

  useEffect(() => {
    const s = getSession(chatbotId)
    setSid(s.sessionId)
    if (s.messages && s.messages.length > 0) {
      setMessages(s.messages)
    } else if (welcome || config?.welcome_message) {
      const msg = welcome || config?.welcome_message
      const wm = { role: 'assistant', content: msg, ts: Date.now() } as Message
      setMessages([wm])
      saveSession(chatbotId, { sessionId: s.sessionId, messages: [wm] })
    }
  }, [chatbotId, config])

  const send = async () => {
    const text = input.trim()
    if (!text) return
    setInput('')
    const um = { role: 'user', content: text, ts: Date.now() } as Message
    setMessages((m) => {
      const nm = [...m, um]
      saveSession(chatbotId, { sessionId: ensureSession(chatbotId, sid, setSid), messages: nm })
      return nm
    })
    setLoading(true)
    try {
      const base = apiBase || ''
      const url = `${base}/api/v1/public/chatbots/${encodeURIComponent(chatbotId)}/chat`
      const res = await fetch(url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ message: text, session_id: ensureSession(chatbotId, sid, setSid) }),
      })
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      const data = await res.json()
      const ans: string = data.response || 'Merhaba!'
      const am = { role: 'assistant', content: ans, ts: Date.now() } as Message
      setMessages((m) => {
        const nm = [...m, am]
        saveSession(chatbotId, { sessionId: ensureSession(chatbotId, sid, setSid), messages: nm })
        return nm
      })
      if (!open) setUnread((u) => u + 1)
    } catch {
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

  const toggle = () => setOpen((v) => {
    const nv = !v
    if (nv) setUnread(0)
    return nv
  })

  return (
    <div className="cbw-container" ref={panelRef}>
      {open ? (
        <ChatDrawer 
          color={color} 
          messages={messages} 
          loading={loading} 
          input={input} 
          setInput={setInput} 
          onSend={send} 
          onClose={toggle}
          botName={config?.bot_display_name}
          botIcon={config?.bot_icon}
        />
      ) : (
        <ChatBubble color={color} unread={unread} onClick={toggle} icon={config?.bot_icon} />
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
function ensureSession(chatbotId: string, sid: string, setSid: (v: string) => void): string {
  if (sid && sid.length > 0) return sid
  const s = getSession(chatbotId)
  setSid(s.sessionId)
  return s.sessionId
}
