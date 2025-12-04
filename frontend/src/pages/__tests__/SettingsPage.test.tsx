import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { ToastProvider } from '@/components/ui/toast'
import SettingsPage from '../SettingsPage'

describe('SettingsPage', () => {
  it('renders static profile information', () => {
    render(
      <ToastProvider>
        <MemoryRouter>
          <SettingsPage />
        </MemoryRouter>
      </ToastProvider>
    )

    expect(screen.getByRole('heading', { name: 'Ayarlar' })).toBeInTheDocument()
    expect(screen.getByText('Profil Bilgileri')).toBeInTheDocument()
    expect(screen.getByDisplayValue('Onur Ceri')).toBeDisabled()
    expect(screen.getByDisplayValue('onur@example.com')).toBeDisabled()
  })
})

