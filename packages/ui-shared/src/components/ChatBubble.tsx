/** @jsxImportSource react */

interface ChatBubbleProps {
  color: string
  unread?: number
  onClick: () => void
  icon?: string
  classNames?: {
    button?: string
    icon?: string
    badge?: string
  }
}

/**
 * ChatBubble component - Floating chat button
 * 
 * Displays a floating button that opens the chat interface.
 * Supports custom icons and unread message badges.
 */
export function ChatBubble({
  color,
  unread = 0,
  onClick,
  icon,
  classNames = {},
}: ChatBubbleProps) {
  return (
    <button
      className={`cbw-bubble ${icon ? 'has-icon' : ''} ${classNames.button || ''}`}
      onClick={onClick}
      aria-label="Sohbeti aç"
      style={!icon ? { background: color } : {}}
    >
      {icon ? (
        <img src={icon} alt="" className={`cbw-bubble-icon ${classNames.icon || ''}`} />
      ) : (
        <svg
          width="32"
          height="32"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2.5"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
        </svg>
      )}
      {unread > 0 && (
        <span className={`cbw-badge ${classNames.badge || ''}`} aria-label={`Okunmamış ${unread}`}>
          {unread}
        </span>
      )}
    </button>
  )
}
