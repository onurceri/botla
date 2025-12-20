import { useMutation, useQueryClient } from '@tanstack/react-query'
import { uploadPDFSource, uploadURLSource, uploadTextSource, deleteSource, refreshSource } from '@/api/source'
import { api } from '@/api/client'
import {
  updateBasicInfo,
  updateAppearance,
  updateModelSettings,
  updateSecuritySettings,
  updateGuardrails,
  updateHandoff,
  updateRefresh,
  updateScrapingConfig
} from '@/api/chatbot'
import { CHATBOT_QUERY_KEY } from '../queries/useChatbot'
import { SOURCES_QUERY_KEY } from '../queries/useSources'

export function useUploadSource(chatbotId: string) {
  const queryClient = useQueryClient()

  const invalidateQueries = () => {
    // Invalidate sources list
    queryClient.invalidateQueries({ queryKey: SOURCES_QUERY_KEY(chatbotId) })
    // Invalidate chatbot (this refreshes suggestions!)
    queryClient.invalidateQueries({ queryKey: CHATBOT_QUERY_KEY(chatbotId) })
  }

  return {
    uploadPDF: useMutation({
      mutationFn: (file: File) => uploadPDFSource(chatbotId, file),
      onSuccess: invalidateQueries,
    }),
    uploadURL: useMutation({
      mutationFn: (url: string) => uploadURLSource(chatbotId, url),
      onSuccess: invalidateQueries,
    }),
    uploadText: useMutation({
      mutationFn: (text: string) => uploadTextSource(chatbotId, text),
      onSuccess: invalidateQueries,
    }),
  }
}

export function useDeleteSource(chatbotId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (sourceId: string) => deleteSource(sourceId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: SOURCES_QUERY_KEY(chatbotId) })
      queryClient.invalidateQueries({ queryKey: CHATBOT_QUERY_KEY(chatbotId) })
    },
  })
}

export function useRefreshSource(chatbotId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (sourceId: string) => refreshSource(sourceId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: SOURCES_QUERY_KEY(chatbotId) })
    },
  })
}

export function useRegenerateSuggestions(chatbotId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async () => {
      await api.post(`/api/v1/chatbots/${chatbotId}/suggestions/regenerate`)
      // Wait for backend processing
      await new Promise(resolve => setTimeout(resolve, 2000))
    },
    onSuccess: () => {
      // Invalidate chatbot to refetch suggestions
      queryClient.invalidateQueries({ queryKey: CHATBOT_QUERY_KEY(chatbotId) })
    },
  })
}

export function useUpdateChatbot(chatbotId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (payload: any) => {
      const { data } = await api.put(`/api/v1/chatbots/${chatbotId}`, payload)
      return data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: CHATBOT_QUERY_KEY(chatbotId) })
    },
  })
}

export function useUpdateBasicInfo(chatbotId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (payload: any) => updateBasicInfo(chatbotId, payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: CHATBOT_QUERY_KEY(chatbotId) }),
  })
}

export function useUpdateAppearance(chatbotId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (payload: any) => updateAppearance(chatbotId, payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: CHATBOT_QUERY_KEY(chatbotId) }),
  })
}

export function useUpdateModelSettings(chatbotId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (payload: any) => updateModelSettings(chatbotId, payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: CHATBOT_QUERY_KEY(chatbotId) }),
  })
}

export function useUpdateSecuritySettings(chatbotId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (payload: any) => updateSecuritySettings(chatbotId, payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: CHATBOT_QUERY_KEY(chatbotId) }),
  })
}

export function useUpdateGuardrails(chatbotId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (payload: any) => updateGuardrails(chatbotId, payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: CHATBOT_QUERY_KEY(chatbotId) }),
  })
}

export function useUpdateHandoff(chatbotId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (payload: any) => updateHandoff(chatbotId, payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: CHATBOT_QUERY_KEY(chatbotId) }),
  })
}

export function useUpdateRefresh(chatbotId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (payload: any) => updateRefresh(chatbotId, payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: CHATBOT_QUERY_KEY(chatbotId) }),
  })
}

export function useUpdateScrapingConfig(chatbotId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (payload: any) => updateScrapingConfig(chatbotId, payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: CHATBOT_QUERY_KEY(chatbotId) }),
  })
}
