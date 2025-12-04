import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { ToastProvider } from '@/components/ui/toast'
import DashboardPage from '../DashboardPage'
import { api } from '@/api/client'
import * as analyticsApi from '@/api/analytics'

describe('DashboardPage', () => {
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

  it('renders stats and recent bots from API', async () => {
    const chart = Array.from({ length: 7 }).map((_, i) => ({
      date: new Date(2024, 0, i + 1).toISOString(),
      conversations: i + 1,
      messages: (i + 1) * 2,
    }))
    vi.spyOn(analyticsApi, 'getAnalytics').mockResolvedValueOnce(chart as any)
    vi.spyOn(api, 'get').mockResolvedValueOnce({ data: [
      { id: 1, name: 'Bot A', model: 'gpt-4o' },
      { id: 2, name: 'Bot B', model: 'gpt-4.1' },
      { id: 3, name: 'Bot C', model: 'gpt-3.5' },
    ] } as any)

    render(
      <ToastProvider>
        <MemoryRouter>
          <DashboardPage />
        </MemoryRouter>
      </ToastProvider>
    )

    expect(screen.getByText('Yükleniyor...')).toBeInTheDocument()

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Dashboard' })).toBeInTheDocument()
    })

    expect(screen.getByText('Toplam Konuşma')).toBeInTheDocument()
    expect(screen.getByText('Toplam Mesaj')).toBeInTheDocument()
    expect(screen.getByText('Aktif Botlar')).toBeInTheDocument()
    expect(screen.getByText('Bot A')).toBeInTheDocument()
    expect(screen.getByText('Bot B')).toBeInTheDocument()
  })

  it('handles fetch errors and shows empty chart state', async () => {
    const errSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    vi.spyOn(analyticsApi, 'getAnalytics').mockRejectedValueOnce(new Error('network'))
    vi.spyOn(api, 'get').mockResolvedValueOnce({ data: [] } as any)

    render(
      <ToastProvider>
        <MemoryRouter>
          <DashboardPage />
        </MemoryRouter>
      </ToastProvider>
    )

    await waitFor(() => {
      expect(errSpy).toHaveBeenCalled()
    })

    expect(screen.queryByText('Yükleniyor...')).not.toBeInTheDocument()
    expect(screen.getByText('Henüz aktivite yok')).toBeInTheDocument()
  })

  it('shows date range and growth badge when data exists', async () => {
    const chart = Array.from({ length: 7 }).map((_, i) => ({
      date: new Date(2024, 4, i + 10).toISOString(),
      conversations: 5,
      messages: 10,
    }))
    vi.spyOn(analyticsApi, 'getAnalytics').mockResolvedValueOnce(chart as any)
    vi.spyOn(api, 'get').mockResolvedValueOnce({ data: [
      { id: 1, name: 'Bot A', model: 'gpt-4o' },
      { id: 2, name: 'Bot B', model: 'gpt-4.1' },
      { id: 3, name: 'Bot C', model: 'gpt-3.5' },
      { id: 4, name: 'Bot D', model: 'gpt-4o-mini' },
    ] } as any)

    render(
      <ToastProvider>
        <MemoryRouter>
          <DashboardPage />
        </MemoryRouter>
      </ToastProvider>
    )

    const rangeText = await screen.findByText((t) => t.includes('10') && t.includes('16'))
    expect(rangeText).toBeInTheDocument()
    expect(screen.getAllByText('+12.5%').length).toBeGreaterThan(0)

    expect(screen.getByText('3')).toBeInTheDocument()
    expect(screen.queryByText('Bot D')).not.toBeInTheDocument()
  })

  it('shows default header when chartData empty and empty bots message', async () => {
    vi.spyOn(analyticsApi, 'getAnalytics').mockResolvedValueOnce([] as any)
    vi.spyOn(api, 'get').mockResolvedValueOnce({ data: [] } as any)

    render(
      <ToastProvider>
        <MemoryRouter>
          <DashboardPage />
        </MemoryRouter>
      </ToastProvider>
    )

    const header = await screen.findByText('Son 7 Gün')
    expect(header).toBeInTheDocument()
    expect(screen.getAllByText('Henüz aktivite yok').length).toBeGreaterThan(0)
    expect(screen.getAllByText('Henüz bir bot oluşturmadınız.').length).toBeGreaterThan(0)
  })

  it('renders XAxis tick labels with Turkish short month', async () => {
    const chart = [
      { date: new Date(2024, 4, 10).toISOString(), conversations: 1, messages: 2 },
      { date: new Date(2024, 4, 11).toISOString(), conversations: 1, messages: 2 },
    ]
    vi.spyOn(analyticsApi, 'getAnalytics').mockResolvedValueOnce(chart as any)
    vi.spyOn(api, 'get').mockResolvedValueOnce({ data: [{ id: 1, name: 'Bot', model: 'gpt-4o' }] } as any)

    render(
      <ToastProvider>
        <MemoryRouter>
          <DashboardPage />
        </MemoryRouter>
      </ToastProvider>
    )

    // Expect tick labels like "10 May" and "11 May" to appear in the DOM
    expect(await screen.findByText((t) => /10\s*May/i.test(t))).toBeInTheDocument()
    expect(screen.getByText((t) => /11\s*May/i.test(t))).toBeInTheDocument()
  })
})
