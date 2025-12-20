import { describe, it, expect, vi } from 'vitest'
import { render, screen, QueryWrapper } from '@/test-utils'
import App from '@/App'

describe('App unauth redirect', () => {
  it('redirects to login when not authenticated and no E2E', async () => {
    // Force VITE_E2E to false for this test
    vi.stubEnv('VITE_E2E', '')
    
    Object.defineProperty(window, 'localStorage', {
      value: {
        getItem: vi.fn().mockReturnValue(null),
        setItem: vi.fn(),
        removeItem: vi.fn(),
      },
      writable: true,
    })
    
    window.history.pushState({}, 'Test page', '/dashboard')
    render(<App />, { wrapper: QueryWrapper })
    
    // Use findByRole to wait for redirect and render
    expect(await screen.findByRole('heading', { name: 'Hoş Geldiniz' })).toBeInTheDocument()
    
    vi.unstubAllEnvs()
  })
})
