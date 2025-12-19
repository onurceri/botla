import { useState, useEffect } from 'react'
import { rateLimitStore } from '@/lib/rateLimit'

export function useRateLimit() {
  const [state, setState] = useState(rateLimitStore.getState())

  useEffect(() => {
    return rateLimitStore.subscribe(setState)
  }, [])

  return state
}
