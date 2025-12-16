import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import App from '@/App'

describe('App unauth redirect', () => {
  it('redirects to login when not authenticated and no E2E', () => {
    Object.defineProperty(window, 'localStorage', {
      value: {
        getItem: vi.fn().mockReturnValue(null),
        setItem: vi.fn(),
        removeItem: vi.fn(),
      },
      writable: true,
    })
    window.history.pushState({}, 'Test page', '/dashboard')
    render(<App />)
    expect(screen.getByRole('heading', { name: 'Hoş Geldiniz' })).toBeInTheDocument()
  })
})
