import { api } from './client'

export const listChatbots = async () => {
  const { data } = await api.get('/api/v1/chatbots')
  return data
}

export const createChatbot = async (payload: unknown) => {
  const { data } = await api.post('/api/v1/chatbots', payload)
  return data
}
export const updateBasicInfo = async (id: string, payload: unknown) => {
  const { data } = await api.put(`/api/v1/chatbots/${id}/basic-info`, payload)
  return data
}

export const updateAppearance = async (id: string, payload: unknown) => {
  const { data } = await api.put(`/api/v1/chatbots/${id}/appearance`, payload)
  return data
}

export const updateModelSettings = async (id: string, payload: unknown) => {
  const { data } = await api.put(`/api/v1/chatbots/${id}/model`, payload)
  return data
}

export const updateSecuritySettings = async (id: string, payload: unknown) => {
  const { data } = await api.put(`/api/v1/chatbots/${id}/security`, payload)
  return data
}

export const updateGuardrails = async (id: string, payload: unknown) => {
  const { data } = await api.put(`/api/v1/chatbots/${id}/guardrails`, payload)
  return data
}

export const updateHandoff = async (id: string, payload: unknown) => {
  const { data } = await api.put(`/api/v1/chatbots/${id}/handoff`, payload)
  return data
}

export const updateRefresh = async (id: string, payload: unknown) => {
  const { data } = await api.put(`/api/v1/chatbots/${id}/refresh`, payload)
  return data
}

export const updateScrapingConfig = async (id: string, payload: unknown) => {
  const { data } = await api.put(`/api/v1/chatbots/${id}/scraping`, payload)
  return data
}

// Deprecated: Use specific update methods above
export const updateChatbot = async (id: string, payload: unknown) => {
  const { data } = await api.put(`/api/v1/chatbots/${id}`, payload)
  return data
}
export const deleteChatbot = async (id: string | number) => {
  const { data } = await api.delete(`/api/v1/chatbots/${id}`)
  return data
}
