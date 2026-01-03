/**
 * Session Utility Unit Tests
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { 
  getSession, 
  saveSession, 
  clearSession, 
  updateSessionMessages,
  ensureSession,
  setSessionId
} from '../utils/session'
import type { ChatMessage } from '../types'

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {}
  return {
    getItem: vi.fn((key: string) => store[key] ?? null),
    setItem: vi.fn((key: string, value: string) => {
      store[key] = value
    }),
    removeItem: vi.fn((key: string) => {
      delete store[key]
    }),
    clear: vi.fn(() => {
      store = {}
    }),
    get store() {
      return store
    }
  }
})()

Object.defineProperty(global, 'localStorage', {
  value: localStorageMock,
  writable: true
})

describe('Session Utilities', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorageMock.clear()
    vi.spyOn(crypto, 'randomUUID').mockReturnValue('a1b2c3d4-e5f6-7890-abcd-ef1234567890')
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('getSession', () => {
    it('creates new session when none exists', () => {
      const session = getSession('test-bot')
      
      expect(session.sessionId).toBe('a1b2c3d4-e5f6-7890-abcd-ef1234567890')
      expect(session.messages).toEqual([])
      expect(localStorageMock.setItem).toHaveBeenCalled()
    })

    it('returns existing session from localStorage', () => {
      const storedSession = JSON.stringify({
        sessionId: 'existing-session',
        messages: [{ role: 'user', content: 'Hello' }]
      })
      localStorageMock.setItem('chatbot_session_test-bot', storedSession)
      
      const session = getSession('test-bot')
      
      expect(session.sessionId).toBe('existing-session')
      expect(session.messages).toHaveLength(1)
    })

    it('creates new session on localStorage parse error', () => {
      localStorageMock.setItem('chatbot_session_test-bot', 'invalid json')
      
      const session = getSession('test-bot')
      
      expect(session.sessionId).toBe('a1b2c3d4-e5f6-7890-abcd-ef1234567890')
      expect(session.messages).toEqual([])
    })

    it('creates new session when stored data is invalid', () => {
      localStorageMock.setItem('chatbot_session_test-bot', JSON.stringify({
        // Missing sessionId or messages
        data: 'invalid'
      }))
      
      const session = getSession('test-bot')
      
      expect(session.sessionId).toBe('a1b2c3d4-e5f6-7890-abcd-ef1234567890')
    })
  })

  describe('saveSession', () => {
    it('saves session to localStorage', () => {
      saveSession('test-bot', {
        sessionId: 'session-123',
        messages: [{ role: 'user', content: 'Hello' }]
      })
      
      expect(localStorageMock.setItem).toHaveBeenCalledWith(
        'chatbot_session_test-bot',
        expect.any(String)
      )
    })

    it('handles localStorage errors gracefully', () => {
      localStorageMock.setItem.mockImplementationOnce(() => {
        throw new Error('Storage error')
      })
      
      // Should not throw
      expect(() => {
        saveSession('test-bot', {
          sessionId: 'session-123',
          messages: []
        })
      }).not.toThrow()
    })
  })

  describe('clearSession', () => {
    it('removes session from localStorage', () => {
      localStorageMock.setItem('chatbot_session_test-bot', '{"sessionId":"123"}')
      
      clearSession('test-bot')
      
      expect(localStorageMock.removeItem).toHaveBeenCalledWith('chatbot_session_test-bot')
    })

    it('handles non-existent session gracefully', () => {
      expect(() => {
        clearSession('non-existent')
      }).not.toThrow()
    })
  })

  describe('updateSessionMessages', () => {
    it('updates session with new messages', () => {
      const messages: ChatMessage[] = [
        { role: 'user', content: 'Hello' },
        { role: 'assistant', content: 'Hi there!' }
      ]
      
      updateSessionMessages('test-bot', 'session-456', messages)
      
      expect(localStorageMock.setItem).toHaveBeenCalledWith(
        'chatbot_session_test-bot',
        expect.stringContaining('session-456')
      )
    })
  })

  describe('ensureSession', () => {
    it('returns existing session ID when valid', () => {
      const setSid = vi.fn()
      
      const result = ensureSession('test-bot', 'existing-id', setSid)
      
      expect(result).toBe('existing-id')
      expect(setSid).not.toHaveBeenCalled()
    })

    it('creates new session when ID is empty', () => {
      const setSid = vi.fn()
      
      const result = ensureSession('test-bot', '', setSid)
      
      expect(result).toBe('a1b2c3d4-e5f6-7890-abcd-ef1234567890')
      expect(setSid).toHaveBeenCalledWith('a1b2c3d4-e5f6-7890-abcd-ef1234567890')
    })

    it('creates new session when ID is undefined', () => {
      const setSid = vi.fn()
      
      const result = ensureSession('test-bot', undefined as unknown as string, setSid)
      
      expect(result).toBe('a1b2c3d4-e5f6-7890-abcd-ef1234567890')
      expect(setSid).toHaveBeenCalledWith('a1b2c3d4-e5f6-7890-abcd-ef1234567890')
    })
  })

  describe('setSessionId', () => {
    it('sets new session ID and clears messages', () => {
      setSessionId('test-bot', 'new-session-id')
      
      expect(localStorageMock.setItem).toHaveBeenCalledWith(
        'chatbot_session_test-bot',
        expect.stringContaining('new-session-id')
      )
      expect(localStorageMock.setItem).toHaveBeenCalledWith(
        'chatbot_session_test-bot',
        expect.stringContaining('[]')
      )
    })
  })
})
