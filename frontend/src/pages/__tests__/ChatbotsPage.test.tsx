import { describe, it, expect, vi, beforeEach } from 'vitest'
import { QueryWrapper } from '@/test-utils'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { ToastProvider } from '@/components/ui/toast'
import ChatbotsPage from '../ChatbotsPage'
import { api } from '@/api/client'

describe('ChatbotsPage', () => {
  beforeEach(() => {
    Object.defineProperty(window, 'localStorage', {
      value: {
        getItem: vi.fn(),
        setItem: vi.fn(),
        removeItem: vi.fn(),
      },
      writable: true,
    })
  })

  // Mock Organization Context
  vi.mock('@/features/organization/context/OrganizationContext', () => ({
    useOrganization: () => ({
      organizations: [],
      currentOrganization: { id: 'org1', name: 'Test Org' },
      workspaces: [],
      currentWorkspace: { id: 'ws1', name: 'Test Workspace' },
      isLoading: false,
      selectOrganization: vi.fn(),
      selectWorkspace: vi.fn(),
    }),
    OrganizationProvider: ({ children }: any) => <>{children}</>,
  }))

  it('renders loading then lists bots', async () => {
    const bots = [
      { id: 1, name: 'Destek Botu', description: 'Müşteri desteği', model: 'gpt-4o-mini' },
      { id: 2, name: 'Satış Botu', description: '', model: 'gpt-4.1' },
    ]
    vi.spyOn(api, 'get').mockResolvedValueOnce({ data: bots } as any)

    render(
      <QueryWrapper>
        <ToastProvider>
          <MemoryRouter>
            <ChatbotsPage />
          </MemoryRouter>
        </ToastProvider>
      </QueryWrapper>,
    )

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Chatbotlarım' })).toBeInTheDocument()
    })

    expect(screen.getByText('Destek Botu')).toBeInTheDocument()
    expect(screen.getByText('Satış Botu')).toBeInTheDocument()
    expect(screen.getAllByRole('button', { name: 'Yönet' }).length).toBeGreaterThanOrEqual(1)
  })

  it('opens menu and closes on outside click', async () => {
    const bots = [{ id: 10, name: 'Test Bot', description: 'Açıklama', model: 'gpt-4o' }]
    vi.spyOn(api, 'get').mockResolvedValueOnce({ data: bots } as any)

    render(
      <QueryWrapper>
        <ToastProvider>
          <MemoryRouter>
            <ChatbotsPage />
          </MemoryRouter>
        </ToastProvider>
      </QueryWrapper>,
    )

    await screen.findByText('Test Bot')

    const menuButton = document.querySelector('[data-menu-trigger="10"]') as HTMLElement
    await userEvent.click(menuButton)

    expect(screen.getByRole('button', { name: 'Sil' })).toBeInTheDocument()

    fireEvent.mouseDown(document.body)
    await waitFor(() => {
      expect(screen.queryByRole('button', { name: 'Sil' })).not.toBeInTheDocument()
    })
  })

  it('deletes bot and removes from list', async () => {
    const bots = [
      { id: 1, name: 'Bot A', description: 'A', model: 'gpt' },
      { id: 2, name: 'Bot B', description: 'B', model: 'gpt' },
    ]
    vi.spyOn(api, 'get').mockResolvedValue({ data: bots } as any)
    const delSpy = vi.spyOn(api, 'delete').mockResolvedValueOnce({ data: {} } as any)
    vi.spyOn(window, 'confirm').mockReturnValue(true)

    render(
      <QueryWrapper>
        <ToastProvider>
          <MemoryRouter>
            <ChatbotsPage />
          </MemoryRouter>
        </ToastProvider>
      </QueryWrapper>,
    )

    await screen.findByText('Bot A')

    const trigger = document.querySelector('[data-menu-trigger="2"]') as HTMLElement
    await userEvent.click(trigger)

    const deleteBtn = await screen.findByRole('button', { name: 'Sil' })
    await userEvent.click(deleteBtn)

    await waitFor(() => {
      expect(delSpy).toHaveBeenCalledWith('/api/v1/chatbots/2')
    })
  })

  it('handles list API error and logs', async () => {
    const errSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    vi.spyOn(api, 'get').mockRejectedValueOnce(new Error('fail'))
    render(
      <QueryWrapper>
        <ToastProvider>
          <MemoryRouter>
            <ChatbotsPage />
          </MemoryRouter>
        </ToastProvider>
      </QueryWrapper>,
    )
    await waitFor(() => {
      expect(errSpy).toHaveBeenCalled()
    })
    expect(screen.queryByText('Yükleniyor...')).not.toBeInTheDocument()
  })

  it('handles delete API error and logs', async () => {
    const bots = [{ id: 3, name: 'Err Bot', description: '', model: 'gpt' }]
    vi.spyOn(api, 'get').mockResolvedValue({ data: bots } as any)
    const errSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    vi.spyOn(api, 'delete').mockRejectedValueOnce(new Error('fail'))
    vi.spyOn(window, 'confirm').mockReturnValue(true)
    render(
      <QueryWrapper>
        <ToastProvider>
          <MemoryRouter>
            <ChatbotsPage />
          </MemoryRouter>
        </ToastProvider>
      </QueryWrapper>,
    )
    await screen.findByText('Err Bot')
    const trigger = document.querySelector('[data-menu-trigger="3"]') as HTMLElement
    await userEvent.click(trigger)
    const deleteBtn = await screen.findByRole('button', { name: 'Sil' })
    await userEvent.click(deleteBtn)
    await waitFor(() => {
      expect(errSpy).toHaveBeenCalled()
    })
  })

  it('closes menu when elements missing (else branch)', async () => {
    const bots = [{ id: 7, name: 'Ghost Bot', description: '', model: 'gpt' }]
    vi.spyOn(api, 'get').mockResolvedValueOnce({ data: bots } as any)
    render(
      <QueryWrapper>
        <ToastProvider>
          <MemoryRouter>
            <ChatbotsPage />
          </MemoryRouter>
        </ToastProvider>
      </QueryWrapper>,
    )
    await screen.findByText('Ghost Bot')
    const trigger = document.querySelector('[data-menu-trigger="7"]') as HTMLElement
    await userEvent.click(trigger)
    const menu = document.querySelector('[data-menu="7"]') as HTMLElement
    expect(menu).toBeTruthy()
    // Remove trigger to simulate missing element
    trigger.parentElement?.removeChild(trigger)
    fireEvent.mouseDown(document.body)
    await waitFor(() => {
      expect(document.querySelector('[data-menu="7"]')).toBeNull()
    })
  })
})
