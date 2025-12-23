import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { ToastProvider } from '@/components/ui/toast'
import { useSourceOps } from '../useSourceOps'
import * as sourceApi from '@/api/source'

describe('useSourceOps', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  it('pollStatus clears on error (catch branch)', async () => {
    const getStatus = vi
      .spyOn(sourceApi, 'getSourceStatus')
      .mockRejectedValueOnce(new Error('fail'))

    const wrapper = ({ children }: any) => <ToastProvider>{children}</ToastProvider>
    const { result } = renderHook(() => useSourceOps('bot-1', false), { wrapper })

    act(() => {
      result.current.pollStatus('s1')
    })

    await act(async () => {
      vi.advanceTimersByTime(1000)
    })

    expect(getStatus).toHaveBeenCalled()
  })

  it('handleDeleteSource shows error toast on failure', async () => {
    vi.spyOn(window, 'confirm').mockReturnValue(true as any)
    vi.spyOn(sourceApi, 'deleteSource').mockRejectedValueOnce(new Error('fail'))

    const wrapper = ({ children }: any) => <ToastProvider>{children}</ToastProvider>
    const { result } = renderHook(() => useSourceOps('bot-1', false), { wrapper })

    await act(async () => {
      await result.current.handleDeleteSource('s1')
    })

    // Toast should render error message in the DOM
    expect(document.body.textContent).toContain('Kaynak silinirken bir hata oluştu.')
  })
})
