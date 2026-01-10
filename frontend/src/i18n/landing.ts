/**
 * Turkish translations for the landing page.
 * All landing page content is centralized here for easy updates.
 */

export const landing = {
  // Navbar
  nav: {
    features: 'Özellikler',
    useCases: 'Kullanım Alanları',
    howItWorks: 'Nasıl Çalışır?',
    pricing: 'Fiyatlandırma',
    faq: 'SSS',
    login: 'Giriş Yap',
    register: 'Ücretsiz Başla',
    dashboard: 'Dashboard',
    goToDashboard: "Dashboard'a Git",
  },

  // Hero Section
  hero: {
    badge: 'GPT-4o & GPT-4o Mini Destekli',
    badgeLive: 'Canlı',
    title: {
      line1: 'Verilerinizi',
      highlight: 'Akıllı Asistana',
      line2: 'Dönüştürün',
    },
    subtitle:
      'Web sitenizi, PDF dokümanlarınızı ve metinlerinizi saniyeler içinde eğitilen yapay zeka destekli bir müşteri asistanına dönüştürün. 7/24 kesintisiz destek, güvenli altyapı.',
    cta: {
      primary: 'Ücretsiz Deneyin',
      primaryAuth: "Dashboard'a Git",
      secondary: 'Nasıl Çalışır?',
    },
    stats: {
      sources: 'Kaynak Türü',
      sourcesValue: '4+',
      security: 'Kurumsal Güvenlik',
      securityValue: 'KVKK',
      languages: 'Dil Desteği',
      languagesValue: '40+',
    },
  },

  // Features Section
  features: {
    badge: 'Yetenekler',
    title: 'Yapay Zeka Destekli',
    titleHighlight: 'Akıllı Altyapı',
    subtitle:
      'RAG (Retrieval-Augmented Generation) teknolojisi ile verilerinizden öğrenen, bağlam odaklı yanıtlar üreten güçlü bir altyapı.',
    items: {
      rag: {
        title: 'RAG Teknolojisi',
        description:
          'Semantik arama ile verilerinizi anlayan akıllı soru-cevap sistemi. Yüksek, orta ve düşük güven seviyelerinde kademeli yanıtlar.',
      },
      sources: {
        title: 'Çoklu Kaynak Desteği',
        description:
          'Web siteleri, PDF dokümanları, metin içerikleri ve XML site haritalarından otomatik veri çekimi ve eğitim.',
      },
      widget: {
        title: 'Özelleştirilebilir Widget',
        description:
          'Renk, font, konum ve marka öğelerini tamamen kontrol edin. Shadow DOM ile stil izolasyonu, mobil uyumlu tasarım.',
      },
      actions: {
        title: 'Akıllı Aksiyonlar',
        description:
          'Webhook ve API çağrıları ile chatbotunuzu güçlendirin. Form gönderimi, bildirim ve basit entegrasyonlar.',
      },
      guardrails: {
        title: 'Güvenlik Önlemleri',
        description:
          'Konu kısıtlamaları, güven eşikleri ve özel yedek mesajlar ile botunuzun hedeften sapmasını engelleyin.',
      },
      analytics: {
        title: 'Detaylı Analitik',
        description:
          'Mesaj sayıları, geri bildirimler, token kullanımı ve konuşma istatistikleri ile performansı takip edin.',
      },
      handoff: {
        title: 'İnsan Desteği',
        description:
          'Karmaşık sorularda otomatik olarak insan temsilciye yönlendirme. E-posta yakalama ve destek talep takibi.',
      },
      multiTenant: {
        title: 'Çoklu Organizasyon',
        description:
          'Organizasyonlar ve çalışma alanları ile ekip bazlı yönetim. Rol tabanlı erişim kontrolü.',
      },
    },
  },

  // Use Cases Section
  useCases: {
    badge: 'Kullanım Alanları',
    title: 'Verilerinizden',
    titleHighlight: 'Öğrenen Asistan',
    subtitle:
      'Web siteniz, PDF\'leriniz veya metin içerikleriniz ile eğitilen bir asistan oluşturun.',
    items: {
      website: {
        title: 'Web Sitesi SSS',
        description:
          'Web sitenizin içeriğini tarayarak ziyaretçilerinizin sorularını otomatik yanıtlayın.',
        features: ['URL tarama', 'Otomatik güncelleme', '7/24 yanıt'],
      },
      docs: {
        title: 'Doküman Asistanı',
        description:
          'PDF ve metin dosyalarınızı yükleyin, içeriklerinden sorulara anında yanıt alın.',
        features: ['PDF yükleme', 'Metin içeriği', 'Anlık yanıt'],
      },
      support: {
        title: 'Destek Botu',
        description:
          'Sık sorulan soruları otomatik yanıtlayın, gerektiğinde e-posta ile yönlendirin.',
        features: ['SSS yanıtları', 'E-posta toplama', 'Geri bildirim'],
      },
      brand: {
        title: 'Markanıza Özel',
        description:
          'Widget renklerini, fontları ve mesajları tamamen özelleştirin, sitenize uyumlu hale getirin.',
        features: ['Renk teması', 'Özel mesajlar', 'Logo ekleme'],
      },
    },
  },

  // How It Works Section
  howItWorks: {
    badge: 'Nasıl Çalışır?',
    title: 'Dakikalar İçinde',
    titleHighlight: 'Yayına Alın',
    subtitle:
      'Karmaşık kurulum süreçlerine son. 3 basit adımda yapay zeka asistanınız hazır.',
    steps: {
      step1: {
        number: '01',
        title: 'Kaynaklarınızı Ekleyin',
        description:
          'PDF yükleyin, web sitenizi taratın veya metin ekleyin. Sistem otomatik olarak içeriği analiz eder ve yapılandırır.',
        bullets: [
          'PDF dosya yükleme',
          'Web sitesi URL tarama',
          'Sitemap entegrasyonu',
          'Metin kopyala-yapıştır',
        ],
        tip: 'URL eklediğinizde sistem sayfaları otomatik keşfedebilir. Include/Exclude filtreleri ile taramayı kontrol edebilirsiniz.',
      },
      step2: {
        number: '02',
        title: 'Kuralları Belirleyin',
        description:
          'Guardrails ile konuşma sınırlarını ayarlayın, güven eşiklerini tanımlayın ve yedek mesajları özelleştirin.',
        bullets: [
          'Konu kısıtlamaları',
          'Güven eşikleri (Yüksek/Orta/Düşük)',
          'Özel yedek mesajlar',
          'İnsan desteğine aktarma',
        ],
        tip: 'Guardrails özelliği ile botun rakip markalardan bahsetmesini veya kapsam dışı konulara girmesini engelleyebilirsiniz.',
      },
      step3: {
        number: '03',
        title: 'Widget ile Yayınlayın',
        description:
          'Tek satır kod ile sitenize ekleyin. Renkleri, fontları ve konumu tamamen özelleştirin.',
        bullets: [
          'Tek satır embed kodu',
          'İzinli domain listesi',
          'Güvenli embed token',
          'Tam görünüm kontrolü',
        ],
        tip: 'Güvenlik için "Allowed Domains" listesini doldurun; böylece botunuz sadece yetkilendirdiğiniz sitelerde çalışır.',
      },
    },
  },

  // Pricing Section
  pricing: {
    badge: 'Fiyatlandırma',
    title: 'Şeffaf ve',
    titleHighlight: 'Esnek Paketler',
    subtitle:
      'Startup, KOBİ veya kurumsal şirket olun. İhtiyacınıza uygun planı seçin.',
    mostPopular: 'En Popüler',
    perMonth: '/ay',
    features: {
      chatbots: 'Chatbot',
      tokens: 'Aylık Token',
      sites: 'Web Sitesi',
      pdfs: 'PDF Dosyası',
      storage: 'Depolama',
      model: 'AI Model',
      dynamicScraping: 'Dinamik Tarama',
      guardrails: 'Guardrails',
      smartFallback: 'Akıllı Yedek',
      handoff: 'İnsan Desteği',
      branding: 'Marka Kaldırma',
      support: 'Destek',
    },
    cta: {
      free: 'Ücretsiz Başla',
      pro: 'Yakında',
      ultra: 'Yakında',
      authenticated: "Dashboard'a Git",
    },
    plans: {
      free: {
        name: 'Başlangıç',
        description: 'Küçük projeler için ideal başlangıç.',
      },
      pro: {
        name: 'Pro',
        description: 'Büyüyen işletmeler için güçlü özellikler.',
      },
      ultra: {
        name: 'Ultra',
        description: 'Kurumsal projeler için sınırsız güç.',
      },
    },
  },

  // FAQ Section
  faq: {
    badge: 'SSS',
    title: 'Sıkça Sorulan',
    titleHighlight: 'Sorular',
    subtitle: 'Merak ettiğiniz her şey burada. Bulamadığınız soruları bize iletebilirsiniz.',
    items: [
      {
        question: 'Botu siteme nasıl eklerim?',
        answer:
          "Size verdiğimiz tek satırlık JavaScript kodunu sitenizin <head> veya <body> etiketleri arasına yapıştırmanız yeterlidir. WordPress, Shopify, Wix, Squarespace ve tüm altyapılarla uyumludur.",
      },
      {
        question: 'Hangi dosya formatlarını destekliyorsunuz?',
        answer:
          "PDF dosyaları, metin içerikleri (TXT / kopyala-yapıştır), web sitesi URL'leri ve XML site haritalarını destekliyoruz. Pro planlarda OCR ile görüntü tabanlı PDF'ler de işlenebilir.",
      },
      {
        question: 'Ücretsiz planda kredi kartı gerekiyor mu?',
        answer:
          'Hayır, Başlangıç (Free) planımızı kullanmak için kredi kartı gerekmez. İstediğiniz kadar ücretsiz kullanabilirsiniz.',
      },
      {
        question: 'Botum yanlış cevap verirse ne olur?',
        answer:
          "Guardrails özelliği sayesinde botun cevap veremediği veya emin olamadığı durumlarda 'Bunu bilmiyorum' demesini veya insan temsilciye yönlendirmesini sağlayabilirsiniz. Güven eşikleri ile kontrol sağlayabilirsiniz.",
      },
      {
        question: 'Verilerim güvende mi?',
        answer:
          'Evet, tüm verileriniz şifrelenerek saklanır. KVKK ve GDPR uyumlu altyapımızda verileriniz sadece chatbotunuzun eğitimi için kullanılır. Row-level security ile tam izolasyon sağlanır.',
      },
      {
        question: 'Chatbot hangi dilleri destekliyor?',
        answer:
          'Chatbot, kullanıcıların sorularını 40+ dilde anlayabilir ve yanıtlayabilir. Türkçe, İngilizce, Almanca, Fransızca ve daha birçok dil desteklenmektedir.',
      },
      {
        question: 'API entegrasyonu yapabilir miyim?',
        answer:
          'Evet, Akıllı Aksiyonlar (Smart Actions) özelliği ile dış API\'lere bağlanabilirsiniz. Sipariş sorgulama, stok kontrolü, randevu alma gibi işlemleri otomatikleştirebilirsiniz.',
      },
      {
        question: 'İçerikler otomatik güncellenir mi?',
        answer:
          'Evet, günlük, haftalık veya aylık otomatik kaynak yenileme planları ile içerikleriniz güncel kalır. ETag desteği ile sadece değişen içerikler güncellenir.',
      },
    ],
  },

  // Security Section
  security: {
    badge: 'Güvenlik',
    title: 'Kurumsal Düzeyde',
    titleHighlight: 'Güvenlik',
    subtitle:
      'KVKK ve GDPR uyumlu altyapı ile verileriniz her zaman güvende.',
    items: {
      jwt: {
        title: 'JWT Kimlik Doğrulama',
        description:
          'Access ve refresh token ile güvenli oturum yönetimi. Token rotasyonu ve otomatik yenileme.',
      },
      ssrf: {
        title: 'SSRF Koruması',
        description:
          'Özel IP aralıkları ve localhost erişimi engellenir. Güvenli URL doğrulama.',
      },
      encryption: {
        title: 'Şifreleme',
        description:
          'Argon2 ile şifre hashleme, SHA-256 ile token depolama. Uçtan uca güvenlik.',
      },
      rls: {
        title: 'Row-Level Security',
        description:
          'Veritabanı seviyesinde izolasyon. Her müşteri kendi verilerine erişir.',
      },
      rateLimit: {
        title: 'Rate Limiting',
        description:
          'DDoS koruması için gelişmiş hız sınırlama. Redis tabanlı sliding window.',
      },
      audit: {
        title: 'Denetim Günlüğü',
        description:
          'Tüm admin işlemleri kayıt altında. KVKK uyumlu veri saklama politikaları.',
      },
    },
  },

  // CTA Section
  cta: {
    title: 'Yapay Zeka Asistanınızı',
    titleHighlight: 'Bugün Oluşturun',
    subtitle:
      'Dakikalar içinde kurulum, günlerce süren geliştirme süreçlerine son. Ücretsiz başlayın, ihtiyacınıza göre büyütün.',
    button: 'Ücretsiz Başlayın',
    buttonAuth: "Dashboard'a Git",
    note: 'Kredi kartı gerekmez',
  },

  // Footer
  footer: {
    description:
      'Yeni nesil web siteleri için geliştirilmiş, verilerinizle eğitilen akıllı müşteri asistanı. Müşterilerinize 7/24 kesintisiz destek sunun.',
    product: {
      title: 'Ürün',
      features: 'Özellikler',
      pricing: 'Fiyatlandırma',
      dashboard: 'Dashboard',
      login: 'Giriş Yap',
      register: 'Kayıt Ol',
    },
    company: {
      title: 'Şirket',
      about: 'Hakkımızda',
      blog: 'Blog',
      contact: 'İletişim',
    },
    legal: {
      title: 'Yasal',
      privacy: 'Gizlilik',
      terms: 'Koşullar',
      kvkk: 'KVKK',
    },
    copyright: '© {year} botla.app. Tüm hakları saklıdır.',
    madeWith: "İstanbul'da ❤️ ile yapıldı",
  },
} as const;

export type LandingTranslations = typeof landing;
