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
    const url = `/api/v1/sources/${sourceId}`
    const opts = etag ? { headers: { 'If-None-Match': etag } } : undefined
    const res = opts ? await api.get(url, opts) : await api.get(url)
    const newEtag = (res as any)?.headers?.['etag'] as string | undefined
    const payload = res.data
    return { data: payload, etag: newEtag, notModified: false }
  } catch (err: any) {
    if (err?.response?.status === 304) {
      if (etag === undefined) return null
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

// Pending URLs types
export interface PendingURL {
  id: string
  url: string
  discovered_at: string
}

export interface ListPendingURLsResponse {
  urls: PendingURL[]
  total: number
  page: number
  per_page: number
}

export interface ApproveResponse {
  approved_count: number
  sources_created: number
}

export interface RejectResponse {
  rejected_count: number
}

export interface ClearResponse {
  cleared_count: number
}

// List pending URLs for a chatbot
export const listPendingURLs = async (
  chatbotId: string,
  page = 1,
  perPage = 20
): Promise<ListPendingURLsResponse> => {
  const { data } = await api.get(
    `/api/v1/chatbots/${chatbotId}/pending-urls?page=${page}&per_page=${perPage}`
  )
  return data as ListPendingURLsResponse
}

// Approve pending URLs (create sources from them)
export const approvePendingURLs = async (
  chatbotId: string,
  urlIds: string[]
): Promise<ApproveResponse> => {
  const { data } = await api.post(`/api/v1/chatbots/${chatbotId}/pending-urls/approve`, {
    url_ids: urlIds,
  })
  return data as ApproveResponse
}

// Reject pending URLs
export const rejectPendingURLs = async (
  chatbotId: string,
  urlIds: string[]
): Promise<RejectResponse> => {
  const { data } = await api.post(`/api/v1/chatbots/${chatbotId}/pending-urls/reject`, {
    url_ids: urlIds,
  })
  return data as RejectResponse
}

// Clear all pending URLs
export const clearPendingURLs = async (chatbotId: string): Promise<ClearResponse> => {
  const { data } = await api.post(`/api/v1/chatbots/${chatbotId}/pending-urls/clear`)
  return data as ClearResponse
}
