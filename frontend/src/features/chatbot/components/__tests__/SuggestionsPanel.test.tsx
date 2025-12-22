import { describe, it, expect, vi } from 'vitest'
import { render, screen, within } from '@testing-library/react'
import SuggestionsPanel from '../SuggestionsPanel'
import userEvent from '@testing-library/user-event'
import * as React from 'react'

describe('SuggestionsPanel', () => {
  it('adds and removes questions', async () => {
    const user = userEvent.setup()
    function Harness() {
      const [enabled, setEnabled] = React.useState(true)
      const [qs, setQs] = React.useState<string[]>([])
      return (
        <SuggestionsPanel 
          suggestionsEnabled={enabled}
          setSuggestionsEnabled={setEnabled}
          suggestedQuestions={qs}
          setSuggestedQuestions={setQs as any}
          allSuggestedQuestions={[]}
        />
      )
    }
    const utils = render(<Harness />)
    const view = within(utils.container)
    const input = view.getByPlaceholderText('Soru ekle...') as HTMLInputElement
    await user.type(input, 'Nasılsın?{Enter}')
    const chip = await view.findByText('Nasılsın?')
    const removeBtn = chip.parentElement!.querySelector('button') as HTMLButtonElement
    await user.click(removeBtn)
    expect(screen.queryByText('Nasılsın?')).not.toBeInTheDocument()
  })

  it('limits to 6 and deduplicates', async () => {
    const user = userEvent.setup()
    function Harness() {
      const [enabled, setEnabled] = React.useState(true)
      const [qs, setQs] = React.useState<string[]>([])
      return (
        <SuggestionsPanel 
          suggestionsEnabled={enabled}
          setSuggestionsEnabled={setEnabled}
          suggestedQuestions={qs}
          setSuggestedQuestions={setQs as any}
          allSuggestedQuestions={[]}
        />
      )
    }
    const utils2 = render(<Harness />)
    const view2 = within(utils2.container)
    const input = view2.getAllByPlaceholderText('Soru ekle...')[0] as HTMLInputElement
    const entries = ['A?', 'B?', 'C?', 'D?', 'E?', 'F?', 'G?']
    for (const q of entries) {
      await user.type(input, `${q}{Enter}`)
    }
    expect(view2.getAllByText(/[A-Z]\?/).length).toBe(7)
    await user.type(input, `a?{Enter}`)
    const texts = view2.getAllByText(/\?$/).map(el => el.textContent?.toLowerCase())
    const uniq = new Set(texts)
    expect(uniq.size).toBe(texts.length)
  })

  it('toggle suggestionsEnabled', async () => {
    function Harness() {
      const [enabled, setEnabled] = React.useState(true)
      return (
        <SuggestionsPanel 
          suggestionsEnabled={enabled}
          setSuggestionsEnabled={setEnabled}
          suggestedQuestions={[]}
          setSuggestedQuestions={() => {}}
          allSuggestedQuestions={[]}
        />
      )
    }
    const utils3 = render(<Harness />)
    const view3 = within(utils3.container)
    const checkbox = view3.getAllByRole('switch')[0] as HTMLButtonElement
    expect(checkbox.checked).toBe(true)
    checkbox.click()
    expect(checkbox.checked).toBe(false)
  })

  it('adds via Ekle button and clears input', async () => {
    const user = userEvent.setup()
    function Harness() {
      const [enabled, setEnabled] = React.useState(true)
      const [qs, setQs] = React.useState<string[]>([])
      return (
        <SuggestionsPanel 
          suggestionsEnabled={enabled}
          setSuggestionsEnabled={setEnabled}
          suggestedQuestions={qs}
          setSuggestedQuestions={setQs as any}
          allSuggestedQuestions={[]}
        />
      )
    }
    const utils = render(<Harness />)
    const view = within(utils.container)
    const input = view.getByPlaceholderText('Soru ekle...') as HTMLInputElement
    await user.type(input, 'Merhaba?')
    const addBtn = view.getByRole('button', { name: /ekle/i })
    const origGet = document.getElementById
    ;(document as any).getElementById = vi.fn(() => input)
    await user.click(addBtn)
    ;(document as any).getElementById = origGet
    expect(await view.findByText('Merhaba?')).toBeInTheDocument()
    expect(input.value).toBe('')
  })
})
