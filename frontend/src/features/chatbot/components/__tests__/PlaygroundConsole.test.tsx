import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, fireEvent, act, cleanup } from '@testing-library/react'
import PlaygroundConsole from '../PlaygroundConsole'

describe('PlaygroundConsole', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    cleanup()
    vi.useRealTimers()
  })

  it('toggles expansion when clicked', () => {
    render(<PlaygroundConsole />)
    
    const header = screen.getByText(/Event Debug Console/i)
    
    // Initially collapsed (h-11)
    const consoleContainer = header.closest('.absolute')
    expect(consoleContainer).toHaveClass('h-11')
    
    // Click to expand
    fireEvent.click(header)
    expect(consoleContainer).toHaveClass('h-80')
    
    // Click to collapse
    fireEvent.click(header)
    expect(consoleContainer).toHaveClass('h-11')
  })

  it('captures and displays widget events', () => {
    render(<PlaygroundConsole />)
    
    // Expand to see logs
    fireEvent.click(screen.getByText(/Event Debug Console/i))
    
    expect(screen.getByText(/Henüz olay kaydedilmedi/i)).toBeInTheDocument()

    // Simulate postMessage from widget
    act(() => {
      window.dispatchEvent(new MessageEvent('message', {
        data: {
          type: 'WIDGET_EVENT_MESSAGE_SENT',
          payload: { content: 'Merhaba bot' }
        }
      }))
    })

    expect(screen.getByText('Kullanıcı mesajı: "Merhaba bot"')).toBeInTheDocument()
    
    // Simulate another event
    act(() => {
      window.dispatchEvent(new MessageEvent('message', {
        data: {
          type: 'WIDGET_EVENT_RESPONSE_RECEIVED',
          payload: { content: 'Size nasıl yardımcı olabilirim?' }
        }
      }))
    })

    expect(screen.getByText('Bot yanıtı: "Size nasıl yardımcı olabilirim?"')).toBeInTheDocument()
  })

  it('shows log count when collapsed', () => {
    render(<PlaygroundConsole />)
    
    act(() => {
      window.dispatchEvent(new MessageEvent('message', {
        data: { type: 'WIDGET_EVENT_CONFIG_LOADED', payload: {} }
      }))
    })

    // Should show badge with "1"
    expect(screen.getByText('1')).toBeInTheDocument()
  })

  it('clears logs when trash icon is clicked', () => {
    render(<PlaygroundConsole />)
    
    // Expand
    fireEvent.click(screen.getByText(/Event Debug Console/i))
    
    // Add a log
    act(() => {
      window.dispatchEvent(new MessageEvent('message', {
        data: { type: 'WIDGET_EVENT_CONFIG_LOADED', payload: {} }
      }))
    })
    
    expect(screen.getByText(/Widget yapılandırması başarıyla yüklendi/i)).toBeInTheDocument()

    // Clear logs
    const clearBtn = screen.getByTitle(/Konsolu Temizle/i)
    fireEvent.click(clearBtn)

    expect(screen.getByText(/Henüz olay kaydedilmedi/i)).toBeInTheDocument()
    expect(screen.queryByText(/Widget yapılandırması başarıyla yüklendi/i)).not.toBeInTheDocument()
  })

  it('displays different event types correctly', () => {
    render(<PlaygroundConsole />)
    fireEvent.click(screen.getByText(/Event Debug Console/i))

    const events = [
      { type: 'ERROR', payload: { message: 'Network timeout' }, expected: 'Hata oluştu: Network timeout' },
      { type: 'HANDOFF', payload: {}, expected: 'Canlı desteğe yönlendirme tetiklendi.' }
    ]

    events.forEach(event => {
      act(() => {
        window.dispatchEvent(new MessageEvent('message', {
          data: { type: `WIDGET_EVENT_${event.type}`, payload: event.payload }
        }))
      })
      expect(screen.getByText(event.expected)).toBeInTheDocument()
    })
  })
})
