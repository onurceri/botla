/** @jsxImportSource react */

interface LoadingIndicatorProps {
  botIcon?: string
  classNames?: {
    row?: string
    avatar?: string
    bubble?: string
    dot?: string
  }
}

/**
 * LoadingIndicator component - Shows typing indicator
 * 
 * Displays an animated loading indicator to show the bot is "typing"
 */
export function LoadingIndicator({ botIcon, classNames = {} }: LoadingIndicatorProps) {
  return (
    <div className={`cbw-loading-row ${classNames.row || ''}`}>
      <div className={`cbw-avatar ${classNames.avatar || ''}`}>
        {botIcon ? (
          <img src={botIcon} alt="" className="cbw-avatar-img" />
        ) : (
          <svg
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M12 8V4H8" />
            <rect width="16" height="12" x="4" y="8" rx="2" />
            <path d="M2 14h2" />
            <path d="M20 14h2" />
            <path d="M15 13v2" />
            <path d="M9 13v2" />
          </svg>
        )}
      </div>
      <div className={`cbw-loading-bubble ${classNames.bubble || ''}`}>
        <span className={`cbw-loading-dot ${classNames.dot || ''}`}></span>
        <span className={`cbw-loading-dot ${classNames.dot || ''}`}></span>
        <span className={`cbw-loading-dot ${classNames.dot || ''}`}></span>
      </div>
    </div>
  )
}
