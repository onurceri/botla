import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, within, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import SourceList from '../SourceList'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

// Mock ResizeObserver for Radix UI
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
}

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  })
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
}

describe('SourceList', () => {
  it('renders sources and calls delete', () => {
    const onDelete = vi.fn()
    const onRefresh = vi.fn()
    const sources = [
      { id: '1', source_type: 'pdf', original_filename: 'file.pdf', status: 'completed', chunk_count: 3 },
    ] as any
    const { container } = render(<SourceList sources={sources} userPlan="pro" onDelete={onDelete} onRefresh={onRefresh} />, { wrapper: createWrapper() })
    expect(screen.getAllByText('file.pdf').length).toBeGreaterThan(0)
    const delBtn = container.querySelector('button[aria-label="Kaynağı Sil"]') as HTMLButtonElement
    fireEvent.click(delBtn)
    expect(onDelete).toHaveBeenCalledWith('1')
  })

  it('renders status variants with proper indicators', () => {
    const onDelete = vi.fn()
    const onRefresh = vi.fn()
    const sources = [
      { id: '1', source_type: 'pdf', original_filename: 'file.pdf', status: 'completed', chunk_count: 3 },
      { id: '2', source_type: 'url', source_url: 'https://a.com', status: 'processing', chunk_count: 1 },
      { id: '3', source_type: 'text', original_filename: 'note', status: 'failed', chunk_count: 0 },
      { id: '4', source_type: 'pdf', original_filename: 'other.pdf', status: 'queued', chunk_count: 2 },
    ] as any
    const { container } = render(<SourceList sources={sources} userPlan="pro" onDelete={onDelete} onRefresh={onRefresh} />, { wrapper: createWrapper() })
    expect(container.querySelector('span.bg-emerald-100')).toBeTruthy()
    // Note: there may be multiple labels due to mobile + desktop views
    expect(screen.getAllByText('failed').length).toBeGreaterThan(0)
    expect(screen.getAllByText('queued').length).toBeGreaterThan(0)
  })

  it('shows refresh button only for URL sources', () => {
    const onDelete = vi.fn()
    const onRefresh = vi.fn()
    const sources = [
      { id: '1', source_type: 'pdf', original_filename: 'file.pdf', status: 'completed', chunk_count: 3 },
      { id: '2', source_type: 'url', source_url: 'https://a.com', status: 'completed', chunk_count: 1 },
    ] as any
    const { container } = render(<SourceList sources={sources} userPlan="pro" onDelete={onDelete} onRefresh={onRefresh} />, { wrapper: createWrapper() })
    const rows = container.querySelectorAll('tbody tr')
    // PDF row should NOT have refresh button
    const pdfRow = rows[0]
    expect(within(pdfRow as HTMLElement).queryByLabelText('Kaynağı Yenile')).toBeNull()
    // URL row should have refresh button
    const urlRow = rows[1]
    expect(within(urlRow as HTMLElement).queryByLabelText('Kaynağı Yenile')).toBeTruthy()
  })

  it('disables refresh button for free plan', () => {
    const onDelete = vi.fn()
    const onRefresh = vi.fn()
    const sources = [
      { id: '1', source_type: 'url', source_url: 'https://a.com', status: 'completed', chunk_count: 1 },
    ] as any
    const { container } = render(<SourceList sources={sources} userPlan="free" onDelete={onDelete} onRefresh={onRefresh} />, { wrapper: createWrapper() })
    const refreshBtn = container.querySelector('button[aria-label="Kaynağı Yenile"]')
    expect(refreshBtn).toBeDisabled()
  })

  it('calls onRefresh when clicking refresh button', () => {
    const onDelete = vi.fn()
    const onRefresh = vi.fn()
    const sources = [
      { id: '1', source_type: 'url', source_url: 'https://a.com', status: 'completed', chunk_count: 1 },
    ] as any
    const { container } = render(<SourceList sources={sources} userPlan="pro" onDelete={onDelete} onRefresh={onRefresh} />, { wrapper: createWrapper() })
    const refreshBtn = container.querySelector('button[aria-label="Kaynağı Yenile"]') as HTMLElement
    fireEvent.click(refreshBtn)
    expect(onRefresh).toHaveBeenCalledWith('1')
  })

  it.skip('shows tooltip for failed sources with error message', async () => {
    const user = userEvent.setup()
    const onDelete = vi.fn()
    const onRefresh = vi.fn()
    const sources = [
      { id: '1', source_type: 'url', source_url: 'https://err.com', status: 'failed', chunk_count: 0, error_message: 'HTTP 403 Forbidden' },
    ] as any
    render(<SourceList sources={sources} userPlan="pro" onDelete={onDelete} onRefresh={onRefresh} />, { wrapper: createWrapper() })
    
    const badges = screen.getAllByText('failed')
    expect(badges.length).toBeGreaterThan(0)
    
    await user.hover(badges[0])
    
    await waitFor(() => {
        expect(screen.getByText('HTTP 403 Forbidden')).toBeInTheDocument()
    })
  })
})
