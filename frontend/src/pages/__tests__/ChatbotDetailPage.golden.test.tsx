import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import ChatbotDetailPage from '../ChatbotDetailPage'
import { ToastProvider } from '@/components/ui/toast'

describe('ChatbotDetailPage golden (new chatbot)', () => {
  it('renders header and basic form for new chatbot', () => {
    render(
      <ToastProvider>
        <MemoryRouter initialEntries={["/chatbots/new"]}>
          <Routes>
            <Route path="/chatbots/:id" element={<ChatbotDetailPage />} />
          </Routes>
        </MemoryRouter>
      </ToastProvider>
    )

    expect(screen.getByText('Yeni Chatbot')).toBeInTheDocument()
    expect(screen.getByText('Temel Bilgiler')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Örn: Müşteri Temsilcisi')).toBeInTheDocument()
  })
})
