/**
 * Widget-specific error translations.
 * Only includes errors that can occur in the chat widget.
 */

export const widgetErrors = {
  en: {
    // Chat/Token errors
    ERR_MONTHLY_TOKENS_EXCEEDED: 'Usage limit reached. Please try again later.',
    CHAT_TIMEOUT_OR_INCOMPLETE: 'Request timed out. Please try again.',
    INTERNAL_ERROR: 'Something went wrong. Please try again.',
    TOO_MANY_REQUESTS: 'Too many requests. Please wait a moment.',
    
    // Handoff errors
    ERR_HANDOFF_EXISTS: 'A support request is already open.',
    ERR_HANDOFF_NOT_ENABLED: 'Support requests are not enabled.',
    ERR_HANDOFF_RATE_LIMITED: 'Please wait before requesting support again.',
  },
  tr: {
    // Chat/Token errors
    ERR_MONTHLY_TOKENS_EXCEEDED: 'Kullanım limiti doldu. Lütfen daha sonra tekrar deneyin.',
    CHAT_TIMEOUT_OR_INCOMPLETE: 'İstek zaman aşımına uğradı. Lütfen tekrar deneyin.',
    INTERNAL_ERROR: 'Bir hata oluştu. Lütfen tekrar deneyin.',
    TOO_MANY_REQUESTS: 'Çok fazla istek. Lütfen biraz bekleyin.',
    
    // Handoff errors
    ERR_HANDOFF_EXISTS: 'Zaten açık bir destek talebi var.',
    ERR_HANDOFF_NOT_ENABLED: 'Destek talepleri etkin değil.',
    ERR_HANDOFF_RATE_LIMITED: 'Tekrar destek talep etmeden önce lütfen bekleyin.',
  },
} as const;

/**
 * Get widget error message for an error code.
 */
export function getWidgetErrorMessage(code: string, lang: string = 'tr'): string {
  const messages = lang === 'tr' ? widgetErrors.tr : widgetErrors.en;
  return (messages as Record<string, string>)[code] ?? 
         (widgetErrors.en as Record<string, string>)[code] ?? 
         (lang === 'tr' ? 'Bir hata oluştu.' : 'An error occurred.');
}
