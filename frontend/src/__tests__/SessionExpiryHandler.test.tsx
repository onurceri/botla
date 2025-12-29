import { render, cleanup } from '@testing-library/react'
import { useEffect, useRef, useCallback } from 'react'
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest'

// Create a standalone test version of SessionExpiryHandler
// This mirrors the implementation in App.tsx
const mockToast = vi.fn()

function SessionExpiryHandler() {
  // Use a ref to always have the latest toast function without causing effect re-runs
  const toastRef = useRef(mockToast)
  useEffect(() => {
    toastRef.current = mockToast
  }, [])

  // Stable event handler using useCallback with no dependencies
  const handleSessionExpired = useCallback(() => {
    toastRef.current('Oturumunuz sona erdi. Lütfen tekrar giriş yapın.', 'error')
  }, [])

  useEffect(() => {
    // Effect runs only once on mount, cleanup uses the same stable function reference
    window.addEventListener('session-expired', handleSessionExpired)
    return () => window.removeEventListener('session-expired', handleSessionExpired)
  }, [handleSessionExpired])

  return null
}

describe('SessionExpiryHandler', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    cleanup()
  })

  it('should handle session-expired event only once per dispatch', () => {
    // Render the component
    render(<SessionExpiryHandler />)

    // Dispatch session-expired event
    window.dispatchEvent(new Event('session-expired'))

    // Toast should be called exactly once
    expect(mockToast).toHaveBeenCalledTimes(1)
    expect(mockToast).toHaveBeenCalledWith(
      'Oturumunuz sona erdi. Lütfen tekrar giriş yapın.',
      'error'
    )
  })

  it('should not register multiple listeners across re-renders', async () => {
    // This test verifies that the useCallback+useRef pattern is working correctly
    const { rerender } = render(<SessionExpiryHandler />)

    // Force re-render
    rerender(<SessionExpiryHandler />)
    rerender(<SessionExpiryHandler />)

    // Dispatch event
    window.dispatchEvent(new Event('session-expired'))

    // Should still only be called once (not multiple times due to re-renders)
    expect(mockToast).toHaveBeenCalledTimes(1)
  })

  it('should properly cleanup listener on unmount', () => {
    const { unmount } = render(<SessionExpiryHandler />)

    // Unmount the component
    unmount()

    // Dispatch event after unmount
    window.dispatchEvent(new Event('session-expired'))

    // Toast should NOT be called since the listener was cleaned up
    expect(mockToast).not.toHaveBeenCalled()
  })

  it('should handle multiple sequential events correctly', () => {
    render(<SessionExpiryHandler />)

    // Dispatch multiple events
    window.dispatchEvent(new Event('session-expired'))
    window.dispatchEvent(new Event('session-expired'))
    window.dispatchEvent(new Event('session-expired'))

    // Toast should be called 3 times (once per dispatch)
    expect(mockToast).toHaveBeenCalledTimes(3)
  })

  it('should not leak listeners when component is mounted/unmounted multiple times', () => {
    // Mount -> Unmount cycle 1
    const { unmount: unmount1 } = render(<SessionExpiryHandler />)
    unmount1()

    // Mount -> Unmount cycle 2
    const { unmount: unmount2 } = render(<SessionExpiryHandler />)
    unmount2()

    // Mount final time
    render(<SessionExpiryHandler />)

    // Dispatch event
    window.dispatchEvent(new Event('session-expired'))

    // Only one listener should be active (from the final mount)
    expect(mockToast).toHaveBeenCalledTimes(1)
  })
})
