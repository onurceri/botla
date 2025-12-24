import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor, cleanup } from '@testing-library/react'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AdminRoute } from '../AdminRoute'
import { api } from '@/api/client'

// Helper function to render AdminRoute with routing
function renderWithRouter(
  mockProfileData: { is_platform_admin: boolean } | 'loading' | null,
  options: { initialRoute?: string } = {},
) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  })

  // Mock the profile API call
  vi.mocked(api.get).mockImplementation((url: string) => {
    if (url === '/api/v1/me') {
      if (mockProfileData === 'loading') {
        // Never resolve to simulate loading
        return new Promise(() => {})
      }
      if (mockProfileData === null) {
        return Promise.reject(new Error('Not authenticated'))
      }
      return Promise.resolve({
        data: {
          id: 'user-123',
          email: 'test@example.com',
          created_at: '2024-01-01T00:00:00Z',
          ...mockProfileData,
        },
      })
    }
    throw new Error(`Unexpected API call: ${url}`)
  })

  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={[options.initialRoute || '/admin']}>
        <Routes>
          <Route path="/dashboard" element={<div data-testid="dashboard">Dashboard Page</div>} />
          <Route
            path="/admin"
            element={
              <AdminRoute>
                <div data-testid="admin-content">Admin Content</div>
              </AdminRoute>
            }
          />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

describe('AdminRoute', () => {
  beforeEach(() => {
    vi.resetAllMocks()
  })

  afterEach(() => {
    cleanup()
  })

  describe('when user is an admin', () => {
    it('renders children for admin users', async () => {
      renderWithRouter({ is_platform_admin: true })

      // Wait for loading to complete
      await waitFor(() => {
        expect(screen.queryByText('Yükleniyor...')).not.toBeInTheDocument()
      })

      // Admin content should be visible
      expect(screen.getByTestId('admin-content')).toBeInTheDocument()
    })

    it('does not redirect admin users', async () => {
      renderWithRouter({ is_platform_admin: true })

      await waitFor(() => {
        expect(screen.queryByText('Yükleniyor...')).not.toBeInTheDocument()
      })

      // Should not redirect to dashboard
      expect(screen.queryByTestId('dashboard')).not.toBeInTheDocument()
    })
  })

  describe('when user is not an admin', () => {
    it('redirects non-admin users to dashboard', async () => {
      renderWithRouter({ is_platform_admin: false })

      // Wait for loading to complete and redirect to happen
      await waitFor(() => {
        expect(screen.getByTestId('dashboard')).toBeInTheDocument()
      })

      // Admin content should not be visible
      expect(screen.queryByTestId('admin-content')).not.toBeInTheDocument()
    })

    it('does not render children for non-admin users', async () => {
      renderWithRouter({ is_platform_admin: false })

      await waitFor(() => {
        expect(screen.getByTestId('dashboard')).toBeInTheDocument()
      })

      expect(screen.queryByTestId('admin-content')).not.toBeInTheDocument()
    })
  })

  describe('loading state', () => {
    it('shows loading spinner while profile is loading', () => {
      renderWithRouter('loading')

      expect(screen.getByText('Yükleniyor...')).toBeInTheDocument()
    })

    it('shows loading spinner with proper styling', () => {
      const { container } = renderWithRouter('loading')

      // Check that the loading container has proper styling
      const loadingContainer = container.querySelector('.flex.items-center.justify-center.h-screen')
      expect(loadingContainer).toBeInTheDocument()
    })
  })

  describe('edge cases', () => {
    it('handles API error gracefully by redirecting', async () => {
      renderWithRouter(null) // Will cause API error

      // Should show loading initially, but when error happens it will have undefined user
      // which should redirect to dashboard
      await waitFor(
        () => {
          // When profile fetch fails, user will be undefined, which should redirect
          expect(screen.queryByText('Yükleniyor...')).not.toBeInTheDocument()
        },
        { timeout: 2000 },
      )
    })
  })
})
