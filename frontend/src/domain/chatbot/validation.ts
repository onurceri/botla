/**
 * Chatbot validation rules and utilities.
 * Centralizes validation logic for chatbot-related operations.
 */

/**
 * Validation result with field-specific errors.
 */
export interface ValidationResult {
  valid: boolean;
  errors: ValidationError[];
}

export interface ValidationError {
  field: string;
  code: string;
  message: string;
}

/**
 * Chatbot name constraints.
 */
export const CHATBOT_NAME_CONSTRAINTS = {
  minLength: 2,
  maxLength: 100,
} as const;

/**
 * Chatbot description constraints.
 */
export const CHATBOT_DESCRIPTION_CONSTRAINTS = {
  maxLength: 500,
} as const;

/**
 * System prompt constraints.
 */
export const SYSTEM_PROMPT_CONSTRAINTS = {
  maxLength: 4000,
} as const;

/**
 * Welcome message constraints.
 */
export const WELCOME_MESSAGE_CONSTRAINTS = {
  maxLength: 500,
} as const;

/**
 * URL constraints.
 */
export const URL_CONSTRAINTS = {
  maxLength: 2048,
  allowedProtocols: ['http:', 'https:'],
} as const;

/**
 * Text source constraints.
 */
export const TEXT_SOURCE_CONSTRAINTS = {
  minLength: 10,
  maxLength: 100000, // Can be overridden by plan limits
} as const;

/**
 * Validate chatbot name.
 */
export function validateChatbotName(name: string): ValidationResult {
  const errors: ValidationError[] = [];
  const trimmed = name.trim();

  if (!trimmed) {
    errors.push({
      field: 'name',
      code: 'ERR_NAME_REQUIRED',
      message: 'Chatbot adı gereklidir',
    });
  } else if (trimmed.length < CHATBOT_NAME_CONSTRAINTS.minLength) {
    errors.push({
      field: 'name',
      code: 'ERR_NAME_TOO_SHORT',
      message: `Chatbot adı en az ${CHATBOT_NAME_CONSTRAINTS.minLength} karakter olmalıdır`,
    });
  } else if (trimmed.length > CHATBOT_NAME_CONSTRAINTS.maxLength) {
    errors.push({
      field: 'name',
      code: 'ERR_NAME_TOO_LONG',
      message: `Chatbot adı en fazla ${CHATBOT_NAME_CONSTRAINTS.maxLength} karakter olmalıdır`,
    });
  }

  return { valid: errors.length === 0, errors };
}

/**
 * Validate chatbot description.
 */
export function validateChatbotDescription(description: string): ValidationResult {
  const errors: ValidationError[] = [];

  if (description.length > CHATBOT_DESCRIPTION_CONSTRAINTS.maxLength) {
    errors.push({
      field: 'description',
      code: 'ERR_DESCRIPTION_TOO_LONG',
      message: `Açıklama en fazla ${CHATBOT_DESCRIPTION_CONSTRAINTS.maxLength} karakter olmalıdır`,
    });
  }

  return { valid: errors.length === 0, errors };
}

/**
 * Validate system prompt.
 */
export function validateSystemPrompt(prompt: string): ValidationResult {
  const errors: ValidationError[] = [];

  if (prompt.length > SYSTEM_PROMPT_CONSTRAINTS.maxLength) {
    errors.push({
      field: 'systemPrompt',
      code: 'ERR_PROMPT_TOO_LONG',
      message: `Sistem promptu en fazla ${SYSTEM_PROMPT_CONSTRAINTS.maxLength} karakter olmalıdır`,
    });
  }

  return { valid: errors.length === 0, errors };
}

/**
 * Validate welcome message.
 */
export function validateWelcomeMessage(message: string): ValidationResult {
  const errors: ValidationError[] = [];

  if (message.length > WELCOME_MESSAGE_CONSTRAINTS.maxLength) {
    errors.push({
      field: 'welcomeMessage',
      code: 'ERR_WELCOME_TOO_LONG',
      message: `Karşılama mesajı en fazla ${WELCOME_MESSAGE_CONSTRAINTS.maxLength} karakter olmalıdır`,
    });
  }

  return { valid: errors.length === 0, errors };
}

