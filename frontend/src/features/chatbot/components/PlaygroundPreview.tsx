import { useRef, useState, useEffect } from 'react'
import { sendChatMessage } from '@/api/chat'

type Props = {
  id: string
  themeColor: string
  chatHeaderColor: string
  chatHeaderTextColor: string
  botMessageColor: string
  botMessageTextColor: string
  userMessageColor: string
  userMessageTextColor: string
  chatFontFamily: string
  position: string
  botDisplayName: string
  botIcon: string
  chatBackgroundColor: string
  welcomeMessage: string
  previewOpen: boolean
  sessionId: string
  suggestionsEnabled: boolean
  suggestedQuestions: string[]
  refreshKey?: number
  hideBranding?: boolean
  customBranding?: { logo_url?: string; text?: string; link?: string } | null
}

type Message = {
  id?: string
  role: 'user' | 'assistant'
  content: string
  ts: number
}

export default function PlaygroundPreview(props: Props) {
  const {
    id,
    themeColor,
    chatHeaderColor,
    chatHeaderTextColor,
    botMessageColor,
    botMessageTextColor,
    userMessageColor,
    userMessageTextColor,
    chatFontFamily,
    position,
    botDisplayName,
    botIcon,
    chatBackgroundColor,
    welcomeMessage,
    sessionId,
    suggestionsEnabled,
    suggestedQuestions,
    refreshKey,
    hideBranding,
    customBranding,
  } = props

  const containerRef = useRef<HTMLDivElement>(null)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const [_error, setError] = useState<string | null>(null)
  const [open, setOpen] = useState(true) // Chat panel open/closed state
  const [currentSuggestionIndex, setCurrentSuggestionIndex] = useState(0) // Carousel index
  const MAX_CHARS = 1000

  // Initialize with welcome message
  useEffect(() => {
    if (welcomeMessage) {
      setMessages([{ role: 'assistant', content: welcomeMessage, ts: Date.now() }])
    } else {
      setMessages([])
    }
  }, [welcomeMessage, refreshKey])

  // Auto-scroll to bottom when messages change
  useEffect(() => {
    if (messagesEndRef.current && typeof messagesEndRef.current.scrollIntoView === 'function') {
      messagesEndRef.current.scrollIntoView({ behavior: 'smooth' })
    }
  }, [messages, loading])

  // Auto-resize textarea
  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
      textareaRef.current.style.height = `${Math.min(textareaRef.current.scrollHeight, 120)}px`
    }
  }, [input])

  // Restore focus after loading
  const prevLoading = useRef(loading)
  useEffect(() => {
    if (prevLoading.current && !loading) {
      setTimeout(() => {
        textareaRef.current?.focus()
      }, 10)
    }
    prevLoading.current = loading
  }, [loading])

  const handleSend = async () => {
    if (loading || !input.trim()) return

    const text = input.trim()
    setInput('')
    setError(null)

    const userMessage: Message = { role: 'user', content: text, ts: Date.now() }
    setMessages((prev) => [...prev, userMessage])
    setLoading(true)

    try {
      const response = await sendChatMessage(id, {
        message: text,
        session_id: sessionId,
      })

      const assistantMessage: Message = {
        id: response.message_id,
        role: 'assistant',
        content: response.response || 'Merhaba!',
        ts: Date.now(),
      }
      setMessages((prev) => [...prev, assistantMessage])
    } catch {
      setError('Bir hata oluştu, lütfen tekrar deneyin.')
      const errorMessage: Message = {
        role: 'assistant',
        content: 'Şu an bir hata oluştu, lütfen tekrar deneyin.',
        ts: Date.now(),
      }
      setMessages((prev) => [...prev, errorMessage])
    } finally {
      setLoading(false)
      if (textareaRef.current) {
        textareaRef.current.style.height = 'auto'
      }
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      if (!loading && input.trim()) {
        handleSend()
      }
    }
  }

  const handleInput = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const val = e.currentTarget.value
    if (val.length <= MAX_CHARS) {
      setInput(val)
    }
  }

  const pickSuggestion = (question: string) => {
    if (loading) return
    setInput(question)
    setTimeout(() => {
      if (!loading) handleSend()
    }, 0)
  }

  const toggleOpen = () => {
    setOpen((prev) => !prev)
  }

  const nextSuggestion = () => {
    setCurrentSuggestionIndex((prev) => (prev + 1) % suggestedQuestions.length)
  }

  const prevSuggestion = () => {
    setCurrentSuggestionIndex((prev) => (prev - 1 + suggestedQuestions.length) % suggestedQuestions.length)
  }

  const hasUserMessages = messages.some((m) => m.role === 'user')
  const showSuggestions = suggestionsEnabled && suggestedQuestions.length > 0 && !hasUserMessages

  return (
    <div ref={containerRef} className="flex-1 relative h-full min-h-[400px]">
      <style>{`
        .playground-chat-panel {
          display: flex;
          flex-direction: column;
          height: 100%;
          background: ${chatBackgroundColor || '#ffffff'};
          font-family: ${chatFontFamily || 'inherit'};
          border-radius: 16px;
          overflow: hidden;
          box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
          position: absolute;
          ${position === 'bottom-left' ? 'left: 20px;' : 'right: 20px;'}
          bottom: 20px;
          width: 380px;
          max-width: calc(100% - 40px);
          max-height: calc(100% - 40px);
        }

        .playground-chat-bubble {
          position: absolute;
          ${position === 'bottom-left' ? 'left: 20px;' : 'right: 20px;'}
          bottom: 20px;
          width: 60px;
          height: 60px;
          border-radius: 50%;
          border: none;
          cursor: pointer;
          display: flex;
          align-items: center;
          justify-content: center;
          box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
          transition: transform 0.2s, box-shadow 0.2s;
          background: ${themeColor || '#3b82f6'};
          color: white;
        }

        .playground-chat-bubble:hover {
          transform: scale(1.05);
          box-shadow: 0 6px 20px rgba(0, 0, 0, 0.2);
        }

        .playground-chat-bubble img {
          width: 100%;
          height: 100%;
          object-fit: cover;
          border-radius: 50%;
        }

        .playground-chat-header {
          background: ${chatHeaderColor || themeColor || '#3b82f6'};
          color: ${chatHeaderTextColor || '#ffffff'};
          padding: 16px 20px;
          display: flex;
          align-items: center;
          gap: 12px;
          flex-shrink: 0;
        }

        .playground-chat-header img {
          width: 32px;
          height: 32px;
          border-radius: 50%;
          object-fit: cover;
        }

        .playground-chat-header-title {
          font-weight: 600;
          font-size: 16px;
          flex: 1;
        }

        .playground-close-btn {
          background: transparent;
          border: none;
          color: ${chatHeaderTextColor || '#ffffff'};
          font-size: 28px;
          line-height: 1;
          cursor: pointer;
          padding: 0;
          width: 28px;
          height: 28px;
          display: flex;
          align-items: center;
          justify-content: center;
          opacity: 0.8;
          transition: opacity 0.2s;
        }

        .playground-close-btn:hover {
          opacity: 1;
        }

        .playground-chat-messages {
          flex: 1;
          overflow-y: auto;
          padding: 16px;
          display: flex;
          flex-direction: column;
          gap: 12px;
        }

        .playground-message-row {
          display: flex;
          gap: 8px;
          align-items: flex-start;
        }

        .playground-message-row.user {
          justify-content: flex-end;
        }

        .playground-message-row.assistant {
          justify-content: flex-start;
        }

        .playground-message {
          padding: 10px 14px;
          border-radius: 12px;
          max-width: 75%;
          word-wrap: break-word;
          line-height: 1.5;
        }

        .playground-message.user {
          background: ${userMessageColor || themeColor || '#3b82f6'};
          color: ${userMessageTextColor || '#ffffff'};
        }

        .playground-message.assistant {
          background: ${botMessageColor || '#f3f4f6'};
          color: ${botMessageTextColor || '#1f2937'};
        }

        .playground-avatar {
          width: 28px;
          height: 28px;
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
          background: ${botMessageColor || '#f3f4f6'};
          color: ${botMessageTextColor || '#1f2937'};
          flex-shrink: 0;
        }

        .playground-suggestions {
          display: flex;
          flex-direction: column;
          gap: 8px;
          width: 100%;
          padding: 0 8px;
          margin-bottom: 8px;
        }

        .playground-suggestions-header {
          font-size: 10px;
          text-transform: uppercase;
          letter-spacing: 0.05em;
          font-weight: 700;
          color: rgba(0, 0, 0, 0.5);
          margin-left: 4px;
          margin-bottom: 2px;
          display: flex;
          align-items: center;
          gap: 4px;
        }

        .playground-suggestions-carousel {
          position: relative;
          width: 100%;
          display: flex;
          align-items: center;
          gap: 8px;
        }

        .playground-carousel-viewport {
          flex: 1;
          overflow: hidden;
          display: flex;
          justify-content: center;
        }

        .playground-carousel-item {
          width: 100%;
          text-align: center;
          justify-content: center;
          white-space: normal;
          max-width: 100%;
          min-width: 0;
          margin: 0;
          animation: playground-fade-in 0.2s ease;
        }

        @keyframes playground-fade-in {
          from { opacity: 0; transform: translateY(2px); }
          to { opacity: 1; transform: translateY(0); }
        }

        .playground-carousel-btn {
          width: 24px;
          height: 24px;
          border-radius: 50%;
          background: white;
          border: 1px solid rgba(0,0,0,0.1);
          box-shadow: 0 1px 2px rgba(0,0,0,0.05);
          display: flex;
          align-items: center;
          justify-content: center;
          cursor: pointer;
          color: #6b7280;
          padding: 0;
          transition: all 0.2s;
          flex-shrink: 0;
        }

        .playground-carousel-btn:hover:not(:disabled) {
          background: #f9fafb;
          color: #111827;
          transform: scale(1.1);
          border-color: rgba(0,0,0,0.15);
        }

        .playground-carousel-btn:active:not(:disabled) {
          transform: scale(0.95);
        }

        .playground-carousel-btn:disabled {
          opacity: 0.3;
          cursor: not-allowed;
        }

        .playground-suggestion {
          text-align: left;
          padding: 8px 14px;
          border-radius: 16px;
          background: rgba(255, 255, 255, 0.8);
          border: 1px solid rgba(0, 0, 0, 0.08);
          font-size: 13px;
          color: #374151;
          cursor: pointer;
          transition: all 0.2s;
          position: relative;
          overflow: hidden;
          box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
          width: 100%;
        }

        .playground-suggestion:hover:not(:disabled) {
          background: #ffffff;
          transform: translateY(-1px);
          box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
          border-color: rgba(0, 0, 0, 0.15);
        }

        .playground-suggestion:active:not(:disabled) {
          transform: translateY(0);
        }

        .playground-suggestion:disabled {
          opacity: 0.5;
          cursor: not-allowed;
        }

        .playground-loading {
          display: flex;
          gap: 4px;
          align-items: center;
          padding: 12px 16px;
        }

        .playground-loading span {
          width: 6px;
          height: 6px;
          background: currentColor;
          border-radius: 50%;
          animation: bounce 1.4s infinite ease-in-out both;
          opacity: 0.7;
        }

        .playground-loading span:nth-child(2) {
          animation-delay: 0.16s;
        }

        .playground-loading span:nth-child(3) {
          animation-delay: 0.32s;
        }

        @keyframes bounce {
          0%, 80%, 100% { transform: scale(0); }
          40% { transform: scale(1); }
        }

        .playground-input-area {
          flex-shrink: 0;
          padding: 12px 16px;
          background: ${chatBackgroundColor || '#ffffff'};
          border-top: 1px solid rgba(0, 0, 0, 0.1);
        }

        .playground-input-wrapper {
          display: flex;
          align-items: flex-end;
          gap: 8px;
          margin-bottom: 8px;
        }

        .playground-input {
          flex: 1;
          padding: 10px 12px;
          border: 1px solid rgba(0, 0, 0, 0.2);
          border-radius: 8px;
          font-family: inherit;
          font-size: 14px;
          resize: none;
          overflow-y: auto;
          background: ${chatBackgroundColor || '#ffffff'};
          color: ${botMessageTextColor || '#1f2937'};
        }

        .playground-input:focus {
          outline: none;
          border-color: ${themeColor || '#3b82f6'};
        }

        .playground-input:disabled {
          opacity: 0.6;
          cursor: not-allowed;
        }

        .playground-send-btn {
          width: 40px;
          height: 40px;
          border-radius: 8px;
          border: none;
          background: ${themeColor || '#3b82f6'};
          color: #ffffff;
          cursor: pointer;
          display: flex;
          align-items: center;
          justify-content: center;
          flex-shrink: 0;
          transition: opacity 0.2s;
        }

        .playground-send-btn:hover:not(:disabled) {
          opacity: 0.9;
        }

        .playground-send-btn:disabled {
          opacity: 0.5;
          cursor: not-allowed;
        }

        .playground-footer {
          display: flex;
          justify-content: space-between;
          align-items: center;
          font-size: 12px;
          color: rgba(0, 0, 0, 0.5);
        }

        .playground-branding {
          font-size: 11px;
        }

        .playground-branding a {
          color: ${themeColor || '#3b82f6'};
          text-decoration: none;
        }

        .playground-branding a:hover {
          text-decoration: underline;
        }

        .playground-branding img {
          height: 14px;
          margin-right: 4px;
          vertical-align: middle;
        }
      `}</style>

      {open ? (
        <div className="playground-chat-panel">
          {/* Header */}
          <div className="playground-chat-header">
            {botIcon && <img src={botIcon} alt="" />}
            <div className="playground-chat-header-title">{botDisplayName || 'Chatbot'}</div>
            <button className="playground-close-btn" onClick={toggleOpen} aria-label="Kapat">
              ×
            </button>
          </div>

          {/* Messages */}
          <div className="playground-chat-messages">
            {messages.map((msg, idx) => (
              <div key={idx} className={`playground-message-row ${msg.role}`}>
                {msg.role === 'assistant' && (
                  <div className="playground-avatar">
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                      <path d="M12 8V4H8" />
                      <rect width="16" height="12" x="4" y="8" rx="2" />
                      <path d="M2 14h2" />
                      <path d="M20 14h2" />
                      <path d="M15 13v2" />
                      <path d="M9 13v2" />
                    </svg>
                  </div>
                )}
                <div className={`playground-message ${msg.role}`}>{msg.content}</div>
              </div>
            ))}

            {/* Suggestions */}
            {showSuggestions && (
              <div className="playground-suggestions">
                <div className="playground-suggestions-header">
                  <span>✨ ÖRNEK SORULAR</span>
                  {suggestedQuestions.length > 1 && (
                    <span style={{ marginLeft: 'auto', fontSize: '9px', opacity: 0.6 }}>
                      {currentSuggestionIndex + 1}/{suggestedQuestions.length}
                    </span>
                  )}
                </div>
                <div className="playground-suggestions-carousel">
                  {suggestedQuestions.length > 1 && (
                    <button
                      className="playground-carousel-btn"
                      onClick={prevSuggestion}
                      disabled={loading}
                      aria-label="Önceki soru"
                    >
                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <path d="M15 18l-6-6 6-6" />
                      </svg>
                    </button>
                  )}

                  <div className="playground-carousel-viewport">
                    <button
                      key={currentSuggestionIndex}
                      className="playground-suggestion playground-carousel-item"
                      onClick={() => pickSuggestion(suggestedQuestions[currentSuggestionIndex])}
                      disabled={loading}
                      aria-label={suggestedQuestions[currentSuggestionIndex]}
                    >
                      {suggestedQuestions[currentSuggestionIndex]}
                    </button>
                  </div>

                  {suggestedQuestions.length > 1 && (
                    <button
                      className="playground-carousel-btn"
                      onClick={nextSuggestion}
                      disabled={loading}
                      aria-label="Sonraki soru"
                    >
                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <path d="M9 18l6-6-6-6" />
                      </svg>
                    </button>
                  )}
                </div>
              </div>
            )}

            {/* Loading indicator */}
            {loading && (
              <div className="playground-message-row assistant">
                <div className="playground-avatar">
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M12 8V4H8" />
                    <rect width="16" height="12" x="4" y="8" rx="2" />
                    <path d="M2 14h2" />
                    <path d="M20 14h2" />
                    <path d="M15 13v2" />
                    <path d="M9 13v2" />
                  </svg>
                </div>
                <div className="playground-message assistant playground-loading">
                  <span></span>
                  <span></span>
                  <span></span>
                </div>
              </div>
            )}

            <div ref={messagesEndRef} />
          </div>

          {/* Input Area */}
          <div className="playground-input-area">
            <div className="playground-input-wrapper">
              <textarea
                ref={textareaRef}
                rows={1}
                className="playground-input"
                placeholder="Mesaj yazın..."
                value={input}
                onChange={handleInput}
                onKeyDown={handleKeyDown}
                disabled={loading}
              />
              <button
                className="playground-send-btn"
                onClick={handleSend}
                disabled={loading || !input.trim()}
                aria-label="Gönder"
              >
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                  <line x1="22" y1="2" x2="11" y2="13"></line>
                  <polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
                </svg>
              </button>
            </div>

            <div className="playground-footer">
              {/* Branding */}
              {hideBranding && customBranding ? (
                <div className="playground-branding">
                  {customBranding.logo_url && <img src={customBranding.logo_url} alt="" />}
                  {customBranding.link ? (
                    <a href={customBranding.link} target="_blank" rel="noreferrer">
                      {customBranding.text || 'Powered by'}
                    </a>
                  ) : (
                    <span>{customBranding.text || ''}</span>
                  )}
                </div>
              ) : !hideBranding ? (
                <div className="playground-branding">
                  Powered by{' '}
                  <a href="https://botla.app" target="_blank" rel="noreferrer">
                    Botla
                  </a>
                </div>
              ) : (
                <div></div>
              )}

              {/* Character counter */}
              <div>
                {input.length} / {MAX_CHARS}
              </div>
            </div>
          </div>
        </div>
      ) : (
        <button className="playground-chat-bubble" onClick={toggleOpen} aria-label="Sohbeti aç">
          {botIcon ? (
            <img src={botIcon} alt="" />
          ) : (
            <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
            </svg>
          )}
        </button>
      )}
    </div>
  )
}
