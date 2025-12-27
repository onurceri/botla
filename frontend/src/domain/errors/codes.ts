/**
 * Error code definitions and categorization.
 * Centralizes error handling logic for the frontend.
 */

/**
 * All known error codes from the backend.
 */
export const ERROR_CODES = {
  // HTTP standard errors
  BAD_REQUEST: 'BAD_REQUEST',
  UNAUTHORIZED: 'UNAUTHORIZED',
  FORBIDDEN: 'FORBIDDEN',
  NOT_FOUND: 'NOT_FOUND',
  CONFLICT: 'CONFLICT',
  TOO_MANY_REQUESTS: 'TOO_MANY_REQUESTS',
  PAYMENT_REQUIRED: 'PAYMENT_REQUIRED',
  INTERNAL_ERROR: 'INTERNAL_ERROR',
  SERVICE_UNAVAILABLE: 'SERVICE_UNAVAILABLE',
  METHOD_NOT_ALLOWED: 'METHOD_NOT_ALLOWED',
  REQUEST_ENTITY_TOO_LARGE: 'REQUEST_ENTITY_TOO_LARGE',
  GONE: 'GONE',

  // Authentication errors
  ERR_EMAIL_REQUIRED: 'ERR_EMAIL_REQUIRED',
  ERR_PASSWORD_REQUIRED: 'ERR_PASSWORD_REQUIRED',
  ERR_EMAIL_AND_PASSWORD_REQUIRED: 'ERR_EMAIL_AND_PASSWORD_REQUIRED',
  ERR_INVALID_EMAIL_FORMAT: 'ERR_INVALID_EMAIL_FORMAT',
  ERR_PASSWORD_TOO_SHORT: 'ERR_PASSWORD_TOO_SHORT',
  ERR_PASSWORD_WEAK: 'ERR_PASSWORD_WEAK',
  ERR_EMAIL_EXISTS: 'ERR_EMAIL_EXISTS',
  ERR_INVALID_CREDENTIALS: 'ERR_INVALID_CREDENTIALS',
  ERR_INVALID_REQUEST_BODY: 'ERR_INVALID_REQUEST_BODY',
  ERR_DATABASE_ERROR: 'ERR_DATABASE_ERROR',
  ERR_FAILED_TO_HASH_PASSWORD: 'ERR_FAILED_TO_HASH_PASSWORD',
  ERR_FAILED_TO_CREATE_USER: 'ERR_FAILED_TO_CREATE_USER',
  ERR_INVALID_ID_FORMAT: 'ERR_INVALID_ID_FORMAT',
  ERR_MISSING_ID: 'ERR_MISSING_ID',

  // Chatbot/Source errors
  ERR_MONTHLY_TOKENS_EXCEEDED: 'ERR_MONTHLY_TOKENS_EXCEEDED',
  ERR_PDF_LIMIT_REACHED: 'ERR_PDF_LIMIT_REACHED',
  ERR_FILE_TOO_LARGE: 'ERR_FILE_TOO_LARGE',
  ERR_READD_COOLDOWN_ACTIVE: 'ERR_READD_COOLDOWN_ACTIVE',
  ERR_DUPLICATE_URL: 'ERR_DUPLICATE_URL',
  ERR_ONLY_URL_REFRESH: 'ERR_ONLY_URL_REFRESH',
  ERR_SOURCE_ALREADY_PROCESSING: 'ERR_SOURCE_ALREADY_PROCESSING',
  ERR_PLAN_REFRESH_UNAVAILABLE: 'ERR_PLAN_REFRESH_UNAVAILABLE',
  ERR_MONTHLY_REFRESH_EXCEEDED: 'ERR_MONTHLY_REFRESH_EXCEEDED',
  ERR_REFRESH_COOLDOWN_ACTIVE: 'ERR_REFRESH_COOLDOWN_ACTIVE',
  ERR_NO_URLS_PROVIDED: 'ERR_NO_URLS_PROVIDED',
  ERR_URL_LIMIT_REACHED: 'ERR_URL_LIMIT_REACHED',
  ERR_MONTHLY_INGESTION_EXCEEDED: 'ERR_MONTHLY_INGESTION_EXCEEDED',
  ERR_SITEMAP_PARSE_FAILED: 'ERR_SITEMAP_PARSE_FAILED',
  ERR_MAX_CHATBOTS_EXCEEDED: 'ERR_MAX_CHATBOTS_EXCEEDED',
  ERR_TEXT_TOO_LONG: 'ERR_TEXT_TOO_LONG',
  ERR_DUPLICATE_CONTENT: 'ERR_DUPLICATE_CONTENT',
  ERR_BLOCKED_URL: 'ERR_BLOCKED_URL',

  // Processing errors
  ERR_EMPTY_URL: 'ERR_EMPTY_URL',
  ERR_EMPTY_CONTENT: 'ERR_EMPTY_CONTENT',
  ERR_SCRAPE_NETWORK: 'ERR_SCRAPE_NETWORK',
  ERR_SCRAPE_TIMEOUT: 'ERR_SCRAPE_TIMEOUT',
  ERR_SCRAPE_FORBIDDEN: 'ERR_SCRAPE_FORBIDDEN',
  ERR_INVALID_URL: 'ERR_INVALID_URL',
  ERR_EMPTY_FILE_PATH: 'ERR_EMPTY_FILE_PATH',
  ERR_PDF_DOWNLOAD_FAILED: 'ERR_PDF_DOWNLOAD_FAILED',
  ERR_PDF_PARSE_FAILED: 'ERR_PDF_PARSE_FAILED',
  ERR_STORAGE_REQUIRED: 'ERR_STORAGE_REQUIRED',
  ERR_CHUNKING_FAILED: 'ERR_CHUNKING_FAILED',
  ERR_EMBEDDING_FAILED: 'ERR_EMBEDDING_FAILED',
  ERR_LLM_NOT_SUPPORTED: 'ERR_LLM_NOT_SUPPORTED',

  // Action errors
  ERR_NAME_AND_ACTION_TYPE_REQUIRED: 'ERR_NAME_AND_ACTION_TYPE_REQUIRED',
  ERR_INVALID_STATUS: 'ERR_INVALID_STATUS',

  // Handoff errors
  ERR_HANDOFF_EXISTS: 'ERR_HANDOFF_EXISTS',
  ERR_HANDOFF_NOT_FOUND: 'ERR_HANDOFF_NOT_FOUND',
  ERR_HANDOFF_EXPIRED: 'ERR_HANDOFF_EXPIRED',
  ERR_HANDOFF_CLOSED: 'ERR_HANDOFF_CLOSED',
  ERR_HANDOFF_RATE_LIMITED: 'ERR_HANDOFF_RATE_LIMITED',
  ERR_HANDOFF_NOT_ENABLED: 'ERR_HANDOFF_NOT_ENABLED',

  // Unknown
  ERR_UNKNOWN: 'ERR_UNKNOWN',
} as const;

