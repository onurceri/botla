import { describe, it, expect, vi, beforeEach } from 'vitest'
import axios from 'axios'
import { api, redirectService, _resetRefreshState, _setWasAuthenticated } from '../client'

describe('axios refresh interceptor', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    window.localStorage.clear()
    redirectService.reset()
    _resetRefreshState()
  })

  it('refreshes token on 401 and retries original request', async () => {
    window.localStorage.setItem('botla_refresh_token', 'rtok')
    const postSpy = vi.spyOn(axios, 'post').mockResolvedValueOnce({
      data: { token: 'newAccess', refresh_token: 'newRefresh' },
    } as any)

    const originalRequest: any = {
      url: '/api/v1/chatbots',
      headers: {},
      method: 'get',
      _retry: false,
    }
    const error: any = { response: { status: 401 }, config: originalRequest }

    const handlers = (api as any).interceptors.response.handlers
    const handler = handlers[handlers.length - 1].rejected
    vi.spyOn(api, 'request').mockResolvedValueOnce({} as any)
    await handler(error).catch(() => {})

    expect(postSpy).toHaveBeenCalled()
    expect(postSpy).toHaveBeenCalled()
  })

  it('coalesces concurrent 401 refresh into a single request', async () => {
    window.localStorage.setItem('botla_refresh_token', 'rtok')
    const postSpy = vi.spyOn(axios, 'post').mockImplementationOnce(async () => {
      await new Promise((r) => setTimeout(r, 10))
      return { data: { token: 'newAccess', refresh_token: 'newRefresh' } } as any
    })

    const handlers = (api as any).interceptors.response.handlers
    const handler = handlers[handlers.length - 1].rejected

    const req1: any = { url: '/x', headers: {}, method: 'get', _retry: false }
    const req2: any = { url: '/y', headers: {}, method: 'get', _retry: false }
    const err1: any = { response: { status: 401 }, config: req1 }
    const err2: any = { response: { status: 401 }, config: req2 }

    vi.spyOn(api, 'request').mockResolvedValue({} as any)
    const p1 = handler(err1).catch(() => {})
    const p2 = handler(err2).catch(() => {})
    await Promise.all([p1, p2])

    expect(postSpy).toHaveBeenCalledTimes(1)
  })

  it('retries all 5 concurrent requests after single token refresh', async () => {
    window.localStorage.setItem('botla_refresh_token', 'rtok')
    
    // Mock the refresh endpoint to succeed after a short delay
    const postSpy = vi.spyOn(axios, 'post').mockImplementationOnce(async () => {
      await new Promise((r) => setTimeout(r, 20))
      return { data: { token: 'newAccess', refresh_token: 'newRefresh' } } as any
    })

    const handlers = (api as any).interceptors.response.handlers
    const handler = handlers[handlers.length - 1].rejected

    // Create 5 concurrent requests that fail with 401
    const requests = Array.from({ length: 5 }, (_, i) => ({
      url: `/api/v1/resource/${i}`,
      headers: {},
      method: 'get',
      _retry: false,
    }))
    const errors = requests.map((req) => ({
      response: { status: 401 },
      config: req,
    }))

    // Mock api.request - axios instance when called as api(config) uses this internally
    vi.spyOn(api, 'request').mockResolvedValue({ data: 'success' } as any)

    // Fire all 5 requests simultaneously using same pattern as coalescing test
    const promises = errors.map((err) => handler(err).catch(() => {}))
    await Promise.all(promises)

    // Verify only 1 refresh happened - this is the key assertion
    expect(postSpy).toHaveBeenCalledTimes(1)
  })

  it('does not retry if already retried', async () => {
    window.localStorage.setItem('botla_refresh_token', 'rtok')
    const postSpy = vi.spyOn(axios, 'post')
    const handlers = (api as any).interceptors.response.handlers
    const handler = handlers[handlers.length - 1].rejected
    const req: any = { url: '/x', headers: {}, method: 'get', _retry: true }
    const err: any = { response: { status: 401 }, config: req }
    await handler(err).catch(() => {})
    expect(postSpy).not.toHaveBeenCalled()
  })

  it('redirects to login on refresh failure and clears tokens', async () => {
    vi.useFakeTimers()
    vi.stubEnv('VITE_E2E', '')
    window.localStorage.setItem('botla_refresh_token', 'rtok')
    vi.spyOn(axios, 'post').mockRejectedValueOnce(new Error('nope'))
    
    // Set up the mock redirect function BEFORE triggering the handler
    const redirectMock = vi.fn()
    redirectService.setRedirectFn(redirectMock)
    
    const handlers = (api as any).interceptors.response.handlers
    const handler = handlers[handlers.length - 1].rejected
    const req: any = { url: '/x', headers: {}, method: 'get', _retry: false }
    const err: any = { response: { status: 401 }, config: req }

    const p = handler(err).catch(() => {})

    // Fast-forward time for the 1500ms delay in handleSessionExpired
    await vi.advanceTimersByTimeAsync(2000)
    await p

    expect(redirectMock).toHaveBeenCalled()
    
    // Cleanup handled by redirectService.reset() in beforeEach
    vi.useRealTimers()
    vi.unstubAllEnvs()
  })

  it('does not trigger session expiry on auth endpoint 401 errors', async () => {
    vi.useFakeTimers()
    vi.stubEnv('VITE_E2E', '')
    
    // Set up the mock redirect function to track if it gets called
    const redirectMock = vi.fn()
    redirectService.setRedirectFn(redirectMock)
    
    const handlers = (api as any).interceptors.response.handlers
    const handler = handlers[handlers.length - 1].rejected
    
    // Test login endpoint
    const loginReq: any = { url: '/api/v1/auth/login', headers: {}, method: 'post', _retry: false }
    const loginErr: any = { response: { status: 401 }, config: loginReq }
    
    await expect(handler(loginErr)).rejects.toEqual(loginErr)
    
    // Fast-forward time - redirect should NOT be called
    await vi.advanceTimersByTimeAsync(2000)
    
    // Verify session expiry was NOT triggered
    expect(redirectMock).not.toHaveBeenCalled()
    
    // Cleanup handled by redirectService.reset() in beforeEach
    vi.useRealTimers()
    vi.unstubAllEnvs()
  })

  it('does not trigger token refresh on /me endpoint 401 (new visitor scenario)', async () => {
    vi.useFakeTimers()
    
    // Set up mocks
    const postSpy = vi.spyOn(axios, 'post')
    const redirectMock = vi.fn()
    redirectService.setRedirectFn(redirectMock)
    const eventSpy = vi.fn()
    window.addEventListener('session-expired', eventSpy)
    
    const handlers = (api as any).interceptors.response.handlers
    const handler = handlers[handlers.length - 1].rejected
    
    // Simulate /me endpoint returning 401 (user not logged in)
    const meReq: any = { url: '/api/v1/me', headers: {}, method: 'get', _retry: false }
    const meErr: any = { response: { status: 401 }, config: meReq }
    
    await expect(handler(meErr)).rejects.toEqual(meErr)
    
    // Fast-forward time
    await vi.advanceTimersByTimeAsync(2000)
    
    // Verify:
    // 1. No refresh was attempted
    expect(postSpy).not.toHaveBeenCalled()
    // 2. No redirect was triggered
    expect(redirectMock).not.toHaveBeenCalled()
    // 3. No session-expired event was dispatched
    expect(eventSpy).not.toHaveBeenCalled()
    
    window.removeEventListener('session-expired', eventSpy)
    vi.useRealTimers()
  })

  it('only shows session expired message if wasAuthenticated is true', async () => {
    vi.useFakeTimers()
    
    // Set up event listener to capture session-expired events
    const eventSpy = vi.fn()
    window.addEventListener('session-expired', eventSpy)
    
    const redirectMock = vi.fn()
    redirectService.setRedirectFn(redirectMock)
    
    // Mock refresh to fail
    vi.spyOn(axios, 'post').mockRejectedValueOnce(new Error('nope'))
    
    const handlers = (api as any).interceptors.response.handlers
    const handler = handlers[handlers.length - 1].rejected
    const req: any = { url: '/api/v1/chatbots', headers: {}, method: 'get', _retry: false }
    const err: any = { response: { status: 401 }, config: req }
    
    // When wasAuthenticated is false (default), session-expired should NOT be dispatched
    const p1 = handler(err).catch(() => {})
    await vi.advanceTimersByTimeAsync(2000)
    await p1
    
    // No session-expired event should be dispatched for unauthenticated user
    expect(eventSpy).not.toHaveBeenCalled()
    
    // Reset for next test
    _resetRefreshState()
    vi.spyOn(axios, 'post').mockRejectedValueOnce(new Error('nope'))
    
    // Set wasAuthenticated to true (simulating logged-in user)
    _setWasAuthenticated(true)
    
    const req2: any = { url: '/api/v1/chatbots', headers: {}, method: 'get', _retry: false }
    const err2: any = { response: { status: 401 }, config: req2 }
    
    const p2 = handler(err2).catch(() => {})
    await vi.advanceTimersByTimeAsync(2000)
    await p2
    
    // Now session-expired event SHOULD be dispatched
    expect(eventSpy).toHaveBeenCalled()
    
    window.removeEventListener('session-expired', eventSpy)
    vi.useRealTimers()
  })

  it('handles ERR_ACCOUNT_DELETED error by clearing localStorage and redirecting', async () => {
    vi.useFakeTimers()
    
    // Set up event listener to capture account-deleted events
    const eventSpy = vi.fn()
    window.addEventListener('account-deleted', eventSpy)
    
    const redirectMock = vi.fn()
    redirectService.setRedirectFn(redirectMock)
    
    // Set some localStorage data
    window.localStorage.setItem('botla_user', '{"email":"test@example.com"}')
    window.localStorage.setItem('botla_last_org_id', 'org-123')
    
    const handlers = (api as any).interceptors.response.handlers
    const handler = handlers[handlers.length - 1].rejected
    
    // Simulate ERR_ACCOUNT_DELETED error
    const req: any = { url: '/api/v1/chatbots', headers: {}, method: 'get', _retry: false }
    const err: any = { 
      response: { 
        status: 403, 
        data: { code: 'ERR_ACCOUNT_DELETED' }
      }, 
      config: req 
    }
    
    const p = handler(err).catch(() => {})
    await vi.advanceTimersByTimeAsync(2000)
    await p
    
    // Verify:
    // 1. account-deleted event was dispatched
    expect(eventSpy).toHaveBeenCalled()
    // 2. localStorage was cleared
    expect(window.localStorage.getItem('botla_user')).toBeNull()
    expect(window.localStorage.getItem('botla_last_org_id')).toBeNull()
    // 3. Redirect was called
    expect(redirectMock).toHaveBeenCalled()
    
    window.removeEventListener('account-deleted', eventSpy)
    vi.useRealTimers()
  })
})
