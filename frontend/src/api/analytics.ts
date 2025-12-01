import { api } from './client'

export const getAnalytics = async () => {
  const { data } = await api.get('/api/v1/analytics')
  return data
}
