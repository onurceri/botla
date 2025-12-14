import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent, within } from '@testing-library/react'
import ColorsSection from '../ColorsSection'

describe('ColorsSection', () => {
  it('updates header and user colors', () => {
    const onToggle = vi.fn()
    const setChatBg = vi.fn()
    const setHeader = vi.fn()
    const setHeaderText = vi.fn()
    const setBotMsg = vi.fn()
    const setBotText = vi.fn()
    const setUserMsg = vi.fn()
    const setUserText = vi.fn()
    render(
      <ColorsSection
        isExpanded={true}
        onToggle={onToggle}
        chatBackgroundColor="#fff"
        setChatBackgroundColor={setChatBg}
        chatHeaderColor="#3b82f6"
        setChatHeaderColor={setHeader}
        chatHeaderTextColor="#ffffff"
        setChatHeaderTextColor={setHeaderText}
        botMessageColor="#eee"
        setBotMessageColor={setBotMsg}
        botMessageTextColor="#000"
        setBotMessageTextColor={setBotText}
        userMessageColor="#222"
        setUserMessageColor={setUserMsg}
        userMessageTextColor="#fff"
        setUserMessageTextColor={setUserText}
      />
    )
    const headerColor = screen.getByLabelText('Header') as HTMLInputElement
    fireEvent.change(headerColor, { target: { value: '#000000' } })
    expect(setHeader).toHaveBeenCalledWith('#000000')
    const userText = screen.getByLabelText('Kullanıcı Yazı') as HTMLInputElement
    fireEvent.change(userText, { target: { value: '#333333' } })
    expect(setUserText).toHaveBeenCalledWith('#333333')
  })

  it('updates all color inputs', () => {
    const setChatBg = vi.fn()
    const setHeader = vi.fn()
    const setHeaderText = vi.fn()
    const setBotMsg = vi.fn()
    const setBotText = vi.fn()
    const setUserMsg = vi.fn()
    const setUserText = vi.fn()
    const utils2 = render(
      <ColorsSection
        isExpanded={true}
        onToggle={() => {}}
        chatBackgroundColor="#ffffff"
        setChatBackgroundColor={setChatBg}
        chatHeaderColor="#000000"
        setChatHeaderColor={setHeader}
        chatHeaderTextColor="#111111"
        setChatHeaderTextColor={setHeaderText}
        botMessageColor="#222222"
        setBotMessageColor={setBotMsg}
        botMessageTextColor="#333333"
        setBotMessageTextColor={setBotText}
        userMessageColor="#444444"
        setUserMessageColor={setUserMsg}
        userMessageTextColor="#555555"
        setUserMessageTextColor={setUserText}
      />
    )
    const view = within(utils2.container)
    const chatBgLabel = view.getByText('Chat Arka Plan')
    const chatBgTextInput = chatBgLabel.parentElement!.querySelectorAll('input')[1] as HTMLInputElement
    fireEvent.change(chatBgTextInput, { target: { value: '#cccccc' } })
    expect(setChatBg).toHaveBeenCalledWith('#cccccc')
    const headerTextLabel = view.getByText('Header Yazı')
    const headerTextInput = headerTextLabel.parentElement!.querySelectorAll('input')[1] as HTMLInputElement
    fireEvent.change(headerTextInput, { target: { value: '#aaaaaa' } })
    expect(setHeaderText).toHaveBeenCalledWith('#aaaaaa')
    const botMsgLabel = view.getByText('Bot Mesaj Arka Planı')
    const botMsgInput = botMsgLabel.parentElement!.querySelectorAll('input')[1] as HTMLInputElement
    fireEvent.change(botMsgInput, { target: { value: '#bbbbbb' } })
    expect(setBotMsg).toHaveBeenCalledWith('#bbbbbb')
    const botTextLabel = view.getByText('Bot Yazı')
    const botTextInput = botTextLabel.parentElement!.querySelectorAll('input')[1] as HTMLInputElement
    fireEvent.change(botTextInput, { target: { value: '#dddddd' } })
    expect(setBotText).toHaveBeenCalledWith('#dddddd')
    const userMsgLabel = view.getByText('Kullanıcı Mesaj Arka Planı')
    const userMsgInput = userMsgLabel.parentElement!.querySelectorAll('input')[1] as HTMLInputElement
    fireEvent.change(userMsgInput, { target: { value: '#eeeeee' } })
    expect(setUserMsg).toHaveBeenCalledWith('#eeeeee')
  })

  it('renders labels and color pickers when expanded', () => {
    const noop = () => {}
    const utils = render(
      <ColorsSection
        isExpanded={true}
        onToggle={noop}
        chatBackgroundColor="#ffffff"
        setChatBackgroundColor={noop}
        chatHeaderColor="#000000"
        setChatHeaderColor={noop}
        chatHeaderTextColor="#111111"
        setChatHeaderTextColor={noop}
        botMessageColor="#222222"
        setBotMessageColor={noop}
        botMessageTextColor="#333333"
        setBotMessageTextColor={noop}
        userMessageColor="#444444"
        setUserMessageColor={noop}
        userMessageTextColor="#555555"
        setUserMessageTextColor={noop}
      />
    )
    const headers = screen.getAllByText('Renkler')
    expect(headers.length).toBeGreaterThan(0)
    const labels = screen.getAllByText(/(Chat Arka Plan|Header$|Header Yazı|Bot Mesaj Arka Planı|Bot Yazı|Kullanıcı Mesaj Arka Planı|Kullanıcı Yazı)/)
    expect(labels.length).toBeGreaterThanOrEqual(7)
    const colorInputs = utils.container.querySelectorAll('input[type="color"]')
    expect(colorInputs.length).toBeGreaterThanOrEqual(6)
  })

})
