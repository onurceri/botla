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
