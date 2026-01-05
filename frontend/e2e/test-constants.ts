/**
 * Test Constants for Turkish UI Text
 * These constants ensure test stability when UI text changes.
 * Use these constants in E2E tests instead of hardcoded Turkish strings.
 */

export const TURKISH = {
  // Auth
  WELCOME: 'Hoş Geldiniz',
  LOGIN: 'Giriş Yap',
  LOGIN_LINK: 'Giriş Yapın',
  LOGOUT: 'Çıkış Yap',
  REGISTER: 'Kayıt Ol',
  REGISTER_LINK: 'Kayıt Olun',
  FORGOT_PASSWORD: 'Şifremi unuttum?',
  PASSWORD: 'Şifre',
  EMAIL: 'Email',
  NAME: 'Ad Soyad',
  PASSWORD_PLACEHOLDER: '••••••••',

  // Validation
  FILL_ALL_FIELDS: 'Lütfen tüm alanları doldurun.',

  // Chatbot
  NEW_CHATBOT: 'Yeni Oluştur',
  CREATE_CHATBOT: 'Oluştur',
  CHATBOT_CREATED: 'Chatbot başarıyla oluşturuldu.',

  // Sources
  DATA_SOURCES_TAB: 'Veri Kaynakları',
  TEXT_SOURCE: 'Metin Gir',
  TEXT_SOURCE_ADDED: 'Metin kaynağı eklendi.',
  URL_SOURCE: 'URL Ekle',
  ADD: 'Ekle',
  ADD_URL: 'URL Ekle',

  // Playground
  PLAYGROUND_TAB: 'Test Alanı',
  OPEN_CHAT: 'Sohbeti aç',
  TYPE_MESSAGE: 'Mesaj yazın...',

  // Navigation
  SETTINGS: 'Ayarlar',
  REPORTS: 'Raporlar',

  // Success/Error messages
  SUCCESS: 'Başarılı',
  ERROR: 'Hata oluştu',
  LOGIN_FAILED: 'Giriş başarısız. Lütfen bilgilerinizi kontrol edin.',

  // Chat
  HELLO: 'Merhaba!',

  // Recent bots
  RECENT_BOTS: 'Son Botlarınız',

  // Session Expired
  SESSION_EXPIRED: 'Oturumunuz süresi doldu',
  SESSION_EXPIRED_MESSAGE: 'Tekrar giriş yapmanız gerekiyor.',
  RELOGIN: 'Tekrar Giriş Yap',
  CANCEL: 'Vazgeç',
}

export const ENGLISH = {
  // Auth
  WELCOME: 'Welcome',
  LOGIN: 'Login',
  LOGOUT: 'Logout',
  REGISTER: 'Register',
  FORGOT_PASSWORD: 'Forgot password?',

  // Session Expired
  SESSION_EXPIRED: 'Your session has expired',
  SESSION_EXPIRED_MESSAGE: 'Please login again to continue.',
  RELOGIN: 'Login Again',
  CANCEL: 'Cancel',
}

/**
 * Test IDs for stable element selection
 * These data-testid attributes should be added to components.
 */