/**
 * Validate URL format and constraints.
 */
export function validateURL(url: string): ValidationResult {
  const errors: ValidationError[] = [];
  const trimmed = url.trim();

  if (!trimmed) {
    errors.push({
      field: 'url',
      code: 'ERR_URL_REQUIRED',
      message: 'URL gereklidir',
    });
    return { valid: false, errors };
  }

  if (trimmed.length > URL_CONSTRAINTS.maxLength) {
    errors.push({
      field: 'url',
      code: 'ERR_URL_TOO_LONG',
      message: `URL en fazla ${URL_CONSTRAINTS.maxLength} karakter olmalıdır`,
    });
    return { valid: false, errors };
  }

  try {
    const parsedUrl = new URL(trimmed);
    
    if (!URL_CONSTRAINTS.allowedProtocols.includes(parsedUrl.protocol)) {
      errors.push({
        field: 'url',
        code: 'ERR_INVALID_URL_PROTOCOL',
        message: 'Sadece HTTP ve HTTPS protokolleri desteklenir',
      });
    }
  } catch {
    errors.push({
      field: 'url',
      code: 'ERR_INVALID_URL_FORMAT',
      message: 'Geçersiz URL formatı',
    });
  }

  return { valid: errors.length === 0, errors };
}

/**
 * Validate text source content.
 */
export function validateTextSource(text: string, maxLength?: number): ValidationResult {
  const errors: ValidationError[] = [];
  const trimmed = text.trim();
  const effectiveMaxLength = maxLength ?? TEXT_SOURCE_CONSTRAINTS.maxLength;

  if (!trimmed) {
    errors.push({
      field: 'text',
      code: 'ERR_TEXT_REQUIRED',
      message: 'Metin içeriği gereklidir',
    });
  } else if (trimmed.length < TEXT_SOURCE_CONSTRAINTS.minLength) {
    errors.push({
      field: 'text',
      code: 'ERR_TEXT_TOO_SHORT',
      message: `Metin en az ${TEXT_SOURCE_CONSTRAINTS.minLength} karakter olmalıdır`,
    });
  } else if (trimmed.length > effectiveMaxLength) {
    errors.push({
      field: 'text',
      code: 'ERR_TEXT_TOO_LONG',
      message: `Metin en fazla ${effectiveMaxLength} karakter olmalıdır`,
    });
  }

  return { valid: errors.length === 0, errors };
}

/**
 * Validate file for upload.
 */
export function validateFile(
  file: File,
  options: {
    maxSizeMB: number;
    allowedTypes: string[];
  }
): ValidationResult {
  const errors: ValidationError[] = [];
  const fileSizeMB = file.size / (1024 * 1024);

  if (fileSizeMB > options.maxSizeMB) {
    errors.push({
      field: 'file',
      code: 'ERR_FILE_TOO_LARGE',
      message: `Dosya boyutu en fazla ${options.maxSizeMB}MB olmalıdır`,
    });
  }

  if (!options.allowedTypes.includes(file.type)) {
    errors.push({
      field: 'file',
      code: 'ERR_INVALID_FILE_TYPE',
      message: 'Desteklenmeyen dosya formatı',
    });
  }

  return { valid: errors.length === 0, errors };
}

/**
 * Combine multiple validation results.
 */
export function combineValidations(...results: ValidationResult[]): ValidationResult {
  const allErrors = results.flatMap((r) => r.errors);
  return {
    valid: allErrors.length === 0,
    errors: allErrors,
  };
}

/**
 * Get first error message from validation result.
 */
export function getFirstError(result: ValidationResult): string | null {
  return result.errors[0]?.message ?? null;
}

/**
 * Get error for a specific field.
 */
export function getFieldError(result: ValidationResult, field: string): string | null {
  const error = result.errors.find((e) => e.field === field);
  return error?.message ?? null;
}
