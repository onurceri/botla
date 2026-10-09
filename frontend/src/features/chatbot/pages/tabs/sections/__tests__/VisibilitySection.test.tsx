import { describe, it, expect } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import type { ReactNode } from 'react'
import VisibilitySection from '../VisibilitySection'
import { ChatbotContext } from '../../../../context/ChatbotContext'
import { useChatbotForm } from '../../../../hooks/useChatbotForm'

function Provider({ children, initialPublic }: { children: ReactNode; initialPublic?: boolean }) {
  const form = useChatbotForm()
  return (
    <ChatbotContext.Provider
      value={{
        ...form,
        isPublic: initialPublic ?? form.isPublic,
        planConfig: {},
        userPlan: 'free',
        availableModels: [],
        isLoading: false,
        refetchChatbot: () => {},
      }}
    >
      {children}
    </ChatbotContext.Provider>
  )
}

describe('VisibilitySection', () => {
  it('shows the bot as public by default and can be switched off', () => {
    render(
      <Provider>
        <VisibilitySection />
      </Provider>,
    )
    const toggle = screen.getByRole('switch', { name: 'Herkese Açık' })
    expect(toggle).toHaveAttribute('aria-checked', 'true')
    expect(screen.getByText(/herkese açık sohbet bağlantısı bu botu sunar/)).toBeInTheDocument()

    fireEvent.click(toggle)

    expect(toggle).toHaveAttribute('aria-checked', 'false')
    expect(screen.getByText(/uç noktalar 404 döner/)).toBeInTheDocument()
  })

  it('reflects a private chatbot', () => {
    render(
      <Provider initialPublic={false}>
        <VisibilitySection />
      </Provider>,
    )
    expect(screen.getByRole('switch', { name: 'Herkese Açık' })).toHaveAttribute(
      'aria-checked',
      'false',
    )
  })
})
