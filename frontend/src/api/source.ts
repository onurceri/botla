import { api } from './client'

export const uploadPDFSource = async (chatbotId: string, file: File) => {
  const form = new FormData()
  form.append('source_type', 'pdf')
  form.append('file', file)
  const { data } = await api.post(`/api/v1/chatbots/${chatbotId}/sources`, form)
  return data as { id: string }
}

export const uploadTextSource = async (chatbotId: string, text: string) => {
  const form = new FormData()
  form.append('source_type', 'text')
  form.append('text', text)
  const { data } = await api.post(`/api/v1/chatbots/${chatbotId}/sources`, form)
  return data as { id: string }
}

export const uploadURLSource = async (chatbotId: string, url: string) => {
  const form = new FormData()
  form.append('source_type', 'url')
  form.append('source_url', url)
  const { data } = await api.post(`/api/v1/chatbots/${chatbotId}/sources`, form)
  return data as { id: string }
}

export const getSourceStatus = async (sourceId: string, etag?: string) => {
  try {
    const res = await api.get(`/api/v1/sources/${sourceId}`, { headers: etag ? { 'If-None-Match': etag } : undefined })
    const newEtag = res.headers['etag'] as string | undefined
    return { data: res.data, etag: newEtag, notModified: false }
  } catch (err: any) {
    if (err?.response?.status === 304) {
      return { data: null, etag, notModified: true }
    }
    throw err
  }
}

export const listSources = async (chatbotId: string) => {
  const { data } = await api.get(`/api/v1/chatbots/${chatbotId}/sources`)
  return data
}

export const deleteSource = async (sourceId: string) => {
  const { data } = await api.delete(`/api/v1/sources/${sourceId}`)
  return data
}

export const refreshSource = async (sourceId: string) => {
  const { data } = await api.post(`/api/v1/sources/${sourceId}/refresh`)
  return data as { id: string }
}

// Sitemap types
export interface SitemapURL {
  loc: string
  lastmod?: string
  changefreq?: string
  priority?: number
}

export interface DiscoverSitemapResponse {
  urls: SitemapURL[]
  total_count: number
}

export interface BulkCreateResponse {
  created_count: number
  skipped_count: number
  errors: string[]
}

// Discover URLs from a sitemap
export const discoverSitemap = async (chatbotId: string, sitemapUrl: string): Promise<DiscoverSitemapResponse> => {
  const { data } = await api.post(`/api/v1/chatbots/${chatbotId}/sitemap/discover`, {
    sitemap_url: sitemapUrl,
  })
  return data as DiscoverSitemapResponse
}

// Bulk create URL sources
export const bulkCreateSources = async (chatbotId: string, urls: string[]): Promise<BulkCreateResponse> => {
  const { data } = await api.post(`/api/v1/chatbots/${chatbotId}/sources/bulk`, {
    urls,
  })
  return data as BulkCreateResponse
}