export type ErrorCode = (typeof ERROR_CODES)[keyof typeof ERROR_CODES];

/**
 * Error categories for grouping and handling.
 */
export type ErrorCategory = 
  | 'auth'
  | 'validation'
  | 'limit'
  | 'network'
  | 'processing'
  | 'permission'
  | 'server'
  | 'unknown';

/**
 * Errors that are considered unrecoverable (require user action like re-login).
 */
const UNRECOVERABLE_ERRORS: ErrorCode[] = [
  'UNAUTHORIZED',
  'FORBIDDEN',
  'ERR_INVALID_CREDENTIALS',
];

/**
 * Errors that are safe to retry.
 */
const RETRYABLE_ERRORS: ErrorCode[] = [
  'TOO_MANY_REQUESTS',
  'INTERNAL_ERROR',
  'SERVICE_UNAVAILABLE',
  'ERR_SCRAPE_TIMEOUT',
  'ERR_SCRAPE_NETWORK',
  'ERR_EMBEDDING_FAILED',
  'ERR_DATABASE_ERROR',
];

/**
 * Errors related to plan limits.
 */
const LIMIT_ERRORS: ErrorCode[] = [
  'ERR_MONTHLY_TOKENS_EXCEEDED',
  'ERR_PDF_LIMIT_REACHED',
  'ERR_URL_LIMIT_REACHED',
  'ERR_MAX_CHATBOTS_EXCEEDED',
  'ERR_MONTHLY_INGESTION_EXCEEDED',
  'ERR_MONTHLY_REFRESH_EXCEEDED',
  'ERR_TEXT_TOO_LONG',
  'ERR_FILE_TOO_LARGE',
  'PAYMENT_REQUIRED',
];

