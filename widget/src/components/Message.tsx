type Msg = { role: 'user' | 'assistant'; content: string; ts?: number }

export function Message({ m }: { m: Msg }) {
  const time = new Date(m.ts || Date.now()).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false })
  return (
    <div className={`cbw-msg ${m.role}`}>
      <div className="cbw-msg-row">
        <div className={`cbw-avatar ${m.role}`}>
          {m.role === 'user' ? (
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{ opacity: 0.8 }}>
              <path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2" />
              <circle cx="12" cy="7" r="4" />
            </svg>
          ) : (
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" style={{ opacity: 0.8 }}>
              <path d="M12 8V4H8" />
              <rect width="16" height="12" x="4" y="8" rx="2" />
              <path d="M2 14h2" />
              <path d="M20 14h2" />
              <path d="M15 13v2" />
              <path d="M9 13v2" />
            </svg>
          )}
        </div>
        <div className="cbw-text">
          <div>{m.content}</div>
          <div className="cbw-ts">{time}</div>
        </div>
      </div>
    </div>
  )
}
