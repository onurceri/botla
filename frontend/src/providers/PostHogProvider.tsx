import { PostHogProvider as PHProvider } from 'posthog-js/react'
import posthog from 'posthog-js'

const POSTHOG_KEY = import.meta.env.VITE_PUBLIC_POSTHOG_KEY
const POSTHOG_HOST = import.meta.env.VITE_PUBLIC_POSTHOG_HOST

interface PostHogProviderProps {
  children: React.ReactNode
}

export function PostHogProvider({ children }: PostHogProviderProps) {
  // Skip PostHog if credentials are not configured
  if (!POSTHOG_KEY || !POSTHOG_HOST) {
    return <>{children}</>
  }

  // Initialize PostHog (only runs once)
  if (typeof window !== 'undefined' && !posthog.__loaded) {
    posthog.init(POSTHOG_KEY, {
      api_host: POSTHOG_HOST,
      person_profiles: 'identified_only',
      capture_pageview: true,
      capture_pageleave: true,
    })
  }

  return <PHProvider client={posthog}>{children}</PHProvider>
}
