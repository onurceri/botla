export function ChatBubble({ color, unread, onClick, icon, isOpen = false }: { 
  color: string; 
  unread: number; 
  onClick: () => void; 
  icon?: string;
  isOpen?: boolean;
}) {
  return (
    <button 
      className={`cbw-bubble ${icon && !isOpen ? 'has-icon' : ''} ${isOpen ? 'is-open' : ''}`} 
      onClick={onClick} 
      aria-label={isOpen ? "Sohbeti kapat" : "Sohbeti aç"} 
      style={!icon || isOpen ? { background: color } : {}}
    >
      {isOpen ? (
        // Show X icon when panel is open (minimize button)
        <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
          <line x1="18" y1="6" x2="6" y2="18"></line>
          <line x1="6" y1="6" x2="18" y2="18"></line>
        </svg>
      ) : icon ? (
        <img src={icon} alt="" className="cbw-bubble-icon" />
      ) : (
        <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
          <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
        </svg>
      )}
      {!isOpen && unread > 0 && <span className="cbw-badge" aria-label={`Okunmamış ${unread}`}>{unread}</span>}
    </button>
  )
}
