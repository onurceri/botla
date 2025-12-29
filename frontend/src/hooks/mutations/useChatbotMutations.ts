import { useMutation, useQueryClient } from '@tanstack/react-query'
import {
  uploadPDFSource,
  uploadURLSource,
  uploadTextSource,
  deleteSource,
  refreshSource,
} from '@/api/source'
import { api } from '@/api/client'
import {
  updateBasicInfo,
  updateAppearance,
  updateModelSettings,
  updateSecuritySettings,
  updateGuardrails,
  updateHandoff,
  updateRefresh,
  updateScrapingConfig,
  deleteChatbot,
  createChatbot,
  getSuggestionJobStatus,
} from '@/api/chatbot'
import type {
  ChatbotUpdateRequest,
  BasicInfoRequest,
  AppearanceRequest,
  ModelSettingsRequest,
  SecuritySettingsRequest,
  GuardrailsRequest,
  HandoffRequest,
  RefreshRequest,
  ScrapingConfigRequest,
  CreateChatbotRequest,
} from '@/types/chatbot'
import { CHATBOTS_QUERY_KEY } from '../queries/useChatbots'
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

  const pollJobStatus = async (): Promise<void> => {
    const maxAttempts = 60
    const pollInterval = 1000
    let attempts = 0

    while (attempts < maxAttempts) {
      const status = await getSuggestionJobStatus(chatbotId)

      if (status.status === 'completed') {
        return
      }

      if (status.status === 'failed') {
        throw new Error(status.error_message || 'Suggestion regeneration failed')
      }

      await new Promise((resolve) => setTimeout(resolve, pollInterval))
      attempts++
    }

    throw new Error('Suggestion regeneration timed out')
  }

  return useMutation({
    mutationFn: async () => {
      await api.post(`/api/v1/chatbots/${chatbotId}/suggestions/regenerate`)
      await pollJobStatus()
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: CHATBOT_QUERY_KEY(chatbotId) })
    },
  })
}

export function useUpdateChatbot(chatbotId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (payload: ChatbotUpdateRequest) => {
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
    mutationFn: (payload: BasicInfoRequest) => updateBasicInfo(chatbotId, payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: CHATBOT_QUERY_KEY(chatbotId) }),
  })
}

export function useUpdateAppearance(chatbotId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (payload: AppearanceRequest) => updateAppearance(chatbotId, payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: CHATBOT_QUERY_KEY(chatbotId) }),
  })
}

export function useUpdateModelSettings(chatbotId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (payload: ModelSettingsRequest) => updateModelSettings(chatbotId, payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: CHATBOT_QUERY_KEY(chatbotId) }),
  })
}

export function useUpdateSecuritySettings(chatbotId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (payload: SecuritySettingsRequest) => updateSecuritySettings(chatbotId, payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: CHATBOT_QUERY_KEY(chatbotId) }),
  })
}

export function useUpdateGuardrails(chatbotId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (payload: GuardrailsRequest) => updateGuardrails(chatbotId, payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: CHATBOT_QUERY_KEY(chatbotId) }),
  })
}

export function useUpdateHandoff(chatbotId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (payload: HandoffRequest) => updateHandoff(chatbotId, payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: CHATBOT_QUERY_KEY(chatbotId) }),
  })
}

export function useUpdateRefresh(chatbotId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (payload: RefreshRequest) => updateRefresh(chatbotId, payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: CHATBOT_QUERY_KEY(chatbotId) }),
  })
}

export function useUpdateScrapingConfig(chatbotId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (payload: ScrapingConfigRequest) => updateScrapingConfig(chatbotId, payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: CHATBOT_QUERY_KEY(chatbotId) }),
  })
}

export function useDeleteChatbot() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string | number) => deleteChatbot(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: CHATBOTS_QUERY_KEY })
    },
  })
}

export function useCreateChatbot() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (payload: CreateChatbotRequest) => createChatbot(payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: CHATBOTS_QUERY_KEY })
    },
  })
}
