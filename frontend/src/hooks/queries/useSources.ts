import { useQuery } from '@tanstack/react-query'
import { listSources } from '@/api/source'

export const SOURCES_QUERY_KEY = (chatbotId: string) => ['chatbot', chatbotId, 'sources'] as const

export function useSources(chatbotId?: string, enabled = true) {
  return useQuery({
    queryKey: SOURCES_QUERY_KEY(chatbotId || ''),
    queryFn: () => listSources(chatbotId!),
    enabled: !!chatbotId && enabled,
    refetchOnMount: 'always',
    staleTime: 0,
  })
}
