import { describe, it, expect, vi, beforeEach } from 'vitest'
import axios from 'axios'
import { api, _resetRedirecting } from '../client'

describe('axios refresh interceptor', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    const store: Record<string, string> = {}
    Object.defineProperty(window, 'localStorage', {
      value: {
        getItem: vi.fn((key: string) => (key in store ? store[key] : null)),
        setItem: vi.fn((key: string, value: string) => {
          store[key] = value
        }),
        removeItem: vi.fn((key: string) => {
          delete store[key]
        }),
      },
      writable: true,
    })
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
    expect(window.localStorage.setItem).toHaveBeenCalledWith('botla_token', 'newAccess')
    expect(window.localStorage.setItem).toHaveBeenCalledWith('botla_refresh_token', 'newRefresh')
    expect(window.localStorage.getItem('botla_token')).toBe('newAccess')
    expect(window.localStorage.getItem('botla_refresh_token')).toBe('newRefresh')
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
    expect(req1.headers.Authorization).toBe(`Bearer ${window.localStorage.getItem('botla_token')}`)
    expect(req2.headers.Authorization).toBe(`Bearer ${window.localStorage.getItem('botla_token')}`)
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
    const getLoc = vi.spyOn(window, 'location', 'get')
    const replaceMock = vi.fn()
    getLoc.mockReturnValue({ ...window.location, replace: replaceMock } as any)
    const handlers = (api as any).interceptors.response.handlers
    const handler = handlers[handlers.length - 1].rejected
    const req: any = { url: '/x', headers: {}, method: 'get', _retry: false }
    const err: any = { response: { status: 401 }, config: req }

    const p = handler(err).catch(() => {})

    // Fast-forward time for the 100ms delay in handleSessionExpired
    vi.advanceTimersByTime(200)
    await p

    expect(window.localStorage.removeItem).toHaveBeenCalledWith('botla_token')
    expect(window.localStorage.removeItem).toHaveBeenCalledWith('botla_refresh_token')

    await vi.waitFor(() => {
      expect(replaceMock).toHaveBeenCalledWith('/login')
    })
    getLoc.mockRestore()
    vi.useRealTimers()
    vi.unstubAllEnvs()
  })
})
