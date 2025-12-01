import { useQuery } from '@tanstack/react-query'
import { listChatbots } from '@/api/chatbot'

export const useChatbots = () => {
  const query = useQuery({ queryKey: ['chatbots'], queryFn: listChatbots })
  return query
}
