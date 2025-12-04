import { Message as MsgComp } from './Message'
import { Suggestions } from './Suggestions'

type Msg = { role: 'user' | 'assistant'; content: string; ts?: number }

export function ChatDrawer(
  { color: _color, messages, loading, input, setInput, onSend, onClose, botName, botIcon, suggestions, onPickSuggestion }:
  { color: string; messages: Msg[]; loading: boolean; input: string; setInput: (v: string) => void; onSend: () => void; onClose: () => void; botName?: string; botIcon?: string; suggestions?: string[]; onPickSuggestion?: (q: string) => void }
) {
  return (
    <div className="cbw-panel" role="dialog" aria-label="Chatbot">
      <div className="cbw-header">
        <div className="cbw-header-title">
            {botIcon && <img src={botIcon} alt="" style={{ width: '28px', height: '28px', borderRadius: '50%', objectFit: 'cover' }} />}
            <span>{botName || 'Chatbot'}</span>
        </div>
        <button className="cbw-close-btn" onClick={onClose} aria-label="Kapat">×</button>
      </div>
      <div className="cbw-messages">
        {messages.map((m, i) => <MsgComp key={i} m={m} />)}
        {(!messages || messages.filter(m => m.role === 'user').length === 0) && suggestions && suggestions.length > 0 && (
          <div className="cbw-msg-row assistant" style={{ justifyContent: 'flex-start', alignItems: 'flex-start' }}>
            <div className="cbw-avatar" style={{ marginTop: '4px' }}>
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M12 8V4H8" />
                <rect width="16" height="12" x="4" y="8" rx="2" />
                <path d="M2 14h2" />
                <path d="M20 14h2" />
                <path d="M15 13v2" />
                <path d="M9 13v2" />
              </svg>
            </div>
            <div style={{ maxWidth: '85%' }}>
              <Suggestions items={suggestions} disabled={!!loading} onPick={(q) => {
                if (onPickSuggestion) onPickSuggestion(q)
              }} />
            </div>
          </div>
        )}
        {loading && (
          <div className="cbw-msg-row assistant" style={{ justifyContent: 'flex-start' }}>
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
            <div className="cbw-msg assistant" style={{ display: 'flex', gap: '4px', alignItems: 'center', padding: '12px 16px' }}>
              <span style={{ width: '6px', height: '6px', background: 'currentColor', borderRadius: '50%', animation: 'cbw-bounce 1.4s infinite ease-in-out both', opacity: 0.7 }}></span>
              <span style={{ width: '6px', height: '6px', background: 'currentColor', borderRadius: '50%', animation: 'cbw-bounce 1.4s infinite ease-in-out both', animationDelay: '0.16s', opacity: 0.7 }}></span>
              <span style={{ width: '6px', height: '6px', background: 'currentColor', borderRadius: '50%', animation: 'cbw-bounce 1.4s infinite ease-in-out both', animationDelay: '0.32s', opacity: 0.7 }}></span>
            </div>
          </div>
        )}
      </div>
      <div className="cbw-input-area">
        <div className="cbw-input-wrapper">
          <input
            type="text"
            className="cbw-input-field"
            placeholder="Mesaj yazın..."
            value={input}
            onChange={(e) => setInput(e.currentTarget.value)}
            onKeyDown={(e) => { if (e.key === 'Enter' && !loading) onSend() }}
            disabled={loading}
          />
          <button className="cbw-send-btn" onClick={onSend} disabled={loading || !input.trim()} aria-label="Gönder">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
              <line x1="22" y1="2" x2="11" y2="13"></line>
              <polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
            </svg>
          </button>
        </div>
        <div className="cbw-brand">Powered by <a href="https://botla.co" target="_blank" rel="noreferrer">Botla</a></div>
      </div>
      <style>{`
        @keyframes cbw-bounce {
          0%, 80%, 100% { transform: scale(0); }
          40% { transform: scale(1); }
        }
      `}</style>
    </div>
  )
}
