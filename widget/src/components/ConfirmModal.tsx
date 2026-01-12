import { useEffect, useRef } from 'react'

interface ConfirmModalProps {
  isOpen: boolean
  icon?: 'refresh' | 'warning' | 'info'
  title: string
  message: string
  confirmText?: string
  cancelText?: string
  onConfirm: () => void
  onCancel: () => void
}

const icons = {
  refresh: (
    <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M21 12a9 9 0 0 0-9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
      <path d="M3 3v5h5" />
      <path d="M3 12a9 9 0 0 0 9 9 9.75 9.75 0 0 0 6.74-2.74L21 16" />
      <path d="M16 16h5v5" />
    </svg>
  ),
  warning: (
    <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z" />
      <path d="M12 9v4" />
      <path d="M12 17h.01" />
    </svg>
  ),
  info: (
    <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="12" cy="12" r="10" />
      <path d="M12 16v-4" />
      <path d="M12 8h.01" />
    </svg>
  )
}

export function ConfirmModal({
  isOpen,
  icon = 'refresh',
  title,
  message,
  confirmText = 'Başlat',
  cancelText = 'Vazgeç',
  onConfirm,
  onCancel
}: ConfirmModalProps) {
  const modalRef = useRef<HTMLDivElement>(null)

  // Handle escape key
  useEffect(() => {
    if (!isOpen) return
    
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onCancel()
    }
    
    document.addEventListener('keydown', handleEscape)
    return () => document.removeEventListener('keydown', handleEscape)
  }, [isOpen, onCancel])

  // Focus trap
  useEffect(() => {
    if (!isOpen || !modalRef.current) return
    
    const firstButton = modalRef.current.querySelector('button')
    firstButton?.focus()
  }, [isOpen])

  if (!isOpen) return null

  return (
    <div className="cbw-modal-overlay" onClick={onCancel} role="dialog" aria-modal="true" aria-labelledby="cbw-modal-title">
      <div 
        ref={modalRef}
        className="cbw-modal" 
        onClick={(e) => e.stopPropagation()}
      >
        <div className="cbw-modal-icon">
          {icons[icon]}
        </div>
        <h3 id="cbw-modal-title" className="cbw-modal-title">{title}</h3>
        <p className="cbw-modal-message">{message}</p>
        <div className="cbw-modal-actions">
          <button 
            className="cbw-modal-btn cbw-modal-btn-cancel" 
            onClick={onCancel}
          >
            {cancelText}
          </button>
          <button 
            className="cbw-modal-btn cbw-modal-btn-confirm" 
            onClick={onConfirm}
          >
            {confirmText}
          </button>
        </div>
      </div>
    </div>
  )
}
