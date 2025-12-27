/**
 * Domain Layer - Re-exports
 * 
 * This module centralizes business logic that was previously scattered across UI components.
 * Import from '@/domain' instead of individual modules for convenience.
 */

// Plans
export {
  // Types
  type PlanCode,
  type PlanTier,
  type PlanFeatures,
  type PlanLimits,
  // Constants
  PLAN_LIMITS,
  PLAN_DISPLAY,
  // Functions
  getPlanLimits,
  normalizePlanCode,
  planCodeToTier,
  getPlanLabel,
  // Limit checks
  type LimitCheckResult,
  canCreateChatbot,
  canAddSource,
  canAddURL,
  canAddPDF,
  canUploadFile,
  canAddText,
  canRefreshSource,
  canUseHandoff,
  canCustomizeBranding,
  hasAnalyticsAccess,
  hasApiAccess,
  canUseOCR,
  canUseDynamicScraping,
  canUseSecureEmbed,
  getPlanFeatures,
  formatRemaining,
  getLimitStatusMessage,
} from './plans';

// Errors
export {
  // Types
  type AppError,
  type ErrorCode,
  type ErrorCategory,
  // Constants
  ERROR_CODES,
  // Functions
  parseError,
  createAppError,
  getUserMessage,
  getErrorAction,
  shouldRedirectToLogin,
  shouldShowUpgrade,
  isRecoverable,
  isRetryable,
  isLimitError,
  isAuthError,
  isValidationError,
  getErrorCategory,
  isKnownErrorCode,
} from './errors';

// Chatbot Validation
export {
  // Types
  type ValidationResult,
  type ValidationError,
  // Constants
  CHATBOT_NAME_CONSTRAINTS,
  CHATBOT_DESCRIPTION_CONSTRAINTS,
  SYSTEM_PROMPT_CONSTRAINTS,
  WELCOME_MESSAGE_CONSTRAINTS,
  URL_CONSTRAINTS,
  TEXT_SOURCE_CONSTRAINTS,
  // Functions
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
} from './chatbot/validation';
