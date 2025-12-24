import { useQuery } from '@tanstack/react-query'
import { api } from '@/api/client'
import type { User } from '@/types/user'

export const PROFILE_QUERY_KEY = ['profile'] as const
export const PLAN_QUERY_KEY = ['plan'] as const
export const USAGE_QUERY_KEY = ['usage'] as const

export function useProfile() {
  return useQuery<User>({
    queryKey: PROFILE_QUERY_KEY,
    queryFn: async () => {
      const { data } = await api.get<User>('/api/v1/me')
      return data
    },
    staleTime: 1000 * 60 * 5, // 5 minutes
    refetchOnWindowFocus: true,
  })
}

export function usePlan() {
  return useQuery({
    queryKey: PLAN_QUERY_KEY,
    queryFn: async () => {
      const { data } = await api.get('/api/v1/me/plan')
      return data
    },
    staleTime: 1000 * 60 * 60, // 1 hour (plan doesn't change often)
    refetchOnWindowFocus: false,
  })
}

export function useUsage() {
  return useQuery({
    queryKey: USAGE_QUERY_KEY,
    queryFn: async () => {
      const { data } = await api.get('/api/v1/me/usage')
      return data
    },
    staleTime: 1000 * 60, // 1 minute (usage changes often)
    refetchOnWindowFocus: true,
  })
}
