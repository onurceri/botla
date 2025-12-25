import { ReactElement, ReactNode, useMemo } from 'react'
import { render, RenderOptions } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { OrganizationProvider } from '@/features/organization/context/OrganizationContext'
import { ToastProvider } from '@/components/ui/toast'
import { vi } from 'vitest'

// Mock organization APIs globally for tests
vi.mock('@/api/organization', () => ({
  getOrganizations: vi.fn().mockResolvedValue([{ id: 'org1', name: 'Test Org' }]),
  createOrganization: vi.fn(),
  updateOrganization: vi.fn(),
  deleteOrganization: vi.fn(),
}))

vi.mock('@/api/workspace', () => ({
  getWorkspaces: vi
    .fn()
    .mockResolvedValue([{ id: 'ws1', name: 'Test Workspace', organization_id: 'org1' }]),
  createWorkspace: vi.fn(),
  updateWorkspace: vi.fn(),
  deleteWorkspace: vi.fn(),
}))

/**
 * Create a test QueryClient with disabled retries for faster tests
 */
export function createTestQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: false, // Don't retry in tests
        gcTime: Infinity, // Keep cache for test duration
      },
      mutations: {
        retry: false,
      },
    },
  })
}

/**
 * Wrapper component that provides QueryClient and OrganizationProvider for tests
 */
export function QueryWrapper({ children }: { children: ReactNode }) {
  const queryClient = useMemo(() => createTestQueryClient(), [])
  return (
    <QueryClientProvider client={queryClient}>
      <ToastProvider>
        <OrganizationProvider>{children}</OrganizationProvider>
      </ToastProvider>
    </QueryClientProvider>
  )
}

/**
 * Custom render that includes QueryClientProvider
 */
export function renderWithQuery(ui: ReactElement, options?: Omit<RenderOptions, 'wrapper'>) {
  return render(ui, { wrapper: QueryWrapper, ...options })
}

/**
 * Re-export everything from testing-library
 */
export * from '@testing-library/react'
