import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { useAutoSave } from '../useAutoSave'
import { api } from '@/api/client'
import { ToastProvider } from '@/components/ui/toast'

vi.mock('@/api/client', () => ({
  api: {
    put: vi.fn().mockResolvedValue({ data: {} }),
  },
}))

const createWrapper = (id: string) => {
  return ({ children }: { children: React.ReactNode }) => (
    <ToastProvider>
      <MemoryRouter initialEntries={[`/chatbots/${id}`]}>
        <Routes>
          <Route path="/chatbots/:id" element={children} />
        </Routes>
      </MemoryRouter>
    </ToastProvider>
  )
}

describe('useAutoSave', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('does not call API when id is "new"', async () => {
    renderHook(() => useAutoSave({ payload: { name: 'Test' }, debounceMs: 100 }), {
      wrapper: createWrapper('new'),
    })

    await act(async () => {
      vi.advanceTimersByTime(200)
    })

    expect(api.put).not.toHaveBeenCalled()
  })

  it('does not call API on initial mount', async () => {
    renderHook(() => useAutoSave({ payload: { name: 'Test' }, debounceMs: 100 }), {
      wrapper: createWrapper('123'),
    })

    await act(async () => {
      vi.advanceTimersByTime(200)
    })

    expect(api.put).not.toHaveBeenCalled()
  })

  it('debounces API calls - only one call after delay', async () => {
    const { rerender } = renderHook(({ payload }) => useAutoSave({ payload, debounceMs: 100 }), {
      wrapper: createWrapper('123'),
      initialProps: { payload: { name: 'Initial' } },
    })

    // First change
    rerender({ payload: { name: 'Changed1' } })

    await act(async () => {
      vi.advanceTimersByTime(50) // Not enough time
    })
    expect(api.put).not.toHaveBeenCalled()

    // Second change before debounce completes
    rerender({ payload: { name: 'Changed2' } })

    await act(async () => {
      vi.advanceTimersByTime(50) // Still not enough
    })
    expect(api.put).not.toHaveBeenCalled()

    // Third change
    rerender({ payload: { name: 'Final' } })

    await act(async () => {
      vi.advanceTimersByTime(150) // Now debounce completes
    })

    // Should only have one API call with the final payload
    expect(api.put).toHaveBeenCalledTimes(1)
    expect(api.put).toHaveBeenCalledWith(
      '/api/v1/chatbots/123',
      { name: 'Final' },
      expect.any(Object),
    )
  })

  it('saves with latest payload even if callback was created with old payload', async () => {
    const { rerender } = renderHook(({ payload }) => useAutoSave({ payload, debounceMs: 100 }), {
      wrapper: createWrapper('456'),
      initialProps: { payload: { message: 'A' } },
    })

    // Simulate rapid typing: A -> AB -> ABC -> ABCD
    rerender({ payload: { message: 'AB' } })
    await act(async () => {
      vi.advanceTimersByTime(20)
    })

    rerender({ payload: { message: 'ABC' } })
    await act(async () => {
      vi.advanceTimersByTime(20)
    })

    rerender({ payload: { message: 'ABCD' } })
    await act(async () => {
      vi.advanceTimersByTime(20)
    })

    rerender({ payload: { message: 'ABCDE' } })

    // No API call yet
    expect(api.put).not.toHaveBeenCalled()

    // Wait for debounce to complete
    await act(async () => {
      vi.advanceTimersByTime(150)
    })

    // Should save with final value
    expect(api.put).toHaveBeenCalledTimes(1)
    expect(api.put).toHaveBeenCalledWith(
      '/api/v1/chatbots/456',
      { message: 'ABCDE' },
      expect.any(Object),
    )
  })

  it('does not make duplicate API calls for same payload value', async () => {
    const { rerender } = renderHook(({ payload }) => useAutoSave({ payload, debounceMs: 100 }), {
      wrapper: createWrapper('789'),
      initialProps: { payload: { name: 'Test' } },
    })

    // First change
    rerender({ payload: { name: 'Changed' } })
    await act(async () => {
      vi.advanceTimersByTime(150)
    })
    expect(api.put).toHaveBeenCalledTimes(1)

    // Same payload value again (should not trigger new save)
    rerender({ payload: { name: 'Changed' } })
    await act(async () => {
      vi.advanceTimersByTime(150)
    })
    expect(api.put).toHaveBeenCalledTimes(1) // Still 1, not 2
  })

  it('respects enabled flag', async () => {
    const { rerender } = renderHook(
      ({ payload, enabled }) => useAutoSave({ payload, debounceMs: 100, enabled }),
      {
        wrapper: createWrapper('xyz'),
        initialProps: { payload: { name: 'Initial' }, enabled: true },
      },
    )

    // Disable and try to change - should not save
    rerender({ payload: { name: 'Changed' }, enabled: false })
    await act(async () => {
      vi.advanceTimersByTime(150)
    })

    expect(api.put).not.toHaveBeenCalled()
  })
})
