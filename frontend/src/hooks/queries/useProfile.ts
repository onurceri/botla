import { useQuery } from '@tanstack/react-query'
import { api } from '@/api/client'

export const PROFILE_QUERY_KEY = ['profile'] as const

export function useProfile() {
  return useQuery({
    queryKey: PROFILE_QUERY_KEY,
    queryFn: async () => {
      const { data } = await api.get('/api/v1/me')
      return data
    },
    staleTime: 1000 * 60 * 5, // 5 minutes (profile doesn't change often)
    refetchOnWindowFocus: true, // Refetch when user returns to tab
  })
}
