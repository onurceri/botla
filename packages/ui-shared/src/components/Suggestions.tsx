/** @jsxImportSource react */
import { useState } from 'react'

interface SuggestionsProps {
  items: string[]
  disabled?: boolean
  onPick: (_question: string) => void
  classNames?: {
    container?: string
    header?: string
    title?: string
    counter?: string
    carousel?: string
    carouselBtn?: string
    viewport?: string
    item?: string
  }
}

/**
 * Suggestions component - Displays suggested questions in a carousel
 * 
 * Shows a list of suggested questions that users can click to quickly
 * ask common questions. Includes navigation controls when multiple
 * suggestions are available.
 */
export function Suggestions({
  items,
  disabled = false,
  onPick,
  classNames = {},
}: SuggestionsProps) {
  if (!items || items.length === 0) return null

  const [currentIndex, setCurrentIndex] = useState(0)

  const next = () => {
    setCurrentIndex((prev) => (prev + 1) % items.length)
  }

  const prev = () => {
    setCurrentIndex((prev) => (prev - 1 + items.length) % items.length)
  }

  return (
    <div className={`cbw-suggestions ${classNames.container || ''}`} aria-label="Önerilen sorular">
      <div className={`cbw-suggestions-header ${classNames.header || ''}`}>
        <span className={`cbw-suggestions-title ${classNames.title || ''}`}>
          <span className="cbw-suggestions-sparkle">✨</span> ÖNERİLEN SORULAR
        </span>
        {items.length > 1 && (
          <span className={`cbw-suggestions-counter ${classNames.counter || ''}`}>
            {currentIndex + 1} / {items.length}
          </span>
        )}
      </div>
      <div className={`cbw-suggestions-carousel ${classNames.carousel || ''}`}>
        {items.length > 1 && (
          <button
            className={`cbw-carousel-btn prev ${classNames.carouselBtn || ''}`}
            onClick={prev}
            disabled={disabled}
            aria-label="Önceki soru"
          >
            <svg
              width="14"
              height="14"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="3"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M15 18l-6-6 6-6" />
            </svg>
          </button>
        )}

        <div className={`cbw-carousel-viewport ${classNames.viewport || ''}`}>
          <button
            key={currentIndex}
            className={`cbw-suggestion cbw-carousel-item ${classNames.item || ''}`}
            onClick={() => onPick(items[currentIndex])}
            disabled={disabled}
            aria-label={items[currentIndex]}
          >
            {items[currentIndex]}
          </button>
        </div>

        {items.length > 1 && (
          <button
            className={`cbw-carousel-btn next ${classNames.carouselBtn || ''}`}
            onClick={next}
            disabled={disabled}
            aria-label="Sonraki soru"
          >
            <svg
              width="14"
              height="14"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="3"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M9 18l6-6-6-6" />
            </svg>
          </button>
        )}
      </div>
    </div>
  )
}
