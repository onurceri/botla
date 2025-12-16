import { useQuery } from '@tanstack/react-query'
import { listChatbots } from '@/api/chatbot'

export const CHATBOTS_QUERY_KEY = ['chatbots'] as const

export function useChatbots(enabled = true) {
  return useQuery({
    queryKey: CHATBOTS_QUERY_KEY,
    queryFn: listChatbots,
    enabled,
    staleTime: 1000 * 30, // 30 seconds
  })
}
