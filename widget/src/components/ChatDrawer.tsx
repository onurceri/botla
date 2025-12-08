import { useEffect, useRef } from 'react'
import { Message as MsgComp } from './Message'
import { Suggestions } from './Suggestions'

type Msg = { role: 'user' | 'assistant'; content: string; ts?: number }

type CustomBranding = {
  logo_url?: string
  text?: string
  link?: string
}

export function ChatDrawer(
  { color: _color, messages, loading, input, setInput, onSend, onClose, botName, botIcon, suggestions, onPickSuggestion, maxChars = 1000, hideBranding = false, customBranding }:
  { color: string; messages: Msg[]; loading: boolean; input: string; setInput: (v: string) => void; onSend: () => void; onClose: () => void; botName?: string; botIcon?: string; suggestions?: string[]; onPickSuggestion?: (q: string) => void; maxChars?: number; hideBranding?: boolean; customBranding?: CustomBranding }
) {
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const MAX_CHARS = maxChars

  const scrollToBottom = () => {
    if (messagesEndRef.current && typeof messagesEndRef.current.scrollIntoView === 'function') {
      messagesEndRef.current.scrollIntoView({ behavior: 'smooth' })
    }
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages, loading])

  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
      textareaRef.current.style.height = `${Math.min(textareaRef.current.scrollHeight, 120)}px`
    }
  }, [input])

  const handleInput = (e: any) => {
    const val = e.currentTarget.value
    if (val.length <= MAX_CHARS) {
      setInput(val)
    }
  }

  const handleKeyDown = (e: any) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      if (!loading && input.trim()) {
        onSend()
        // Reset height after send
        if (textareaRef.current) textareaRef.current.style.height = 'auto'
      }
    }
  }

  return (
    <div className="cbw-panel" role="dialog" aria-label="Chatbot">
      <div className="cbw-header">
        <div className="cbw-header-title">
            {botIcon && <img src={botIcon} alt="" style={{ width: '28px', height: '28px', borderRadius: '50%', objectFit: 'cover' }} />}
            <span>{botName || 'Chatbot'}</span>
        </div>
        <button className="cbw-close-btn" onClick={onClose} aria-label="Kapat">×</button>
      </div>
      <div className="cbw-messages">
        {messages.map((m, i) => <MsgComp key={i} m={m} />)}
        {(!messages || messages.filter(m => m.role === 'user').length === 0) && suggestions && suggestions.length > 0 && (
          <div className="cbw-msg-row assistant" style={{ justifyContent: 'flex-start', alignItems: 'flex-start' }}>
            <div className="cbw-avatar" style={{ marginTop: '4px' }}>
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M12 8V4H8" />
                <rect width="16" height="12" x="4" y="8" rx="2" />
                <path d="M2 14h2" />
                <path d="M20 14h2" />
                <path d="M15 13v2" />
                <path d="M9 13v2" />
              </svg>
            </div>
            <div style={{ maxWidth: '85%' }}>
              <Suggestions items={suggestions} disabled={!!loading} onPick={(q) => {
                if (onPickSuggestion) onPickSuggestion(q)
              }} />
            </div>
          </div>
        )}
        {loading && (
          <div className="cbw-msg-row assistant" style={{ justifyContent: 'flex-start' }}>
            <div className="cbw-avatar">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M12 8V4H8" />
                <rect width="16" height="12" x="4" y="8" rx="2" />
                <path d="M2 14h2" />
                <path d="M20 14h2" />
                <path d="M15 13v2" />
                <path d="M9 13v2" />
              </svg>
            </div>
            <div className="cbw-msg assistant" style={{ display: 'flex', gap: '4px', alignItems: 'center', padding: '12px 16px' }}>
              <span style={{ width: '6px', height: '6px', background: 'currentColor', borderRadius: '50%', animation: 'cbw-bounce 1.4s infinite ease-in-out both', opacity: 0.7 }}></span>
              <span style={{ width: '6px', height: '6px', background: 'currentColor', borderRadius: '50%', animation: 'cbw-bounce 1.4s infinite ease-in-out both', animationDelay: '0.16s', opacity: 0.7 }}></span>
              <span style={{ width: '6px', height: '6px', background: 'currentColor', borderRadius: '50%', animation: 'cbw-bounce 1.4s infinite ease-in-out both', animationDelay: '0.32s', opacity: 0.7 }}></span>
            </div>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>
      <div className="cbw-input-area">
        <div className="cbw-input-wrapper" style={{ alignItems: 'flex-end' }}>
          <textarea
            ref={textareaRef}
            rows={1}
            className="cbw-input-field"
            placeholder="Mesaj yazın..."
            value={input}
            onInput={handleInput}
            onKeyDown={handleKeyDown}
            disabled={loading}
            style={{ resize: 'none', overflowY: 'auto' }}
          />
          <button className="cbw-send-btn" onClick={onSend} disabled={loading || !input.trim()} aria-label="Gönder" style={{ marginBottom: '2px' }}>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
              <line x1="22" y1="2" x2="11" y2="13"></line>
              <polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
            </svg>
          </button>
        </div>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          {/* Branding Footer */}
          {hideBranding && customBranding ? (
            <div className="cbw-brand">
              {customBranding.logo_url && <img src={customBranding.logo_url} alt="" style={{ height: '16px', marginRight: '4px', verticalAlign: 'middle' }} />}
              {customBranding.link ? (
                <a href={customBranding.link} target="_blank" rel="noreferrer">{customBranding.text || 'Powered by'}</a>
              ) : (
                <span>{customBranding.text || ''}</span>
              )}
            </div>
          ) : !hideBranding ? (
            <div className="cbw-brand">Powered by <a href="https://botla.co" target="_blank" rel="noreferrer">Botla</a></div>
          ) : (
            <div></div>
          )}
          <div className="cbw-char-limit">{input.length} / {MAX_CHARS}</div>
        </div>
      </div>
      <style>{`
        @keyframes cbw-bounce {
          0%, 80%, 100% { transform: scale(0); }
          40% { transform: scale(1); }
        }
      `}</style>
    </div>
  )
}
