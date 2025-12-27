import type { ChatMessage, SessionData } from '../types'
import { STORAGE_PREFIX } from '../constants'

function storageKey(chatbotId: string): string {
  return `${STORAGE_PREFIX}${chatbotId}`
}

export function getSession(chatbotId: string): SessionData {
  try {
    const raw = localStorage.getItem(storageKey(chatbotId))
    if (raw) {
      const parsed = JSON.parse(raw) as SessionData
      if (parsed.sessionId && Array.isArray(parsed.messages)) {
        return parsed
      }
    }
  } catch (error) {
    console.warn('[Widget] Failed to parse session:', error)
  }
  
  const newSession: SessionData = {
    sessionId: crypto.randomUUID(),
    messages: []
  }
  saveSession(chatbotId, newSession)
  return newSession
}

export function saveSession(chatbotId: string, data: SessionData): void {
  try {
    localStorage.setItem(storageKey(chatbotId), JSON.stringify(data))
  } catch (error) {
    console.warn('[Widget] Failed to save session:', error)
  }
}

export function clearSession(chatbotId: string): void {
  try {
    localStorage.removeItem(storageKey(chatbotId))
  } catch (error) {
    console.warn('[Widget] Failed to clear session:', error)
  }
}

export function updateSessionMessages(
  chatbotId: string, 
  sessionId: string, 
  messages: ChatMessage[]
): void {
  saveSession(chatbotId, { sessionId, messages })
}

export function ensureSession(
  chatbotId: string, 
  currentSid: string, 
  setSid: (v: string) => void
): string {
  if (currentSid && currentSid.length > 0) return currentSid
  const session = getSession(chatbotId)
  setSid(session.sessionId)
  return session.sessionId
}
export function setSessionId(chatbotId: string, sessionId: string) {
  saveSession(chatbotId, { sessionId, messages: [] })
}
