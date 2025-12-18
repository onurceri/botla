import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { SaveIndicator } from '../SaveIndicator'

describe('SaveIndicator', () => {
  it('renders nothing when there is no state', () => {
    const { container } = render(
      <SaveIndicator isSaving={false} lastSavedAt={null} error={null} />
    )
    expect(container.textContent).toBe('')
  })

  it('renders saving state', () => {
    render(<SaveIndicator isSaving lastSavedAt={null} error={null} />)
    expect(screen.getByText('Kaydediliyor...')).toBeInTheDocument()
  })

  it('renders success state', () => {
    render(
      <SaveIndicator
        isSaving={false}
        lastSavedAt={new Date()}
        error={null}
      />
    )
    expect(screen.getByText('Kaydedildi')).toBeInTheDocument()
  })

  it('renders error state and retry text', () => {
    render(
      <SaveIndicator
        isSaving={false}
        lastSavedAt={null}
        error="Hata - Tekrar deneniyor..."
      />
    )
    expect(screen.getByText('Tekrar deneniyor...')).toBeInTheDocument()
  })

  it('renders error state as generic message for long errors', () => {
    render(
      <SaveIndicator
        isSaving={false}
        lastSavedAt={null}
        error="This model is not available on your plan"
      />
    )
    expect(screen.getByText('Kaydedilemedi')).toBeInTheDocument()
  })
})
