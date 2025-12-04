import axios from 'axios'

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080',
  timeout: 30000,
})

let refreshPromise: Promise<void> | null = null

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('botla_token')
  if (token) {
    config.headers = config.headers || {}
    config.headers.Authorization = `Bearer ${token}`
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
        const refreshToken = localStorage.getItem('botla_refresh_token')
        if (!refreshToken) throw new Error('No refresh token')

        if (!refreshPromise) {
          refreshPromise = axios
            .post(`${api.defaults.baseURL}/api/v1/auth/refresh`, { refresh_token: refreshToken })
            .then(({ data }) => {
              localStorage.setItem('botla_token', data.token)
              localStorage.setItem('botla_refresh_token', data.refresh_token)
            })
            .finally(() => {
              refreshPromise = null
            })
        }

        await refreshPromise
        const newToken = localStorage.getItem('botla_token')
        originalRequest.headers = originalRequest.headers || {}
        originalRequest.headers.Authorization = `Bearer ${newToken}`
        return api(originalRequest)
      } catch (refreshErr) {
        localStorage.removeItem('botla_token')
        localStorage.removeItem('botla_refresh_token')
        if (!import.meta.env.VITE_E2E) {
          window.location.replace('/login')
        }
        return Promise.reject(refreshErr)
      }
    }
    return Promise.reject(err)
  },
)
