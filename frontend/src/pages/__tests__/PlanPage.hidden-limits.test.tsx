import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import PlanPage from '../PlanPage'
import { api } from '@/api/client'

describe('PlanPage hidden limits', () => {
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

  it('renders monthly ingestions and embedding token usage from /me', async () => {
    vi.spyOn(api, 'get').mockResolvedValueOnce({
      data: {
        plan_code: 'pro',
        plan_price: 199,
        plan_currency: 'TRY',
        config: {
          max_monthly_ingestions: 50,
          max_monthly_embedding_tokens: 250000,
          files: { ocr_enabled: true, max_size_mb: 20, max_files_per_bot: 20, max_files_total: 100, total_storage_mb: 500 },
          scraping: { dynamic_enabled: true, max_urls_per_bot: 10, max_pages_per_crawl: 10 },
          chat: { allowed_models: ['gpt-4o'], max_monthly_tokens: 1000000, rag: { top_k: 5, max_context_tokens: 4000 } },
          refresh: { enabled: true, max_monthly: 5 },
        },
        usage: {
          files_count: 0,
          max_files_count_in_one_bot: 0,
          storage_used_mb: 0,
          urls_count: 0,
          max_urls_count_in_one_bot: 0,
          tokens_used: 0,
          ingestions_used: 7,
          ingestion_embedding_tokens: 12345,
          refresh_count: 0,
        },
      },
      headers: { 'x-ratelimit-limit': '100', 'x-ratelimit-remaining': '80' },
    } as any)

    render(
      <MemoryRouter>
        <PlanPage />
      </MemoryRouter>
    )

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Plan ve Faturalandırma' })).toBeInTheDocument()
    })

    expect(screen.getByText('Aylık Kaynak Ekleme')).toBeInTheDocument()
    expect(screen.getByText((t) => /7\s*\/\s*50\s*kullanıldı/.test(t))).toBeInTheDocument()

    expect(screen.getByText('Aylık Embedding Token Kullanımı')).toBeInTheDocument()
    expect(screen.getByText((t) => /12,345\s*\/\s*250,000/.test(t))).toBeInTheDocument()
  })
})
