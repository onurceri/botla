import { describe, it, expect, vi } from 'vitest'
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { ToastProvider } from '@/components/ui/toast'
import ChatbotDetailPage from '../ChatbotDetailPage'
import * as sourceApi from '@/api/source'

vi.mock('@/features/chatbot/hooks/useSourceOps', () => {
  const refreshSources = vi.fn()
  const pollStatus = vi.fn()
  return {
    useSourceOps: () => ({
      sources: [],
      refreshSources,
      pollStatus,
      handleDeleteSource: vi.fn(),
    })
  }
})

describe('ChatbotDetailPage sources upload flow', () => {
  it('uploads URL source and calls refresh and poll', async () => {
    vi.spyOn(sourceApi, 'uploadURLSource').mockResolvedValueOnce({ id: 's1' } as any)
    vi.spyOn((await import('@/api/client')).api, 'get').mockImplementation((url: string) => {
      if (url.includes('/api/v1/me')) return Promise.resolve({ data: { subscription_plan: 'pro' } } as any)
      if (url.includes('/api/v1/chatbots/abc')) return Promise.resolve({ data: { id: 'abc', name: 'Bot' } } as any)
      if (url.includes('/api/v1/chatbots/abc/sources')) return Promise.resolve({ data: [] } as any)
      return Promise.resolve({ data: {} } as any)
    })
    const utils = render(
      <ToastProvider>
        <MemoryRouter initialEntries={["/chatbots/abc"]}>
          <Routes>
            <Route path="/chatbots/:id" element={<ChatbotDetailPage />} />
          </Routes>
        </MemoryRouter>
      </ToastProvider>
    )
    const view = within(utils.container)
    const sourcesTab = await view.findByRole('tab', { name: /Veri Kaynakları/ })
    await userEvent.click(sourcesTab)
    const urlBtns = view.getAllByText('Web Sitesi')
    await userEvent.click(urlBtns[urlBtns.length - 1])
    const input = await view.findByPlaceholderText('https://example.com')
    await userEvent.type(input, 'https://example.com')
    const addBtn = view.getByRole('button', { name: 'Ekle' })
    await userEvent.click(addBtn)
    const mod = await import('@/features/chatbot/hooks/useSourceOps') as any
    expect(mod.useSourceOps().refreshSources).toHaveBeenCalled()
    expect(mod.useSourceOps().pollStatus).toHaveBeenCalledWith('s1')
    expect(sourceApi.uploadURLSource).toHaveBeenCalledWith('abc', 'https://example.com')
  })

  it('uploads PDF source and calls refresh and poll', async () => {
    vi.clearAllMocks()
    vi.spyOn(sourceApi, 'uploadPDFSource').mockResolvedValueOnce({ id: 'p1' } as any)
    vi.spyOn((await import('@/api/client')).api, 'get').mockImplementation((url: string) => {
      if (url.includes('/api/v1/me')) return Promise.resolve({ data: { subscription_plan: 'pro' } } as any)
      if (url.includes('/api/v1/chatbots/abc')) return Promise.resolve({ data: { id: 'abc', name: 'Bot' } } as any)
      if (url.includes('/api/v1/chatbots/abc/sources')) return Promise.resolve({ data: [] } as any)
      return Promise.resolve({ data: {} } as any)
    })
    const utils = render(
      <ToastProvider>
        <MemoryRouter initialEntries={["/chatbots/abc"]}>
          <Routes>
            <Route path="/chatbots/:id" element={<ChatbotDetailPage />} />
          </Routes>
        </MemoryRouter>
      </ToastProvider>
    )
    const view = within(utils.container)
    const sourcesTab = await view.findByRole('tab', { name: /Veri Kaynakları/ })
    await userEvent.click(sourcesTab)
    const pdfBtns = view.getAllByText('PDF Yükle')
    await userEvent.click(pdfBtns[pdfBtns.length - 1])
    const fileInput = utils.container.querySelector('#pdf-upload') as HTMLInputElement
    const file = new File(['%PDF-1.4'], 'doc.pdf', { type: 'application/pdf' })
    Object.defineProperty(fileInput, 'files', { value: [file] })
    const { fireEvent } = await import('@testing-library/react')
    fireEvent.change(fileInput)
    expect(sourceApi.uploadPDFSource).toHaveBeenCalledWith('abc', file)
    expect(await screen.findByText('PDF başarıyla yüklendi. İşleniyor...')).toBeInTheDocument()
  })
})
