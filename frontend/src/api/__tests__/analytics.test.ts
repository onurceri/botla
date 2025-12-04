import { describe, it, expect, vi } from 'vitest'
import { getAnalytics } from '../analytics'
import { api } from '../client'

describe('api/analytics', () => {
  it('calls endpoint and returns data', async () => {
    const payload = [{ date: '2024-05-10', conversations: 1, messages: 2 }]
    const spy = vi.spyOn(api, 'get').mockResolvedValueOnce({ data: payload } as any)
    const data = await getAnalytics()
    expect(spy).toHaveBeenCalledWith('/api/v1/analytics')
    expect(data).toEqual(payload)
  })
})

