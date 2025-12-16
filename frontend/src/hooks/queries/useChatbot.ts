import { useQuery } from '@tanstack/react-query'
import { api } from '@/api/client'

export const CHATBOT_QUERY_KEY = (id: string) => ['chatbot', id] as const

export function useChatbot(chatbotId?: string, enabled = true) {
  return useQuery({
    queryKey: CHATBOT_QUERY_KEY(chatbotId || ''),
    queryFn: async () => {
      const { data } = await api.get(`/api/v1/chatbots/${chatbotId}`)
      return data
    },
    enabled: !!chatbotId && enabled,
    refetchOnMount: 'always', // Always refetch when component mounts (fixes stale data)
    staleTime: 0, // Data becomes stale immediately
  })
}
