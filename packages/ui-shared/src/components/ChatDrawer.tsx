/** @jsxImportSource react */
import React, { useEffect, useRef } from 'react'
import { Message } from './Message'
import { Suggestions } from './Suggestions'
import { LoadingIndicator } from './LoadingIndicator'
import type { ChatMessage, CustomBranding } from '../types'

interface ChatDrawerProps {
  messages: ChatMessage[]
  loading: boolean
  input: string
  setInput: (_v: string) => void
  onSend: () => void
  onClose: () => void
  botName?: string
  botIcon?: string
  suggestions?: string[]
  onPickSuggestion?: (_q: string) => void
  maxChars?: number
  hideBranding?: boolean
  customBranding?: CustomBranding
  onFeedback?: (_id: string, _isPositive: boolean) => void
  onSubmitEmail?: (_requestId: string, _email: string) => Promise<void>
  isPreviewMode?: boolean
  marketingUrl?: string
  classNames?: {
    panel?: string
    header?: string
    headerTitle?: string
    headerIcon?: string
    closeBtn?: string
    messagesArea?: string
    suggestionsContainer?: string
    inputArea?: string
    inputWrapper?: string
    inputField?: string
    sendBtn?: string
    inputFooter?: string
    branding?: string
    charLimit?: string
  }
}

const DEFAULT_MARKETING_URL = 'https://botla.app'

/**
 * ChatDrawer component - Main chat interface
 * 
 * Displays the full chat interface including messages, input field,
 * suggestions, and branding. Handles message scrolling and auto-focus.
 */
export function ChatDrawer({
  messages,
  loading,
  input,
  setInput,
  onSend,
  onClose,
  botName,
  botIcon,
  suggestions,
  onPickSuggestion,
  maxChars = 1000,
  hideBranding = false,
  customBranding,
  onFeedback,
  onSubmitEmail,
  isPreviewMode = false,
  marketingUrl = DEFAULT_MARKETING_URL,
  classNames = {},
}: ChatDrawerProps) {
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const textareaRef = useRef<HTMLTextAreaElement>(null)

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

  const handleInput = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const val = e.currentTarget.value
    if (val.length <= maxChars) {
      setInput(val)
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      if (!loading && input.trim()) {
        onSend()
        // Reset height after send
        if (textareaRef.current) textareaRef.current.style.height = 'auto'
      }
    }
  }

  const showSuggestions =
    (!messages ||
      (messages.filter((m) => m.role === 'user').length === 0 &&
        !messages.some((m) => m.type === 'handoff'))) &&
    suggestions &&
    suggestions.length > 0

  return (
    <div
      className={`cbw-panel ${isPreviewMode ? 'cbw-preview-panel' : ''} ${classNames.panel || ''}`}
      role="dialog"
      aria-label="Chatbot"
    >
      <div className={`cbw-header ${classNames.header || ''}`}>
        <div className={`cbw-header-title ${classNames.headerTitle || ''}`}>
          {botIcon ? (
            <img src={botIcon} alt="" className={`cbw-header-icon ${classNames.headerIcon || ''}`} />
          ) : (
            <div className={`cbw-header-icon-default ${classNames.headerIcon || ''}`}>
              <svg
                width="20"
                height="20"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2.5"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
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
        <button
          className={`cbw-close-btn ${classNames.closeBtn || ''}`}
          onClick={onClose}
          aria-label="Kapat"
        >
          <svg
            width="18"
            height="18"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="3"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <line x1="18" y1="6" x2="6" y2="18"></line>
            <line x1="6" y1="6" x2="18" y2="18"></line>
          </svg>
        </button>
      </div>

      <div className={`cbw-messages ${classNames.messagesArea || ''}`}>
        {messages.map((m, i) => (
          <Message
            key={i}
            message={m}
            onFeedback={onFeedback}
            onSubmitEmail={onSubmitEmail}
            botIcon={botIcon}
          />
        ))}
        {showSuggestions && (
          <div className={`cbw-suggestions-container ${classNames.suggestionsContainer || ''}`}>
            <Suggestions
              items={suggestions}
              disabled={!!loading}
              onPick={(q) => {
                if (onPickSuggestion) onPickSuggestion(q)
              }}
            />
          </div>
        )}
        {loading && <LoadingIndicator botIcon={botIcon} />}
        <div ref={messagesEndRef} />
      </div>

      <div className={`cbw-input-area ${classNames.inputArea || ''}`}>
        <div className={`cbw-input-wrapper ${classNames.inputWrapper || ''}`}>
          <textarea
            ref={textareaRef}
            rows={1}
            className={`cbw-input-field ${classNames.inputField || ''}`}
            placeholder="Mesaj yazın..."
            value={input}
            onChange={handleInput}
            onKeyDown={handleKeyDown}
            disabled={loading}
          />
          <button
            className={`cbw-send-btn ${classNames.sendBtn || ''}`}
            onClick={onSend}
            disabled={loading || !input.trim()}
            aria-label="Gönder"
          >
            <svg
              width="18"
              height="18"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="3"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M22 2L11 13"></path>
              <path d="M22 2L15 22L11 13L2 9L22 2Z"></path>
            </svg>
          </button>
        </div>
        <div className={`cbw-input-footer ${classNames.inputFooter || ''}`}>
          {/* Branding Footer */}
          {hideBranding && customBranding ? (
            <div className={`cbw-brand-custom ${classNames.branding || ''}`}>
              {customBranding.logo_url && (
                <img src={customBranding.logo_url} alt="" className="cbw-brand-logo" />
              )}
              {customBranding.link ? (
                <a href={customBranding.link} target="_blank" rel="noreferrer" className="cbw-brand-text">
                  {typeof customBranding.text === 'string' ? customBranding.text : 'Powered by'}
                </a>
              ) : (
                <span className="cbw-brand-text">
                  {typeof customBranding.text === 'string' ? customBranding.text : ''}
                </span>
              )}
            </div>
          ) : !hideBranding ? (
            <div className={`cbw-brand-default ${classNames.branding || ''}`}>
              Powered by{' '}
              <a href={marketingUrl} target="_blank" rel="noreferrer">
                Botla
              </a>
            </div>
          ) : (
            <div></div>
          )}
          <div className={`cbw-char-limit ${classNames.charLimit || ''}`}>
            {input.length} / {maxChars}
          </div>
        </div>
      </div>
    </div>
  )
}
