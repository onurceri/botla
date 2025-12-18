import { describe, it, expect, vi } from 'vitest'
import { renderHook } from '@testing-library/react'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { useAutoSave } from '../useAutoSave'
import { api } from '@/api/client'
import { ToastProvider } from '@/components/ui/toast'

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
  it('does not call API when id is "new"', async () => {
    const putSpy = vi.spyOn(api, 'put')

    renderHook(() => useAutoSave({ payload: { name: 'Test' }, debounceMs: 0 }), {
      wrapper: createWrapper('new'),
    })

    await new Promise((resolve) => setTimeout(resolve, 50))

    expect(putSpy).not.toHaveBeenCalled()
  })
})
