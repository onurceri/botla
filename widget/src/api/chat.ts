/**
 * Chat API module for widget communication with the backend.
 * Provides a testable interface for sending chat messages.
 */

export interface SendMessageResult {
  message?: string
  response?: string
  messageId?: string
  handoffRequestId?: string
  error?: string
  rateLimited?: boolean
}

export interface SendMessageOptions {
  apiBase?: string
  sessionId?: string
  embedToken?: string
  captchaToken?: string
}

const MAX_MESSAGE_LENGTH = 4000

/**
 * Sends a message to the chatbot API.
 * 
 * @param chatbotId - The chatbot ID to send the message to
 * @param message - The message content
 * @param options - Additional options like API base URL, session ID, etc.
 * @returns Promise with the result containing either response or error
 */
export async function sendMessage(
  chatbotId: string,
  message: string,
  options: SendMessageOptions = {}
): Promise<SendMessageResult> {
  const { apiBase = '', sessionId, embedToken, captchaToken } = options

  // Truncate overly long messages
  let processedMessage = message
  if (message.length > MAX_MESSAGE_LENGTH) {
    processedMessage = message.slice(0, MAX_MESSAGE_LENGTH)
  }

  const url = `${apiBase}/api/v1/public/chatbots/${encodeURIComponent(chatbotId)}/chat`

  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  if (embedToken) {
    headers['X-Embed-Token'] = embedToken
  }

  try {
    const res = await fetch(url, {
      method: 'POST',
      headers,
      body: JSON.stringify({
        message: processedMessage,
        session_id: sessionId,
        captcha_token: captchaToken,
      }),
    })

    if (!res.ok) {
      // Handle specific HTTP status codes
      if (res.status === 404) {
        let errorData
        try {
          errorData = await res.json()
        } catch {
          errorData = { error: 'ERR_NOT_FOUND' }
        }
        return { error: errorData.error || 'ERR_NOT_FOUND' }
      }

      if (res.status === 429) {
        let errorData
        try {
          errorData = await res.json()
        } catch {
          errorData = { error: 'ERR_RATE_LIMITED' }
        }
        return { error: errorData.error || 'ERR_RATE_LIMITED', rateLimited: true }
      }

      return { error: `HTTP ${res.status}` }
    }

    const data = await res.json()
    return {
      message: processedMessage,
      response: data.response,
      messageId: data.message_id,
      handoffRequestId: data.handoff_request_id,
    }
  } catch (err) {
    const error = err instanceof Error ? err : new Error(String(err))
    // Check if it's a network error
    if (error.message.toLowerCase().includes('network') || 
        error.message.toLowerCase().includes('fetch') ||
        error.message.toLowerCase().includes('failed')) {
      return { error: 'network error: ' + error.message }
    }
    return { error: error.message }
  }
}

export { MAX_MESSAGE_LENGTH }
