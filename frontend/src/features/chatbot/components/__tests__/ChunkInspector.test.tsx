import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import ChunkInspector from '../ChunkInspector'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { getSourceChunks } from '@/api/source'

// Mock API
vi.mock('@/api/source', () => ({
  getSourceChunks: vi.fn()
}))

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

describe('ChunkInspector', () => {
  it('renders loading state initially', () => {
    // Mock infinite loading
    (getSourceChunks as any).mockImplementation(() => new Promise(() => {}))
    
    render(<ChunkInspector sourceId="1" open={true} onOpenChange={vi.fn()} />, {
      wrapper: createWrapper(),
    })
    
    // Radix Dialog might render 'Parçalar yükleniyor...' inside
    // Note: Radix UI Dialog requires 'open' to be true to render content
    // And it might render in a Portal. screen.getByText searches document.body.
    
    expect(screen.getByText('Parçalar yükleniyor...')).toBeInTheDocument()
  })

  it('renders chunks when loaded', async () => {
    (getSourceChunks as any).mockResolvedValue({
      chunks: [
        { id: 'c1', score: 0.9, payload: { original_text: 'Test Chunk Content', chunk_index: 0, source_id: '1', created_at: '' } }
      ],
      next_cursor: null
    })

    render(<ChunkInspector sourceId="1" open={true} onOpenChange={vi.fn()} />, {
      wrapper: createWrapper(),
    })

    expect(await screen.findByText('Test Chunk Content')).toBeInTheDocument()
  })
})
