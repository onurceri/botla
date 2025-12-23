import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import HandoffRequestsTab from '../HandoffRequestsTab'
import { ChatbotContext } from '../../../context/ChatbotContext'
import * as handoffApi from '@/api/handoff'

// Mock react-router-dom
vi.mock('react-router-dom', () => ({
  useParams: () => ({ id: 'chatbot-123' }),
}))

// Mock API
vi.mock('@/api/handoff', () => ({
  getHandoffRequests: vi.fn(),
  getHandoffRequestDetail: vi.fn(),
  updateHandoffStatus: vi.fn(),
}))

// Mock components that might cause issues
vi.mock('markdown-to-jsx', () => ({
  default: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="markdown">{children}</div>
  ),
}))

describe('HandoffRequestsTab', () => {
  const mockPlanConfig = {
    guardrails: {
      can_use_escalate_fallback: true,
    },
  }

  const mockRequests = [
    {
      id: 'req-1',
      chatbot_id: 'chatbot-123',
      session_id: 'sess-1',
      user_email: 'test@example.com',
      status: 'pending',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    },
  ]

  const mockDetail = {
    request: mockRequests[0],
    messages: [
      {
        id: 'msg-1',
        role: 'user',
        content: '**Hello** markdown',
        created_at: new Date().toISOString(),
      },
      {
        id: 'msg-2',
        role: 'assistant',
        content: '*Hi* there',
        created_at: new Date().toISOString(),
      },
    ],
  }

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(handoffApi.getHandoffRequests).mockResolvedValue(mockRequests as any)
    vi.mocked(handoffApi.getHandoffRequestDetail).mockResolvedValue(mockDetail as any)
  })

  const renderComponent = (planConfig = mockPlanConfig) => {
    return render(
      <ChatbotContext.Provider
        value={
          {
            planConfig: planConfig as any,
            isLoading: false,
            error: null,
            refreshChatbot: vi.fn(),
          } as any
        }
      >
        <HandoffRequestsTab />
      </ChatbotContext.Provider>,
    )
  }

  it('renders requests list', async () => {
    renderComponent()

    await waitFor(() => {
      expect(screen.getByText('test@example.com')).toBeInTheDocument()
    })
  })

  it('opens detail and renders markdown messages', async () => {
    renderComponent()

    await waitFor(() => {
      expect(screen.getAllByText('test@example.com')[0]).toBeInTheDocument()
    })

    // Click on the request card - handle potential multiple elements by taking the first one
    const emailElement = screen.getAllByText('test@example.com')[0]
    fireEvent.click(emailElement.closest('.cursor-pointer')!)

    // Wait for detail to load
    await waitFor(() => {
      expect(screen.getByText('Konuşma Geçmişi')).toBeInTheDocument()
    })

    // Check if markdown content is passed to the mock
    const markdownElements = screen.getAllByTestId('markdown')
    expect(markdownElements.length).toBe(2)
    expect(markdownElements[0]).toHaveTextContent('**Hello** markdown')
    expect(markdownElements[1]).toHaveTextContent('*Hi* there')
  })

  it('shows upgrade message when feature is not available', () => {
    renderComponent({
      guardrails: {
        can_use_escalate_fallback: false,
      },
    })

    expect(screen.getByText('Bu Özellik Planınızda Mevcut Değil')).toBeInTheDocument()
  })
})
