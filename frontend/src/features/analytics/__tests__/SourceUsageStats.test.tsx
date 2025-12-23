import { describe, it, expect, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { SourceUsageStats } from '../SourceUsageStats'
import * as analyticsApi from '@/api/analytics'

// Mock the API
vi.mock('@/api/analytics', () => ({
  getSourceUsageStats: vi.fn(),
}))

describe('SourceUsageStats', () => {
  const mockStats = [
    {
      source_id: '1',
      source_name: 'Test Source PDF',
      source_type: 'pdf',
      times_used: 15,
      avg_relevance: 0.85,
      positive_feedback: 10,
      negative_feedback: 1,
      last_used: '2023-01-01T12:00:00Z',
    },
    {
      source_id: '2',
      source_name: 'Website Home',
      source_type: 'url',
      times_used: 42,
      avg_relevance: 0.92,
      positive_feedback: 30,
      negative_feedback: 0,
      last_used: '2023-01-02T12:00:00Z',
    },
  ]

  it('renders loading state initially', () => {
    vi.mocked(analyticsApi.getSourceUsageStats).mockResolvedValue([])
    const { container } = render(<SourceUsageStats chatbotId="123" />)
    // Check for skeleton loaders
    const skeletons = container.querySelectorAll('.animate-pulse')
    expect(skeletons.length).toBeGreaterThan(0)
  })

  it('renders empty state when no data', async () => {
    vi.mocked(analyticsApi.getSourceUsageStats).mockResolvedValue([])
    render(<SourceUsageStats chatbotId="123" />)

    await waitFor(() => {
      expect(screen.getByText(/Henüz veri yok/i)).toBeInTheDocument()
    })
  })

  it('renders stats cards correctly', async () => {
    vi.mocked(analyticsApi.getSourceUsageStats).mockResolvedValue(mockStats)
    render(<SourceUsageStats chatbotId="123" />)

    await waitFor(() => {
      expect(screen.getByText('Test Source PDF')).toBeInTheDocument()
      expect(screen.getByText('Website Home')).toBeInTheDocument()

      // Check for stats
      expect(screen.getByText('15')).toBeInTheDocument() // times used
      expect(screen.getByText('42')).toBeInTheDocument()

      // Check for badges
      expect(screen.getByText('pdf')).toBeInTheDocument()
      expect(screen.getByText('url')).toBeInTheDocument()
    })
  })
})
