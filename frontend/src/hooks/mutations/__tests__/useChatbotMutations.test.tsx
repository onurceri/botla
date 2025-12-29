import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useRegenerateSuggestions } from '../useChatbotMutations'
import { api } from '@/api/client'

const { mockGetSuggestionJobStatus } = vi.hoisted(() => {
  return { mockGetSuggestionJobStatus: vi.fn() }
})

vi.mock('@/api/chatbot', async () => {
  const actual = await vi.importActual('@/api/chatbot')
  return {
    ...actual,
    getSuggestionJobStatus: mockGetSuggestionJobStatus,
  }
})

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

describe('useRegenerateSuggestions', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.restoreAllMocks()
    mockGetSuggestionJobStatus.mockReset()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('triggers regeneration and polls until completed', async () => {
    const postSpy = vi.spyOn(api, 'post').mockResolvedValue({
      data: { job_id: 'job-123' },
    })

    mockGetSuggestionJobStatus
      .mockResolvedValueOnce({
        job_id: 'job-123',
        status: 'pending',
        suggested_questions: [],
      })
      .mockResolvedValueOnce({
        job_id: 'job-123',
        status: 'processing',
        suggested_questions: [],
      })
      .mockResolvedValueOnce({
        job_id: 'job-123',
        status: 'completed',
        suggested_questions: [['Question 1'], ['Question 2']],
      })

    const { result } = renderHook(() => useRegenerateSuggestions('chatbot-1'), {
      wrapper: createWrapper(),
    })

    expect(result.current.mutateAsync).toBeDefined()

    await result.current.mutateAsync()

    expect(postSpy).toHaveBeenCalledWith(
      '/api/v1/chatbots/chatbot-1/suggestions/regenerate',
    )

    await waitFor(() => {
      expect(mockGetSuggestionJobStatus).toHaveBeenCalledTimes(3)
    })

    expect(result.current.isSuccess).toBe(true)
  })

  it('throws error when job fails', async () => {
    vi.spyOn(api, 'post').mockResolvedValue({
      data: { job_id: 'job-456' },
    })

    mockGetSuggestionJobStatus
      .mockResolvedValueOnce({
        job_id: 'job-456',
        status: 'pending',
        suggested_questions: [],
      })
      .mockResolvedValueOnce({
        job_id: 'job-456',
        status: 'failed',
        suggested_questions: [],
        error_message: 'Processing failed due to invalid sources',
      })

    const { result } = renderHook(() => useRegenerateSuggestions('chatbot-2'), {
      wrapper: createWrapper(),
    })

    try {
      await result.current.mutateAsync()
    } catch {}

    await waitFor(() => {
      expect(mockGetSuggestionJobStatus).toHaveBeenCalledTimes(2)
    })

    expect(result.current.isError).toBe(true)
  })

  it('handles API error during regeneration request', async () => {
    vi.spyOn(api, 'post').mockRejectedValue(new Error('Network error'))

    const { result } = renderHook(() => useRegenerateSuggestions('chatbot-4'), {
      wrapper: createWrapper(),
    })

    let errorThrown = false
    try {
      await result.current.mutateAsync()
    } catch {
      errorThrown = true
    }

    expect(errorThrown).toBe(true)
    await waitFor(() => {
      expect(result.current.isError).toBe(true)
    })
  })

  it('handles API error during status polling', async () => {
    vi.spyOn(api, 'post').mockResolvedValue({
      data: { job_id: 'job-101' },
    })

    mockGetSuggestionJobStatus.mockRejectedValue(
      new Error('Failed to fetch status'),
    )

    const { result } = renderHook(() => useRegenerateSuggestions('chatbot-5'), {
      wrapper: createWrapper(),
    })

    let errorThrown = false
    try {
      await result.current.mutateAsync()
    } catch {
      errorThrown = true
    }

    expect(errorThrown).toBe(true)
    await waitFor(() => {
      expect(result.current.isError).toBe(true)
    })
  })

  it('polls and reaches max attempts before timing out', async () => {
    vi.spyOn(api, 'post').mockResolvedValue({
      data: { job_id: 'job-789' },
    })

    let callCount = 0
    const maxAttempts = 3
    mockGetSuggestionJobStatus.mockImplementation(() => {
      callCount++
      if (callCount >= maxAttempts) {
        return Promise.reject(new Error('Suggestion regeneration timed out'))
      }
      return Promise.resolve({
        job_id: 'job-789',
        status: 'processing',
        suggested_questions: [],
      })
    })

    const { result } = renderHook(() => useRegenerateSuggestions('chatbot-3'), {
      wrapper: createWrapper(),
    })

    await expect(result.current.mutateAsync()).rejects.toThrow(
      'Suggestion regeneration timed out',
    )

    expect(callCount).toBe(maxAttempts)
  })
})
