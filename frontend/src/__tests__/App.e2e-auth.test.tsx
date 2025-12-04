import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import App from '@/App'
import * as analyticsApi from '@/api/analytics'
import { api } from '@/api/client'

describe('App E2E auth bypass', () => {
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

  afterEach(() => {
    vi.unstubAllEnvs()
  })

  it('renders protected routes when VITE_E2E is true without token', async () => {
    vi.spyOn(window.localStorage, 'getItem').mockReturnValue(null)
    vi.stubEnv('VITE_E2E', '1')
    vi.spyOn(analyticsApi, 'getAnalytics').mockResolvedValueOnce([
      { date: new Date().toISOString(), conversations: 1, messages: 2 },
    ] as any)
    vi.spyOn(api, 'get').mockResolvedValueOnce({ data: [{ id: 1, name: 'Bot', model: 'gpt-4o' }] } as any)
    render(<App />)
    expect(screen.queryByRole('heading', { name: /Giriş Yap/i })).not.toBeInTheDocument()
    expect(await screen.findByRole('heading', { name: 'Dashboard' })).toBeInTheDocument()
  })
})
