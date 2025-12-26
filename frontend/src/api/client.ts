import axios from 'axios'
import { rateLimitStore } from '@/lib/rateLimit'

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080',
  timeout: 30000,
  withCredentials: true,
})

let refreshPromise: Promise<void> | null = null
let isRedirecting = false
export const _resetRedirecting = () => {
  isRedirecting = false
}

// Exported for testing - allows mocking the redirect behavior
export let _redirectToLogin = () => {
  if (typeof window !== 'undefined') {
    window.location.href = '/login'
  }
}
export const _setRedirectToLogin = (fn: () => void) => {
  _redirectToLogin = fn
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
    _redirectToLogin()
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
      try {
        if (!refreshPromise) {
          refreshPromise = axios
            .post(
              `${api.defaults.baseURL}/api/v1/auth/refresh`, 
              {}, 
              { withCredentials: true }
            )
            .then(() => {
              // Cookies set by server
            })
            .finally(() => {
              refreshPromise = null
            })
        }

        await refreshPromise
        return api(originalRequest)
      } catch {
        handleSessionExpired()
        return Promise.reject(err)
      }
    }
    return Promise.reject(err)
  },
)
