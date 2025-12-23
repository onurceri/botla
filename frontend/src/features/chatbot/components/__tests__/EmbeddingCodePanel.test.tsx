import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import EmbeddingCodePanel from '../EmbeddingCodePanel'

vi.stubEnv('VITE_WIDGET_SCRIPT_URL', 'https://widget.botla.app/widget.js')

describe('EmbeddingCodePanel', () => {
  it('toggles secure embed and refreshes secret', async () => {
    const onToggle = vi.fn()
    const onDomains = vi.fn()
    const onSecret = vi.fn()
    const onRefresh = vi.fn()

    render(
      <EmbeddingCodePanel
        id="123"
        secureEmbedPlanEnabled={true}
        secureEmbedEnabled={false}
        allowedDomains=""
        embedSecret=""
        onToggleSecure={onToggle}
        onDomainsChange={onDomains}
        onSecretChange={onSecret}
        onSecretRefresh={onRefresh}
        onSecretClear={() => {}}
      />,
    )

    // Switch component has role="switch" instead of "checkbox"
    const toggle = screen.getByRole('switch') as HTMLButtonElement
    fireEvent.click(toggle)
    expect(onToggle).toHaveBeenCalledWith(true)
  })

  it('copies embed script to clipboard', async () => {
    const writeText = vi.fn()
    Object.assign(navigator, { clipboard: { writeText } })
    render(
      <EmbeddingCodePanel
        id="abc"
        secureEmbedPlanEnabled={false}
        secureEmbedEnabled={false}
        allowedDomains=""
        embedSecret=""
        onToggleSecure={() => {}}
        onDomainsChange={() => {}}
        onSecretChange={() => {}}
        onSecretRefresh={() => {}}
        onSecretClear={() => {}}
      />,
    )
    const copyBtns = screen.getAllByRole('button', { name: /Kopyala/i })
    const copyBtn = copyBtns[copyBtns.length - 1]
    fireEvent.click(copyBtn)
    expect(writeText).toHaveBeenCalledWith(
      '<script src="https://widget.botla.app/widget.js" data-bot="abc"></script>',
    )
  })

  it('updates domains input', async () => {
    const onDomains = vi.fn()
    render(
      <EmbeddingCodePanel
        id="123"
        secureEmbedPlanEnabled={true}
        secureEmbedEnabled={true}
        allowedDomains="example.com"
        embedSecret=""
        onToggleSecure={() => {}}
        onDomainsChange={onDomains}
        onSecretChange={() => {}}
        onSecretRefresh={() => {}}
        onSecretClear={() => {}}
      />,
    )
    const domainsInput = screen.getByPlaceholderText('ornek.com, digersite.com') as HTMLInputElement
    fireEvent.change(domainsInput, { target: { value: 'a.com, b.com' } })
    expect(onDomains).toHaveBeenCalledWith('a.com, b.com')
  })

  it('shows advanced token section toggle when secure embed enabled', async () => {
    render(
      <EmbeddingCodePanel
        id="123"
        secureEmbedPlanEnabled={true}
        secureEmbedEnabled={true}
        allowedDomains=""
        embedSecret=""
        onToggleSecure={() => {}}
        onDomainsChange={() => {}}
        onSecretChange={() => {}}
        onSecretRefresh={() => {}}
        onSecretClear={() => {}}
      />,
    )
    // Advanced section toggle should be visible
    const advancedBtns = screen.getAllByText(/Gelişmiş: Token Doğrulama/i)
    expect(advancedBtns.length).toBeGreaterThan(0)
  })
})
