import { describe, it, expect } from 'vitest';
import {
  parseError,
  createAppError,
  getUserMessage,
  getErrorAction,
  shouldRedirectToLogin,
  shouldShowUpgrade,
} from '../index';

describe('domain/errors', () => {
  describe('parseError', () => {
    it('should parse string error', () => {
      const result = parseError('ERR_INVALID_CREDENTIALS');
      expect(result.code).toBe('ERR_INVALID_CREDENTIALS');
      expect(result.message).toBe('ERR_INVALID_CREDENTIALS');
      expect(result.recoverable).toBe(false);
      expect(result.category).toBe('auth');
    });

    it('should parse object with error field', () => {
      const result = parseError({ error: 'ERR_PDF_LIMIT_REACHED' });
      expect(result.code).toBe('ERR_PDF_LIMIT_REACHED');
      expect(result.isLimitError).toBe(true);
    });

    it('should parse object with code field', () => {
      const result = parseError({ code: 'BAD_REQUEST' });
      expect(result.code).toBe('BAD_REQUEST');
      expect(result.category).toBe('validation');
    });

    it('should parse Axios-like error', () => {
      const axiosError = {
        response: {
          data: {
            code: 'UNAUTHORIZED',
            message: 'Invalid token',
          },
        },
      };
      const result = parseError(axiosError);
      expect(result.code).toBe('UNAUTHORIZED');
      expect(result.recoverable).toBe(false);
    });

    it('should handle null/undefined', () => {
      expect(parseError(null).code).toBe('ERR_UNKNOWN');
      expect(parseError(undefined).code).toBe('ERR_UNKNOWN');
    });

    it('should set retryable flag correctly', () => {
      expect(parseError('TOO_MANY_REQUESTS').retryable).toBe(true);
      expect(parseError('INTERNAL_ERROR').retryable).toBe(true);
      expect(parseError('BAD_REQUEST').retryable).toBe(false);
    });

    it('should set isLimitError flag correctly', () => {
      expect(parseError('ERR_MAX_CHATBOTS_EXCEEDED').isLimitError).toBe(true);
      expect(parseError('PAYMENT_REQUIRED').isLimitError).toBe(true);
      expect(parseError('BAD_REQUEST').isLimitError).toBe(false);
    });
  });

  describe('createAppError', () => {
    it('should create AppError from code', () => {
      const error = createAppError('ERR_PDF_LIMIT_REACHED');
      expect(error.code).toBe('ERR_PDF_LIMIT_REACHED');
      expect(error.isLimitError).toBe(true);
      expect(error.category).toBe('limit');
    });

    it('should use provided language', () => {
      const errorTr = createAppError('ERR_INVALID_CREDENTIALS', 'tr');
      const errorEn = createAppError('ERR_INVALID_CREDENTIALS', 'en');
      // Messages should differ by language
      expect(errorTr.userMessage).toBeTruthy();
      expect(errorEn.userMessage).toBeTruthy();
    });
  });

  describe('getUserMessage', () => {
    it('should return Turkish message by default', () => {
      const message = getUserMessage('ERR_INVALID_CREDENTIALS');
      expect(message).toBe('Geçersiz e-posta veya şifre');
    });

    it('should return English message when requested', () => {
      const message = getUserMessage('ERR_INVALID_CREDENTIALS', 'en');
      expect(message).toBe('Invalid email or password');
    });

    it('should return fallback for unknown codes', () => {
      const message = getUserMessage('SOME_UNKNOWN_ERROR', 'tr');
      expect(message).toBe('Bir hata oluştu');
    });

    it('should return English fallback', () => {
      const message = getUserMessage('SOME_UNKNOWN_ERROR', 'en');
      expect(message).toBe('An error occurred');
    });
  });

  describe('getErrorAction', () => {
    it('should suggest login for auth errors', () => {
      const error = parseError('UNAUTHORIZED');
      expect(getErrorAction(error)).toBe('login');
    });

    it('should suggest upgrade for limit errors', () => {
      const error = parseError('ERR_MAX_CHATBOTS_EXCEEDED');
      expect(getErrorAction(error)).toBe('upgrade');
    });

    it('should suggest fix for validation errors', () => {
      const error = parseError('BAD_REQUEST');
      expect(getErrorAction(error)).toBe('fix');
    });

    it('should suggest retry for retryable network errors', () => {
      const error = parseError('ERR_SCRAPE_NETWORK');
      expect(getErrorAction(error)).toBe('retry');
    });

    it('should suggest contact for permission errors', () => {
      // FORBIDDEN is a permission error - user should contact support
      const error = parseError('FORBIDDEN');
      expect(getErrorAction(error)).toBe('contact');
    });

    it('should suggest dismiss for unknown errors', () => {
      const error = parseError('SOME_UNKNOWN');
      expect(getErrorAction(error)).toBe('dismiss');
    });
  });

  describe('shouldRedirectToLogin', () => {
    it('should return true for unrecoverable auth errors', () => {
      const error = parseError('UNAUTHORIZED');
      expect(shouldRedirectToLogin(error)).toBe(true);
    });

    it('should return false for recoverable auth errors', () => {
      // Password weak is an auth error but recoverable (user can try again)
      const error = parseError('ERR_PASSWORD_WEAK');
      expect(shouldRedirectToLogin(error)).toBe(false);
    });

    it('should return false for non-auth errors', () => {
      const error = parseError('ERR_PDF_LIMIT_REACHED');
      expect(shouldRedirectToLogin(error)).toBe(false);
    });
  });

  describe('shouldShowUpgrade', () => {
    it('should return true for limit errors', () => {
      expect(shouldShowUpgrade(parseError('ERR_MAX_CHATBOTS_EXCEEDED'))).toBe(true);
      expect(shouldShowUpgrade(parseError('ERR_PDF_LIMIT_REACHED'))).toBe(true);
      expect(shouldShowUpgrade(parseError('PAYMENT_REQUIRED'))).toBe(true);
    });

    it('should return false for non-limit errors', () => {
      expect(shouldShowUpgrade(parseError('BAD_REQUEST'))).toBe(false);
      expect(shouldShowUpgrade(parseError('UNAUTHORIZED'))).toBe(false);
    });
  });
});
