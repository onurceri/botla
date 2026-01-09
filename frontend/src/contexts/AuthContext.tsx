import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { api, _setWasAuthenticated } from '@/api/client'
import type { User } from '@/types/user'

/**
 * Authentication Context
 * 
 * Provides a global auth state based on cookie-based authentication.
 * Uses React Query to fetch and cache user data from /api/v1/me.
 * 
 * ## Architecture
 * - Backend sets HttpOnly cookies (`botla_token`, `botla_refresh_token`)
 * - Frontend uses `withCredentials: true` to automatically include cookies
 * - Auth state is determined by the API response, not localStorage
 * 
 * ## Usage
 * ```tsx
 * function MyComponent() {
 *   const { user, isLoading, isAuthenticated, logout } = useAuth()
 *   
 *   if (isLoading) return <Spinner />
 *   if (!isAuthenticated) return <Navigate to="/login" />
 *   
 *   return <div>Hello, {user.name}!</div>
 * }
 * ```
 */

export const AUTH_QUERY_KEY = ['auth', 'user'] as const

interface AuthContextValue {
  user: User | null
  isLoading: boolean
  isAuthenticated: boolean
  error: Error | null
  logout: () => Promise<void>
  refetch: () => void
}

const AuthContext = createContext<AuthContextValue | null>(null)

/**
 * Fetch current user from API
 * This is the single source of truth for auth state
 */
async function fetchCurrentUser(): Promise<User | null> {
  try {
    const { data } = await api.get<User>('/api/v1/me')
    return data
  } catch (error: unknown) {
    // 401 means not authenticated - this is expected, not an error
    if (error && typeof error === 'object' && 'response' in error) {
      const axiosError = error as { response?: { status?: number } }
      if (axiosError.response?.status === 401) {
        return null
      }
    }
    throw error
  }
}

interface AuthProviderProps {
  children: ReactNode
}

export function AuthProvider({ children }: AuthProviderProps) {
  const queryClient = useQueryClient()
  const [isInitialized, setIsInitialized] = useState(false)

  const {
    data: user,
    isLoading: queryLoading,
    error,
    refetch,
  } = useQuery<User | null>({
    queryKey: AUTH_QUERY_KEY,
    queryFn: fetchCurrentUser,
    staleTime: 1000 * 60 * 5, // 5 minutes
    retry: false, // Don't retry on 401
    refetchOnWindowFocus: true,
    refetchOnReconnect: true,
  })

  // Mark as initialized once first query completes
  useEffect(() => {
    if (!queryLoading) {
      setIsInitialized(true)
    }
  }, [queryLoading])

  // Track when user becomes authenticated
  // This is used to determine if session expiry should be shown
  useEffect(() => {
    if (user) {
      _setWasAuthenticated(true)
    }
  }, [user])

  const logout = async () => {
    try {
      await api.post('/api/v1/auth/logout')
    } catch {
      // Ignore logout errors - cookies will be cleared by server
    } finally {
      // Clear cached auth data
      queryClient.setQueryData(AUTH_QUERY_KEY, null)
      queryClient.invalidateQueries({ queryKey: AUTH_QUERY_KEY })
      
      // Clear any legacy localStorage data
      if (typeof window !== 'undefined') {
        localStorage.removeItem('botla_token')
        localStorage.removeItem('botla_refresh_token')
        localStorage.removeItem('botla_user')
      }
    }
  }

  // Show loading during initial auth check
  const isLoading = !isInitialized

  const value: AuthContextValue = {
    user: user ?? null,
    isLoading,
    isAuthenticated: !!user,
    error: error as Error | null,
    logout,
    refetch: () => { refetch() },
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

/**
 * Hook to access auth context
 * Must be used within AuthProvider
 */
export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}

/**
 * Hook to check if user is a platform admin
 */
export function useIsAdmin(): boolean {
  const { user } = useAuth()
  return user?.is_platform_admin ?? false
}