/**
 * Errors related to authentication.
 */
const AUTH_ERRORS: ErrorCode[] = [
  'UNAUTHORIZED',
  'ERR_INVALID_CREDENTIALS',
  'ERR_EMAIL_REQUIRED',
  'ERR_PASSWORD_REQUIRED',
  'ERR_EMAIL_AND_PASSWORD_REQUIRED',
  'ERR_EMAIL_EXISTS',
  'ERR_PASSWORD_TOO_SHORT',
  'ERR_PASSWORD_WEAK',
  'ERR_INVALID_EMAIL_FORMAT',
];

/**
 * Errors related to validation.
 */
const VALIDATION_ERRORS: ErrorCode[] = [
  'BAD_REQUEST',
  'ERR_INVALID_REQUEST_BODY',
  'ERR_INVALID_ID_FORMAT',
  'ERR_MISSING_ID',
  'ERR_INVALID_URL',
  'ERR_NO_URLS_PROVIDED',
  'ERR_DUPLICATE_URL',
  'ERR_DUPLICATE_CONTENT',
  'ERR_BLOCKED_URL',
];

/**
 * Check if an error is recoverable.
 */
export function isRecoverable(code: string): boolean {
  return !UNRECOVERABLE_ERRORS.includes(code as ErrorCode);
}

/**
 * Check if an error is retryable.
 */
export function isRetryable(code: string): boolean {
  return RETRYABLE_ERRORS.includes(code as ErrorCode);
}

/**
 * Check if an error is a limit error (requires plan upgrade).
 */
export function isLimitError(code: string): boolean {
  return LIMIT_ERRORS.includes(code as ErrorCode);
}

/**
 * Check if an error is an authentication error.
 */
export function isAuthError(code: string): boolean {
  return AUTH_ERRORS.includes(code as ErrorCode);
}

/**
 * Check if an error is a validation error.
 */
export function isValidationError(code: string): boolean {
  return VALIDATION_ERRORS.includes(code as ErrorCode);
}

/**
 * Get the category of an error code.
 */
export function getErrorCategory(code: string): ErrorCategory {
  if (AUTH_ERRORS.includes(code as ErrorCode)) return 'auth';
  if (VALIDATION_ERRORS.includes(code as ErrorCode)) return 'validation';
  if (LIMIT_ERRORS.includes(code as ErrorCode)) return 'limit';
  if (['ERR_SCRAPE_NETWORK', 'ERR_SCRAPE_TIMEOUT', 'SERVICE_UNAVAILABLE'].includes(code)) return 'network';
  if (['FORBIDDEN', 'ERR_HANDOFF_NOT_ENABLED'].includes(code)) return 'permission';
  if (['INTERNAL_ERROR', 'ERR_DATABASE_ERROR'].includes(code)) return 'server';
  if (code.startsWith('ERR_')) return 'processing';
  return 'unknown';
}

/**
 * Check if a string is a known error code.
 */
export function isKnownErrorCode(code: string): code is ErrorCode {
  return Object.values(ERROR_CODES).includes(code as ErrorCode);
}
