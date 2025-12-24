import { describe, it, expect, vi } from 'vitest'
import { QueryWrapper } from '@/test-utils'
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { ToastProvider } from '@/components/ui/toast'
import ChatbotDetailPage from '../ChatbotDetailPage'
import SourcesTab from '@/features/chatbot/pages/tabs/SourcesTab'
import { api } from '@/api/client'

vi.mock('@/features/organization/context/OrganizationContext', () => ({
  useOrganization: () => ({
    currentWorkspace: { id: 'ws-1' },
    isLoading: false,
  }),
  OrganizationProvider: ({ children }: any) => children,
}))

vi.mock('@/api/source', () => ({
  listSources: vi.fn().mockResolvedValue([
    {
      id: 'src1',
      source_type: 'text',
      original_filename: 'inline.txt',
      status: 'completed',
      chunk_count: 3,
    },
  ]),
  deleteSource: vi.fn().mockResolvedValue(undefined),
  uploadPDFSource: vi.fn(),
  uploadTextSource: vi.fn(),
  uploadURLSource: vi.fn(),
  getSourceStatus: vi.fn(),
  listPendingURLs: vi.fn().mockResolvedValue([]),
}))

describe('ChatbotDetailPage sources', () => {
  it('deletes a source and refreshes list', async () => {
    const user = userEvent.setup()
    vi.spyOn(api, 'get').mockImplementation((url: string) => {
      if (url.includes('/api/v1/me/plan')) {
        return Promise.resolve({
          data: {
            code: 'pro',
            features: {
              files: {
                max_files_per_bot: 999,
                max_size_mb: 50,
                ocr_enabled: true,
                max_files_total: 999,
                total_storage_mb: 1000,
              },
              scraping: { max_urls_per_bot: 999 },
            },
            available_models: [],
          },
        } as any)
      }
      if (url.includes('/api/v1/chatbots/123')) {
        return Promise.resolve({ data: { id: '123', name: 'Bot' } } as any)
      }
      return Promise.resolve({ data: {} } as any)
    })
    vi.spyOn(window, 'confirm').mockReturnValue(true)

    render(
      <QueryWrapper>
        <ToastProvider>
          <MemoryRouter initialEntries={['/chatbots/123']}>
            <Routes>
              <Route path="/chatbots/:id" element={<ChatbotDetailPage />}>
                <Route path="sources" element={<SourcesTab />} />
              </Route>
            </Routes>
          </MemoryRouter>
        </ToastProvider>
      </QueryWrapper>,
    )

    const srcTabs = await screen.findAllByRole('link', { name: /Kaynaklar/i })
    await user.click(srcTabs[srcTabs.length - 1])
    const sourceHeads = await screen.findAllByText('inline.txt')
    const sourceCard = sourceHeads[0].closest('[data-testid="source-card"]')!
    const delBtn = within(sourceCard as HTMLElement).getByRole('button', { name: /Kaynağı Sil/i })
    await user.click(delBtn)

    // Confirm delete in AlertDialog
    const confirmBtn = await screen.findByRole('button', { name: 'Sil' })
    await user.click(confirmBtn)

    const { deleteSource } = await import('@/api/source')
    expect(deleteSource).toHaveBeenCalledWith('src1')
  })

  it('polls source status until terminal state and refreshes list', async () => {
    const user = userEvent.setup()
    vi.spyOn(globalThis, 'setInterval').mockImplementation((cb: any) => {
      Promise.resolve().then(cb)
      Promise.resolve().then(cb)
      Promise.resolve().then(cb)
      return 1 as any
    })
    vi.spyOn(globalThis, 'clearInterval').mockImplementation(() => {})
    vi.spyOn(api, 'get').mockImplementation((url: string) => {
      if (url.includes('/api/v1/me/plan')) {
        return Promise.resolve({
          data: {
            code: 'pro',
            features: {
              files: {
                max_files_per_bot: 999,
                max_size_mb: 50,
                ocr_enabled: true,
                max_files_total: 999,
                total_storage_mb: 1000,
              },
              scraping: { max_urls_per_bot: 999 },
            },
            available_models: [],
          },
        } as any)
      }
      if (url.includes('/api/v1/chatbots/123')) {
        return Promise.resolve({ data: { id: '123', name: 'Bot' } } as any)
      }
      return Promise.resolve({ data: {} } as any)
    })
    const { uploadTextSource, getSourceStatus, listSources } = await import('@/api/source')
    ;(uploadTextSource as any).mockResolvedValueOnce({ id: 'sid1' })
    ;(getSourceStatus as any)
      .mockResolvedValueOnce({ status: 'pending' })
      .mockResolvedValueOnce({ status: 'processing' })
    ;(getSourceStatus as any).mockResolvedValue({ status: 'completed' })

    render(
      <QueryWrapper>
        <ToastProvider>
          <MemoryRouter initialEntries={['/chatbots/123']}>
            <Routes>
              <Route path="/chatbots/:id" element={<ChatbotDetailPage />}>
                <Route path="sources" element={<SourcesTab />} />
              </Route>
            </Routes>
          </MemoryRouter>
        </ToastProvider>
      </QueryWrapper>,
    )

    const srcTabs2 = await screen.findAllByRole('link', { name: /Kaynaklar/i })
    await user.click(srcTabs2[srcTabs2.length - 1])
    const pdfBtn = await screen.findByRole('button', { name: 'PDF Yükle' })
    const uploaderButtonRow = pdfBtn.parentElement as HTMLElement
    await user.click(within(uploaderButtonRow).getByRole('button', { name: 'Metin' }))
    const textarea = await screen.findByPlaceholderText(/Metin içeriğinizi buraya yapıştırın/i)
    await user.type(textarea, 'hello')
    const buttons = screen.getAllByRole('button', { name: 'Ekle' })
    await user.click(buttons[buttons.length - 1])

    await Promise.resolve()
    await Promise.resolve()
    await Promise.resolve()

    expect(getSourceStatus).toHaveBeenCalled()
    const calls = (listSources as any).mock.calls.length
    expect(calls).toBeGreaterThanOrEqual(2)
    const cells = await screen.findAllByText('inline.txt')
    expect(cells.length).toBeGreaterThan(0)
  })
})
