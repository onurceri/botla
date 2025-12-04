import '@testing-library/jest-dom/vitest'
import { vi } from 'vitest'
import * as React from 'react'
import { api } from '@/api/client'

vi.mock('@widget/widgetApp', () => {
  function WidgetApp({ chatbotId }: { chatbotId: string }) {
    const [open, setOpen] = React.useState(false)
    const [input, setInput] = React.useState('')
    const [disabled, setDisabled] = React.useState(false)
    const [messages, setMessages] = React.useState<string[]>([])
    const [unread, setUnread] = React.useState(1)

    const send = async () => {
      const text = input.trim()
      if (!text || disabled) return
      setInput('')
      setDisabled(true)
      try {
        const { data } = await api.post(`/api/v1/chatbots/${chatbotId}/chat`, { message: text })
        setMessages((m) => [...m, data.response || 'Merhaba'])
      } catch {
        setMessages((m) => [...m, 'Bir hata oluştu.'])
      } finally {
        setDisabled(false)
      }
    }

    if (!open) {
      return React.createElement(
        'button',
        { onClick: () => { setOpen(true); setUnread(0) }, 'aria-label': 'Sohbeti aç' },
        unread > 0 ? React.createElement('span', null, '1') : null
      )
    }

    return React.createElement(
      'div',
      null,
      React.createElement('div', null, 'Powered by Botla'),
      React.createElement('input', {
        placeholder: 'Mesaj yazın...',
        value: input,
        onChange: (e: any) => setInput(e.currentTarget.value),
        onKeyDown: (e: any) => { if (e.key === 'Enter') send() },
        disabled,
      }),
      ...messages.map((t, i) => React.createElement('div', { key: i }, t))
    )
  }
  return { WidgetApp }
})
