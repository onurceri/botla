import { useCallback } from 'react';
import { toast } from 'sonner';
import { parseError, shouldRedirectToLogin, shouldShowUpgrade, type AppError } from '@/domain';

/**
 * Hook to handle API errors with localized messages.
 * Uses the domain layer for centralized error handling.
 */
export function useApiError() {
  const showError = useCallback((error: { code?: string; message?: string } | string) => {
    // Get language preference (default to Turkish)
    const lang = localStorage.getItem('language') ?? 'tr';
    
    const appError = parseError(error, lang);
    toast.error(appError.userMessage);
    
    return appError.userMessage;
  }, []);

  const handleApiError = useCallback((error: unknown): AppError => {
    // Get language preference (default to Turkish)
    const lang = localStorage.getItem('language') ?? 'tr';
    
    const appError = parseError(error, lang);
    
    // Show the error toast
    toast.error(appError.userMessage);
    
    // Handle special cases
    if (shouldRedirectToLogin(appError)) {
      // Navigation should be handled by the caller
      console.warn('Auth error detected - redirect to login recommended');
    }
    
    if (shouldShowUpgrade(appError)) {
      // Upgrade prompt should be handled by the caller
      console.info('Limit error detected - upgrade prompt recommended');
    }
    
    return appError;
  }, []);

  return { showError, handleApiError };
}

