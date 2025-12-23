import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor, cleanup } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { ToastProvider } from '@/components/ui/toast'
import OnboardingWizard from '../OnboardingWizard'
import { api } from '@/api/client'

const renderOnboarding = () => {
  return render(
    <ToastProvider>
      <MemoryRouter>
        <OnboardingWizard />
      </MemoryRouter>
    </ToastProvider>,
  )
}

// Helper to get the main action button (İleri or Botu Oluştur)
const getActionButton = () => {
  const buttons = screen.getAllByRole('button')
  return buttons.find(
    (btn) =>
      btn.textContent?.includes('İleri') ||
      btn.textContent?.includes('Botu Oluştur') ||
      btn.textContent?.includes('Botu Görüntüle'),
  )!
}

// Helper to get the back button
const getBackButton = () => {
  const buttons = screen.getAllByRole('button')
  return buttons.find((btn) => btn.textContent?.includes('Geri'))
}

describe('OnboardingWizard', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    // Mock onboarding state calls to prevent unexpected call errors
    vi.spyOn(api, 'get').mockImplementation((url) => {
      if (url === '/api/v1/me/onboarding') {
        return Promise.resolve({ data: { step: 0, data: {} } })
      }
      return Promise.reject(new Error('Unexpected GET ' + url))
    })
    vi.spyOn(api, 'put').mockImplementation((url) => {
      if (url === '/api/v1/me/onboarding') {
        return Promise.resolve({ data: { ok: true } })
      }
      return Promise.reject(new Error('Unexpected PUT ' + url))
    })
  })

  afterEach(() => {
    cleanup()
  })

  describe('Adım 1 - Bot Adlandırma', () => {
    it('ilk adımı doğru şekilde render eder', () => {
      renderOnboarding()

      expect(screen.getByText('Botunuza İsim Verin')).toBeInTheDocument()
      expect(screen.getByLabelText('Bot Adı')).toBeInTheDocument()
      expect(getActionButton()).toBeInTheDocument()
    })

    it('geçersiz isimle ilerlemeye izin vermez', async () => {
      const user = userEvent.setup()
      renderOnboarding()

      const input = screen.getByLabelText('Bot Adı')
      await user.type(input, 'A') // Tek karakter - geçersiz

      const nextButton = getActionButton()
      expect(nextButton).toBeDisabled()
    })

    it('geçerli isimle ilerlemeye izin verir', async () => {
      const user = userEvent.setup()
      renderOnboarding()

      const input = screen.getByLabelText('Bot Adı')
      await user.type(input, 'Test Bot')

      const nextButton = getActionButton()
      expect(nextButton).not.toBeDisabled()
    })
  })

  describe('Adım 2 - Bilgi Kaynağı', () => {
    it('ikinci adıma geçiş yapabilir', async () => {
      const user = userEvent.setup()
      renderOnboarding()

      // Adım 1'i tamamla
      await user.type(screen.getByLabelText('Bot Adı'), 'Test Bot')
      await user.click(getActionButton())

      // Adım 2'de olduğumuzu doğrula
      await waitFor(() => {
        expect(screen.getByText('Bilgi Kaynağı Ekleyin')).toBeInTheDocument()
      })
    })
  })

  describe('Adım 3 - Kişilik Belirleme', () => {
    it('üçüncü adıma geçiş yapabilir', async () => {
      const user = userEvent.setup()
      renderOnboarding()

      // Adım 1'i tamamla
      await user.type(screen.getByLabelText('Bot Adı'), 'Test Bot')
      await user.click(getActionButton())

      // Adım 2'yi tamamla
      await waitFor(() => {
        expect(screen.getByLabelText('İçerik')).toBeInTheDocument()
      })
      await user.type(
        screen.getByLabelText('İçerik'),
        'Bu bir test içeriğidir. Botun öğrenmesi için yeterli karakter sayısına ulaşmak gerekiyor.',
      )
      await user.click(getActionButton())

      // Adım 3'te olduğumuzu doğrula
      await waitFor(() => {
        expect(screen.getByText('Kişiliğini Belirleyin')).toBeInTheDocument()
      })
    })
  })

  describe('Bot Oluşturma', () => {
    it('başarılı bot oluşturma akışı', async () => {
      const user = userEvent.setup()

      // API çağrılarını mock'la
      vi.spyOn(api, 'post')
        .mockResolvedValueOnce({ data: { id: 'test-bot-id', name: 'Test Bot' } })
        .mockResolvedValueOnce({ data: {} })

      renderOnboarding()

      // Adım 1'i tamamla
      await user.type(screen.getByLabelText('Bot Adı'), 'Test Bot')
      await user.click(getActionButton())

      // Adım 2'yi tamamla
      await waitFor(() => {
        expect(screen.getByLabelText('İçerik')).toBeInTheDocument()
      })
      await user.type(
        screen.getByLabelText('İçerik'),
        'Bu bir test içeriğidir. Botun öğrenmesi için yeterli karakter sayısına ulaşmak gerekiyor.',
      )
      await user.click(getActionButton())

      // Adım 3'ü tamamla ve bot oluştur
      await waitFor(() => {
        expect(screen.getByText('Kişiliğini Belirleyin')).toBeInTheDocument()
      })
      await user.click(getActionButton())

      // Başarı mesajını bekle
      await waitFor(() => {
        expect(screen.getByText((content, element) => {
          return /Tebrikler|başarıyla oluşturuldu/i.test(element?.textContent || '');
        })).toBeInTheDocument()
      }, { timeout: 3000 })
    })
  })

  describe('Navigasyon', () => {
    it('geri butonu ile önceki adıma dönebilir', async () => {
      const user = userEvent.setup()
      renderOnboarding()

      // Adım 1'i tamamla ve ilerle
      await user.type(screen.getByLabelText('Bot Adı'), 'Test Bot')
      await user.click(getActionButton())

      // Adım 2'de olduğumuzu doğrula
      await waitFor(() => {
        expect(screen.getByText('Bilgi Kaynağı Ekleyin')).toBeInTheDocument()
      })

      // Geri butonuna tıkla
      const backButton = getBackButton()
      expect(backButton).toBeDefined()
      await user.click(backButton!)

      // Adım 1'e döndüğümüzü doğrula
      await waitFor(() => {
        expect(screen.getByText('Botunuza İsim Verin')).toBeInTheDocument()
      })
    })
  })
})
