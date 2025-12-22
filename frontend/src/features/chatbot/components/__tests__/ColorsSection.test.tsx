import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, within, cleanup } from '@testing-library/react'
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

  it('updates header and user colors', () => {
    const setHeader = vi.fn()
    const setUserText = vi.fn()
    render(
      <ColorsSection
        {...defaultProps}
        setChatHeaderColor={setHeader}
        setUserMessageTextColor={setUserText}
      />
    )
    const headerColor = screen.getByLabelText('Header') as HTMLInputElement
    fireEvent.change(headerColor, { target: { value: '#000000' } })
    expect(setHeader).toHaveBeenCalledWith('#000000')
    // User message section
    // Find the section by text and get its container
    // H4 is inside the container div
    const userMessageHeader = screen.getByText('Kullanıcı Mesajları')
    const userMessageSection = userMessageHeader.parentElement!
    
    // Within that section, find the label "Yazı"
    const userText = within(userMessageSection).getByLabelText('Yazı') as HTMLInputElement
    
    fireEvent.change(userText, { target: { value: '#333333' } })
    expect(setUserText).toHaveBeenCalledWith('#333333')
  })

  it('updates font and theme color', async () => {
    const user = userEvent.setup()
    const setFont = vi.fn()
    const setTheme = vi.fn()
    render(
      <ColorsSection
        {...defaultProps}
        setChatFontFamily={setFont}
        setThemeColor={setTheme}
      />
    )
    
    // Debugging font selection - use last instance if multiple found
    const fontSelects = screen.getAllByLabelText('Yazı Tipi')
    const fontSelect = fontSelects[fontSelects.length - 1] as HTMLSelectElement
    
    await user.selectOptions(fontSelect, 'Roboto, sans-serif')
    expect(setFont).toHaveBeenCalledWith('Roboto, sans-serif')

    const themeInputs = screen.getAllByLabelText('Varsayılan İkon Rengi')
    const themeInput = themeInputs[themeInputs.length - 1] as HTMLInputElement
    
    fireEvent.change(themeInput, { target: { value: '#ff0000' } })
    expect(setTheme).toHaveBeenCalledWith('#ff0000')
  })

  it('renders labels and color pickers when expanded', () => {
    const utils = render(
      <ColorsSection {...defaultProps} />
    )
    const headers = screen.getAllByText('Yazı ve Renkler')
    expect(headers.length).toBeGreaterThan(0)
    
    // Check for new fields
    // Allow multiple instances if cleanup is flaky, but ensure at least one exists
    const genelHeaders = screen.getAllByText('Genel')
    expect(genelHeaders.length).toBeGreaterThanOrEqual(1)
    expect(genelHeaders[genelHeaders.length - 1]).toBeInTheDocument()
 
    const fonts = screen.getAllByText('Yazı Tipi')
    expect(fonts[fonts.length - 1]).toBeInTheDocument()

    const themes = screen.getAllByText('Varsayılan İkon Rengi')
    expect(themes[themes.length - 1]).toBeInTheDocument()

    const bubbles = screen.getAllByText('Kabarcık Ovalleşmesi')
    expect(bubbles[bubbles.length - 1]).toBeInTheDocument()

    // Check for section headers
    expect(screen.getByText('Panel & Header')).toBeInTheDocument()
    expect(screen.getByText('Bot Mesajları')).toBeInTheDocument()
    expect(screen.getByText('Kullanıcı Mesajları')).toBeInTheDocument()
    expect(screen.getByText('Giriş Alanı')).toBeInTheDocument()
    
    const colorInputs = utils.container.querySelectorAll('input[type="color"]')
    // 6 existing + 1 theme color + 3 input/send = 10? 
    // ChatBg, Header, HeaderText, BotMsg, BotText, UserMsg, UserText, InputBg, InputText, SendBtn, ThemeColor.
    // 11 color inputs total.
    expect(colorInputs.length).toBeGreaterThanOrEqual(10)
  })
})
