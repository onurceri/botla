import { useMutation, useQueryClient } from '@tanstack/react-query'
import { uploadPDFSource, uploadURLSource, uploadTextSource, deleteSource, refreshSource } from '@/api/source'
import { api } from '@/api/client'
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
