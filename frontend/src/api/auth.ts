import { api } from './client'

export const login = async (email: string, password: string) => {
  const { data } = await api.post('/api/v1/auth/login', { email, password })
  return data as { token: string; refresh_token: string }
}

export const register = async (email: string, password: string, fullName: string) => {
  const { data } = await api.post('/api/v1/auth/register', { email, password, full_name: fullName })
  return data as { token: string; refresh_token: string }
}

export const refreshToken = async (token: string) => {
  const { data } = await api.post('/api/v1/auth/refresh', { refresh_token: token })
  return data as { token: string; refresh_token: string }
}

export const logout = async (token: string) => {
  await api.post('/api/v1/auth/logout', { refresh_token: token })
}

export const protectedPing = async () => {
  const { data } = await api.get('/api/v1/protected')
  return data as { user_id: string; status: string }
}
