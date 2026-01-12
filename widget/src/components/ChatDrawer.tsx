import { useEffect, useRef } from 'react'
import { Message as MsgComp } from './Message'
import { Suggestions } from './Suggestions'
import { ConfirmModal } from './ConfirmModal'
import type { ChatMessage, CustomBranding } from '../types'

const MARKETING_URL = import.meta.env.VITE_MARKETING_URL || 'https://botla.app'
const LONG_CONVERSATION_THRESHOLD = 15 // Show warning after 15 user messages

interface ChatDrawerProps {
  messages: ChatMessage[]
  loading: boolean
  input: string
  setInput: (v: string) => void
  onSend: () => void
  onClose: () => void
  onResetSession?: () => void
  showResetConfirm?: boolean
  onResetConfirm?: () => void
  onResetCancel?: () => void
  botName?: string
  botIcon?: string
  suggestions?: string[]
  onPickSuggestion?: (q: string) => void
  maxChars?: number
  hideBranding?: boolean
  customBranding?: CustomBranding
  onFeedback?: (id: string, isPositive: boolean) => void
  onSubmitEmail?: (requestId: string, email: string) => Promise<void>
  isPreviewMode?: boolean
}

export function ChatDrawer({
  messages,
  loading,
  input,
  setInput,
  onSend,
  onClose,
  onResetSession,
  showResetConfirm = false,
  onResetConfirm,
  onResetCancel,
  botName,
  botIcon,
  suggestions,
  onPickSuggestion,
  maxChars = 1000,
  hideBranding = false,
  customBranding,
  onFeedback,
  onSubmitEmail,
  isPreviewMode = false
}: ChatDrawerProps) {
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const MAX_CHARS = maxChars

  // Count user messages for long conversation warning
  const userMessageCount = messages.filter(m => m.role === 'user').length
  const showLongConversationWarning = userMessageCount >= LONG_CONVERSATION_THRESHOLD

  const scrollToBottom = () => {
    if (messagesEndRef.current && typeof messagesEndRef.current.scrollIntoView === 'function') {
      messagesEndRef.current.scrollIntoView({ behavior: 'smooth' })
    }
  }

  // Restore focus after loading finishes (message sent)
  const prevLoading = useRef(loading)
  useEffect(() => {
    if (prevLoading.current && !loading) {
      // Small timeout to ensure DOM is ready and enabled
      setTimeout(() => {
        textareaRef.current?.focus()
      }, 10)
    }
    prevLoading.current = loading
  }, [loading])

  useEffect(() => {
    scrollToBottom()
  }, [messages, loading])

  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
      textareaRef.current.style.height = `${Math.min(textareaRef.current.scrollHeight, 120)}px`
    }
  }, [input])

  const handleInput = (e: Event) => {
    const target = e.currentTarget as HTMLTextAreaElement
    const val = target.value
    if (val.length <= MAX_CHARS) {
      setInput(val)
    }
  }

  const handleKeyDown = (e: KeyboardEvent) => {
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
    <div className={`cbw-panel ${isPreviewMode ? 'cbw-preview-panel' : ''}`} role="dialog" aria-label="Chatbot">
      <div className="cbw-header">
        <div className="cbw-header-title">
            {botIcon ? (
              <img src={botIcon} alt="" className="cbw-header-icon" />
            ) : (
              <div className="cbw-header-icon-default">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M12 8V4H8" />
                  <rect width="16" height="12" x="4" y="8" rx="2" />
                  <path d="M2 14h2" />
                  <path d="M20 14h2" />
                  <path d="M15 13v2" />
                  <path d="M9 13v2" />
                </svg>
              </div>
            )}
            <span>{botName || 'Chatbot'}</span>
        </div>
        <div className="cbw-header-actions">
          {onResetSession && (
            <button className="cbw-reset-btn" onClick={onResetSession} aria-label="Yeni Konuşma" title="Yeni Konuşma Başlat">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                <path d="M21 12a9 9 0 0 0-9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
                <path d="M3 3v5h5" />
                <path d="M3 12a9 9 0 0 0 9 9 9.75 9.75 0 0 0 6.74-2.74L21 16" />
                <path d="M16 16h5v5" />
              </svg>
            </button>
          )}
          <button className="cbw-close-btn" onClick={onClose} aria-label="Kapat">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
              <line x1="18" y1="6" x2="6" y2="18"></line>
              <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg>
          </button>
        </div>
      </div>
      <div className="cbw-messages">
        {/* Long conversation warning */}
        {showLongConversationWarning && onResetSession && (
          <div className="cbw-conversation-warning">
            <div className="cbw-conversation-warning-icon">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <circle cx="12" cy="12" r="10" />
                <path d="M12 16v-4" />
                <path d="M12 8h.01" />
              </svg>
            </div>
            <div className="cbw-conversation-warning-content">
              <span className="cbw-conversation-warning-text">
                Daha iyi sonuçlar için yeni konuşma başlatabilirsiniz.
              </span>
            </div>
            <button className="cbw-conversation-warning-action" onClick={onResetSession}>
              Yeni
            </button>
          </div>
        )}
        {messages.map((m, i) => <MsgComp key={i} m={m} onFeedback={onFeedback} onSubmitEmail={onSubmitEmail} botIcon={botIcon} />)}
        {(!messages || (messages.filter(m => m.role === 'user').length === 0 && !messages.some(m => m.type === 'handoff'))) && suggestions && suggestions.length > 0 && (
          <div className="cbw-suggestions-container">

            <Suggestions items={suggestions} disabled={!!loading} onPick={(q) => {
              if (onPickSuggestion) onPickSuggestion(q)
            }} />
          </div>
        )}
        {loading && (
          <div className="cbw-msg-row">
            <div className="cbw-avatar">
              {botIcon ? (
                <img src={botIcon} alt="" className="cbw-avatar-img" />
              ) : (
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M12 8V4H8" />
                  <rect width="16" height="12" x="4" y="8" rx="2" />
                  <path d="M2 14h2" />
                  <path d="M20 14h2" />
                  <path d="M15 13v2" />
                  <path d="M9 13v2" />
                </svg>
              )}
            </div>
            <div className="cbw-loading-bubble">
              <span className="cbw-loading-dot"></span>
              <span className="cbw-loading-dot"></span>
              <span className="cbw-loading-dot"></span>
            </div>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>
      <div className="cbw-input-area">
        <div className="cbw-input-wrapper">
          <textarea
            ref={textareaRef}
            rows={1}
            className="cbw-input-field"
            placeholder="Mesaj yazın..."
            value={input}
            onInput={handleInput}
            onKeyDown={handleKeyDown}
            disabled={loading}
          />
          <button className="cbw-send-btn" onClick={onSend} disabled={loading || !input.trim()} aria-label="Gönder">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
              <path d="M22 2L11 13"></path>
              <path d="M22 2L15 22L11 13L2 9L22 2Z"></path>
            </svg>
          </button>
        </div>
        <div className="cbw-input-footer">
           <div className="cbw-char-limit">{input.length} / {MAX_CHARS}</div>
          {/* Branding Footer */}
          {hideBranding && customBranding ? (
            <div className="cbw-brand">
              {customBranding.link ? (
                <a href={customBranding.link} target="_blank" rel="noreferrer">{typeof customBranding.text === 'string' ? customBranding.text : 'Powered by'}</a>
              ) : (
                <span>{typeof customBranding.text === 'string' ? customBranding.text : ''}</span>
              )}
            </div>
          ) : !hideBranding ? (
             <div className="cbw-brand">Powered by <a href={MARKETING_URL} target="_blank" rel="noreferrer">Botla</a></div>
          ) : (
            <div></div>
          )}
        </div>
      </div>
      <style>{`
        /* Local overrides if needed */
      `}</style>
      {showResetConfirm && onResetConfirm && onResetCancel && (
        <ConfirmModal
          isOpen={showResetConfirm}
          icon="refresh"
          title="Yeni Konuşma Başlat"
          message="Mevcut konuşmanız silinecek."
          confirmText="Başlat"
          cancelText="Vazgeç"
          onConfirm={onResetConfirm}
          onCancel={onResetCancel}
        />
      )}
    </div>
  )
}
