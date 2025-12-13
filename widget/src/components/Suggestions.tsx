import { useState } from 'react'

export function Suggestions({ items, disabled, onPick }: { items: string[]; disabled: boolean; onPick: (q: string) => void }) {
  if (!items || items.length === 0) return null

  const [currentIndex, setCurrentIndex] = useState(0)

  const next = () => {
    setCurrentIndex((prev) => (prev + 1) % items.length)
  }

  const prev = () => {
    setCurrentIndex((prev) => (prev - 1 + items.length) % items.length)
  }

  return (
    <div className="cbw-suggestions" aria-label="Önerilen sorular">
      <div className="cbw-suggestions-header">
        <span>✨ ÖRNEK SORULAR</span>
        {items.length > 1 && (
          <span style={{ marginLeft: 'auto', fontSize: '9px', opacity: 0.6 }}>
            {currentIndex + 1}/{items.length}
          </span>
        )}
      </div>
      <div className="cbw-suggestions-carousel">
        {items.length > 1 && (
          <button 
            className="cbw-carousel-btn prev" 
            onClick={prev}
            disabled={disabled}
            aria-label="Önceki soru"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M15 18l-6-6 6-6" />
            </svg>
          </button>
        )}
        
        <div className="cbw-carousel-viewport">
          <button
            key={currentIndex}
            className="cbw-suggestion cbw-carousel-item"
            onClick={() => onPick(items[currentIndex])}
            disabled={disabled}
            aria-label={items[currentIndex]}
          >
            {items[currentIndex]}
          </button>
        </div>

        {items.length > 1 && (
          <button 
            className="cbw-carousel-btn next" 
            onClick={next}
            disabled={disabled}
            aria-label="Sonraki soru"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M9 18l6-6-6-6" />
            </svg>
          </button>
        )}
      </div>
    </div>
  )
}
