import { api } from './client'

export const uploadPDFSource = async (chatbotId: string, file: File) => {
  const form = new FormData()
  form.append('source_type', 'pdf')
  form.append('file', file)
  const { data } = await api.post(`/api/v1/chatbots/${chatbotId}/sources`, form)
  return data as { id: string }
}

export const uploadTextSource = async (chatbotId: string, text: string) => {
  const form = new FormData()
  form.append('source_type', 'text')
  form.append('text', text)
  const { data } = await api.post(`/api/v1/chatbots/${chatbotId}/sources`, form)
  return data as { id: string }
}

export const uploadURLSource = async (chatbotId: string, url: string) => {
  const form = new FormData()
  form.append('source_type', 'url')
  form.append('source_url', url)
  const { data } = await api.post(`/api/v1/chatbots/${chatbotId}/sources`, form)
  return data as { id: string }
}

export const getSourceStatus = async (sourceId: string) => {
  const { data } = await api.get(`/api/v1/sources/${sourceId}`)
  return data
}

export const listSources = async (chatbotId: string) => {
  const { data } = await api.get(`/api/v1/chatbots/${chatbotId}/sources`)
  return data
}

export const deleteSource = async (sourceId: string) => {
  const { data } = await api.delete(`/api/v1/sources/${sourceId}`)
  return data
}
