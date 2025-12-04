type Msg = { role: 'user' | 'assistant'; content: string; ts?: number }

export function Message({ m }: { m: Msg }) {
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
        {m.content}
        <div className="cbw-ts">{time}</div>
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
