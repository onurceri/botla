import { describe, it, expect, vi } from 'vitest'
import { uploadPDFSource, uploadTextSource, uploadURLSource, getSourceStatus, listSources, deleteSource } from '../source'
import { api } from '../client'

describe('api/source', () => {
  it('uploadPDFSource posts FormData and returns id', async () => {
    const file = new File(['hello'], 'doc.pdf', { type: 'application/pdf' })
    const post = vi.spyOn(api, 'post').mockResolvedValueOnce({ data: { id: 'pdf-1' } } as any)
    const id = await uploadPDFSource('bot-1', file)
    expect(id).toEqual({ id: 'pdf-1' })
    const [url, form] = post.mock.calls.at(-1) as [string, FormData]
    expect(url).toBe('/api/v1/chatbots/bot-1/sources')
    expect(form).toBeInstanceOf(FormData)
    expect((form as FormData).get('source_type')).toBe('pdf')
    expect((form as FormData).get('file')).toBe(file)
  })

  it('uploadTextSource posts text payload', async () => {
    const post = vi.spyOn(api, 'post').mockResolvedValueOnce({ data: { id: 'txt-1' } } as any)
    const id = await uploadTextSource('bot-2', 'Merhaba')
    expect(id).toEqual({ id: 'txt-1' })
    const [url, form] = post.mock.calls.at(-1) as [string, FormData]
    expect(url).toBe('/api/v1/chatbots/bot-2/sources')
    expect((form as FormData).get('source_type')).toBe('text')
    expect((form as FormData).get('text')).toBe('Merhaba')
  })

  it('uploadURLSource posts url payload', async () => {
    const post = vi.spyOn(api, 'post').mockResolvedValueOnce({ data: { id: 'url-1' } } as any)
    const id = await uploadURLSource('bot-3', 'https://example.com')
    expect(id).toEqual({ id: 'url-1' })
    const [url, form] = post.mock.calls.at(-1) as [string, FormData]
    expect(url).toBe('/api/v1/chatbots/bot-3/sources')
    expect((form as FormData).get('source_type')).toBe('url')
    expect((form as FormData).get('source_url')).toBe('https://example.com')
  })

  it('getSourceStatus fetches by id', async () => {
    const get = vi.spyOn(api, 'get').mockResolvedValueOnce({ data: { id: 's1', status: 'ready' } } as any)
    const data = await getSourceStatus('s1')
    expect(data).toEqual({ id: 's1', status: 'ready' })
    expect(get).toHaveBeenCalledWith('/api/v1/sources/s1')
  })

  it('listSources fetches by chatbot id', async () => {
    const get = vi.spyOn(api, 'get').mockResolvedValueOnce({ data: [{ id: 's1' }] } as any)
    const data = await listSources('bot-4')
    expect(data).toEqual([{ id: 's1' }])
    expect(get).toHaveBeenCalledWith('/api/v1/chatbots/bot-4/sources')
  })

  it('deleteSource calls delete by id', async () => {
    const del = vi.spyOn(api, 'delete').mockResolvedValueOnce({ data: {} } as any)
    const data = await deleteSource('s2')
    expect(data).toEqual({})
    expect(del).toHaveBeenCalledWith('/api/v1/sources/s2')
  })
})
