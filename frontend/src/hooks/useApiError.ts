import { useCallback } from 'react';
import { toast } from 'sonner';
import { getErrorMessage } from '../i18n/errors';

/**
 * Hook to handle API errors with localized messages.
 * Uses the current language from localStorage or defaults to 'tr'.
 */
export function useApiError() {
  const showError = useCallback((error: { code?: string; message?: string } | string) => {
    // Get language preference (default to Turkish)
    const lang = localStorage.getItem('language') ?? 'tr';
    
    let code: string;
    if (typeof error === 'string') {
      code = error;
    } else {
      code = error.code ?? 'INTERNAL_ERROR';
    }
    
    const message = getErrorMessage(code, lang);
    toast.error(message);
    
    return message;
  }, []);

  const handleApiError = useCallback((error: unknown) => {
    // Handle axios/fetch error responses
    if (error && typeof error === 'object') {
      const err = error as { response?: { data?: { code?: string } }; code?: string; message?: string };
      
      // Axios error with response
      if (err.response?.data?.code) {
        return showError({ code: err.response.data.code });
      }
      
      // Direct error object
      if (err.code) {
        return showError({ code: err.code });
      }
      
      // Fallback to message
      if (err.message) {
        showError({ code: 'INTERNAL_ERROR' });
        return err.message;
      }
    }
    
    // Unknown error
    return showError({ code: 'INTERNAL_ERROR' });
  }, [showError]);

  return { showError, handleApiError };
}
