export function ChatBubble({ color, unread, onClick, icon }: { color: string; unread: number; onClick: () => void; icon?: string }) {
  return (
    <button 
      className={`cbw-bubble ${icon ? 'has-icon' : ''}`} 
      onClick={onClick} 
      aria-label="Sohbeti aç" 
      style={!icon ? { background: color } : {}}
    >
      {icon ? (
        <img src={icon} alt="" className="cbw-bubble-icon" />
      ) : (
        <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
          <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
        </svg>
      )}
      {unread > 0 && <span className="cbw-badge" aria-label={`Okunmamış ${unread}`}>{unread}</span>}
    </button>
  )
}
