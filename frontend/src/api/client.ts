import axios from 'axios'
import { rateLimitStore } from '@/lib/rateLimit'

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080',
  timeout: 30000,
  withCredentials: true,
})

let refreshPromise: Promise<void> | null = null
let isRedirecting = false

// Queue for failed requests waiting for token refresh
interface FailedRequest {
  resolve: (value: unknown) => void
  reject: (reason: unknown) => void
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  config: any
}

let failedQueue: FailedRequest[] = []

/**
 * Process the failed request queue after token refresh
 * @param error - If provided, reject all requests with this error. If null, retry all requests.
 */
const processQueue = (error: Error | null = null) => {
  failedQueue.forEach((request) => {
    if (error) {
      request.reject(error)
    } else {
      // Retry the request - cookies will automatically include the new token
      request.resolve(api(request.config))
    }
  })
  failedQueue = []
}

/**
 * Singleton service to manage redirect behavior.
 * This pattern provides better encapsulation and test isolation compared to mutable globals.
 */
class RedirectService {
  private static instance: RedirectService | null = null
  private redirectFn: () => void
  
  // Store the default function for reset capability
  private readonly defaultRedirectFn = () => {
    if (typeof window !== 'undefined') {
      window.location.href = '/login'
    }
  }
  
  private constructor() {
    this.redirectFn = this.defaultRedirectFn
  }
  
  static getInstance(): RedirectService {
    if (!RedirectService.instance) {
      RedirectService.instance = new RedirectService()
    }
    return RedirectService.instance
  }
  
  /**
   * Set a custom redirect function (useful for testing)
   */
  setRedirectFn(fn: () => void): void {
    this.redirectFn = fn
  }
  
  /**
   * Execute the redirect function
   */
  redirect(): void {
    this.redirectFn()
  }
  
  /**
   * Reset redirect function to default (call in test teardown)
   */
  reset(): void {
    this.redirectFn = this.defaultRedirectFn
    isRedirecting = false
  }
  
  /**
   * Reset the singleton instance (for complete test isolation)
   */
  static resetInstance(): void {
    if (RedirectService.instance) {
      RedirectService.instance.redirectFn = RedirectService.instance.defaultRedirectFn
    }
    isRedirecting = false
  }
}

// Export the singleton instance
export const redirectService = RedirectService.getInstance()

// Legacy exports for backward compatibility (deprecated, prefer redirectService)
export const _resetRedirecting = () => {
  isRedirecting = false
}

// Reset refresh state for test isolation
export const _resetRefreshState = () => {
  failedQueue = []
  refreshPromise = null
  isRedirecting = false
}

export const _setRedirectToLogin = (fn: () => void) => {
  redirectService.setRedirectFn(fn)
}

export const _redirectToLogin = () => {
  redirectService.redirect()
}

// Helper to check if a stored value is valid (not undefined, null, or "undefined" string)
// const isValidToken = ... (removed as we use cookies now)

// Clear tokens and redirect to login (only once)
const handleSessionExpired = () => {
  if (isRedirecting) return
  isRedirecting = true

  // Cookies are HttpOnly and cannot be cleared by JS.
  // We rely on the server to clear them on /logout or they will expire.
  // We can try to call logout endpoint here, but session expired usually means token is already invalid.
  
  // Dispatch event for app-level handling (e.g., showing toast)
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('session-expired'))
  }

  // Skip redirect only in E2E mode
  const isE2E = import.meta.env.VITE_E2E === '1' || import.meta.env.VITE_E2E === 'true'
  if (isE2E) {
    return
  }

  // Small delay to allow toast to show before redirect
  setTimeout(() => {
    redirectService.redirect()
  }, 1500)
}

api.interceptors.request.use((config) => {
  const storage = typeof window !== 'undefined' ? window.localStorage : null
  // Token is handled via HttpOnly cookies automatically
  
  config.headers = config.headers || {}
  
  const orgId = storage?.getItem('botla_last_org_id')
  if (orgId) {
    config.headers['X-Organization-ID'] = orgId
    const wsId = storage?.getItem(`botla_last_ws_id_${orgId}`)
    if (wsId) {
      config.headers['X-Workspace-ID'] = wsId
    }
  }
  return config
})

api.interceptors.response.use(
  (res) => {
    rateLimitStore.updateFromHeaders(res.headers)
    return res
  },
  async (err) => {
    const originalRequest = err.config

    // Capture rate limits from error responses too
    if (err.response?.headers) {
      rateLimitStore.updateFromHeaders(err.response.headers)
    }

    // If already redirecting, just reject without further processing
    if (isRedirecting) {
      return Promise.reject(err)
    }

    // Skip token refresh logic for auth endpoints - 401 on these means invalid credentials, not session expiry
    const authEndpoints = ['/api/v1/auth/login', '/api/v1/auth/register', '/api/v1/auth/refresh']
    const isAuthEndpoint = authEndpoints.some((endpoint) => originalRequest?.url?.includes(endpoint))
    if (isAuthEndpoint) {
      return Promise.reject(err)
    }

    if (err.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true
      
      // If a refresh is already in progress, queue this request
      if (refreshPromise) {
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject, config: originalRequest })
        })
      }

      // Start the refresh process
      refreshPromise = axios
        .post(
          `${api.defaults.baseURL}/api/v1/auth/refresh`, 
          {}, 
          { withCredentials: true }
        )
        .then(() => {
          // Refresh successful - process queued requests
          processQueue(null)
        })
        .catch((refreshError) => {
          // Refresh failed - reject all queued requests
          processQueue(refreshError)
          handleSessionExpired()
          throw refreshError
        })
        .finally(() => {
          refreshPromise = null
        })

      try {
        await refreshPromise
        // Retry the original request that triggered the refresh
        return api(originalRequest)
      } catch {
        return Promise.reject(err)
      }
    }
    return Promise.reject(err)
  },
)
