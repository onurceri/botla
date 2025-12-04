import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import IdentitySection from '../IdentitySection'

describe('IdentitySection', () => {
  it('toggles and updates fields', () => {
    const onToggle = vi.fn()
    const setName = vi.fn()
    const setIcon = vi.fn()
    const setWelcome = vi.fn()
    render(
      <IdentitySection
        isExpanded={true}
        onToggle={onToggle}
        botDisplayName="Bot"
        setBotDisplayName={setName}
        botIcon=""
        setBotIcon={setIcon}
        welcomeMessage="Selam"
        setWelcomeMessage={setWelcome}
      />
    )
    const nameInput = screen.getByLabelText('Bot Görünen Adı') as HTMLInputElement
    fireEvent.change(nameInput, { target: { value: 'Destek' } })
    expect(setName).toHaveBeenCalledWith('Destek')
    const toggleBtn = screen.getByRole('button', { name: /Kimlik/i })
    fireEvent.click(toggleBtn)
    expect(onToggle).toHaveBeenCalled()
    const iconInput = screen.getByLabelText('Bot İkon URL') as HTMLInputElement
    fireEvent.change(iconInput, { target: { value: 'https://img.png' } })
    expect(setIcon).toHaveBeenCalledWith('https://img.png')
    const welcomeInput = screen.getByLabelText('Karşılama Mesajı') as HTMLInputElement
    fireEvent.change(welcomeInput, { target: { value: 'Hoş Geldiniz' } })
    expect(setWelcome).toHaveBeenCalledWith('Hoş Geldiniz')
  })
})
