/**
 * Centralized error handling utilities.
 * Provides consistent error parsing and user message generation.
 */

import { getErrorMessage as getI18nErrorMessage } from '@/i18n/errors';
import { 
  isRecoverable, 
  isRetryable, 
  isLimitError,
  getErrorCategory,
  type ErrorCategory,
} from './codes';

export * from './codes';

/**
 * Structured error representation for the application.
 */
export interface AppError {
  /** Original error code from the backend */
  code: string;
  /** Raw error message */
  message: string;
  /** Localized user-friendly message */
  userMessage: string;
  /** Whether the user can continue without action */
  recoverable: boolean;
  /** Whether the request can be retried */
  retryable: boolean;
  /** Whether this is a plan limit error */
  isLimitError: boolean;
  /** Error category for handling */
  category: ErrorCategory;
}

/**
 * API error response structure from backend.
 */
interface ApiErrorResponse {
  error?: string;
  code?: string;
  message?: string;
  details?: unknown;
}

/**
 * Parse an unknown error into a structured AppError.
 * Handles various error formats from Axios, fetch, and direct objects.
 */
export function parseError(error: unknown, lang: string = 'tr'): AppError {
  const code = extractErrorCode(error);
  const message = extractErrorMessage(error) || code;
  
  return {
    code,
    message,
    userMessage: getUserMessage(code, lang),
    recoverable: isRecoverable(code),
    retryable: isRetryable(code),
    isLimitError: isLimitError(code),
    category: getErrorCategory(code),
  };
}

/**
 * Extract error code from various error formats.
 */
function extractErrorCode(error: unknown): string {
  if (!error) return 'ERR_UNKNOWN';
  
  if (typeof error === 'string') {
    return error;
  }
  
  if (typeof error === 'object' && error !== null) {
    const err = error as Record<string, unknown>;
    
    // Axios error with response
    if (err.response && typeof err.response === 'object') {
      const response = err.response as Record<string, unknown>;
      const data = response.data as ApiErrorResponse | undefined;
      if (data?.code) return data.code;
      if (data?.error) return data.error;
    }
    
    // Direct error object
    if (typeof err.code === 'string') return err.code;
    if (typeof err.error === 'string') return err.error;
  }
  
  return 'ERR_UNKNOWN';
}

/**
 * Extract error message from various error formats.
 */
function extractErrorMessage(error: unknown): string | null {
  if (!error) return null;
  
  if (typeof error === 'string') return error;
  
  if (error instanceof Error) {
    return error.message;
  }
  
  if (typeof error === 'object' && error !== null) {
    const err = error as Record<string, unknown>;
    
    // Axios error with response
    if (err.response && typeof err.response === 'object') {
      const response = err.response as Record<string, unknown>;
      const data = response.data as ApiErrorResponse | undefined;
      if (data?.message) return data.message;
    }
    
    if (typeof err.message === 'string') return err.message;
  }
  
  return null;
}

/**
 * Get localized user message for an error code.
 */
export function getUserMessage(code: string, lang: string = 'tr'): string {
  const message = getI18nErrorMessage(code, lang);
  
  // If the i18n lookup returned the code itself, provide a fallback
  if (message === code) {
    return lang === 'tr' ? 'Bir hata oluştu' : 'An error occurred';
  }
  
  return message;
}

/**
 * Create an AppError from just a code.
 */
export function createAppError(code: string, lang: string = 'tr'): AppError {
  return {
    code,
    message: code,
    userMessage: getUserMessage(code, lang),
    recoverable: isRecoverable(code),
    retryable: isRetryable(code),
    isLimitError: isLimitError(code),
    category: getErrorCategory(code),
  };
}

/**
 * Get action suggestion based on error category.
 */
export function getErrorAction(error: AppError): 'retry' | 'login' | 'upgrade' | 'fix' | 'contact' | 'dismiss' {
  switch (error.category) {
    case 'auth':
      return 'login';
    case 'limit':
      return 'upgrade';
    case 'validation':
      return 'fix';
    case 'network':
    case 'server':
      return error.retryable ? 'retry' : 'contact';
    case 'permission':
      return 'contact';
    default:
      return 'dismiss';
  }
}

/**
 * Check if error should trigger a redirect to login.
 */
export function shouldRedirectToLogin(error: AppError): boolean {
  return error.category === 'auth' && !error.recoverable;
}

/**
 * Check if error should show upgrade prompt.
 */
export function shouldShowUpgrade(error: AppError): boolean {
  return error.isLimitError;
}
