import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import SourcesTab from '../SourcesTab'
import { ChatbotProvider } from '../../context/ChatbotContext'

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
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={['/chatbots/abc/sources']}>
        <ChatbotProvider chatbotId="abc">
          {children}
        </ChatbotProvider>
      </MemoryRouter>
    </QueryClientProvider>
  )
}

const mockSources = [
  {
    id: '1',
    source_type: 'pdf',
    original_filename: 'file.pdf',
    status: 'completed',
    chunk_count: 3,
  },
  {
    id: '2',
    source_type: 'url',
    source_url: 'https://example.com',
    status: 'processing',
    chunk_count: 0,
  },
]

vi.mock('../../hooks/useSourceOps', () => ({
  useSourceOps: () => ({
    sources: mockSources,
    refreshSources: vi.fn(),
    pollStatus: vi.fn(),
    handleDeleteSource: vi.fn(),
    handleRefreshSource: vi.fn(),
    refreshingId: undefined,
  }),
}))

vi.mock('@/hooks/mutations/useChatbotMutations', () => ({
  useUploadSource: () => ({
    uploadPDF: { mutateAsync: vi.fn() },
    uploadURL: { mutateAsync: vi.fn() },
    uploadText: { mutateAsync: vi.fn() },
  }),
  useUpdateScrapingConfig: () => ({ mutateAsync: vi.fn() }),
  useUpdateRefresh: () => ({ mutateAsync: vi.fn() }),
}))

describe('SourcesTab', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders stats and sources', async () => {
    render(<SourcesTab />, { wrapper: createWrapper() })

    expect(screen.getByText('Bilgi Bankası')).toBeInTheDocument()
    expect(screen.getByText('2')).toBeInTheDocument() // Total count
    expect(screen.getByText('file.pdf')).toBeInTheDocument()
    expect(screen.getByText('https://example.com')).toBeInTheDocument()
  })

  it('filters sources by search query', async () => {
    const user = userEvent.setup()
    render(<SourcesTab />, { wrapper: createWrapper() })

    const searchInput = screen.getByPlaceholderText('Kaynak ara...')
    await user.type(searchInput, 'pdf')

    expect(screen.getByText('file.pdf')).toBeInTheDocument()
    expect(screen.queryByText('https://example.com')).not.toBeInTheDocument()
  })

  it('filters sources by type', async () => {
    const user = userEvent.setup()
    render(<SourcesTab />, { wrapper: createWrapper() })

    const urlFilter = screen.getByRole('button', { name: /URL/ })
    await user.click(urlFilter)

    expect(screen.queryByText('file.pdf')).not.toBeInTheDocument()
    expect(screen.getByText('https://example.com')).toBeInTheDocument()
  })

  it('collapses and expands upload section', async () => {
    const user = userEvent.setup()
    render(<SourcesTab />, { wrapper: createWrapper() })

    // Initially expanded (SourceUploader is visible)
    expect(screen.getByText('PDF Yükle')).toBeInTheDocument()

    const toggleBtn = screen.getByRole('button', { name: /Yeni Kaynak Ekle/ })
    await user.click(toggleBtn)

    // Should collapse
    expect(screen.queryByText('PDF Yükle')).not.toBeInTheDocument()
  })
})
