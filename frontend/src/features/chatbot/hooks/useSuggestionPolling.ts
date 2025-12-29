import { useEffect, useRef } from 'react'
import { listSources } from '@/api/source'

export function useSuggestionPolling(chatbotId: string, refetchChatbot: () => void) {
  const pollingRef = useRef<NodeJS.Timeout | null>(null)
  const startTimeRef = useRef<number | null>(null)
  const MAX_POLLING_DURATION = 5 * 60 * 1000 // 5 minutes

  useEffect(() => {
    if (!chatbotId) return

    const checkStatus = async () => {
      try {
        const sources = await listSources(chatbotId)
        const processingSources = sources.filter((s: any) =>
          ['pending', 'processing', 'queued'].includes(s.status),
        )

        if (processingSources.length === 0) {
          // If we were polling and now everything is done, refetch chatbot and stop polling
          if (pollingRef.current) {
            refetchChatbot()
            stopPolling()
          }
          return
        }

        // Check if we hit the timeout
        if (startTimeRef.current && Date.now() - startTimeRef.current > MAX_POLLING_DURATION) {
            // Timeout reached. One last refetch to get whatever we have, then stop.
            refetchChatbot()
            stopPolling()
            return
        }

        // Start polling if not already started
        if (!pollingRef.current) {
          startTimeRef.current = Date.now()
          pollingRef.current = setInterval(checkStatus, 5000)
        }
      } catch (error) {
        // Silently fail on error and stop polling to be safe
        stopPolling()
      }
    }

    const stopPolling = () => {
      if (pollingRef.current) {
        clearInterval(pollingRef.current)
        pollingRef.current = null
        startTimeRef.current = null
      }
    }

    // Initial check
    checkStatus()

    return () => stopPolling()
  }, [chatbotId, refetchChatbot])
}
