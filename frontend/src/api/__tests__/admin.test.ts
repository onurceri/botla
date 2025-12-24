import { describe, it, expect, vi, beforeEach } from 'vitest'
import * as adminApi from '../admin'
import { api } from '../client'

describe('api/admin', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('getOverviewStats', () => {
    it('calls the correct endpoint', async () => {
      const payload = {
        total_users: 10,
        total_organizations: 5,
        total_chatbots: 3,
        total_messages: 100,
      }
      const spy = vi.spyOn(api, 'get').mockResolvedValueOnce({ data: payload } as any)
      
      const response = await adminApi.getOverviewStats()
      
      expect(spy).toHaveBeenCalledWith('/api/v1/admin/stats/overview')
      expect(response).toEqual(payload)
    })
  })

  describe('getDetailedHealth', () => {
    it('calls the correct endpoint', async () => {
      const payload = {
        status: 'healthy',
        version: '1.0.0',
        uptime: '1d',
        environment: 'production',
        dependencies: [],
      }
      const spy = vi.spyOn(api, 'get').mockResolvedValueOnce({ data: payload } as any)
      
      const response = await adminApi.getDetailedHealth()
      
      expect(spy).toHaveBeenCalledWith('/api/v1/admin/health/detailed')
      expect(response).toEqual(payload)
    })
  })

  describe('listUsers', () => {
    it('calls the correct endpoint with params', async () => {
      const payload = {
        users: [],
        total: 0,
      }
      const params = { limit: 10, offset: 0, email: 'test@example.com' }
      const spy = vi.spyOn(api, 'get').mockResolvedValueOnce({ data: payload } as any)
      
      const response = await adminApi.listUsers(params)
      
      expect(spy).toHaveBeenCalledWith('/api/v1/admin/users', { params })
      expect(response).toEqual(payload)
    })
  })

  describe('getUser', () => {
    it('calls the correct endpoint', async () => {
      const userId = 'user-123'
      const payload = { id: userId, email: 'user@example.com' }
      const spy = vi.spyOn(api, 'get').mockResolvedValueOnce({ data: payload } as any)
      
      const response = await adminApi.getUser(userId)
      
      expect(spy).toHaveBeenCalledWith(`/api/v1/admin/users/${userId}`)
      expect(response).toEqual(payload)
    })
  })

  describe('updateUser', () => {
    it('calls the correct endpoint with patch data', async () => {
      const userId = 'user-123'
      const data = { full_name: 'New Name', is_platform_admin: true }
      const spy = vi.spyOn(api, 'patch').mockResolvedValueOnce({ data: { status: 'updated' } } as any)
      
      const response = await adminApi.updateUser(userId, data)
      
      expect(spy).toHaveBeenCalledWith(`/api/v1/admin/users/${userId}`, data)
      expect(response.status).toEqual('updated')
    })
  })

  describe('listOrganizations', () => {
    it('calls the correct endpoint with params', async () => {
      const payload = {
        organizations: [],
        total: 0,
      }
      const params = { limit: 10, offset: 0, name: 'Org' }
      const spy = vi.spyOn(api, 'get').mockResolvedValueOnce({ data: payload } as any)
      
      const response = await adminApi.listOrganizations(params)
      
      expect(spy).toHaveBeenCalledWith('/api/v1/admin/organizations', { params })
      expect(response).toEqual(payload)
    })
  })

  describe('getOrganization', () => {
    it('calls the correct endpoint', async () => {
      const orgId = 'org-123'
      const payload = { id: orgId, name: 'My Org' }
      const spy = vi.spyOn(api, 'get').mockResolvedValueOnce({ data: payload } as any)
      
      const response = await adminApi.getOrganization(orgId)
      
      expect(spy).toHaveBeenCalledWith(`/api/v1/admin/organizations/${orgId}`)
      expect(response).toEqual(payload)
    })
  })

  describe('Queue endpoints', () => {
    it('getQueues calls the correct endpoint', async () => {
      const spy = vi.spyOn(api, 'get').mockResolvedValueOnce({ data: [] } as any)
      await adminApi.getQueues()
      expect(spy).toHaveBeenCalledWith('/api/v1/admin/queues')
    })

    it('getStuckJobs calls the correct endpoint', async () => {
      const spy = vi.spyOn(api, 'get').mockResolvedValueOnce({ data: [] } as any)
      await adminApi.getStuckJobs('1h')
      expect(spy).toHaveBeenCalledWith('/api/v1/admin/queues/stuck', { params: { threshold: '1h' } })
    })

    it('retryJob calls the correct endpoint', async () => {
      const spy = vi.spyOn(api, 'post').mockResolvedValueOnce({ data: {} } as any)
      await adminApi.retryJob('job-1')
      expect(spy).toHaveBeenCalledWith('/api/v1/admin/queues/job-1/retry')
    })

    it('deleteJob calls the correct endpoint', async () => {
      const spy = vi.spyOn(api, 'delete').mockResolvedValueOnce({ data: {} } as any)
      await adminApi.deleteJob('job-1')
      expect(spy).toHaveBeenCalledWith('/api/v1/admin/queues/job-1')
    })
  })

  describe('Error endpoints', () => {
    it('listErrors calls the correct endpoint', async () => {
      const spy = vi.spyOn(api, 'get').mockResolvedValueOnce({ data: { data: [], total: 0 } } as any)
      await adminApi.listErrors({ page: 1, severity: 'error' })
      expect(spy).toHaveBeenCalledWith('/api/v1/admin/errors', { params: { page: 1, severity: 'error' } })
    })

    it('getError calls the correct endpoint', async () => {
      const spy = vi.spyOn(api, 'get').mockResolvedValueOnce({ data: {} } as any)
      await adminApi.getError('error-1')
      expect(spy).toHaveBeenCalledWith('/api/v1/admin/errors/error-1')
    })
  })

  describe('Chatbot endpoints', () => {
    it('listChatbots calls the correct endpoint', async () => {
      const spy = vi.spyOn(api, 'get').mockResolvedValueOnce({ data: { data: [], total: 0 } } as any)
      await adminApi.listChatbots({ page: 1, search: 'bot' })
      expect(spy).toHaveBeenCalledWith('/api/v1/admin/chatbots', { params: { page: 1, search: 'bot' } })
    })

    it('forceRefreshChatbot calls the correct endpoint', async () => {
      const spy = vi.spyOn(api, 'post').mockResolvedValueOnce({ data: {} } as any)
      await adminApi.forceRefreshChatbot('bot-1')
      expect(spy).toHaveBeenCalledWith('/api/v1/admin/chatbots/bot-1/force-refresh')
    })
  })

  describe('Audit endpoints', () => {
    it('listAuditLogs calls the correct endpoint', async () => {
      const spy = vi.spyOn(api, 'get').mockResolvedValueOnce({ data: { data: [], total: 0 } } as any)
      await adminApi.listAuditLogs({ page: 1 })
      expect(spy).toHaveBeenCalledWith('/api/v1/admin/audit-logs', { params: { page: 1 } })
    })
  })
})
