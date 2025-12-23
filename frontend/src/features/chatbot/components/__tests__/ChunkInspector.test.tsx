import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import ChunkInspector from '../ChunkInspector'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import * as sourceApi from '@/api/source'

// Mock the API
vi.mock('@/api/source', () => ({
  getSourceChunks: vi.fn(),
}))

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: false,
    },
  },
})

const Wrapper = ({ children }: { children: React.ReactNode }) => (
  <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
)

describe('ChunkInspector', () => {
  const mockChunks = [
    {
      id: 'chunk-1',
      score: 0.9,
      payload: {
        source_id: 'src-1',
        original_text: 'This is the first chunk of text content.',
        chunk_index: 0,
        created_at: new Date().toISOString(),
      },
    },
    {
      id: 'chunk-2',
      score: 0.8,
      payload: {
        source_id: 'src-1',
        original_text: 'This is the second chunk with some more details.',
        chunk_index: 1,
        created_at: new Date().toISOString(),
      },
    },
  ]

  beforeEach(() => {
    vi.resetAllMocks()
    queryClient.clear()
    // Force cleanup of document body for Dialog portals
    document.body.innerHTML = ''
  })

  it('renders nothing when closed', () => {
    render(
      <Wrapper>
        <ChunkInspector sourceId="src-1" open={false} onOpenChange={vi.fn()} />
      </Wrapper>,
    )
    expect(screen.queryByText('Kaynak Parçaları')).not.toBeInTheDocument()
  })

  it('fetches and displays chunks when open', async () => {
    vi.mocked(sourceApi.getSourceChunks).mockResolvedValue({
      chunks: mockChunks,
      next_cursor: null,
    })

    render(
      <Wrapper>
        <ChunkInspector sourceId="src-1" open={true} onOpenChange={vi.fn()} />
      </Wrapper>,
    )

    expect(screen.getByText('Kaynak Parçaları')).toBeInTheDocument()

    // Wait for chunks to load
    await waitFor(() => {
      expect(screen.getByText('This is the first chunk of text content.')).toBeInTheDocument()
      expect(
        screen.getByText('This is the second chunk with some more details.'),
      ).toBeInTheDocument()
    })

    expect(screen.getByText('#0')).toBeInTheDocument()
    expect(screen.getByText('#1')).toBeInTheDocument()
  })

  it('handles empty state', async () => {
    vi.mocked(sourceApi.getSourceChunks).mockResolvedValue({
      chunks: [],
      next_cursor: null,
    })

    render(
      <Wrapper>
        <ChunkInspector sourceId="src-1" open={true} onOpenChange={vi.fn()} />
      </Wrapper>,
    )

    await waitFor(() => {
      expect(screen.getByText('Bu kaynak için parça bulunamadı.')).toBeInTheDocument()
    })
  })

  it('filters chunks via search', async () => {
    vi.mocked(sourceApi.getSourceChunks).mockResolvedValue({
      chunks: mockChunks,
      next_cursor: null,
    })

    const user = userEvent.setup()
    render(
      <Wrapper>
        <ChunkInspector sourceId="src-1" open={true} onOpenChange={vi.fn()} />
      </Wrapper>,
    )

    await waitFor(() => {
      expect(screen.getByText('This is the first chunk of text content.')).toBeInTheDocument()
    })

    // Search for "second"
    const searchInputs = screen.getAllByPlaceholderText('Yüklenen parçalarda ara...')
    const searchInput = searchInputs[0]
    await user.type(searchInput, 'second')

    await waitFor(() => {
      expect(screen.queryByText('This is the first chunk of text content.')).not.toBeInTheDocument()

      // Check for highlighted text
      const highlighted = screen.getByText('second', { selector: 'span' })
      expect(highlighted).toBeInTheDocument()
      expect(highlighted.className).toContain('bg-yellow-200') // or whatever class is used

      // Check surrounding text presence (might be split)
      expect(screen.getByText(/This is the/, { selector: 'p' })).toBeInTheDocument()
      expect(
        screen.getByText(/chunk with some more details/, { selector: 'p' }),
      ).toBeInTheDocument()
    })
  })

  it('loads more pages when "Load More" is clicked', async () => {
    // Ensure we start fresh with unique ID to avoid cache
    const uniqueId = 'src-load-more'

    vi.mocked(sourceApi.getSourceChunks)
      .mockResolvedValueOnce({
        chunks: [mockChunks[0]],
        next_cursor: 'page-2',
      })
      .mockResolvedValueOnce({
        chunks: [mockChunks[1]],
        next_cursor: null,
      })

    const user = userEvent.setup()
    render(
      <Wrapper>
        <ChunkInspector sourceId={uniqueId} open={true} onOpenChange={vi.fn()} />
      </Wrapper>,
    )

    // First page loaded
    await waitFor(() => {
      expect(screen.getByText('This is the first chunk of text content.')).toBeInTheDocument()
    })

    // Use waitFor to ensure stable state
    await waitFor(() => {
      expect(
        screen.queryByText('This is the second chunk with some more details.'),
      ).not.toBeInTheDocument()
    })

    // Click load more
    const loadMoreBtn = screen.getByRole('button', { name: /Daha Fazla Yükle/i })
    await user.click(loadMoreBtn)

    // Second page loaded
    await waitFor(() => {
      expect(
        screen.getByText('This is the second chunk with some more details.'),
      ).toBeInTheDocument()
    })
  })

  it('handles API error', async () => {
    vi.mocked(sourceApi.getSourceChunks).mockRejectedValue(new Error('Failed to fetch'))

    render(
      <Wrapper>
        <ChunkInspector sourceId="src-1" open={true} onOpenChange={vi.fn()} />
      </Wrapper>,
    )

    await waitFor(() => {
      expect(screen.getByText('Parçalar yüklenemedi.')).toBeInTheDocument()
    })
  })
})
