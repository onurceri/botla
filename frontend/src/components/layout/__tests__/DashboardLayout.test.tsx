import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import DashboardLayout from '../DashboardLayout'
import { Button } from '@/components/ui/button'
import App from '@/App'

describe('DashboardLayout', () => {
  beforeEach(() => {
    Object.defineProperty(window, 'localStorage', {
      value: {
        getItem: vi.fn(),
        setItem: vi.fn(),
        removeItem: vi.fn(),
      },
      writable: true,
    })
  })

  it('toggles sidebar mode and persists to localStorage', async () => {
    vi.spyOn(window.localStorage, 'getItem').mockReturnValue('pinned')
    const setSpy = vi.spyOn(window.localStorage, 'setItem')

    render(
      <MemoryRouter initialEntries={["/"]}>
        <Routes>
          <Route path="/" element={<DashboardLayout />}> 
            <Route index element={<Button>İçerik</Button>} />
          </Route>
        </Routes>
      </MemoryRouter>
    )

    const toggle = screen.getByTitle('Sabit → Hover')
    toggle.click()
    expect(setSpy).toHaveBeenCalledWith('botla_sidebar_mode', 'hover')
  })

  it('logs out and navigates to login', async () => {
    vi.spyOn(window.localStorage, 'getItem').mockImplementation((key: string) => {
      if (key === 'botla_token') return 'tok'
      return null
    })
    render(<App />)
    const logoutBtns = screen.getAllByRole('button', { name: 'Logout' })
    logoutBtns[logoutBtns.length - 1].click()
    expect(await screen.findByRole('heading', { name: 'Hoş Geldiniz' })).toBeInTheDocument()
  })

  it('opens mobile menu overlay and closes on click', async () => {
    vi.spyOn(window.localStorage, 'getItem').mockReturnValue('pinned')
    const { container } = render(
      <MemoryRouter initialEntries={["/"]}>
        <Routes>
          <Route path="/" element={<DashboardLayout />}> 
            <Route index element={<Button>İçerik</Button>} />
          </Route>
        </Routes>
      </MemoryRouter>
    )
    const header = container.querySelector('header')!
    const menuBtn = header.querySelector('button') as HTMLButtonElement
    menuBtn.click()
    await new Promise((r) => setTimeout(r, 0))
    const overlay = container.querySelector('.fixed.inset-0') as HTMLDivElement
    expect(overlay).not.toBeNull()
    overlay.click()
    await new Promise((r) => setTimeout(r, 0))
    expect(container.querySelector('.fixed.inset-0')).toBeNull()
  })

  it('shows breadcrumb label for Chatbots route', async () => {
    vi.spyOn(window.localStorage, 'getItem').mockReturnValue('pinned')
    const { container } = render(
      <MemoryRouter initialEntries={["/chatbots"]}>
        <Routes>
          <Route path="/" element={<DashboardLayout />}> 
            <Route path="/chatbots" element={<Button>List</Button>} />
          </Route>
        </Routes>
      </MemoryRouter>
    )
    const header2 = container.querySelector('header')!
    const label = header2.querySelector('.text-foreground.font-medium') as HTMLElement
    expect(label.textContent).toBe('Chatbots')
  })
})
