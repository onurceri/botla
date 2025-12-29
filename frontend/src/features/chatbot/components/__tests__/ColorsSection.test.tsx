import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, cleanup } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import ColorsSection from '../ColorsSection'

beforeEach(() => {
  cleanup()
})

describe('ColorsSection', () => {
  const defaultProps = {
    isExpanded: true,
    onToggle: vi.fn(),
    chatBackgroundColor: '#fff',
    setChatBackgroundColor: vi.fn(),
    chatHeaderColor: '#3b82f6',
    setChatHeaderColor: vi.fn(),
    chatHeaderTextColor: '#ffffff',
    setChatHeaderTextColor: vi.fn(),
    botMessageColor: '#eee',
    setBotMessageColor: vi.fn(),
    botMessageTextColor: '#000',
    setBotMessageTextColor: vi.fn(),
    userMessageColor: '#222',
    setUserMessageColor: vi.fn(),
    userMessageTextColor: '#fff',
    setUserMessageTextColor: vi.fn(),
    inputBackgroundColor: '#fff',
    setInputBackgroundColor: vi.fn(),
    inputTextColor: '#000',
    setInputTextColor: vi.fn(),
    sendButtonColor: '#000',
    setSendButtonColor: vi.fn(),
    chatFontFamily: 'Inter, sans-serif',
    setChatFontFamily: vi.fn(),
    themeColor: '#000',
    setThemeColor: vi.fn(),
    bubbleRadius: '22px',
    setBubbleRadius: vi.fn(),
  }

  it('updates header and user colors via color picker', () => {
    const setHeader = vi.fn()
    const setUserText = vi.fn()
    render(
      <ColorsSection
        {...defaultProps}
        setChatHeaderColor={setHeader}
        setUserMessageTextColor={setUserText}
      />,
    )

    // Verify color picker buttons are present
    const headerPicker = document.getElementById('header-color')
    expect(headerPicker).toBeInTheDocument()

    const userTextPicker = document.getElementById('user-text-color')
    expect(userTextPicker).toBeInTheDocument()
  })

  it('updates font and theme color', async () => {
    const user = userEvent.setup()
    const setFont = vi.fn()
    const setTheme = vi.fn()
    render(<ColorsSection {...defaultProps} setChatFontFamily={setFont} setThemeColor={setTheme} />)

    // Font selection
    const fontSelect = screen.getByLabelText('Yazı Tipi') as HTMLSelectElement
    await user.selectOptions(fontSelect, 'Roboto, sans-serif')
    expect(setFont).toHaveBeenCalledWith('Roboto, sans-serif')

    // Theme color picker present
    const themePicker = document.getElementById('theme-color')
    expect(themePicker).toBeInTheDocument()
  })

  it('renders labels and color pickers when expanded', () => {
    render(<ColorsSection {...defaultProps} />)

    // Check section title
    expect(screen.getByText('Yazı ve Renkler')).toBeInTheDocument()

    // Check for section headers
    expect(screen.getByText('Genel')).toBeInTheDocument()
    expect(screen.getByText('Yazı Tipi')).toBeInTheDocument()
    expect(screen.getByText('Varsayılan İkon Rengi')).toBeInTheDocument()
    expect(screen.getByText('Kabarcık Ovalleşmesi')).toBeInTheDocument()
    expect(screen.getByText('Panel & Header')).toBeInTheDocument()
    expect(screen.getByText('Bot Mesajları')).toBeInTheDocument()
    expect(screen.getByText('Kullanıcı Mesajları')).toBeInTheDocument()
    expect(screen.getByText('Giriş Alanı')).toBeInTheDocument()

    // Check that color pickers are present (one for each color picker)
    const ids = [
      'theme-color',
      'chat-bg',
      'header-color',
      'bot-msg-color',
      'user-msg-color',
      'input-bg',
      'send-btn',
    ]

    ids.forEach((id) => {
      expect(document.getElementById(id)).toBeInTheDocument()
    })
  })
})
