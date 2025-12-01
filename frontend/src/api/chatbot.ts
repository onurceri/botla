import { api } from './client'

export const listChatbots = async () => {
  const { data } = await api.get('/api/v1/chatbots')
  return data
}

export const createChatbot = async (payload: unknown) => {
  const { data } = await api.post('/api/v1/chatbots', payload)
  return data
}
