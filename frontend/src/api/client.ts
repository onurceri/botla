import axios from 'axios'
import { rateLimitStore } from '@/lib/rateLimit'

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080',
  timeout: 30000,
})

let refreshPromise: Promise<void> | null = null
let isRedirecting = false
export const _resetRedirecting = () => { isRedirecting = false }

// Helper to check if a stored value is valid (not undefined, null, or "undefined" string)
const isValidToken = (token: string | null | undefined): token is string => {
  return token !== null && token !== undefined && token !== 'undefined' && token !== 'null' && token.length > 0
}

// Clear tokens and redirect to login (only once)
const handleSessionExpired = () => {
  if (isRedirecting) return
  isRedirecting = true
  
  const storage = typeof window !== 'undefined' ? window.localStorage : null
  storage?.removeItem('botla_token')
  storage?.removeItem('botla_refresh_token')
  
  // Dispatch event for app-level handling (e.g., showing toast)
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('session-expired'))
  }
  
  if (!import.meta.env.VITE_E2E) {
    if (typeof window !== 'undefined') {
      // Small delay to allow toast to show before redirect
      setTimeout(() => {
        window.location.replace('/login')
      }, 100)
    }
  }
}

api.interceptors.request.use((config) => {
  const storage = typeof window !== 'undefined' ? window.localStorage : null
  const token = storage?.getItem('botla_token')
  if (isValidToken(token)) {
    config.headers = config.headers || {}
    config.headers.Authorization = `Bearer ${token}`

    const orgId = storage?.getItem('botla_last_org_id')
    if (orgId) {
      config.headers['X-Organization-ID'] = orgId
      const wsId = storage?.getItem(`botla_last_ws_id_${orgId}`)
      if (wsId) {
        config.headers['X-Workspace-ID'] = wsId
      }
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

    if (err.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true
      try {
        const storage = typeof window !== 'undefined' ? window.localStorage : null
        const refreshToken = storage?.getItem('botla_refresh_token')
        
        // Check for valid refresh token
        if (!isValidToken(refreshToken)) {
          throw new Error('No refresh token')
        }

        if (!refreshPromise) {
          refreshPromise = axios
            .post(`${api.defaults.baseURL}/api/v1/auth/refresh`, { refresh_token: refreshToken })
            .then(({ data }) => {
              // Validate tokens before storing
              if (isValidToken(data.token) && isValidToken(data.refresh_token)) {
                storage?.setItem('botla_token', data.token)
                storage?.setItem('botla_refresh_token', data.refresh_token)
              } else {
                throw new Error('Invalid tokens received from refresh')
              }
            })
            .finally(() => {
              refreshPromise = null
            })
        }

        await refreshPromise
        const newToken = storage?.getItem('botla_token')
        if (!isValidToken(newToken)) {
          throw new Error('Token refresh failed')
        }
        originalRequest.headers = originalRequest.headers || {}
        originalRequest.headers.Authorization = `Bearer ${newToken}`
        return api(originalRequest)
      } catch {
        handleSessionExpired()
        return Promise.reject(err)
      }
    }
    return Promise.reject(err)
  },
)


