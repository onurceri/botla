export interface ErrorMessage {
  title: string
  description: string
  suggestion?: string
  icon?: 'warning' | 'error' | 'info'
}

export const ERROR_MESSAGES: Record<string, ErrorMessage> = {
  // URL-specific errors
  ERR_EMPTY_URL: {
    title: 'Geçersiz URL',
    description: 'URL adresi boş veya geçersiz.',
    suggestion: 'Lütfen geçerli bir web adresi girin.',
    icon: 'error',
  },
  ERR_EMPTY_CONTENT: {
    title: 'İçerik Bulunamadı',
    description: 'Web sitesinden içerik çıkarılamadı.',
    suggestion:
      'Bu site JavaScript ile içerik yüklüyor olabilir. Lütfen farklı bir URL deneyin veya destek ekibiyle iletişime geçin.',
    icon: 'warning',
  },
  ERR_SCRAPE_NETWORK: {
    title: 'Bağlantı Hatası',
    description: 'Web sitesine bağlanılamadı.',
    suggestion: "URL'nin doğru olduğundan ve sitenin erişilebilir olduğundan emin olun.",
    icon: 'error',
  },
  ERR_SCRAPE_TIMEOUT: {
    title: 'Zaman Aşımı',
    description: 'Web sitesi yanıt vermedi.',
    suggestion: 'Lütfen daha sonra tekrar deneyin veya farklı bir URL kullanın.',
    icon: 'warning',
  },
  ERR_SCRAPE_FORBIDDEN: {
    title: 'Erişim Engellendi',
    description: 'Web sitesi erişimi engelledi (403 Forbidden).',
    suggestion:
      'Bu site botları engelliyor olabilir. Lütfen site yöneticisiyle iletişime geçin veya farklı bir kaynak kullanın.',
    icon: 'error',
  },
  ERR_INVALID_URL: {
    title: 'Geçersiz URL Formatı',
    description: 'Girilen URL formatı geçersiz.',
    suggestion: 'Lütfen geçerli bir web adresi girin (örn: https://example.com)',
    icon: 'error',
  },

  // PDF-specific errors
  ERR_EMPTY_FILE_PATH: {
    title: 'Dosya Bulunamadı',
    description: 'Dosya yolu boş veya geçersiz.',
    suggestion: 'Lütfen dosyayı tekrar yükleyin.',
    icon: 'error',
  },
  ERR_PDF_DOWNLOAD_FAILED: {
    title: 'İndirme Hatası',
    description: 'PDF dosyası indirilemedi.',
    suggestion: 'Lütfen dosyayı tekrar yükleyin veya farklı bir dosya deneyin.',
    icon: 'error',
  },
  ERR_PDF_PARSE_FAILED: {
    title: 'PDF Ayrıştırma Hatası',
    description: 'PDF dosyası okunamadı.',
    suggestion:
      'Dosya bozuk veya şifreli olabilir. Lütfen geçerli bir PDF dosyası yükleyin.',
    icon: 'error',
  },

  // Text-specific errors
  ERR_STORAGE_REQUIRED: {
    title: 'Depolama Hatası',
    description: 'Dosya depolama servisi kullanılamıyor.',
    suggestion: 'Lütfen daha sonra tekrar deneyin veya destek ekibiyle iletişime geçin.',
    icon: 'error',
  },

  // Common processing errors
  ERR_CHUNKING_FAILED: {
    title: 'İşleme Hatası',
    description: 'İçerik parçalara ayrılamadı.',
    suggestion: 'Lütfen farklı bir kaynak deneyin veya destek ekibiyle iletişime geçin.',
    icon: 'error',
  },
  ERR_EMBEDDING_FAILED: {
    title: 'Vektörleme Hatası',
    description: 'İçerik vektörlenemedi.',
    suggestion: 'Lütfen daha sonra tekrar deneyin veya destek ekibiyle iletişime geçin.',
    icon: 'error',
  },
  ERR_LLM_NOT_SUPPORTED: {
    title: 'Sistem Hatası',
    description: 'AI modeli desteklenmiyor.',
    suggestion: 'Lütfen destek ekibiyle iletişime geçin.',
    icon: 'error',
  },
}

/**
 * Get user-friendly error message for a given error code
 * @param errorCode - The error code from the backend
 * @returns ErrorMessage object with title, description, and suggestion
 */
export function getErrorMessage(errorCode: string): ErrorMessage {
  const message = ERROR_MESSAGES[errorCode]
  if (message) {
    return message
  }

  // Default error message for unknown codes
  return {
    title: 'Bilinmeyen Hata',
    description: errorCode || 'Bir hata oluştu.',
    suggestion: 'Lütfen daha sonra tekrar deneyin veya destek ekibiyle iletişime geçin.',
    icon: 'error',
  }
}

/**
 * Check if an error code is a known error
 * @param errorCode - The error code to check
 * @returns true if the error code is known
 */
export function isKnownError(errorCode: string): boolean {
  return errorCode in ERROR_MESSAGES
}
