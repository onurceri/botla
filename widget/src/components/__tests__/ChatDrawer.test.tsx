import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, within } from '@testing-library/preact'
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

  it('shows Botla branding when hideBranding is false, even if custom exists', async () => {
    const { container } = render(
      <ChatDrawer 
        color="#3b82f6"
        messages={[]}
        loading={false}
        input=""
        setInput={() => {}}
        onSend={() => {}}
        onClose={() => {}}
        hideBranding={false}
        customBranding={{ text: 'ACME', link: 'https://example.com' }}
      />
    )
    const scoped = within(container as HTMLElement)
    expect(await scoped.findByText(/Powered by/i)).toBeDefined()
    expect(await scoped.findByText('Botla')).toBeDefined()
  })

  it('shows custom branding when hideBranding is true and custom exists', async () => {
    const { container } = render(
      <ChatDrawer 
        color="#3b82f6"
        messages={[]}
        loading={false}
        input=""
        setInput={() => {}}
        onSend={() => {}}
        onClose={() => {}}
        hideBranding={true}
        customBranding={{ text: 'ACME', link: 'https://example.com' }}
      />
    )
    const scoped = within(container as HTMLElement)
    expect(await scoped.findByText('ACME')).toBeDefined()
    // Botla default should not be visible
    expect(scoped.queryByText(/Powered by/i)).toBeNull()
  })

  it('shows no branding when hideBranding is true and custom is absent', async () => {
    const { container } = render(
      <ChatDrawer 
        color="#3b82f6"
        messages={[]}
        loading={false}
        input=""
        setInput={() => {}}
        onSend={() => {}}
        onClose={() => {}}
        hideBranding={true}
      />
    )
    const scoped = within(container as HTMLElement)
    expect(scoped.queryByText(/Powered by/i)).toBeNull()
  })
})
