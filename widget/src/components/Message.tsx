import { useState } from 'react'

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
}

export function Message({ m, onFeedback, onSubmitEmail }: Props) {
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
      <div className="cbw-msg-row assistant" style={{ justifyContent: 'flex-start' }}>
        <div className="cbw-avatar">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <path d="M3 11h3a2 2 0 0 1 2 2v3a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-5Zm0 0a9 9 0 1 1 18 0m0 0v5a2 2 0 0 1-2 2h-1a2 2 0 0 1-2-2v-3a2 2 0 0 1 2-2h3Z"/>
          </svg>
        </div>
        <div className="cbw-msg assistant cbw-handoff-card" style={{ 
          background: 'linear-gradient(135deg, var(--cbw-color) 0%, color-mix(in srgb, var(--cbw-color) 80%, #000) 100%)',
          borderRadius: '16px',
          padding: '16px',
          maxWidth: '85%',
          boxShadow: '0 4px 12px rgba(0,0,0,0.15)'
        }}>
          {submitted ? (
            <div style={{ textAlign: 'center', padding: '12px 0' }}>
              <div style={{ fontSize: '28px', marginBottom: '12px' }}>✅</div>
              <div style={{ fontWeight: 600, fontSize: '16px', marginBottom: '6px', color: 'var(--cbw-bot-msg-text-color, #fff)' }}>
                Talebiniz alındı!
              </div>
              <div style={{ fontSize: '14px', opacity: 0.9, color: 'var(--cbw-bot-msg-text-color, #fff)', lineHeight: '1.4' }}>
                En kısa sürede sizinle iletişime geçeceğiz.
              </div>
            </div>
          ) : (
            <>
              <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '16px' }}>
                <div style={{ 
                  background: 'rgba(255,255,255,0.2)', 
                  borderRadius: '50%', 
                  padding: '8px',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center'
                }}>
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{ color: 'var(--cbw-bot-msg-text-color, #fff)' }}>
                    <path d="M3 11h3a2 2 0 0 1 2 2v3a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-5Zm0 0a9 9 0 1 1 18 0m0 0v5a2 2 0 0 1-2 2h-1a2 2 0 0 1-2-2v-3a2 2 0 0 1 2-2h3Z"/>
                  </svg>
                </div>
                <span style={{ fontWeight: 600, fontSize: '15px', color: 'var(--cbw-bot-msg-text-color, #fff)', lineHeight: '1.3' }}>
                  Destek Talebi Oluşturun
                </span>
              </div>
              <div style={{ fontSize: '14px', marginBottom: '16px', color: 'var(--cbw-bot-msg-text-color, #fff)', opacity: 0.95, lineHeight: '1.5' }}>
                Size en kısa sürede dönüş yapabilmemiz için lütfen e-posta adresinizi paylaşın.
              </div>
              <form onSubmit={handleSubmitEmail} style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="ornek@email.com"
                  disabled={submitting}
                  style={{
                    padding: '12px 14px',
                    borderRadius: '10px',
                    border: '1px solid rgba(255,255,255,0.2)',
                    fontSize: '14px',
                    outline: 'none',
                    background: 'rgba(255,255,255,0.95)',
                    color: '#333',
                    width: '100%',
                    boxSizing: 'border-box',
                    boxShadow: '0 2px 4px rgba(0,0,0,0.05)'
                  }}
                />
                {error && <div style={{ fontSize: '13px', color: '#ff8a8a', background: 'rgba(0,0,0,0.2)', padding: '6px 10px', borderRadius: '6px' }}>{error}</div>}
                <button
                  type="submit"
                  disabled={submitting || !email.includes('@')}
                  style={{
                    padding: '12px 16px',
                    borderRadius: '10px',
                    border: 'none',
                    background: '#fff',
                    color: 'var(--cbw-color, #000)',
                    fontWeight: 600,
                    fontSize: '14px',
                    cursor: submitting ? 'not-allowed' : 'pointer',
                    opacity: submitting ? 0.7 : 1,
                    transition: 'all 0.2s',
                    boxShadow: '0 4px 12px rgba(0,0,0,0.1)'
                  }}
                >
                  {submitting ? 'Gönderiliyor...' : 'Gönder'}
                </button>
              </form>
            </>
          )}
          <div className="cbw-ts" style={{ marginTop: '8px', textAlign: 'right', fontSize: '11px', opacity: 0.7, color: 'var(--cbw-bot-msg-text-color, #fff)' }}>{time}</div>
        </div>
      </div>
    )
  }

  // Regular message
  return (
    <div className={`cbw-msg-row ${isUser ? 'user' : 'assistant'}`} style={{ justifyContent: isUser ? 'flex-end' : 'flex-start' }}>
      {!isUser && (
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
      )}
      
      <div className={`cbw-msg ${m.role}`}>
        <div className="cbw-content">{m.content}</div>
        <div className="cbw-footer" style={{ display: 'flex', alignItems: 'center', justifyContent: 'flex-end', gap: '8px', marginTop: '4px' }}>
          {!isUser && m.id && onFeedback && (
            <div className="cbw-feedback" style={{ display: 'flex', gap: '4px' }}>
              <button 
                className={`cbw-feedback-btn ${m.feedback === true ? 'active' : ''}`}
                onClick={() => onFeedback(m.id!, true)}
                title="Yararlı"
                style={{ opacity: m.feedback === false ? 0.3 : 1, border: 'none', background: 'none', cursor: 'pointer', padding: 0, color: 'inherit' }}
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill={m.feedback === true ? "currentColor" : "none"} stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M7 10v12" />
                  <path d="M15 5.88 14 10h5.83a2 2 0 0 1 1.92 2.56l-2.33 8A2 2 0 0 1 17.5 22H4a2 2 0 0 1-2-2v-8a2 2 0 0 1 2-2h2.76a2 2 0 0 0 1.79-1.11L12 2h0a3.13 3.13 0 0 1 3 3.88Z" />
                </svg>
              </button>
              <button 
                className={`cbw-feedback-btn ${m.feedback === false ? 'active' : ''}`}
                onClick={() => onFeedback(m.id!, false)}
                title="Yararlı değil"
                style={{ opacity: m.feedback === true ? 0.3 : 1, border: 'none', background: 'none', cursor: 'pointer', padding: 0, color: 'inherit' }}
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill={m.feedback === false ? "currentColor" : "none"} stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M17 14V2" />
                  <path d="M9 18.12 10 14H4.17a2 2 0 0 1-1.92-2.56l2.33-8A2 2 0 0 1 6.5 2H20a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2h-2.76a2 2 0 0 0-1.79 1.11L12 22h0a3.13 3.13 0 0 1-3-3.88Z" />
                </svg>
              </button>
            </div>
          )}
          <div className="cbw-ts">{time}</div>
        </div>
      </div>

      {isUser && (
        <div className="cbw-avatar">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2" />
            <circle cx="12" cy="7" r="4" />
          </svg>
        </div>
      )}
    </div>
  )
}
