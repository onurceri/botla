import { useState, useEffect, useRef, useCallback } from 'react'
import { useParams } from 'react-router-dom'
import { api } from '@/api/client'
import { useToastErrors } from './useToastErrors'
import { AUTO_SAVE_DEFAULT_FALLBACK, AUTO_SAVE_RETRY_SUFFIX, getTurkishErrorMessage } from '@/lib/errorMessages'

type AutoSaveState = {
  isSaving: boolean
  lastSavedAt: Date | null
  error: string | null
}

type UseAutoSaveOptions = {
  payload: Record<string, unknown>
  enabled?: boolean
  debounceMs?: number
  onSuccess?: () => void
  onError?: (error: string) => void
  maxRetries?: number
  retryDelayMs?: number
  saveFn?: (id: string, payload: any) => Promise<any>
}

const DEFAULT_MAX_RETRIES = 2
const DEFAULT_RETRY_DELAY_MS = 3000
const SUCCESS_VISIBILITY_MS = 3000

export function useAutoSave({
  payload,
  enabled = true,
  debounceMs = 800,
  onSuccess,
  onError,
  maxRetries = DEFAULT_MAX_RETRIES,
  retryDelayMs = DEFAULT_RETRY_DELAY_MS,
  saveFn,
}: UseAutoSaveOptions): AutoSaveState {
  const { id } = useParams()
  const toasts = useToastErrors()

  const [state, setState] = useState<AutoSaveState>({
    isSaving: false,
    lastSavedAt: null,
    error: null,
  })

  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const abortRef = useRef<AbortController | null>(null)
  const prevPayloadRef = useRef<string | null>(null)
  const payloadRef = useRef(payload) // Store latest payload in ref
  const retryCountRef = useRef<number>(0)
  const retryTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const successTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const lastToastedErrorRef = useRef<string | null>(null)

  // Keep payloadRef in sync with latest payload
  useEffect(() => {
    payloadRef.current = payload
  }, [payload])

  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current)
        timeoutRef.current = null
      }
      if (retryTimeoutRef.current) {
        clearTimeout(retryTimeoutRef.current)
        retryTimeoutRef.current = null
      }
      if (successTimeoutRef.current) {
        clearTimeout(successTimeoutRef.current)
        successTimeoutRef.current = null
      }
      if (abortRef.current) {
        abortRef.current.abort()
        abortRef.current = null
      }
    }
  }, [])

  useEffect(() => {
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      if (timeoutRef.current || state.isSaving) {
        e.preventDefault()
        e.returnValue = ''
      }
    }

    window.addEventListener('beforeunload', handleBeforeUnload)
    return () => window.removeEventListener('beforeunload', handleBeforeUnload)
  }, [state.isSaving])

  const save = useCallback(
    async (isRetry = false) => {
      if (!id || id === 'new') return

      if (abortRef.current) {
        abortRef.current.abort()
      }
      abortRef.current = new AbortController()

      setState((s) => ({ ...s, isSaving: true, error: null }))

      try {
        if (saveFn) {
           await saveFn(id, payloadRef.current)
        } else {
           await api.put(`/api/v1/chatbots/${id}`, payloadRef.current, {
             signal: abortRef.current.signal,
           })
        }

        retryCountRef.current = 0
        lastToastedErrorRef.current = null
        if (successTimeoutRef.current) {
          clearTimeout(successTimeoutRef.current)
          successTimeoutRef.current = null
        }
        setState((s) => ({
          ...s,
          isSaving: false,
          lastSavedAt: new Date(),
          error: null,
        }))
        successTimeoutRef.current = setTimeout(() => {
          setState((s) => ({
            ...s,
            lastSavedAt: null,
          }))
          successTimeoutRef.current = null
        }, SUCCESS_VISIBILITY_MS)
        onSuccess?.()
      } catch (e: any) {
        if (e.name === 'CanceledError' || e.name === 'AbortError') {
          return
        }

        const errorMsg = getTurkishErrorMessage(e, AUTO_SAVE_DEFAULT_FALLBACK)

        if (!isRetry && retryCountRef.current < maxRetries) {
          retryCountRef.current += 1
          setState((s) => ({
            ...s,
            isSaving: false,
            error: `${errorMsg}${AUTO_SAVE_RETRY_SUFFIX}`,
          }))

          retryTimeoutRef.current = setTimeout(() => {
            save(true)
          }, retryDelayMs)
          return
        }

        retryCountRef.current = 0
        setState((s) => ({
          ...s,
          isSaving: false,
          error: errorMsg,
        }))
        if (lastToastedErrorRef.current !== errorMsg) {
          toasts.error(errorMsg)
          lastToastedErrorRef.current = errorMsg
        }
        onError?.(errorMsg)
      }
    },
    [id, onSuccess, onError, maxRetries, retryDelayMs, toasts, saveFn]
  )

  // Store save function in ref to avoid stale closures and prevent effect re-runs
  const saveRef = useRef(save)
  useEffect(() => {
    saveRef.current = save
  }, [save])

  useEffect(() => {
    if (!enabled) return
    if (!id || id === 'new') return

    const payloadStr = JSON.stringify(payload)

    // Initialize prevPayloadRef on first run to prevent auto-save on mount
    if (prevPayloadRef.current === null) {
      prevPayloadRef.current = payloadStr
      return
    }

    if (payloadStr === prevPayloadRef.current) return
    prevPayloadRef.current = payloadStr

    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current)
    }

    timeoutRef.current = setTimeout(() => {
      timeoutRef.current = null
      saveRef.current()
    }, debounceMs)
  }, [payload, enabled, debounceMs, id])

  return state
}
