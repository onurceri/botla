import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
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
    const input = screen.getByPlaceholderText('Mesaj yazın')
    fireEvent.keyDown(input, { key: 'Enter', code: 'Enter' })
    expect(onSend).not.toHaveBeenCalled()
  })
})
