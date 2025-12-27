import { describe, it, expect } from 'vitest';
import {
  CHATBOT_NAME_CONSTRAINTS,
  CHATBOT_DESCRIPTION_CONSTRAINTS,
  SYSTEM_PROMPT_CONSTRAINTS,
  WELCOME_MESSAGE_CONSTRAINTS,
  URL_CONSTRAINTS,
  TEXT_SOURCE_CONSTRAINTS,
  validateChatbotName,
  validateChatbotDescription,
  validateSystemPrompt,
  validateWelcomeMessage,
  validateURL,
  validateTextSource,
  validateFile,
  combineValidations,
  getFirstError,
  getFieldError,
} from '../validation';

describe('domain/chatbot/validation', () => {
  describe('constraints', () => {
    it('should define chatbot name constraints', () => {
      expect(CHATBOT_NAME_CONSTRAINTS.minLength).toBe(2);
      expect(CHATBOT_NAME_CONSTRAINTS.maxLength).toBe(100);
    });

    it('should define description constraints', () => {
      expect(CHATBOT_DESCRIPTION_CONSTRAINTS.maxLength).toBe(500);
    });

    it('should define system prompt constraints', () => {
      expect(SYSTEM_PROMPT_CONSTRAINTS.maxLength).toBe(4000);
    });

    it('should define welcome message constraints', () => {
      expect(WELCOME_MESSAGE_CONSTRAINTS.maxLength).toBe(500);
    });

    it('should define URL constraints', () => {
      expect(URL_CONSTRAINTS.maxLength).toBe(2048);
      expect(URL_CONSTRAINTS.allowedProtocols).toContain('https:');
      expect(URL_CONSTRAINTS.allowedProtocols).toContain('http:');
    });

    it('should define text source constraints', () => {
      expect(TEXT_SOURCE_CONSTRAINTS.minLength).toBe(10);
      expect(TEXT_SOURCE_CONSTRAINTS.maxLength).toBeGreaterThan(1000);
    });
  });

  describe('validateChatbotName', () => {
    it('should accept valid name', () => {
      const result = validateChatbotName('My Chatbot');
      expect(result.valid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    it('should reject empty name', () => {
      const result = validateChatbotName('');
      expect(result.valid).toBe(false);
      expect(result.errors[0].code).toBe('ERR_NAME_REQUIRED');
    });

    it('should reject whitespace-only name', () => {
      const result = validateChatbotName('   ');
      expect(result.valid).toBe(false);
      expect(result.errors[0].code).toBe('ERR_NAME_REQUIRED');
    });

    it('should reject name that is too short', () => {
      const result = validateChatbotName('A');
      expect(result.valid).toBe(false);
      expect(result.errors[0].code).toBe('ERR_NAME_TOO_SHORT');
    });

    it('should reject name that is too long', () => {
      const longName = 'A'.repeat(101);
      const result = validateChatbotName(longName);
      expect(result.valid).toBe(false);
      expect(result.errors[0].code).toBe('ERR_NAME_TOO_LONG');
    });

    it('should trim and validate', () => {
      const result = validateChatbotName('  AB  ');
      expect(result.valid).toBe(true);
    });
  });

  describe('validateChatbotDescription', () => {
    it('should accept valid description', () => {
      const result = validateChatbotDescription('This is my chatbot description.');
      expect(result.valid).toBe(true);
    });

    it('should accept empty description', () => {
      const result = validateChatbotDescription('');
      expect(result.valid).toBe(true);
    });

    it('should reject description that is too long', () => {
      const longDesc = 'A'.repeat(501);
      const result = validateChatbotDescription(longDesc);
      expect(result.valid).toBe(false);
      expect(result.errors[0].code).toBe('ERR_DESCRIPTION_TOO_LONG');
    });
  });

  describe('validateSystemPrompt', () => {
    it('should accept valid prompt', () => {
      const result = validateSystemPrompt('You are a helpful assistant.');
      expect(result.valid).toBe(true);
    });

    it('should accept empty prompt', () => {
      const result = validateSystemPrompt('');
      expect(result.valid).toBe(true);
    });

    it('should reject prompt that is too long', () => {
      const longPrompt = 'A'.repeat(4001);
      const result = validateSystemPrompt(longPrompt);
      expect(result.valid).toBe(false);
      expect(result.errors[0].code).toBe('ERR_PROMPT_TOO_LONG');
    });
  });

  describe('validateWelcomeMessage', () => {
    it('should accept valid message', () => {
      const result = validateWelcomeMessage('Welcome! How can I help?');
      expect(result.valid).toBe(true);
    });

    it('should reject message that is too long', () => {
      const longMessage = 'A'.repeat(501);
      const result = validateWelcomeMessage(longMessage);
      expect(result.valid).toBe(false);
      expect(result.errors[0].code).toBe('ERR_WELCOME_TOO_LONG');
    });
  });

  describe('validateURL', () => {
    it('should accept valid HTTPS URL', () => {
      const result = validateURL('https://example.com');
      expect(result.valid).toBe(true);
    });

    it('should accept valid HTTP URL', () => {
      const result = validateURL('http://example.com');
      expect(result.valid).toBe(true);
    });

    it('should accept URL with path', () => {
      const result = validateURL('https://example.com/path/to/page');
      expect(result.valid).toBe(true);
    });

    it('should reject empty URL', () => {
      const result = validateURL('');
      expect(result.valid).toBe(false);
      expect(result.errors[0].code).toBe('ERR_URL_REQUIRED');
    });

    it('should reject invalid URL format', () => {
      const result = validateURL('not-a-url');
      expect(result.valid).toBe(false);
      expect(result.errors[0].code).toBe('ERR_INVALID_URL_FORMAT');
    });

    it('should reject non-HTTP protocols', () => {
      const result = validateURL('ftp://example.com');
      expect(result.valid).toBe(false);
      expect(result.errors[0].code).toBe('ERR_INVALID_URL_PROTOCOL');
    });

    it('should reject URL that is too long', () => {
      const longUrl = 'https://example.com/' + 'a'.repeat(2050);
      const result = validateURL(longUrl);
      expect(result.valid).toBe(false);
      expect(result.errors[0].code).toBe('ERR_URL_TOO_LONG');
    });

    it('should trim URL', () => {
      const result = validateURL('  https://example.com  ');
      expect(result.valid).toBe(true);
    });
  });

  describe('validateTextSource', () => {
    it('should accept valid text', () => {
      const text = 'This is a valid text source content.';
      const result = validateTextSource(text);
      expect(result.valid).toBe(true);
    });

    it('should reject empty text', () => {
      const result = validateTextSource('');
      expect(result.valid).toBe(false);
      expect(result.errors[0].code).toBe('ERR_TEXT_REQUIRED');
    });

    it('should reject text that is too short', () => {
      const result = validateTextSource('Short');
      expect(result.valid).toBe(false);
      expect(result.errors[0].code).toBe('ERR_TEXT_TOO_SHORT');
    });

    it('should reject text that exceeds max length', () => {
      const longText = 'A'.repeat(100001);
      const result = validateTextSource(longText);
      expect(result.valid).toBe(false);
      expect(result.errors[0].code).toBe('ERR_TEXT_TOO_LONG');
    });

    it('should use custom max length when provided', () => {
      const text = 'A'.repeat(6000);
      const resultWithDefault = validateTextSource(text); // default is 100000
      const resultWithCustom = validateTextSource(text, 5000);
      expect(resultWithDefault.valid).toBe(true);
      expect(resultWithCustom.valid).toBe(false);
    });
  });

  describe('validateFile', () => {
    it('should accept valid PDF file', () => {
      const file = new File(['content'], 'test.pdf', { type: 'application/pdf' });
      const result = validateFile(file, {
        maxSizeMB: 10,
        allowedTypes: ['application/pdf'],
      });
      expect(result.valid).toBe(true);
    });

    it('should reject file that is too large', () => {
      // Create a file-like object with specific size
      const content = new ArrayBuffer(11 * 1024 * 1024); // 11MB
      const file = new File([content], 'large.pdf', { type: 'application/pdf' });
      const result = validateFile(file, {
        maxSizeMB: 10,
        allowedTypes: ['application/pdf'],
      });
      expect(result.valid).toBe(false);
      expect(result.errors[0].code).toBe('ERR_FILE_TOO_LARGE');
    });

    it('should reject unsupported file type', () => {
      const file = new File(['content'], 'test.exe', { type: 'application/x-msdownload' });
      const result = validateFile(file, {
        maxSizeMB: 10,
        allowedTypes: ['application/pdf'],
      });
      expect(result.valid).toBe(false);
      expect(result.errors[0].code).toBe('ERR_INVALID_FILE_TYPE');
    });
  });

  describe('combineValidations', () => {
    it('should return valid when all are valid', () => {
      const result1 = validateChatbotName('Valid Name');
      const result2 = validateChatbotDescription('Valid description');
      const combined = combineValidations(result1, result2);
      expect(combined.valid).toBe(true);
      expect(combined.errors).toHaveLength(0);
    });

    it('should combine errors from multiple validations', () => {
      const result1 = validateChatbotName('');
      const result2 = validateURL('invalid');
      const combined = combineValidations(result1, result2);
      expect(combined.valid).toBe(false);
      expect(combined.errors.length).toBeGreaterThanOrEqual(2);
    });
  });

  describe('getFirstError', () => {
    it('should return first error message', () => {
      const result = validateChatbotName('');
      const error = getFirstError(result);
      expect(error).toBe('Chatbot adı gereklidir');
    });

    it('should return null for valid result', () => {
      const result = validateChatbotName('Valid Name');
      const error = getFirstError(result);
      expect(error).toBeNull();
    });
  });

  describe('getFieldError', () => {
    it('should return error for specific field', () => {
      const result = validateChatbotName('');
      const error = getFieldError(result, 'name');
      expect(error).toBe('Chatbot adı gereklidir');
    });

    it('should return null for field with no error', () => {
      const result = validateChatbotName('Valid Name');
      const error = getFieldError(result, 'name');
      expect(error).toBeNull();
    });

    it('should return null for non-existent field', () => {
      const result = validateChatbotName('');
      const error = getFieldError(result, 'description');
      expect(error).toBeNull();
    });
  });
});
