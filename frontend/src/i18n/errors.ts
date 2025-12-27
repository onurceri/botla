/**
 * Error code translations for API responses.
 * Backend returns machine-readable codes, frontend handles display messages.
 */

export type ErrorCode = keyof typeof errorMessages.en;

export const errorMessages = {
  en: {
    // HTTP standard errors
    BAD_REQUEST: 'Invalid request',
    UNAUTHORIZED: 'Unauthorized',
    FORBIDDEN: 'Access denied',
    NOT_FOUND: 'Resource not found',
    CONFLICT: 'Conflict with existing resource',
    TOO_MANY_REQUESTS: 'Too many requests. Please try again later.',
    PAYMENT_REQUIRED: 'Upgrade required',
    INTERNAL_ERROR: 'An error occurred. Please try again.',
    SERVICE_UNAVAILABLE: 'Service unavailable',
    METHOD_NOT_ALLOWED: 'Method not allowed',
    REQUEST_ENTITY_TOO_LARGE: 'Request too large',
    GONE: 'Resource no longer available',

    // Authentication errors
    ERR_EMAIL_REQUIRED: 'Email is required',
    ERR_PASSWORD_REQUIRED: 'Password is required',
    ERR_EMAIL_AND_PASSWORD_REQUIRED: 'Email and password are required',
    ERR_INVALID_EMAIL_FORMAT: 'Invalid email format',
    ERR_PASSWORD_TOO_SHORT: 'Password must be at least 8 characters',
    ERR_PASSWORD_WEAK: 'Password must contain uppercase, lowercase, number, and special character (@$!%*?&)',
    ERR_EMAIL_EXISTS: 'An account with this email already exists',
    ERR_INVALID_CREDENTIALS: 'Invalid email or password',
    ERR_INVALID_REQUEST_BODY: 'Invalid request',
    ERR_DATABASE_ERROR: 'A server error occurred. Please try again.',
    ERR_FAILED_TO_HASH_PASSWORD: 'A server error occurred',
    ERR_FAILED_TO_CREATE_USER: 'Failed to create account. Please try again.',
    ERR_INVALID_ID_FORMAT: 'Invalid ID format',
    ERR_MISSING_ID: 'ID is required',

    // Chatbot/Source errors
    ERR_MONTHLY_TOKENS_EXCEEDED: 'Monthly usage limit reached',
    ERR_PDF_LIMIT_REACHED: 'PDF limit reached for this chatbot',
    ERR_FILE_TOO_LARGE: 'File exceeds maximum size limit',
    ERR_READD_COOLDOWN_ACTIVE: 'Please wait before re-adding this source',
    ERR_DUPLICATE_URL: 'This URL has already been added',
    ERR_ONLY_URL_REFRESH: 'Only URL sources can be refreshed',
    ERR_SOURCE_ALREADY_PROCESSING: 'Source is already being processed',
    ERR_PLAN_REFRESH_UNAVAILABLE: 'Refresh is not available on your plan',
    ERR_MONTHLY_REFRESH_EXCEEDED: 'Monthly refresh limit reached',
    ERR_REFRESH_COOLDOWN_ACTIVE: 'Please wait before refreshing again',
    ERR_NO_URLS_PROVIDED: 'No URLs provided',
    ERR_URL_LIMIT_REACHED: 'URL limit reached for this chatbot',
    ERR_MONTHLY_INGESTION_EXCEEDED: 'Monthly ingestion limit reached',
    ERR_SITEMAP_PARSE_FAILED: 'Failed to parse sitemap',
    ERR_MAX_CHATBOTS_EXCEEDED: 'Maximum chatbot limit reached',
    ERR_TEXT_TOO_LONG: 'Text exceeds maximum length',
    ERR_DUPLICATE_CONTENT: 'This content has already been added to this chatbot',
    ERR_BLOCKED_URL: 'This URL is not allowed for security reasons',

    // Action errors
    ERR_NAME_AND_ACTION_TYPE_REQUIRED: 'Name and action type are required',
    ERR_INVALID_STATUS: 'Invalid status',

    // Handoff errors
    ERR_HANDOFF_EXISTS: 'A support request already exists for this conversation',
    ERR_HANDOFF_NOT_FOUND: 'Support request not found',
    ERR_HANDOFF_EXPIRED: 'This support request has expired',
    ERR_HANDOFF_CLOSED: 'This support request is already closed',
    ERR_HANDOFF_RATE_LIMITED: 'Too many support requests. Please try again later.',
    ERR_HANDOFF_NOT_ENABLED: 'Support requests are not enabled for this chatbot',

    // Privacy/Export errors
    ERR_MISSING_REQUEST_ID: 'Request ID is required',
    ERR_MISSING_EXPORT_ID: 'Export ID is required',
    ERR_MISSING_USER_ID: 'User ID is required',
    ERR_PRIVACY_REQUEST_NOT_FOUND: 'Privacy request not found',
    ERR_NOT_EXPORT_REQUEST: 'Not an export request',
    ERR_EXPORT_NOT_READY: 'Export is not ready yet',
    ERR_EXPORT_URL_MISSING: 'Export URL is missing',
    ERR_EXPORT_EXPIRED: 'Export has expired',
    ERR_STORAGE_NOT_CONFIGURED: 'Storage is not configured',
    ERR_FAILED_TO_GENERATE_URL: 'Failed to generate download URL',
    ERR_FAILED_TO_DOWNLOAD: 'Failed to download file',
    ERR_INVALID_ACTION: 'Invalid action',

    // Admin errors
    ERR_FAILED_TO_FETCH_STATS: 'Failed to fetch statistics',
    ERR_FAILED_TO_LIST_USERS: 'Failed to list users',
    ERR_USER_NOT_FOUND: 'User not found',
    ERR_FAILED_TO_GET_USER: 'Failed to get user details',
    ERR_FAILED_TO_UPDATE_USER: 'Failed to update user',
    ERR_FAILED_TO_LIST_ORGS: 'Failed to list organizations',
    ERR_ORG_NOT_FOUND: 'Organization not found',
    ERR_FAILED_TO_GET_ORG: 'Failed to get organization details',
    ERR_MISSING_JOB_ID: 'Job ID is required',
    ERR_FAILED_TO_FETCH_JOBS: 'Failed to fetch jobs',
    ERR_FAILED_TO_RETRY_JOB: 'Failed to retry job',
    ERR_FAILED_TO_DELETE_JOB: 'Failed to delete job',

    // Workspace/Organization errors
    ERR_NOT_WORKSPACE_MEMBER: 'You are not a member of this workspace',
    ERR_NOT_ORG_MEMBER: 'You are not a member of this organization',
    ERR_WORKSPACE_CHECK_ERROR: 'Failed to verify workspace membership',
    ERR_MEMBERSHIP_CHECK_ERROR: 'Failed to verify membership',
    ERR_GET_PLAN_ERROR: 'Failed to get plan details',
    ERR_CREATE_BOT_ERROR: 'Failed to create chatbot',
  },
  tr: {
    // HTTP standard errors
    BAD_REQUEST: 'Geçersiz istek',
    UNAUTHORIZED: 'Yetkisiz erişim',
    FORBIDDEN: 'Erişim reddedildi',
    NOT_FOUND: 'Kaynak bulunamadı',
    CONFLICT: 'Mevcut kaynak ile çakışma',
    TOO_MANY_REQUESTS: 'Çok fazla istek. Lütfen daha sonra tekrar deneyin.',
    PAYMENT_REQUIRED: 'Yükseltme gerekli',
    INTERNAL_ERROR: 'Bir hata oluştu. Lütfen tekrar deneyin.',
    SERVICE_UNAVAILABLE: 'Servis kullanılamıyor',
    METHOD_NOT_ALLOWED: 'İzin verilmeyen metod',
    REQUEST_ENTITY_TOO_LARGE: 'İstek çok büyük',
    GONE: 'Kaynak artık mevcut değil',

    // Authentication errors
    ERR_EMAIL_REQUIRED: 'E-posta adresi gereklidir',
    ERR_PASSWORD_REQUIRED: 'Şifre gereklidir',
    ERR_EMAIL_AND_PASSWORD_REQUIRED: 'E-posta ve şifre gereklidir',
    ERR_INVALID_EMAIL_FORMAT: 'Geçersiz e-posta formatı',
    ERR_PASSWORD_TOO_SHORT: 'Şifre en az 8 karakter olmalıdır',
    ERR_PASSWORD_WEAK: 'Şifre büyük harf, küçük harf, rakam ve özel karakter (@$!%*?&) içermelidir',
    ERR_EMAIL_EXISTS: 'Bu e-posta adresi zaten kayıtlı',
    ERR_INVALID_CREDENTIALS: 'Geçersiz e-posta veya şifre',
    ERR_INVALID_REQUEST_BODY: 'Geçersiz istek',
    ERR_DATABASE_ERROR: 'Sunucu hatası oluştu. Lütfen tekrar deneyin.',
    ERR_FAILED_TO_HASH_PASSWORD: 'Sunucu hatası oluştu',
    ERR_FAILED_TO_CREATE_USER: 'Hesap oluşturulamadı. Lütfen tekrar deneyin.',
    ERR_INVALID_ID_FORMAT: 'Geçersiz ID formatı',
    ERR_MISSING_ID: 'ID gereklidir',

    // Chatbot/Source errors
    ERR_MONTHLY_TOKENS_EXCEEDED: 'Aylık kullanım limiti aşıldı',
    ERR_PDF_LIMIT_REACHED: 'Bu chatbot için PDF limiti doldu',
    ERR_FILE_TOO_LARGE: 'Dosya boyutu çok büyük',
    ERR_READD_COOLDOWN_ACTIVE: 'Bu kaynağı yeniden eklemeden önce lütfen bekleyin',
    ERR_DUPLICATE_URL: 'Bu URL zaten eklenmiş',
    ERR_ONLY_URL_REFRESH: 'Yalnızca URL kaynakları yenilenebilir',
    ERR_SOURCE_ALREADY_PROCESSING: 'Kaynak zaten işleniyor',
    ERR_PLAN_REFRESH_UNAVAILABLE: 'Planınızda yenileme özelliği mevcut değil',
    ERR_MONTHLY_REFRESH_EXCEEDED: 'Aylık yenileme limiti aşıldı',
    ERR_REFRESH_COOLDOWN_ACTIVE: 'Yeniden yenilemeden önce lütfen bekleyin',
    ERR_NO_URLS_PROVIDED: 'URL sağlanmadı',
    ERR_URL_LIMIT_REACHED: 'Bu chatbot için URL limiti doldu',
    ERR_MONTHLY_INGESTION_EXCEEDED: 'Aylık içe aktarma limiti aşıldı',
    ERR_SITEMAP_PARSE_FAILED: 'Site haritası ayrıştırılamadı',
    ERR_MAX_CHATBOTS_EXCEEDED: 'Maksimum chatbot limitine ulaşıldı',
    ERR_TEXT_TOO_LONG: 'Metin çok uzun',
    ERR_DUPLICATE_CONTENT: 'Bu içerik zaten bu chatbota eklenmiş',
    ERR_BLOCKED_URL: 'Bu URL güvenlik nedeniyle engellenmiştir',

    // Action errors
    ERR_NAME_AND_ACTION_TYPE_REQUIRED: 'İsim ve aksiyon türü gereklidir',
    ERR_INVALID_STATUS: 'Geçersiz durum',

    // Handoff errors
    ERR_HANDOFF_EXISTS: 'Bu konuşma için zaten bir destek talebi var',
    ERR_HANDOFF_NOT_FOUND: 'Destek talebi bulunamadı',
    ERR_HANDOFF_EXPIRED: 'Bu destek talebinin süresi doldu',
    ERR_HANDOFF_CLOSED: 'Bu destek talebi zaten kapatıldı',
    ERR_HANDOFF_RATE_LIMITED: 'Çok fazla destek talebi. Lütfen daha sonra tekrar deneyin.',
    ERR_HANDOFF_NOT_ENABLED: 'Bu chatbot için destek talepleri etkin değil',

    // Privacy/Export errors
    ERR_MISSING_REQUEST_ID: 'İstek ID\'si gereklidir',
    ERR_MISSING_EXPORT_ID: 'Dışa aktarım ID\'si gereklidir',
    ERR_MISSING_USER_ID: 'Kullanıcı ID\'si gereklidir',
    ERR_PRIVACY_REQUEST_NOT_FOUND: 'Gizlilik talebi bulunamadı',
    ERR_NOT_EXPORT_REQUEST: 'Dışa aktarım talebi değil',
    ERR_EXPORT_NOT_READY: 'Dışa aktarım henüz hazır değil',
    ERR_EXPORT_URL_MISSING: 'Dışa aktarım URL\'si eksik',
    ERR_EXPORT_EXPIRED: 'Dışa aktarımın süresi doldu',
    ERR_STORAGE_NOT_CONFIGURED: 'Depolama yapılandırılmamış',
    ERR_FAILED_TO_GENERATE_URL: 'İndirme URL\'si oluşturulamadı',
    ERR_FAILED_TO_DOWNLOAD: 'Dosya indirilemedi',
    ERR_INVALID_ACTION: 'Geçersiz işlem',

    // Admin errors
    ERR_FAILED_TO_FETCH_STATS: 'İstatistikler alınamadı',
    ERR_FAILED_TO_LIST_USERS: 'Kullanıcılar listelenemedi',
    ERR_USER_NOT_FOUND: 'Kullanıcı bulunamadı',
    ERR_FAILED_TO_GET_USER: 'Kullanıcı detayları alınamadı',
    ERR_FAILED_TO_UPDATE_USER: 'Kullanıcı güncellenemedi',
    ERR_FAILED_TO_LIST_ORGS: 'Organizasyonlar listelenemedi',
    ERR_ORG_NOT_FOUND: 'Organizasyon bulunamadı',
    ERR_FAILED_TO_GET_ORG: 'Organizasyon detayları alınamadı',
    ERR_MISSING_JOB_ID: 'İş ID\'si gereklidir',
    ERR_FAILED_TO_FETCH_JOBS: 'İşler alınamadı',
    ERR_FAILED_TO_RETRY_JOB: 'İş yeniden denenemedi',
    ERR_FAILED_TO_DELETE_JOB: 'İş silinemedi',

    // Workspace/Organization errors
    ERR_NOT_WORKSPACE_MEMBER: 'Bu çalışma alanının üyesi değilsiniz',
    ERR_NOT_ORG_MEMBER: 'Bu organizasyonun üyesi değilsiniz',
    ERR_WORKSPACE_CHECK_ERROR: 'Çalışma alanı üyeliği doğrulanamadı',
    ERR_MEMBERSHIP_CHECK_ERROR: 'Üyelik doğrulanamadı',
    ERR_GET_PLAN_ERROR: 'Plan detayları alınamadı',
    ERR_CREATE_BOT_ERROR: 'Chatbot oluşturulamadı',
  },
} as const;

/**
 * Get localized error message for an error code.
 * Falls back to English if translation not found, then to the code itself.
 */
export function getErrorMessage(code: string, lang: string = 'en'): string {
  const messages = lang === 'tr' ? errorMessages.tr : errorMessages.en;
  return (messages as Record<string, string>)[code] ?? 
         (errorMessages.en as Record<string, string>)[code] ?? 
         code;
}

/**
 * Type guard to check if a string is a known error code.
 */
export function isKnownErrorCode(code: string): code is ErrorCode {
  return code in errorMessages.en;
}
