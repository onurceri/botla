import { useState, createElement } from 'react'
import Markdown from 'markdown-to-jsx'

type Msg = { 
  id?: string
  role: 'user' | 'assistant'
  content: string
  ts?: number
  feedback?: boolean
  type?: 'welcome' | 'handoff' | 'normal'
  handoffRequestId?: string
  emailSubmitted?: boolean
}

type Props = {
  m: Msg
  onFeedback?: (id: string, isPositive: boolean) => void
  onSubmitEmail?: (requestId: string, email: string) => Promise<void>
  botIcon?: string
}

export function Message({ m, onFeedback, onSubmitEmail, botIcon }: Props) {
  const [email, setEmail] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [submitted, setSubmitted] = useState(m.emailSubmitted || false)
  const [error, setError] = useState('')
  
  const time = new Date(m.ts || Date.now()).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false })
  const isUser = m.role === 'user'
  const isHandoff = m.type === 'handoff'

  const handleSubmitEmail = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!email.trim() || !email.includes('@') || !m.handoffRequestId || !onSubmitEmail) return
    
    setSubmitting(true)
    setError('')
    
    try {
      await onSubmitEmail(m.handoffRequestId, email.trim())
      setSubmitted(true)
    } catch {
      setError('E-posta gönderilemedi, lütfen tekrar deneyin.')
    } finally {
      setSubmitting(false)
    }
  }

  // Handoff card with email form
  if (isHandoff && m.handoffRequestId && onSubmitEmail) {
    return (
      <div className="cbw-msg-row assistant">
        <div className="cbw-avatar">
          {botIcon ? (
            <img src={botIcon} alt="" className="cbw-avatar-img" />
          ) : (
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M3 11h3a2 2 0 0 1 2 2v3a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-5Zm0 0a9 9 0 1 1 18 0m0 0v5a2 2 0 0 1-2 2h-1a2 2 0 0 1-2-2v-3a2 2 0 0 1 2-2h3Z"/>
            </svg>
          )}
        </div>
        <div className="cbw-msg assistant cbw-handoff-card">
          {/* Decorative glass circles */}
          <div className="cbw-handoff-glass-circle"></div>
          
          {submitted ? (
            <div className="cbw-handoff-success">
              <div className="cbw-handoff-success-sparkle">✨</div>
              <div className="cbw-handoff-success-title">
                Talebiniz alındı!
              </div>
              <div className="cbw-handoff-success-text">
                En kısa sürede sizinle iletişime geçeceğiz.
              </div>
            </div>
          ) : (
            <>
              <div className="cbw-handoff-header">
                <div className="cbw-handoff-icon-wrapper">
                  <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                    <path d="M3 11h3a2 2 0 0 1 2 2v3a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-5Zm0 0a9 9 0 1 1 18 0m0 0v5a2 2 0 0 1-2 2h-1a2 2 0 0 1-2-2v-3a2 2 0 0 1 2-2h3Z"/>
                  </svg>
                </div>
                <span className="cbw-handoff-title">
                  Destek Talebi
                </span>
              </div>
              <div className="cbw-handoff-description">
                Size dönüş yapabilmemiz için lütfen e-postanızı paylaşın.
              </div>
              <form onSubmit={handleSubmitEmail} className="cbw-handoff-form">
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="e-posta@adresiniz.com"
                  disabled={submitting}
                  className="cbw-handoff-input"
                />
                {error && <div className="cbw-handoff-error">{error}</div>}
                <button
                  type="submit"
                  disabled={submitting || !email.includes('@')}
                  className="cbw-handoff-submit"
                >
                  {submitting ? 'Gönderiliyor...' : 'Gönder'}
                </button>
              </form>
            </>
          )}
          <div className="cbw-handoff-ts">{time}</div>
        </div>
      </div>
    )
  }

  // Regular message
  return (
    <div className={`cbw-msg-row ${isUser ? 'user' : 'assistant'}`}>
      {!isUser && (
        <div className="cbw-avatar">
          {botIcon ? (
            <img src={botIcon} alt="" className="cbw-avatar-img" />
          ) : (
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
              <path d="M12 8V4H8" />
              <rect width="16" height="12" x="4" y="8" rx="2" />
              <path d="M2 14h2" />
              <path d="M20 14h2" />
              <path d="M15 13v2" />
              <path d="M9 13v2" />
            </svg>
          )}
        </div>
      )}
      
      <div className={`cbw-msg ${m.role}`}>
        <div className="cbw-msg-content">
          <Markdown options={{ createElement }}>{m.content}</Markdown>
        </div>
        <div className="cbw-msg-footer">
          {!isUser && m.id && onFeedback && (
            <div className="cbw-feedback-container">
              <button 
                className={`cbw-feedback-btn ${m.feedback === true ? 'active positive' : ''}`}
                onClick={() => onFeedback(m.id!, true)}
                title="Yararlı"
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill={m.feedback === true ? "currentColor" : "none"} stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M7 10v12" />
                  <path d="M15 5.88 14 10h5.83a2 2 0 0 1 1.92 2.56l-2.33 8A2 2 0 0 1 17.5 22H4a2 2 0 0 1-2-2v-8a2 2 0 0 1 2-2h2.76a2 2 0 0 0 1.79-1.11L12 2h0a3.13 3.13 0 0 1 3 3.88Z" />
                </svg>
              </button>
              <button 
                className={`cbw-feedback-btn ${m.feedback === false ? 'active negative' : ''}`}
                onClick={() => onFeedback(m.id!, false)}
                title="Yararlı değil"
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill={m.feedback === false ? "currentColor" : "none"} stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M17 14V2" />
                  <path d="M9 18.12 10 14H4.17a2 2 0 0 1-1.92-2.56l2.33-8A2 2 0 0 1 6.5 2H20a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2h-2.76a2 2 0 0 0-1.79-1.11L12 22h0a3.13 3.13 0 0 1-3-3.88Z" />
                </svg>
              </button>
            </div>
          )}
          <div className="cbw-ts">{time}</div>
        </div>
      </div>

      {isUser && (
        <div className="cbw-avatar">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
            <path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2" />
            <circle cx="12" cy="7" r="4" />
          </svg>
        </div>
      )}
    </div>
  )
}
