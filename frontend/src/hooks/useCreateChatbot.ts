import { useMutation, useQueryClient } from '@tanstack/react-query'
import { createChatbot } from '@/api/chatbot'

export const useCreateChatbot = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (payload: unknown) => createChatbot(payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['chatbots'] })
    },
  })
}
