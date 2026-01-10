/**
 * Turkish translations for the Privacy Settings page.
 * All privacy-related content is centralized here for easy updates.
 */

export const privacy = {
  // Page Header
  page: {
    title: 'Gizlilik Ayarları',
    description: 'Veri işleme izinlerinizi, veri dışa aktarma ve hesap silme işlemlerinizi buradan yönetebilirsiniz.',
  },

  // Request Status Labels
  status: {
    pending: 'İstek alındı',
    processing: 'İşleniyor',
    completed: 'Tamamlandı',
    denied: 'Reddedildi',
    loading: 'Bekliyor...',
  },

  // Consents Section
  consents: {
    title: 'İzinler ve Tercihler',
    description: 'Hangi verilerinizin nasıl işleneceğine karar verin.',
    marketing: {
      label: 'Pazarlama İletişimi',
      description: 'Kampanya ve duyurulardan haberdar olmak istiyorum.',
    },
    analytics: {
      label: 'Analitik Veriler',
      description: 'Hizmet kalitesini artırmak için anonim kullanım verilerimi paylaş.',
    },
    personalization: {
      label: 'Kişiselleştirme',
      description: 'Bana özel içerik ve öneriler sunulmasını kabul ediyorum.',
    },
    thirdParty: {
      label: 'Üçüncü Taraf Paylaşımı',
      description: 'Verilerimin iş ortakları ve üçüncü taraflarla paylaşılmasını kabul ediyorum.',
    },
  },

  // Data Export Section
  export: {
    title: 'Veri Dışa Aktarma',
    description: 'Tüm kişisel verilerinizin bir kopyasını indirin. İşlem tamamlandığında buradan indirebilirsiniz.',
    button: 'Verilerimi İndir',
    preparing: 'Hazırlanıyor...',
    statusLabel: 'Durum',
    downloadButton: 'İndirmeyi Başlat',
    denied: 'Talebiniz reddedildi.',
    history: {
      title: 'Dışa Aktarma Geçmişi',
      empty: 'Henüz dışa aktarma talebiniz yok.',
      type: 'Talep Türü',
      status: 'Durum',
      date: 'Tarih',
      actions: 'İşlemler',
      delete: 'Sil',
      deleteConfirm: 'Bu dışa aktarma talebini silmek istediğinizden emin misiniz?',
      deleteSuccess: 'Dışa aktarma talebi silindi.',
      deleteError: 'Silme işlemi başarısız oldu.',
    },
    rateLimit: {
      title: 'Dışa Aktarma Sınırlaması',
      message: '24 saatte bir dışa aktarma talebinde bulunabilirsiniz.',
      nextAvailable: 'Sonraki dışa aktarma: {time}',
      lastRequest: 'Son dışa aktarma: {time}',
    },
    activeRequest: {
      title: 'Bekleyen Talep',
      message: 'Zaten bekleyen veya işlenmekte olan bir dışa aktarma talebiniz var.',
      waitMessage: 'Mevcut talep tamamlandıktan sonra yeni talep oluşturabilirsiniz.',
    },
  },

  // Data Correction Section
  correction: {
    title: 'Veri Düzeltme',
    description: 'Kişisel verilerinizde bir yanlışlık olduğunu düşünüyorsanız düzeltme talebinde bulunun.',
    label: 'Düzeltme Detayları',
    placeholder: 'Hangi verinin nasıl düzeltilmesini istediğinizi açıklayın...',
    charCount: '{count} / {max}',
    maxLength: 1000,
    button: 'Düzeltme Talebi Gönder',
    sending: 'Gönderiliyor...',
    history: {
      title: 'Düzeltme Geçmişi',
      empty: 'Henüz düzeltme talebiniz yok.',
      reason: 'Düzeltme Talebi',
      status: 'Durum',
      date: 'Tarih',
      actions: 'İşlemler',
      viewDetails: 'Detay',
      detailsTitle: 'Düzeltme Talebi Detayı',
      delete: 'Sil',
      deleteConfirm: 'Bu düzeltme talebini silmek istediğinizden emin misiniz?',
      deleteSuccess: 'Düzeltme talebi silindi.',
      deleteError: 'Silme işlemi başarısız oldu.',
    },
    rateLimit: {
      title: 'Düzeltme Sınırlaması',
      message: '24 saatte bir düzeltme talebinde bulunabilirsiniz.',
      nextAvailable: 'Sonraki düzeltme: {time}',
      lastRequest: 'Son düzeltme talebi: {time}',
    },
    activeRequest: {
      title: 'Bekleyen Talep',
      message: 'Zaten bekleyen veya işlenmekte olan bir düzeltme talebiniz var.',
      waitMessage: 'Mevcut talep tamamlandıktan sonra yeni talep oluşturabilirsiniz.',
    },
  },

  // Delete Account Section
  delete: {
    title: 'Hesabı Sil',
    description: 'Hesabınızı ve tüm verilerinizi kalıcı olarak silin. Bu işlem geri alınamaz.',
    button: 'Hesabımı Sil',
    dialog: {
      title: 'Emin misiniz?',
      description: 'Bu işlem geri alınamaz. Hesabınız ve tüm verileriniz sunucularımızdan kalıcı olarak silinecektir.',
      reasonLabel: 'Silme Nedeni (İsteğe bağlı)',
      reasonPlaceholder: 'Bize neden ayrıldığınızı söylemek ister misiniz?',
      cancel: 'İptal',
      confirmButton: 'Evet, Hesabımı Sil',
      deleting: 'Siliniyor...',
    },
  },

  // Toast Messages
  toast: {
    consentsUpdated: 'Gizlilik tercihleriniz güncellendi.',
    consentsError: 'Tercihler güncellenirken bir hata oluştu.',
    exportRequested: 'Veri dışa aktarma talebiniz alındı.',
    exportError: 'Dışa aktarma talebi oluşturulurken bir hata oluştu.',
    exportActiveExists: 'Zaten bekleyen bir dışa aktarma talebiniz var.',
    exportRateLimit: '24 saatte bir dışa aktarma talebinde bulunabilirsiniz.',
    downloadError: 'Dışa aktarma indirme sırasında bir hata oluştu.',
    correctionRequested: 'Veri düzeltme talebiniz alındı.',
    correctionError: 'Düzeltme talebi oluşturulurken bir hata oluştu.',
    deleteRequested: 'Hesap silme talebiniz alındı.',
    deleteError: 'Hesap silme talebi oluşturulurken bir hata oluştu.',
  },
} as const;

export type PrivacyTranslations = typeof privacy;

/**
 * Get localized status label for privacy request status.
 */
export function getStatusLabel(status?: string): string {
  if (!status) return privacy.status.loading;
  return (privacy.status as Record<string, string>)[status] ?? status;
}
