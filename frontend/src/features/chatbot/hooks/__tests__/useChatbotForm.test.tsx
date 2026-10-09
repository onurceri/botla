import { describe, it, expect } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useChatbotForm } from '../../hooks/useChatbotForm'

describe('useChatbotForm branding payload', () => {
  it('sets custom_branding to null when hideBranding is false', () => {
    const { result } = renderHook(() => useChatbotForm())
    act(() => {
      result.current.setHideBranding(false)
      result.current.setCustomBranding({ text: 'ACME', link: 'https://example.com' })
    })
    const payload = result.current.buildPayload()
    expect(payload.hide_branding).toBe(false)
    expect(payload.custom_branding).toBeNull()
  })

  it('includes custom_branding when hideBranding is true', () => {
    const { result } = renderHook(() => useChatbotForm())
    act(() => {
      result.current.setHideBranding(true)
      result.current.setCustomBranding({ text: 'ACME', link: 'https://example.com' })
    })
    const payload = result.current.buildPayload()
    expect(payload.hide_branding).toBe(true)
    expect(payload.custom_branding).toEqual({ text: 'ACME', link: 'https://example.com' })
  })
})

describe('useChatbotForm visibility payload', () => {
  it('defaults to public', () => {
    const { result } = renderHook(() => useChatbotForm())
    expect(result.current.isPublic).toBe(true)
    expect(result.current.buildVisibilityPayload()).toEqual({ is_public: true })
  })

  it('loads is_public from the server and treats a missing value as public', () => {
    const { result } = renderHook(() => useChatbotForm())
    act(() => result.current.setFromServer({ is_public: false }))
    expect(result.current.buildVisibilityPayload()).toEqual({ is_public: false })
    act(() => result.current.setFromServer({}))
    expect(result.current.isPublic).toBe(true)
  })

  it('keeps is_public out of the full payload so it is only saved by the toggle', () => {
    const { result } = renderHook(() => useChatbotForm())
    expect('is_public' in result.current.buildPayload()).toBe(false)
  })
})
