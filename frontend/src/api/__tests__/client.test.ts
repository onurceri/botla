import { describe, it, expect, vi, beforeEach } from 'vitest'
import axios from 'axios'
import { api, _resetRedirecting, _setRedirectToLogin } from '../client'

describe('axios refresh interceptor', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    window.localStorage.clear()
    _resetRedirecting()
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
    _setRedirectToLogin(redirectMock)
    
    const handlers = (api as any).interceptors.response.handlers
    const handler = handlers[handlers.length - 1].rejected
    const req: any = { url: '/x', headers: {}, method: 'get', _retry: false }
    const err: any = { response: { status: 401 }, config: req }

    const p = handler(err).catch(() => {})

    // Fast-forward time for the 1500ms delay in handleSessionExpired
    await vi.advanceTimersByTimeAsync(2000)
    await p

    expect(redirectMock).toHaveBeenCalled()
    
    // Restore original redirect function
    _setRedirectToLogin(() => {
      if (typeof window !== 'undefined') {
        window.location.href = '/login'
      }
    })
    vi.useRealTimers()
    vi.unstubAllEnvs()
  })

  it('does not trigger session expiry on auth endpoint 401 errors', async () => {
    vi.useFakeTimers()
    vi.stubEnv('VITE_E2E', '')
    
    // Set up the mock redirect function to track if it gets called
    const redirectMock = vi.fn()
    _setRedirectToLogin(redirectMock)
    
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
    
    // Restore original redirect function
    _setRedirectToLogin(() => {
      if (typeof window !== 'undefined') {
        window.location.href = '/login'
      }
    })
    vi.useRealTimers()
    vi.unstubAllEnvs()
  })
})
