import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import App from '@/App'

describe('PrivateRoute', () => {
  beforeEach(() => {
    Object.defineProperty(window, 'localStorage', {
      value: {
        getItem: vi.fn(),
        setItem: vi.fn(),
        removeItem: vi.fn(),
      },
      writable: true,
    })
  })
  it('renders protected layout when authenticated', () => {
    vi.spyOn(window.localStorage, 'getItem').mockReturnValue('tok')
    render(<App />)
    expect(screen.queryByRole('heading', { name: /Giriş Yap/i })).not.toBeInTheDocument()
  })
})
