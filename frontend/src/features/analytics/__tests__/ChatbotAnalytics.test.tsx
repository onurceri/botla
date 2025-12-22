import { describe, it, expect, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { ChatbotAnalytics } from '../ChatbotAnalytics'
import * as analyticsApi from '@/api/analytics'

// Mock the API
vi.mock('@/api/analytics', () => ({
  getChatbotAnalyticsOverview: vi.fn(),
  getChatbotAnalyticsTrends: vi.fn(),
  getSourceUsageStats: vi.fn()
}))

// Mock ResponsiveContainer to avoid sizing issues in tests
vi.mock('recharts', async (importOriginal) => {
  const original = await importOriginal() as any
  return {
    ...original,
    ResponsiveContainer: ({ children }: any) => <div style={{ width: 800, height: 300 }}>{children}</div>
  }
})

describe('ChatbotAnalytics', () => {
  const mockOverview = {
    total_conversations: 120,
    total_messages: 450,
    total_tokens_used: 15000,
    avg_positive_feedback: 0.95
  }

  const mockTrends = [
    { date: '2023-01-01', total_conversations: 10, total_messages: 40 },
    { date: '2023-01-02', total_conversations: 15, total_messages: 55 }
  ]

  it('renders loading state initially', () => {
    vi.mocked(analyticsApi.getChatbotAnalyticsOverview).mockResolvedValue({})
    vi.mocked(analyticsApi.getChatbotAnalyticsTrends).mockResolvedValue([])
    
    const { container } = render(<ChatbotAnalytics chatbotId="123" />)
    const skeletons = container.querySelectorAll('.animate-pulse')
    expect(skeletons.length).toBeGreaterThan(0)
  })

  it('renders overview stats correctly', async () => {
    vi.mocked(analyticsApi.getChatbotAnalyticsOverview).mockResolvedValue(mockOverview)
    vi.mocked(analyticsApi.getChatbotAnalyticsTrends).mockResolvedValue({ daily: mockTrends })
    vi.mocked(analyticsApi.getSourceUsageStats).mockResolvedValue([])

    render(<ChatbotAnalytics chatbotId="123" />)

    await waitFor(() => {
      expect(analyticsApi.getChatbotAnalyticsOverview).toHaveBeenCalled()
    })

    // Wait for stats to appear
    await waitFor(() => {
        expect(screen.getByText('120')).toBeInTheDocument() 
    }, { timeout: 3000 })
    
    expect(screen.getByText('450')).toBeInTheDocument() 
    expect(screen.getByText('15,000')).toBeInTheDocument() 
    expect(screen.getByText(/95%/)).toBeInTheDocument() 
  })
})
