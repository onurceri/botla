import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import ActionLogs from '../ActionLogs'
import * as actionApi from '@/api/action'

// Mock the API module
vi.mock('@/api/action', () => ({
  getActionLogs: vi.fn(),
}))

describe('ActionLogs', () => {
  const mockLogs = [
    {
      id: 'log-1',
      chatbot_id: 'bot-1',
      action_id: 'act-1',
      status: 'success',
      duration_ms: 120,
      created_at: '2023-10-27T10:00:00Z',
      request_payload: { q: 'test' },
      response_payload: { a: 'ok' },
    },
    {
      id: 'log-2',
      chatbot_id: 'bot-1',
      action_id: 'act-1',
      status: 'failed',
      duration_ms: 500,
      created_at: '2023-10-27T10:05:00Z',
      error_message: 'Network Error',
    },
  ]

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders loading state initially', async () => {
    // Return a promise that doesn't resolve immediately to test loading state if needed
    // But since useEffect runs on mount, we might miss the initial true state if we await.
    // However, the button should be disabled.
    ;(actionApi.getActionLogs as any).mockImplementation(() => new Promise(() => {})) // Never resolves
    render(<ActionLogs chatbotId="bot-1" />)

    const refreshBtn = screen.getByText('Yenile')
    expect(refreshBtn).toBeDisabled()
  })

  it('renders logs after fetch', async () => {
    ;(actionApi.getActionLogs as any).mockResolvedValue({ logs: mockLogs, page: 1, limit: 20 })
    render(<ActionLogs chatbotId="bot-1" />)

    await waitFor(() => {
      expect(screen.getByText('Başarılı')).toBeInTheDocument()
    })
    expect(screen.getByText('Hata')).toBeInTheDocument()
    expect(screen.getByText('120 ms')).toBeInTheDocument()
  })

  it('opens details dialog when clicked', async () => {
    ;(actionApi.getActionLogs as any).mockResolvedValue({ logs: mockLogs })
    render(<ActionLogs chatbotId="bot-1" />)

    await waitFor(() => {
      expect(screen.getByText('Başarılı')).toBeInTheDocument()
    })

    const buttons = screen.getAllByText('İncele')
    fireEvent.click(buttons[0])

    expect(await screen.findByText('Aksiyon Detayı')).toBeInTheDocument()
    // JSON.stringify formatting check might be fragile, checking for content existence
    expect(screen.getByText((content) => content.includes('"q": "test"'))).toBeInTheDocument()
  })

  it('shows empty state', async () => {
    ;(actionApi.getActionLogs as any).mockResolvedValue({ logs: [] })
    render(<ActionLogs chatbotId="bot-1" />)

    await waitFor(() => {
      expect(screen.getByText('Henüz bir kayıt yok.')).toBeInTheDocument()
    })
  })
})
