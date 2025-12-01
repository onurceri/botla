import { Message as MsgComp } from './Message'

type Msg = { role: 'user' | 'assistant'; content: string; ts?: number }

export function ChatDrawer(
  { color, messages, loading, input, setInput, onSend, onClose, botName, botIcon }:
  { color: string; messages: Msg[]; loading: boolean; input: string; setInput: (v: string) => void; onSend: () => void; onClose: () => void; botName?: string; botIcon?: string }
) {
  return (
    <div className="cbw-panel" role="dialog" aria-label="Chatbot">
      <div className="cbw-header">
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            {botIcon && <img src={botIcon} alt="" style={{ width: '24px', height: '24px', borderRadius: '50%', objectFit: 'cover' }} />}
            <span>{botName || 'Chatbot'}</span>
        </div>
        <button className="cbw-bubble" style={{ background: 'transparent', color: 'inherit', boxShadow: 'none', width: 'auto', height: 'auto', padding: '4px' }} onClick={onClose} aria-label="Kapat">×</button>
      </div>
      <div className="cbw-messages">
        {messages.map((m, i) => <MsgComp key={i} m={m} />)}
        {loading && <div className="cbw-msg assistant">Yazıyor…</div>}
      </div>
      <div className="cbw-input">
        <input
          type="text"
          placeholder="Mesaj yazın"
          value={input}
          onChange={(e) => setInput(e.currentTarget.value)}
          onKeyDown={(e) => { if (e.key === 'Enter' && !loading) onSend() }}
          disabled={loading}
        />
        <button onClick={onSend} disabled={loading} style={{ background: color }}>Gönder</button>
      </div>
      <div className="cbw-brand">Powered by <a href="https://botla.co" target="_blank" rel="noreferrer">Botla</a></div>
    </div>
  )
}
