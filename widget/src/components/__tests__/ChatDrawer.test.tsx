import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/preact'
import { ChatDrawer } from '../ChatDrawer'

describe('ChatDrawer', () => {
  it('does not send on Enter when loading', async () => {
    const onSend = vi.fn()
    render(
      <ChatDrawer 
        color="#3b82f6"
        messages={[]}
        loading={true}
        input="hello"
        setInput={() => {}}
        onSend={onSend}
        onClose={() => {}}
      />
    )
    const input = screen.getByPlaceholderText('Mesaj yazın...')
    fireEvent.keyDown(input, { key: 'Enter', code: 'Enter' })
    expect(onSend).not.toHaveBeenCalled()
  })

  it('renders suggestions when no user message', async () => {
    const onSend = vi.fn()
    render(
      <ChatDrawer 
        color="#3b82f6"
        messages={[{ role: 'assistant', content: 'Merhaba' }]}
        loading={false}
        input=""
        setInput={() => {}}
        onSend={onSend}
        onClose={() => {}}
        botName="Bot"
        suggestions={["S1", "S2"]}
        onPickSuggestion={() => {}}
      />
    )
    expect(await screen.findByRole('button', { name: 'S1' })).toBeDefined()
  })
})
