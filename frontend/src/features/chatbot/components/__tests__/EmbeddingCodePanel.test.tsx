import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import EmbeddingCodePanel from '../EmbeddingCodePanel'

describe('EmbeddingCodePanel', () => {
  it('toggles secure embed and refreshes secret', async () => {
    const onToggle = vi.fn()
    const onDomains = vi.fn()
    const onSecret = vi.fn()
    const onRefresh = vi.fn()

    render(
      <EmbeddingCodePanel
        id="123"
        userPlan="pro"
        secureEmbedEnabled={false}
        allowedDomains=""
        embedSecret=""
        onToggleSecure={onToggle}
        onDomainsChange={onDomains}
        onSecretChange={onSecret}
        onSecretRefresh={onRefresh}
      />
    )

    const checkbox = screen.getByRole('checkbox') as HTMLInputElement
    fireEvent.click(checkbox)
    expect(onToggle).toHaveBeenCalledWith(true)
  })

  it('copies embed script to clipboard', async () => {
    const writeText = vi.fn()
    Object.assign(navigator, { clipboard: { writeText } })
    render(
      <EmbeddingCodePanel
        id="abc"
        userPlan="free"
        secureEmbedEnabled={false}
        allowedDomains=""
        embedSecret=""
        onToggleSecure={() => {}}
        onDomainsChange={() => {}}
        onSecretChange={() => {}}
        onSecretRefresh={() => {}}
      />
    )
    const copyBtns = screen.getAllByRole('button', { name: /Kopyala/i })
    const copyBtn = copyBtns[copyBtns.length - 1]
    fireEvent.click(copyBtn)
    expect(writeText).toHaveBeenCalledWith('<script src="https://cdn.botla.co/widget.js" data-bot="abc"></script>')
  })

  it('updates domains and secret and refreshes', async () => {
    const onToggle = vi.fn()
    const onDomains = vi.fn()
    const onSecret = vi.fn()
    const onRefresh = vi.fn()
    render(
      <EmbeddingCodePanel
        id="123"
        userPlan="pro"
        secureEmbedEnabled={true}
        allowedDomains="example.com"
        embedSecret="secret"
        onToggleSecure={onToggle}
        onDomainsChange={onDomains}
        onSecretChange={onSecret}
        onSecretRefresh={onRefresh}
      />
    )
    const domainsInput = screen.getByPlaceholderText('example.com, another.com') as HTMLInputElement
    fireEvent.change(domainsInput, { target: { value: 'a.com, b.com' } })
    expect(onDomains).toHaveBeenCalledWith('a.com, b.com')
    const secretInput = screen.getByPlaceholderText('Gizli anahtar') as HTMLInputElement
    fireEvent.change(secretInput, { target: { value: 'new' } })
    expect(onSecret).toHaveBeenCalledWith('new')
    const refreshBtn = screen.getByRole('button', { name: 'Yenile' })
    fireEvent.click(refreshBtn)
    expect(onRefresh).toHaveBeenCalled()
  })
})
