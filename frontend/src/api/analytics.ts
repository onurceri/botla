import { api } from './client'

export const getAnalytics = async () => {
  const { data } = await api.get('/api/v1/analytics')
  return data
}

export const getChatbotAnalyticsOverview = async (chatbotId: string) => {
  const { data } = await api.get(`/api/v1/chatbots/${chatbotId}/analytics/overview`)
  return data
}

export const getChatbotAnalyticsTrends = async (chatbotId: string, days = 30) => {
  const { data } = await api.get(`/api/v1/chatbots/${chatbotId}/analytics/trends?days=${days}`)
  return data
}

export const getSourceUsageStats = async (chatbotId: string, days = 30) => {
  const { data } = await api.get(`/api/v1/chatbots/${chatbotId}/analytics/sources?days=${days}`)
  return data
}
