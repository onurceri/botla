type Msg = { id?: string; role: 'user' | 'assistant'; content: string; ts?: number; feedback?: boolean }

export function Message({ m, onFeedback }: { m: Msg; onFeedback?: (id: string, isPositive: boolean) => void }) {
  const time = new Date(m.ts || Date.now()).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false })
  const isUser = m.role === 'user'
  
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
