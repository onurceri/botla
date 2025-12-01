import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { fireEvent } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import AnalyticsPage from '../AnalyticsPage'
import * as api from '@/api/analytics'

const renderWithClient = (ui: React.ReactNode) => {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>)
}

describe('AnalyticsPage', () => {
  it('shows empty state when no data', async () => {
    vi.spyOn(api, 'getAnalytics').mockResolvedValueOnce([])
    renderWithClient(<AnalyticsPage />)
    expect(await screen.findByText('Veri Yok')).toBeInTheDocument()
  })

  it('shows error state when request fails', async () => {
    vi.spyOn(api, 'getAnalytics').mockRejectedValueOnce(new Error('fail'))
    renderWithClient(<AnalyticsPage />)
    expect(await screen.findByText('Veriler alınırken bir hata oluştu.')).toBeInTheDocument()
  })

  it('shows totals for fetched data', async () => {
    vi.spyOn(api, 'getAnalytics').mockResolvedValueOnce([
      { date: '2025-11-20', messages: 10, conversations: 3 },
      { date: '2025-11-21', messages: 5, conversations: 2 },
    ])
    renderWithClient(<AnalyticsPage />)
    expect(await screen.findByText('15')).toBeInTheDocument()
    expect(screen.getByText('5')).toBeInTheDocument()
  })

  it('applies date range filter and recalculates totals', async () => {
    vi.spyOn(api, 'getAnalytics').mockResolvedValueOnce([
      { date: '2025-11-20', messages: 10, conversations: 3 },
      { date: '2025-11-21', messages: 5, conversations: 2 },
      { date: '2025-11-22', messages: 8, conversations: 4 },
    ])
    const { container } = renderWithClient(<AnalyticsPage />)

    await screen.findByText('23')
    const inputs = container.querySelectorAll('input[type="date"]')
    const start = inputs[0] as HTMLInputElement
    const end = inputs[1] as HTMLInputElement
    fireEvent.change(start, { target: { value: '2025-11-21' } })
    fireEvent.change(end, { target: { value: '2025-11-22' } })

    expect(await screen.findByText('13')).toBeInTheDocument()
    expect(screen.getByText('6')).toBeInTheDocument()
  })
})
