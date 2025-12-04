import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import React from 'react'
import SettingsPage from '@/pages/SettingsPage'

vi.mock('@/api/client', () => {
  return {
    api: {
      get: vi.fn(() => Promise.resolve({
        data: { subscription_plan: 'pro' },
        headers: { 'x-ratelimit-limit': '10', 'x-ratelimit-remaining': '9' },
      })),
    },
  }
})

describe('SettingsPage Plan sekmesi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('başlangıçta Profil görünür ve Plan tıklanınca Plan içerik görünür', async () => {
    render(<SettingsPage />)
    expect(screen.getByText('Profil Bilgileri')).toBeInTheDocument()
    expect(screen.queryByText('Plan Özeti')).not.toBeInTheDocument()

    const planButtons = screen.getAllByRole('button', { name: /Plan/i })
    await userEvent.click(planButtons[0])
    expect(await screen.findByText('Plan Özeti')).toBeInTheDocument()
    expect(screen.getByText('Kota ve Limitler')).toBeInTheDocument()
  })

  it('plan ve rate limit header bilgilerini gösterir', async () => {
    render(<SettingsPage />)
    const planButtons2 = screen.getAllByRole('button', { name: /Plan/i })
    await userEvent.click(planButtons2[0])

    // Plan adı
    expect(await screen.findByText('pro')).toBeInTheDocument()
    // Rate limit kalan/toplam
    expect(screen.getByText(/9\/10 \(kalan\/toplam\)/)).toBeInTheDocument()
  })
})
