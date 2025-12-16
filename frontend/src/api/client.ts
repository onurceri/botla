import axios from 'axios'

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080',
  timeout: 30000,
})

let refreshPromise: Promise<void> | null = null

api.interceptors.request.use((config) => {
  const storage = typeof window !== 'undefined' ? window.localStorage : null
  const token = storage?.getItem('botla_token')
  if (token) {
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
  (res) => res,
  async (err) => {
    const originalRequest = err.config
    if (err.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true
      try {
        const storage = typeof window !== 'undefined' ? window.localStorage : null
        const refreshToken = storage?.getItem('botla_refresh_token')
        if (!refreshToken) throw new Error('No refresh token')

        if (!refreshPromise) {
          refreshPromise = axios
            .post(`${api.defaults.baseURL}/api/v1/auth/refresh`, { refresh_token: refreshToken })
            .then(({ data }) => {
              storage?.setItem('botla_token', data.token)
              storage?.setItem('botla_refresh_token', data.refresh_token)
            })
            .finally(() => {
              refreshPromise = null
            })
        }

        await refreshPromise
        const newToken = storage?.getItem('botla_token')
        originalRequest.headers = originalRequest.headers || {}
        originalRequest.headers.Authorization = `Bearer ${newToken}`
        return api(originalRequest)
      } catch (refreshErr) {
        const storage = typeof window !== 'undefined' ? window.localStorage : null
        storage?.removeItem('botla_token')
        storage?.removeItem('botla_refresh_token')
        if (!import.meta.env.VITE_E2E) {
          if (typeof window !== 'undefined') {
            window.location.replace('/login')
          }
        }
        return Promise.reject(refreshErr)
      }
    }
    return Promise.reject(err)
  },
)
