import { describe, it, expect, vi } from 'vitest'
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import SourceCard, { Source } from '../SourceCard'

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

const mockSource = (overrides?: Partial<Source>): Source => ({
  id: '1',
  source_type: 'pdf',
  original_filename: 'test-document.pdf',
  status: 'completed',
  chunk_count: 10,
  ...overrides,
})

describe('SourceCard', () => {
  it('renders source card with correct source type icon for PDF', () => {
    const onDelete = vi.fn()
    const onRefresh = vi.fn()
    const { container } = render(
      <SourceCard source={mockSource()} userPlan="pro" onDelete={onDelete} onRefresh={onRefresh} />,
      { wrapper: createWrapper() },
    )

    const card = container.querySelector('[data-testid="source-card"]')!
    expect(within(card as HTMLElement).getByText('PDF Doküman')).toBeInTheDocument()
    expect(within(card as HTMLElement).getByText('test-document.pdf')).toBeInTheDocument()
    expect(within(card as HTMLElement).getByText('10')).toBeInTheDocument()
  })

  it('renders source card with correct source type icon for URL', () => {
    const onDelete = vi.fn()
    const onRefresh = vi.fn()
    const { container } = render(
      <SourceCard
        source={mockSource({
          source_type: 'url',
          source_url: 'https://example.com',
          original_filename: undefined,
        })}
        userPlan="pro"
        onDelete={onDelete}
        onRefresh={onRefresh}
      />,
      { wrapper: createWrapper() },
    )

    const card = container.querySelector('[data-testid="source-card"]')!
    expect(within(card as HTMLElement).getByText('Web Sayfası')).toBeInTheDocument()
    expect(within(card as HTMLElement).getByText('https://example.com')).toBeInTheDocument()
  })

  it('renders source card with correct source type icon for text', () => {
    const onDelete = vi.fn()
    const onRefresh = vi.fn()
    const { container } = render(
      <SourceCard
        source={mockSource({ source_type: 'text', original_filename: 'My Notes' })}
        userPlan="pro"
        onDelete={onDelete}
        onRefresh={onRefresh}
      />,
      { wrapper: createWrapper() },
    )

    const card = container.querySelector('[data-testid="source-card"]')!
    expect(within(card as HTMLElement).getByText('Metin')).toBeInTheDocument()
    expect(within(card as HTMLElement).getByText('My Notes')).toBeInTheDocument()
  })

  it('renders completed status badge correctly', () => {
    const onDelete = vi.fn()
    const onRefresh = vi.fn()
    const { container } = render(
      <SourceCard
        source={mockSource({ status: 'completed' })}
        userPlan="pro"
        onDelete={onDelete}
        onRefresh={onRefresh}
      />,
      { wrapper: createWrapper() },
    )

    const card = container.querySelector('[data-testid="source-card"]')!
    expect(within(card as HTMLElement).getByText('Tamamlandı')).toBeInTheDocument()
  })

  it('renders processing status badge correctly', () => {
    const onDelete = vi.fn()
    const onRefresh = vi.fn()
    const { container } = render(
      <SourceCard
        source={mockSource({ status: 'processing' })}
        userPlan="pro"
        onDelete={onDelete}
        onRefresh={onRefresh}
      />,
      { wrapper: createWrapper() },
    )

    const card = container.querySelector('[data-testid="source-card"]')!
    expect(within(card as HTMLElement).getByText('İşleniyor')).toBeInTheDocument()
  })

  it('renders failed status badge correctly', () => {
    const onDelete = vi.fn()
    const onRefresh = vi.fn()
    const { container } = render(
      <SourceCard
        source={mockSource({ status: 'failed' })}
        userPlan="pro"
        onDelete={onDelete}
        onRefresh={onRefresh}
      />,
      { wrapper: createWrapper() },
    )

    const card = container.querySelector('[data-testid="source-card"]')!
    expect(within(card as HTMLElement).getByText('Başarısız')).toBeInTheDocument()
  })

  it('renders queued status badge correctly', () => {
    const onDelete = vi.fn()
    const onRefresh = vi.fn()
    const { container } = render(
      <SourceCard
        source={mockSource({ status: 'queued' })}
        userPlan="pro"
        onDelete={onDelete}
        onRefresh={onRefresh}
      />,
      { wrapper: createWrapper() },
    )

    const card = container.querySelector('[data-testid="source-card"]')!
    expect(within(card as HTMLElement).getByText('Beklemede')).toBeInTheDocument()
  })

  it('calls onDelete when delete is confirmed', async () => {
    const user = userEvent.setup()
    const onDelete = vi.fn()
    const onRefresh = vi.fn()
    const { container } = render(
      <SourceCard source={mockSource()} userPlan="pro" onDelete={onDelete} onRefresh={onRefresh} />,
      { wrapper: createWrapper() },
    )

    const card = container.querySelector('[data-testid="source-card"]')!
    const deleteBtn = within(card as HTMLElement).getByLabelText('Kaynağı Sil')
    await user.click(deleteBtn)

    // Should show confirmation dialog
    expect(screen.getByText('Bu kaynağı silmek istediğinizden emin misiniz?')).toBeInTheDocument()

    // Click confirm
    const confirmBtn = screen.getByRole('button', { name: 'Sil' })
    await user.click(confirmBtn)

    expect(onDelete).toHaveBeenCalledWith('1')
  })

  it('does not call onDelete when delete is cancelled', async () => {
    const user = userEvent.setup()
    const onDelete = vi.fn()
    const onRefresh = vi.fn()
    const { container } = render(
      <SourceCard source={mockSource()} userPlan="pro" onDelete={onDelete} onRefresh={onRefresh} />,
      { wrapper: createWrapper() },
    )

    const card = container.querySelector('[data-testid="source-card"]')!
    const deleteBtn = within(card as HTMLElement).getByLabelText('Kaynağı Sil')
    await user.click(deleteBtn)

    // Click cancel
    const cancelBtn = screen.getByRole('button', { name: 'İptal' })
    await user.click(cancelBtn)

    expect(onDelete).not.toHaveBeenCalled()
  })

  it('shows refresh button only for URL sources', () => {
    const onDelete = vi.fn()
    const onRefresh = vi.fn()

    // PDF should NOT have refresh button
    const { container: pdfContainer } = render(
      <SourceCard
        source={mockSource({ source_type: 'pdf' })}
        userPlan="pro"
        onDelete={onDelete}
        onRefresh={onRefresh}
      />,
      { wrapper: createWrapper() },
    )
    const pdfCard = pdfContainer.querySelector('[data-testid="source-card"]')!
    expect(within(pdfCard as HTMLElement).queryByLabelText('Kaynağı Yenile')).toBeNull()

    // URL should have refresh button
    const { container: urlContainer } = render(
      <SourceCard
        source={mockSource({ source_type: 'url', source_url: 'https://example.com' })}
        userPlan="pro"
        onDelete={onDelete}
        onRefresh={onRefresh}
      />,
      { wrapper: createWrapper() },
    )
    const urlCard = urlContainer.querySelector('[data-testid="source-card"]')!
    expect(within(urlCard as HTMLElement).getByLabelText('Kaynağı Yenile')).toBeInTheDocument()
  })

  it('disables refresh button for free plan', () => {
    const onDelete = vi.fn()
    const onRefresh = vi.fn()
    const { container } = render(
      <SourceCard
        source={mockSource({ source_type: 'url', source_url: 'https://example.com' })}
        userPlan="free"
        onDelete={onDelete}
        onRefresh={onRefresh}
      />,
      { wrapper: createWrapper() },
    )

    const card = container.querySelector('[data-testid="source-card"]')!
    const refreshBtn = within(card as HTMLElement).getByLabelText('Kaynağı Yenile')
    expect(refreshBtn).toBeDisabled()
  })

  it('calls onRefresh when clicking refresh button', async () => {
    const user = userEvent.setup()
    const onDelete = vi.fn()
    const onRefresh = vi.fn()
    const { container } = render(
      <SourceCard
        source={mockSource({ source_type: 'url', source_url: 'https://example.com' })}
        userPlan="pro"
        onDelete={onDelete}
        onRefresh={onRefresh}
      />,
      { wrapper: createWrapper() },
    )

    const card = container.querySelector('[data-testid="source-card"]')!
    const refreshBtn = within(card as HTMLElement).getByLabelText('Kaynağı Yenile')
    await user.click(refreshBtn)

    expect(onRefresh).toHaveBeenCalledWith('1')
  })

  it('shows capability summary dialog when clicked', async () => {
    const user = userEvent.setup()
    const onDelete = vi.fn()
    const onRefresh = vi.fn()
    const { container } = render(
      <SourceCard
        source={mockSource({ capability_summary: 'This source contains information about...' })}
        userPlan="pro"
        onDelete={onDelete}
        onRefresh={onRefresh}
      />,
      { wrapper: createWrapper() },
    )

    const card = container.querySelector('[data-testid="source-card"]')!
    const summaryBtn = within(card as HTMLElement).getByText('Özet')
    await user.click(summaryBtn)

    expect(screen.getByText('This source contains information about...')).toBeInTheDocument()
  })

  it('does not show summary button when no capability_summary', () => {
    const onDelete = vi.fn()
    const onRefresh = vi.fn()
    const { container } = render(
      <SourceCard
        source={mockSource({ capability_summary: undefined })}
        userPlan="pro"
        onDelete={onDelete}
        onRefresh={onRefresh}
      />,
      { wrapper: createWrapper() },
    )

    const card = container.querySelector('[data-testid="source-card"]')!
    expect(within(card as HTMLElement).queryByText('Özet')).toBeNull()
  })

  it('displays chunk count correctly', () => {
    const onDelete = vi.fn()
    const onRefresh = vi.fn()
    const { container } = render(
      <SourceCard
        source={mockSource({ chunk_count: 42 })}
        userPlan="pro"
        onDelete={onDelete}
        onRefresh={onRefresh}
      />,
      { wrapper: createWrapper() },
    )

    const card = container.querySelector('[data-testid="source-card"]')!
    expect(within(card as HTMLElement).getByText(/42/)).toBeInTheDocument()
    expect(within(card as HTMLElement).getByText(/parça/)).toBeInTheDocument()
  })

  it('applies processing ring style when status is processing', () => {
    const onDelete = vi.fn()
    const onRefresh = vi.fn()
    const { container } = render(
      <SourceCard
        source={mockSource({ status: 'processing' })}
        userPlan="pro"
        onDelete={onDelete}
        onRefresh={onRefresh}
      />,
      { wrapper: createWrapper() },
    )

    const card = container.querySelector('[data-testid="source-card"]')
    expect(card).toHaveClass('ring-2')
  })
})
