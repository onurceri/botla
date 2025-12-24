import '@testing-library/jest-dom/vitest'
import { beforeEach, afterEach, vi } from 'vitest'
import { cleanup } from '@testing-library/react'
import * as React from 'react'
import { api } from '@/api/client'

// Mock PostHog to prevent analytics calls in tests
vi.mock('posthog-js', () => ({
  default: {
    init: vi.fn(),
    capture: vi.fn(),
    identify: vi.fn(),
    reset: vi.fn(),
    __loaded: false,
  },
}))

vi.mock('posthog-js/react', () => ({
  PostHogProvider: ({ children }: { children: React.ReactNode }) => children,
  usePostHog: () => ({
    capture: vi.fn(),
    identify: vi.fn(),
  }),
}))

const createUnexpectedHttpCallError = (method: string, args: unknown[]) => {
  const prettyArgs = args.map((a) => {
    try {
      return typeof a === 'string' ? a : JSON.stringify(a)
    } catch {
      return String(a)
    }
  })
  return new Error(
    `Unexpected ${method} call in tests: ${prettyArgs.join(' ')}. Mock the call in the test.`,
  )
}

beforeEach(() => {
  const client = api as any
  const methods = ['get', 'post', 'put', 'patch', 'delete', 'request']

  for (const method of methods) {
    const fn = client?.[method]
    if (typeof fn !== 'function') continue
    if (!vi.isMockFunction(fn)) {
      vi.spyOn(client, method)
    }
    vi.mocked(client[method]).mockImplementation((...args: unknown[]) => {
      throw createUnexpectedHttpCallError(`api.${method}`, args)
    })
  }
})

afterEach(() => {
  cleanup()
})

const localStorageStore: Record<string, string> = {}
Object.defineProperty(window, 'localStorage', {
  value: {
    getItem: vi.fn((key: string) => (key in localStorageStore ? localStorageStore[key] : null)),
    setItem: vi.fn((key: string, value: string) => {
      localStorageStore[key] = String(value)
    }),
    removeItem: vi.fn((key: string) => {
      delete localStorageStore[key]
    }),
    clear: vi.fn(() => {
      for (const key of Object.keys(localStorageStore)) {
        delete localStorageStore[key]
      }
    }),
  },
  writable: true,
})

// Mock IntersectionObserver
class IntersectionObserverMock {
  disconnect = vi.fn()
  observe = vi.fn()
  takeRecords = vi.fn()
  unobserve = vi.fn()
}

vi.stubGlobal('IntersectionObserver', IntersectionObserverMock)

// Mock ResizeObserver
class ResizeObserverMock {
  disconnect = vi.fn()
  observe = vi.fn()
  unobserve = vi.fn()
}

vi.stubGlobal('ResizeObserver', ResizeObserverMock)

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
        {
          onClick: () => {
            setOpen(true)
            setUnread(0)
          },
          'aria-label': 'Sohbeti aç',
        },
        unread > 0 ? React.createElement('span', null, '1') : null,
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
        onKeyDown: (e: any) => {
          if (e.key === 'Enter') send()
        },
        disabled,
      }),
      ...messages.map((t, i) => React.createElement('div', { key: i }, t)),
    )
  }
  return { WidgetApp }
})
