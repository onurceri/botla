import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import OverviewPanel from '../OverviewPanel'
import { ChatbotContext, type ModelInfo } from '../../context/ChatbotContext'
import { useChatbotForm } from '../../hooks/useChatbotForm'

// Mock context wrapper
const mockAvailableModels: ModelInfo[] = [
  { id: 'gpt-4o', name: 'GPT-4o', provider: 'openai', max_tokens: 128000, supported_features: [] },
  {
    id: 'gpt-4o-mini',
    name: 'GPT-4o Mini',
    provider: 'openai',
    max_tokens: 128000,
    supported_features: [],
  },
]

const MockChatbotProvider = ({ children }: { children: React.ReactNode }) => {
  const form = useChatbotForm()
  return (
    <ChatbotContext.Provider
      value={{
        ...form,
        planConfig: {},
        userPlan: 'free',
        availableModels: mockAvailableModels,
        isLoading: false,
      }}
    >
      {children}
    </ChatbotContext.Provider>
  )
}

describe('OverviewPanel', () => {
  it('renders and allows editing of name and custom instruction', () => {
    const setName = vi.fn()
    const setCustomInstruction = vi.fn()
    const setModel = vi.fn()
    const setTemperature = vi.fn()
    const setMaxTokens = vi.fn()

    render(
      <MockChatbotProvider>
        <OverviewPanel
          name="Bot"
          setName={setName}
          customInstruction="Merhaba"
          setCustomInstruction={setCustomInstruction}
          model="gpt-4o"
          setModel={setModel}
          temperature={0.5}
          setTemperature={setTemperature}
          maxTokens={1024}
          setMaxTokens={setMaxTokens}
        />
      </MockChatbotProvider>,
    )

    expect(screen.getByText(/Kimlik/i)).toBeInTheDocument()
    const nameInput = screen.getByDisplayValue('Bot') as HTMLInputElement
    fireEvent.change(nameInput, { target: { value: 'Destek Botu' } })
    expect(setName).toHaveBeenCalledWith('Destek Botu')

    const promptTextarea = screen.getByPlaceholderText(
      'Botunuza özel davranış kuralları ekleyin...',
    ) as HTMLTextAreaElement
    fireEvent.change(promptTextarea, { target: { value: 'Yeni talimat' } })
    expect(setCustomInstruction).toHaveBeenCalledWith('Yeni talimat')
  })

  it('renders available models from context', () => {
    render(
      <MockChatbotProvider>
        <OverviewPanel
          name="Bot"
          setName={vi.fn()}
          customInstruction=""
          setCustomInstruction={vi.fn()}
          model="gpt-4o"
          setModel={vi.fn()}
          temperature={0.7}
          setTemperature={vi.fn()}
          maxTokens={512}
          setMaxTokens={vi.fn()}
        />
      </MockChatbotProvider>,
    )

    // Check that model options are rendered from context
    expect(screen.getAllByText('GPT-4o')[0]).toBeInTheDocument()
    expect(screen.getAllByText('GPT-4o Mini')[0]).toBeInTheDocument()
  })
})
