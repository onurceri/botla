import { describe, it, expect, vi } from 'vitest'
import { render, screen, within } from '@testing-library/react'
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
  it('renders sources as cards and calls delete', async () => {
    const user = userEvent.setup()
    const onDelete = vi.fn()
    const onRefresh = vi.fn()
    const sources = [
      { id: '1', source_type: 'pdf', original_filename: 'file.pdf', status: 'completed', chunk_count: 3 },
    ] as any
    const { container } = render(<SourceList sources={sources} userPlan="pro" onDelete={onDelete} onRefresh={onRefresh} />, { wrapper: createWrapper() })
    
    expect(screen.getAllByText('file.pdf').length).toBeGreaterThan(0)
    
    // Find delete button within the source card
    const card = container.querySelector('[data-testid="source-card"]')!
    const delBtn = within(card as HTMLElement).getByLabelText('Kaynağı Sil')
    await user.click(delBtn)
    
    // Confirm delete in dialog
    const confirmBtn = screen.getByRole('button', { name: 'Sil' })
    await user.click(confirmBtn)
    
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
    render(<SourceList sources={sources} userPlan="pro" onDelete={onDelete} onRefresh={onRefresh} />, { wrapper: createWrapper() })
    
    expect(screen.getAllByText('Tamamlandı').length).toBeGreaterThan(0)
    expect(screen.getAllByText('Başarısız').length).toBeGreaterThan(0)
    expect(screen.getAllByText('Beklemede').length).toBeGreaterThan(0)
  })

  it('shows refresh button only for URL sources', () => {
    const onDelete = vi.fn()
    const onRefresh = vi.fn()
    const sources = [
      { id: '1', source_type: 'pdf', original_filename: 'file.pdf', status: 'completed', chunk_count: 3 },
      { id: '2', source_type: 'url', source_url: 'https://a.com', status: 'completed', chunk_count: 1 },
    ] as any
    const { container } = render(<SourceList sources={sources} userPlan="pro" onDelete={onDelete} onRefresh={onRefresh} />, { wrapper: createWrapper() })
    
    const cards = container.querySelectorAll('[data-testid="source-card"]')
    
    // PDF card (first) should NOT have refresh button
    const pdfCardRefreshBtn = within(cards[0] as HTMLElement).queryByLabelText('Kaynağı Yenile')
    expect(pdfCardRefreshBtn).toBeNull()
    
    // URL card (second) should have refresh button
    const urlCardRefreshBtn = within(cards[1] as HTMLElement).queryByLabelText('Kaynağı Yenile')
    expect(urlCardRefreshBtn).toBeTruthy()
  })

  it('disables refresh button for free plan', () => {
    const onDelete = vi.fn()
    const onRefresh = vi.fn()
    const sources = [
      { id: '1', source_type: 'url', source_url: 'https://a.com', status: 'completed', chunk_count: 1 },
    ] as any
    const { container } = render(<SourceList sources={sources} userPlan="free" onDelete={onDelete} onRefresh={onRefresh} />, { wrapper: createWrapper() })
    
    const card = container.querySelector('[data-testid="source-card"]')!
    const refreshBtn = within(card as HTMLElement).getByLabelText('Kaynağı Yenile')
    expect(refreshBtn).toBeDisabled()
  })

  it('calls onRefresh when clicking refresh button', async () => {
    const user = userEvent.setup()
    const onDelete = vi.fn()
    const onRefresh = vi.fn()
    const sources = [
      { id: '1', source_type: 'url', source_url: 'https://a.com', status: 'completed', chunk_count: 1 },
    ] as any
    const { container } = render(<SourceList sources={sources} userPlan="pro" onDelete={onDelete} onRefresh={onRefresh} />, { wrapper: createWrapper() })
    
    const card = container.querySelector('[data-testid="source-card"]')!
    const refreshBtn = within(card as HTMLElement).getByLabelText('Kaynağı Yenile')
    await user.click(refreshBtn)
    
    expect(onRefresh).toHaveBeenCalledWith('1')
  })

  it('renders empty state when no sources', () => {
    const onDelete = vi.fn()
    const onRefresh = vi.fn()
    render(<SourceList sources={[]} userPlan="pro" onDelete={onDelete} onRefresh={onRefresh} />, { wrapper: createWrapper() })
    
    expect(screen.getByTestId('empty-state')).toBeInTheDocument()
    expect(screen.getByText('Henüz kaynak eklenmemiş')).toBeInTheDocument()
  })

  it('renders cards in a grid layout', () => {
    const onDelete = vi.fn()
    const onRefresh = vi.fn()
    const sources = [
      { id: '1', source_type: 'pdf', original_filename: 'file1.pdf', status: 'completed', chunk_count: 3 },
      { id: '2', source_type: 'pdf', original_filename: 'file2.pdf', status: 'completed', chunk_count: 5 },
    ] as any
    const { container } = render(<SourceList sources={sources} userPlan="pro" onDelete={onDelete} onRefresh={onRefresh} />, { wrapper: createWrapper() })
    
    const grid = container.querySelector('[data-testid="source-list"]')!
    expect(grid).toHaveClass('grid')
    expect(grid).toHaveClass('grid-cols-1')
    expect(grid).toHaveClass('md:grid-cols-2')
    expect(grid).toHaveClass('lg:grid-cols-3')
  })
})
