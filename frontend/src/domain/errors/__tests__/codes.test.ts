import { describe, it, expect } from 'vitest';
import {
  ERROR_CODES,
  isRecoverable,
  isRetryable,
  isLimitError,
  isAuthError,
  isValidationError,
  getErrorCategory,
  isKnownErrorCode,
} from '../codes';

describe('domain/errors/codes', () => {
  describe('ERROR_CODES', () => {
    it('should define all HTTP standard errors', () => {
      expect(ERROR_CODES.BAD_REQUEST).toBe('BAD_REQUEST');
      expect(ERROR_CODES.UNAUTHORIZED).toBe('UNAUTHORIZED');
      expect(ERROR_CODES.FORBIDDEN).toBe('FORBIDDEN');
      expect(ERROR_CODES.NOT_FOUND).toBe('NOT_FOUND');
      expect(ERROR_CODES.TOO_MANY_REQUESTS).toBe('TOO_MANY_REQUESTS');
      expect(ERROR_CODES.INTERNAL_ERROR).toBe('INTERNAL_ERROR');
    });

    it('should define authentication errors', () => {
      expect(ERROR_CODES.ERR_INVALID_CREDENTIALS).toBe('ERR_INVALID_CREDENTIALS');
      expect(ERROR_CODES.ERR_EMAIL_EXISTS).toBe('ERR_EMAIL_EXISTS');
      expect(ERROR_CODES.ERR_PASSWORD_WEAK).toBe('ERR_PASSWORD_WEAK');
    });

    it('should define limit errors', () => {
      expect(ERROR_CODES.ERR_MAX_CHATBOTS_EXCEEDED).toBe('ERR_MAX_CHATBOTS_EXCEEDED');
      expect(ERROR_CODES.ERR_PDF_LIMIT_REACHED).toBe('ERR_PDF_LIMIT_REACHED');
      expect(ERROR_CODES.ERR_URL_LIMIT_REACHED).toBe('ERR_URL_LIMIT_REACHED');
    });

    it('should define processing errors', () => {
      expect(ERROR_CODES.ERR_SCRAPE_NETWORK).toBe('ERR_SCRAPE_NETWORK');
      expect(ERROR_CODES.ERR_PDF_PARSE_FAILED).toBe('ERR_PDF_PARSE_FAILED');
      expect(ERROR_CODES.ERR_EMBEDDING_FAILED).toBe('ERR_EMBEDDING_FAILED');
    });
  });

  describe('isRecoverable', () => {
    it('should return false for unrecoverable errors', () => {
      expect(isRecoverable('UNAUTHORIZED')).toBe(false);
      expect(isRecoverable('FORBIDDEN')).toBe(false);
      expect(isRecoverable('ERR_INVALID_CREDENTIALS')).toBe(false);
    });

    it('should return true for recoverable errors', () => {
      expect(isRecoverable('BAD_REQUEST')).toBe(true);
      expect(isRecoverable('TOO_MANY_REQUESTS')).toBe(true);
      expect(isRecoverable('ERR_PDF_LIMIT_REACHED')).toBe(true);
      expect(isRecoverable('ERR_UNKNOWN')).toBe(true);
    });
  });

  describe('isRetryable', () => {
    it('should return true for retryable errors', () => {
      expect(isRetryable('TOO_MANY_REQUESTS')).toBe(true);
      expect(isRetryable('INTERNAL_ERROR')).toBe(true);
      expect(isRetryable('SERVICE_UNAVAILABLE')).toBe(true);
      expect(isRetryable('ERR_SCRAPE_TIMEOUT')).toBe(true);
      expect(isRetryable('ERR_SCRAPE_NETWORK')).toBe(true);
    });

    it('should return false for non-retryable errors', () => {
      expect(isRetryable('BAD_REQUEST')).toBe(false);
      expect(isRetryable('UNAUTHORIZED')).toBe(false);
      expect(isRetryable('ERR_INVALID_CREDENTIALS')).toBe(false);
      expect(isRetryable('ERR_PDF_LIMIT_REACHED')).toBe(false);
    });
  });

  describe('isLimitError', () => {
    it('should return true for limit errors', () => {
      expect(isLimitError('ERR_MONTHLY_TOKENS_EXCEEDED')).toBe(true);
      expect(isLimitError('ERR_PDF_LIMIT_REACHED')).toBe(true);
      expect(isLimitError('ERR_URL_LIMIT_REACHED')).toBe(true);
      expect(isLimitError('ERR_MAX_CHATBOTS_EXCEEDED')).toBe(true);
      expect(isLimitError('PAYMENT_REQUIRED')).toBe(true);
    });

    it('should return false for non-limit errors', () => {
      expect(isLimitError('BAD_REQUEST')).toBe(false);
      expect(isLimitError('UNAUTHORIZED')).toBe(false);
      expect(isLimitError('ERR_INVALID_URL')).toBe(false);
    });
  });

  describe('isAuthError', () => {
    it('should return true for auth errors', () => {
      expect(isAuthError('UNAUTHORIZED')).toBe(true);
      expect(isAuthError('ERR_INVALID_CREDENTIALS')).toBe(true);
      expect(isAuthError('ERR_EMAIL_REQUIRED')).toBe(true);
      expect(isAuthError('ERR_PASSWORD_WEAK')).toBe(true);
    });

    it('should return false for non-auth errors', () => {
      expect(isAuthError('BAD_REQUEST')).toBe(false);
      expect(isAuthError('ERR_PDF_LIMIT_REACHED')).toBe(false);
    });
  });

  describe('isValidationError', () => {
    it('should return true for validation errors', () => {
      expect(isValidationError('BAD_REQUEST')).toBe(true);
      expect(isValidationError('ERR_INVALID_REQUEST_BODY')).toBe(true);
      expect(isValidationError('ERR_INVALID_URL')).toBe(true);
      expect(isValidationError('ERR_DUPLICATE_URL')).toBe(true);
    });

    it('should return false for non-validation errors', () => {
      expect(isValidationError('UNAUTHORIZED')).toBe(false);
      expect(isValidationError('ERR_PDF_LIMIT_REACHED')).toBe(false);
    });
  });

  describe('getErrorCategory', () => {
    it('should categorize auth errors', () => {
      expect(getErrorCategory('UNAUTHORIZED')).toBe('auth');
      expect(getErrorCategory('ERR_INVALID_CREDENTIALS')).toBe('auth');
    });

    it('should categorize validation errors', () => {
      expect(getErrorCategory('BAD_REQUEST')).toBe('validation');
      expect(getErrorCategory('ERR_INVALID_URL')).toBe('validation');
    });

    it('should categorize limit errors', () => {
      expect(getErrorCategory('ERR_PDF_LIMIT_REACHED')).toBe('limit');
      expect(getErrorCategory('PAYMENT_REQUIRED')).toBe('limit');
    });

    it('should categorize network errors', () => {
      expect(getErrorCategory('ERR_SCRAPE_NETWORK')).toBe('network');
      expect(getErrorCategory('SERVICE_UNAVAILABLE')).toBe('network');
    });

    it('should categorize permission errors', () => {
      expect(getErrorCategory('FORBIDDEN')).toBe('permission');
    });

    it('should categorize server errors', () => {
      expect(getErrorCategory('INTERNAL_ERROR')).toBe('server');
      expect(getErrorCategory('ERR_DATABASE_ERROR')).toBe('server');
    });

    it('should categorize processing errors', () => {
      expect(getErrorCategory('ERR_PDF_PARSE_FAILED')).toBe('processing');
      expect(getErrorCategory('ERR_CHUNKING_FAILED')).toBe('processing');
    });

    it('should return unknown for unrecognized errors', () => {
      expect(getErrorCategory('SOME_RANDOM_ERROR')).toBe('unknown');
    });
  });

  describe('isKnownErrorCode', () => {
    it('should return true for known error codes', () => {
      expect(isKnownErrorCode('BAD_REQUEST')).toBe(true);
      expect(isKnownErrorCode('ERR_INVALID_CREDENTIALS')).toBe(true);
      expect(isKnownErrorCode('ERR_PDF_LIMIT_REACHED')).toBe(true);
    });

    it('should return false for unknown error codes', () => {
      expect(isKnownErrorCode('RANDOM_ERROR')).toBe(false);
      expect(isKnownErrorCode('')).toBe(false);
    });
  });
});