export const TEST_IDS = {
  // Login Page
  LOGIN_PAGE: 'login-page',
  LOGIN_EMAIL_INPUT: 'login-page-email-input',
  LOGIN_PASSWORD_INPUT: 'login-page-password-input',
  LOGIN_SUBMIT_BUTTON: 'login-page-submit-button',
  LOGIN_FORGOT_PASSWORD_LINK: 'login-page-forgot-password-link',
  LOGIN_ERROR_MESSAGE: 'login-page-error-message',
  LOGIN_TITLE: 'login-page-title',
  LOGIN_REMEMBER_ME_CHECKBOX: 'login-page-remember-me-checkbox',

  // Register Page
  REGISTER_PAGE: 'register-page',
  REGISTER_NAME_INPUT: 'register-page-name-input',
  REGISTER_EMAIL_INPUT: 'register-page-email-input',
  REGISTER_PASSWORD_INPUT: 'register-page-password-input',
  REGISTER_SUBMIT_BUTTON: 'register-page-submit-button',
  REGISTER_ERROR_MESSAGE: 'register-page-error-message',
  REGISTER_TITLE: 'register-page-title',

  // Chatbots Page
  CHATBOTS_PAGE: 'chatbots-page',
  CHATBOTS_CREATE_BUTTON: 'chatbots-page-create-button',
  CHATBOTS_LIST: 'chatbots-page-list',
  CHATBOT_CARD: 'chatbot-card',
  CHATBOT_MANAGE_BUTTON: 'chatbot-manage-button',

  // Chatbot Detail Page
  CHATBOT_DETAIL_PAGE: 'chatbot-detail-page',
  CHATBOT_SOURCES_TAB: 'chatbot-sources-tab',
  CHATBOT_PLAYGROUND_TAB: 'chatbot-playground-tab',
  CHATBOT_SETTINGS_TAB: 'chatbot-settings-tab',

  // Sources
  SOURCE_UPLOADER: 'source-uploader',
  SOURCE_TEXT_OPTION: 'source-text-option',
  SOURCE_URL_OPTION: 'source-url-option',
  SOURCE_ADD_BUTTON: 'source-add-button',

  // Playground
  PLAYGROUND_CONTAINER: 'playground-container',
  PLAYGROUND_CHAT_OPEN_BUTTON: 'playground-chat-open-button',
  PLAYGROUND_MESSAGE_INPUT: 'playground-message-input',
  PLAYGROUND_CHAT_WINDOW: 'playground-chat-window',

  // Common
  LOADING_SPINNER: 'loading-spinner',
  ERROR_MESSAGE: 'error-message',
  SUCCESS_MESSAGE: 'success-message',
  USER_MENU: 'user-menu',

  // User Menu & Dropdown
  USER_AVATAR: 'user-avatar',
  USER_MENU_AVATAR: 'user-menu-avatar',
  USER_MENU_DROPDOWN: 'user-menu-dropdown',
  DROPDOWN_USER_MENU: 'dropdown-user-menu',

  // Menu Items
  MENU_ITEM_PROFILE: 'menu-item-profile',
  MENU_ITEM_SETTINGS: 'menu-item-settings',
  MENU_ITEM_HELP: 'menu-item-help',
  MENU_ITEM_LOGOUT: 'menu-item-logout',

  // Session Expired Modal
  MODAL_SESSION_EXPIRED: 'modal-session-expired',
  SESSION_EXPIRED_TITLE: 'session-expired-title',
  SESSION_EXPIRED_MESSAGE: 'session-expired-message',
  BTN_RELOGIN: 'btn-relogin',
  BTN_SESSION_CANCEL: 'btn-session-cancel',

  // Token Refresh Indicators
  TOKEN_REFRESHING: 'token-refreshing',
  TOKEN_REFRESH_SUCCESS: 'token-refresh-success',
  TOKEN_REFRESH_ERROR: 'token-refresh-error',

  // Session Status
  SESSION_STATUS: 'session-status',
  SESSION_EXPIRY_TIME: 'session-expiry-time',
  SESSION_INFO_TOOLTIP: 'session-info-tooltip',

  // Remember Me
  REMEMBER_ME_CHECKBOX: 'remember-me-checkbox',
  REMEMBER_ME_LABEL: 'remember-me-label',

  // Dashboard Page
  PAGE_DASHBOARD: 'page-dashboard',
}

/**
 * Page URL patterns for navigation
 */
export const PAGE_URLS = {
  LOGIN: '/login',
  REGISTER: '/register',
  DASHBOARD: '/dashboard',
  CHATBOTS: '/dashboard/chatbots',
  CHATBOT_DETAIL_REGEX: /\/dashboard\/chatbots\/[a-zA-Z0-9_-]+$/,
  LOGIN_REGEX: /\/login(\/)?$/,
}

/**
 * Data-TestID Usage Guide for Base Components
 *
 * For base UI components (Button, Input, Select, etc.), pass data-testid via props.
 * The components support arbitrary HTML attributes through ...props.
 *
 * @example
 * // In React component
 * <Button data-testid="submit-button">Submit</Button>
 * <Input data-testid="email-input" />
 *
 * // In E2E test
 * await page.getByTestId('submit-button').click()
 * await page.getByTestId('email-input').fill('test@example.com')
 *
 * @example Common base component patterns
 * // Submit button pattern
 * <Button type="submit" data-testid="login-page-submit-button">Giriş Yap</Button>
 *
 * // Form input pattern
 * <Input
 *   id="email"
 *   data-testid="login-page-email-input"
 *   placeholder="Email"
 * />
 *
 * // Icon button pattern
 * <Button variant="ghost" size="icon" data-testid="close-button">
 *   <X className="h-4 w-4" />
 * </Button>
 */
