export function ChatBubble({ color, unread, onClick, icon }: { color: string; unread: number; onClick: () => void; icon?: string }) {
  return (
    <button className="cbw-bubble" onClick={onClick} aria-label="Sohbeti aç" style={{ background: icon ? 'transparent' : color }}>
      {icon ? <img src={icon} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover', borderRadius: '50%' }} /> : '💬'}
      {unread > 0 && <span className="cbw-badge" aria-label={`Okunmamış ${unread}`}>{unread}</span>}
    </button>
  )
}
