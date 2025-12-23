import { describe, it, expect, afterEach } from 'vitest'
import * as React from 'react'
import { render, screen, cleanup } from '@testing-library/react'
import { ToastProvider, useToast } from '../toast'

function AutoFire() {
  const { toast } = useToast()
  React.useEffect(() => {
    toast('Mesaj', 'info', 500)
  }, [toast])
  return null
}

describe('Toast component', () => {
  afterEach(() => {
    cleanup()
  })
  it('auto-dismisses after duration', async () => {
    render(
      <ToastProvider>
        <AutoFire />
      </ToastProvider>,
    )
    expect(await screen.findByText('Mesaj')).toBeInTheDocument()
    await new Promise((r) => setTimeout(r, 600))
    expect(screen.queryByText('Mesaj')).not.toBeInTheDocument()
  })

  it('dismisses when close button clicked', async () => {
    render(
      <ToastProvider>
        <AutoFire />
      </ToastProvider>,
    )
    const toastText = await screen.findByText('Mesaj')
    const closeBtn = toastText.parentElement!.querySelector('button') as HTMLButtonElement
    closeBtn.click()
    await new Promise((r) => setTimeout(r, 0))
    expect(screen.queryByText('Mesaj')).not.toBeInTheDocument()
  })
})
