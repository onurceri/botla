type Msg = { role: 'user' | 'assistant'; content: string; ts?: number }

export function Message({ m }: { m: Msg }) {
  const time = new Date(m.ts || Date.now()).toLocaleTimeString()
  return (
    <div className={`cbw-msg ${m.role}`}>
      <div className="cbw-msg-row">
        <div className={`cbw-avatar ${m.role}`}>{m.role === 'user' ? '👤' : '🤖'}</div>
        <div className="cbw-text">
          <div>{m.content}</div>
          <div className="cbw-ts">{time}</div>
        </div>
      </div>
    </div>
  )
}
