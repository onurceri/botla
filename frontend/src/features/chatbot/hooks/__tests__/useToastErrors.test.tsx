import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { ToastProvider } from '@/components/ui/toast'
import { useToastErrors } from '../useToastErrors'

function Demo() {
  const t = useToastErrors()
  return (
    <div>
      <button onClick={() => t.success('Başarılı!')}>success</button>
      <button onClick={() => t.error('Hata!')}>error</button>
      <button onClick={() => t.info('Bilgi!')}>info</button>
    </div>
  )
}

describe('useToastErrors', () => {
  it('shows success, error, and info toasts', async () => {
    render(
      <ToastProvider>
        <Demo />
      </ToastProvider>
    )

    screen.getByText('success').click()
    expect(await screen.findByText('Başarılı!')).toBeInTheDocument()

    screen.getByText('error').click()
    expect(await screen.findByText('Hata!')).toBeInTheDocument()

    screen.getByText('info').click()
    expect(await screen.findByText('Bilgi!')).toBeInTheDocument()
  })
})

