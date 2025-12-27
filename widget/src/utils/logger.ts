import { DEBUG_STORAGE_KEY } from '../constants'

const LOG_PREFIX = '[Botla Widget]'

function shouldLog(): boolean {
  try {
    return import.meta.env.DEV || 
           localStorage.getItem(DEBUG_STORAGE_KEY) === '1'
  } catch {
    return false
  }
}

export const logger = {
  debug: (message: string, ...args: unknown[]) => {
    if (shouldLog()) {
      console.debug(`${LOG_PREFIX} ${message}`, ...args)
    }
  },
  
  info: (message: string, ...args: unknown[]) => {
    if (shouldLog()) {
      console.info(`${LOG_PREFIX} ${message}`, ...args)
    }
  },
  
  warn: (message: string, ...args: unknown[]) => {
    console.warn(`${LOG_PREFIX} ${message}`, ...args)
  },
  
  error: (message: string, error?: unknown, ...args: unknown[]) => {
    console.error(`${LOG_PREFIX} ${message}`, error, ...args)
  }
}
