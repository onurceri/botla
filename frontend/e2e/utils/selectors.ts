/**
 * E2E Test Selectors
 * 
 * This module exports all data-testid constants for stable element selection in Playwright tests.
 * All element IDs follow the naming convention: {prefix}-{component}-{action}
 * 
 * Element Type Prefixes:
 * - btn: Buttons
 * - input: Text inputs
 * - select: Select dropdowns
 * - link: Navigation links
 * - tab: Tab navigation
 * - modal: Modal dialogs
 * - toast: Toast notifications
 * - dropdown: Dropdown menus
 * - checkbox: Checkboxes
 * - radio: Radio buttons
 * - toggle: Toggle switches
 * - card: Card components
 * - page: Page containers
 * - header: Headers
 * - footer: Footers
 * - nav: Navigation elements
 * - sidebar: Sidebar components
 * - menu: Menu items
 * - icon: Icon buttons
 * - avatar: User avatars
 * - badge: Badges
 * - table: Table elements
 * - row: Table rows
 * - cell: Table cells
 * - form: Form containers
 * - label: Form labels
 * - error: Error messages
 * - success: Success messages
 * - loading: Loading indicators
 * - empty: Empty states
 * - search: Search inputs
 * - filter: Filter elements
 * - sort: Sort elements
 * - pagination: Pagination elements
 * - breadcrumb: Breadcrumb navigation
 * - tooltip: Tooltips
 * - dialog: Dialogs
 * - alert: Alert messages
 * - progress: Progress indicators
 * - spinner: Loading spinners
 * - skeleton: Skeleton loaders
 * - accordion: Accordion components
 * - carousel: Carousel elements
 * - slider: Slider components
 * - chart: Chart components
 * - graph: Graph elements
 * - widget: Widget components
 * - container: Container elements
 * - wrapper: Wrapper elements
 * - section: Section elements
 * - group: Element groups
 * - item: List items
 * - detail: Detail views
 * - list: List containers
 * - grid: Grid layouts
 * - layout: Layout containers
 * - main: Main content areas
 * - content: Content containers
 * - title: Title elements
 * - heading: Heading elements
 * - text: Text elements
 * - paragraph: Paragraph elements
 * - label: Label elements
 * - description: Description elements
 * - message: Message elements
 * - notification: Notification elements
 * - status: Status indicators
 * - indicator: Indicator elements
 * - tag: Tag elements
 * - chip: Chip components
 * - pill: Pill components
 * - badge: Badge components
 * - marker: Marker elements
 * - point: Point elements
 * - line: Line elements
 * - bar: Bar elements
 * - axis: Axis elements
 * - legend: Legend elements
 * - tooltip: Tooltip elements
 * - handle: Handle elements
 * - track: Track elements
 * - thumb: Thumb elements
 * - rail: Rail elements
 * - step: Step elements
 * - steps: Steps container
 * - wizard: Wizard components
 * - form: Form elements
 * - field: Form fields
 * - control: Form controls
 * - validation: Validation messages
 * - help: Help text
 * - placeholder: Placeholder text
 * - value: Display values
 * - count: Count indicators
 * - total: Total indicators
 * - limit: Limit indicators
 * - threshold: Threshold indicators
 * - setting: Setting elements
 * - option: Option elements
 * - choice: Choice elements
 * - selection: Selection elements
 * - result: Result elements
 * - output: Output elements
 * - input: Input elements
 * - output: Output elements
 * - display: Display elements
 * - view: View elements
 * - preview: Preview elements
 * - thumbnail: Thumbnail elements
 * - image: Image elements
 * - picture: Picture elements
 * - video: Video elements
 * - audio: Audio elements
 * - media: Media elements
 * - file: File elements
 * - upload: Upload elements
 * - download: Download elements
 * - attachment: Attachment elements
 * - document: Document elements
 * - folder: Folder elements
 * - directory: Directory elements
 * - tree: Tree elements
 * - node: Tree nodes
 * - leaf: Leaf nodes
 * - branch: Branch nodes
 * - root: Root elements
 * - level: Level indicators
 * - depth: Depth indicators
 * - hierarchy: Hierarchy elements
 * - structure: Structure elements
 * - organization: Organization elements
 * - relationship: Relationship elements
 * - connection: Connection elements
 * - link: Link elements
 * - edge: Edge elements
 * - vertex: Vertex elements
 * - graph: Graph elements
 * - network: Network elements
 * - cluster: Cluster elements
 * - cloud: Cloud elements
 * - server: Server elements
 * - client: Client elements
 * - endpoint: Endpoint elements
 * - service: Service elements
 * - api: API elements
 * - request: Request elements
 * - response: Response elements
 * - data: Data elements
 * - payload: Payload elements
 * - body: Body elements
 * - header: Header elements
 * - footer: Footer elements
 * - meta: Meta elements
 * - config: Configuration elements
 * - setting: Setting elements
 * - option: Option elements
 * - preference: Preference elements
 * - customization: Customization elements
 * - theme: Theme elements
 * - style: Style elements
 * - appearance: Appearance elements
 * - design: Design elements
 * - layout: Layout elements
 * - position: Position elements
 * - placement: Placement elements
 * - alignment: Alignment elements
 * - spacing: Spacing elements
 * - sizing: Sizing elements
 * - dimension: Dimension elements
 * - size: Size elements
 * - width: Width elements
 * - height: Height elements
 * - length: Length elements
 * - depth: Depth elements
 * - scale: Scale elements
 * - zoom: Zoom elements
 * - pan: Pan elements
 * - rotate: Rotate elements
 * - transform: Transform elements
 * - animation: Animation elements
 * - transition: Transition elements
 * - motion: Motion elements
 * - effect: Effect elements
 * - filter: Filter elements
 * - blur: Blur elements
 * - shadow: Shadow elements
 * - opacity: Opacity elements
 * - transparency: Transparency elements
 * - color: Color elements
 * - palette: Palette elements
 * - scheme: Scheme elements
 * - mode: Mode elements
 * - brightness: Brightness elements
 * - contrast: Contrast elements
 * - saturation: Saturation elements
 * - hue: Hue elements
 * - tint: Tint elements
 * - shade: Shade elements
 * - tone: Tone elements
 */

// ============================================================================
// PAGE SELECTORS
// ============================================================================

export const PAGES = {
  /** Login page container */
  LOGIN_PAGE: 'page-login',
  
  /** Register page container */
  REGISTER_PAGE: 'page-register',
  
  /** Dashboard page container */
  DASHBOARD_PAGE: 'page-dashboard',
  
  /** Chatbots list page container */
  CHATBOTS_PAGE: 'page-chatbots',
  
  /** Chatbot detail page container */
  CHATBOT_DETAIL_PAGE: 'page-chatbot-detail',
  
  /** Chatbot settings page container */
  CHATBOT_SETTINGS_PAGE: 'page-chatbot-settings',
  
  /** Chatbot sources page container */
  CHATBOT_SOURCES_PAGE: 'page-chatbot-sources',
  
  /** Chatbot playground page container */
  CHATBOT_PLAYGROUND_PAGE: 'page-chatbot-playground',
  
  /** Sources page container */
  SOURCES_PAGE: 'page-sources',
  
  /** Analytics page container */
  ANALYTICS_PAGE: 'page-analytics',
  
  /** Settings page container */
  SETTINGS_PAGE: 'page-settings',
  
  /** Profile page container */
  PROFILE_PAGE: 'page-profile',
  
  /** Organization page container */
  ORGANIZATION_PAGE: 'page-organization',
  
  /** Workspace page container */
  WORKSPACE_PAGE: 'page-workspace',
}

// ============================================================================
// AUTH SELECTORS
// ============================================================================

export const AUTH = {
  // Login Form
  LOGIN_FORM: 'form-login',
  LOGIN_TITLE: 'title-login',
  LOGIN_SUBTITLE: 'subtitle-login',
  LOGIN_EMAIL_INPUT: 'input-login-email',
  LOGIN_PASSWORD_INPUT: 'input-login-password',
  LOGIN_SUBMIT_BUTTON: 'btn-login-submit',
  LOGIN_FORGOT_PASSWORD_LINK: 'link-login-forgot-password',
  LOGIN_REMEMBER_ME_CHECKBOX: 'checkbox-login-remember-me',
  LOGIN_ERROR_MESSAGE: 'error-login-message',
  LOGIN_SUCCESS_MESSAGE: 'success-login-message',
  
  // Register Form
  REGISTER_FORM: 'form-register',
  REGISTER_TITLE: 'title-register',
  REGISTER_SUBTITLE: 'subtitle-register',
  REGISTER_NAME_INPUT: 'input-register-name',
  REGISTER_EMAIL_INPUT: 'input-register-email',
  REGISTER_PASSWORD_INPUT: 'input-register-password',
  REGISTER_CONFIRM_PASSWORD_INPUT: 'input-register-confirm-password',
  REGISTER_TERMS_CHECKBOX: 'checkbox-register-terms',
  REGISTER_SUBMIT_BUTTON: 'btn-register-submit',
  REGISTER_ERROR_MESSAGE: 'error-register-message',
  REGISTER_SUCCESS_MESSAGE: 'success-register-message',
  
  // Logout
  LOGOUT_BUTTON: 'btn-logout',
  LOGOUT_CONFIRM_BUTTON: 'btn-logout-confirm',
  LOGOUT_CANCEL_BUTTON: 'btn-logout-cancel',
  LOGOUT_MODAL: 'modal-logout-confirm',
  
  // Password Reset
  PASSWORD_RESET_FORM: 'form-password-reset',
  PASSWORD_RESET_EMAIL_INPUT: 'input-password-reset-email',
  PASSWORD_RESET_SUBMIT_BUTTON: 'btn-password-reset-submit',
  PASSWORD_RESET_SUCCESS_MESSAGE: 'success-password-reset-message',
  
  // Session
  SESSION_TIMEOUT_WARNING: 'warning-session-timeout',
  SESSION_EXTEND_BUTTON: 'btn-session-extend',
  SESSION_LOGOUT_BUTTON: 'btn-session-logout',
}

// ============================================================================
// CHATBOT SELECTORS
// ============================================================================

export const CHATBOTS = {
  // Chatbots List
  CHATBOTS_LIST: 'list-chatbots',
  CHATBOTS_LIST_CONTAINER: 'container-chatbots-list',
  CHATBOTS_EMPTY_STATE: 'empty-chatbots-list',
  CHATBOTS_LOADING: 'loading-chatbots-list',
  CHATBOTS_ERROR_MESSAGE: 'error-chatbots-list',
  CHATBOTS_SEARCH_INPUT: 'input-chatbots-search',
  CHATBOTS_FILTER_SELECT: 'select-chatbots-filter',
  CHATBOTS_SORT_SELECT: 'select-chatbots-sort',
  CHATBOTS_CREATE_BUTTON: 'btn-chatbots-create',
  CHATBOTS_CREATE_MODAL: 'modal-chatbots-create',
  CHATBOTS_CREATE_NAME_INPUT: 'input-chatbots-create-name',
  CHATBOTS_CREATE_DESCRIPTION_INPUT: 'input-chatbots-create-description',
  CHATBOTS_CREATE_SUBMIT_BUTTON: 'btn-chatbots-create-submit',
  CHATBOTS_CREATE_CANCEL_BUTTON: 'btn-chatbots-create-cancel',
  
  // Chatbot Card
  CHATBOT_CARD: 'card-chatbot',
  CHATBOT_CARD_NAME: 'name-chatbot-card',
  CHATBOT_CARD_DESCRIPTION: 'description-chatbot-card',
  CHATBOT_CARD_STATUS: 'status-chatbot-card',
  CHATBOT_CARD_CREATED_AT: 'created-at-chatbot-card',
  CHATBOT_CARD_MANAGE_BUTTON: 'btn-chatbot-manage',
  CHATBOT_CARD_SETTINGS_BUTTON: 'btn-chatbot-settings',
  CHATBOT_CARD_DELETE_BUTTON: 'btn-chatbot-delete',
  CHATBOT_CARD_DUPLICATE_BUTTON: 'btn-chatbot-duplicate',
  CHATBOT_CARD_SHARE_BUTTON: 'btn-chatbot-share',
  CHATBOT_CARD_EMBED_BUTTON: 'btn-chatbot-embed',
  
  // Chatbot Detail Header
  CHATBOT_HEADER: 'header-chatbot-detail',
  CHATBOT_NAME: 'title-chatbot-name',
  CHATBOT_DESCRIPTION: 'description-chatbot',
  CHATBOT_STATUS_BADGE: 'badge-chatbot-status',
  CHATBOT_VERSION: 'text-chatbot-version',
  CHATBOT_CREATED_DATE: 'date-chatbot-created',
  
  // Chatbot Tabs
  CHATBOT_OVERVIEW_TAB: 'tab-chatbot-overview',
  CHATBOT_SOURCES_TAB: 'tab-chatbot-sources',
  CHATBOT_PLAYGROUND_TAB: 'tab-chatbot-playground',
  CHATBOT_SETTINGS_TAB: 'tab-chatbot-settings',
  CHATBOT_ANALYTICS_TAB: 'tab-chatbot-analytics',
  CHATBOT_INTEGRATIONS_TAB: 'tab-chatbot-integrations',
  CHATBOT_CHUNKS_TAB: 'tab-chatbot-chunks',
  
  // Chatbot Settings
  CHATBOT_SETTINGS_NAME_INPUT: 'input-chatbot-settings-name',
  CHATBOT_SETTINGS_DESCRIPTION_INPUT: 'input-chatbot-settings-description',
  CHATBOT_SETTINGS_WELCOME_MESSAGE: 'input-chatbot-settings-welcome-message',
  CHATBOT_SETTINGS_LANGUAGE_SELECT: 'select-chatbot-settings-language',
  CHATBOT_SETTINGS_SAVE_BUTTON: 'btn-chatbot-settings-save',
  CHATBOT_SETTINGS_RESET_BUTTON: 'btn-chatbot-settings-reset',
  CHATBOT_SETTINGS_DELETE_BUTTON: 'btn-chatbot-settings-delete',
  CHATBOT_SETTINGS_DANGER_ZONE: 'zone-chatbot-settings-danger',
  
  // Chatbot Overview
  CHATBOT_OVERVIEW_STATS: 'stats-chatbot-overview',
  CHATBOT_OVERVIEW_TOTAL_MESSAGES: 'stat-chatbot-total-messages',
  CHATBOT_OVERVIEW_TOTAL_CONVERSATIONS: 'stat-chatbot-total-conversations',
  CHATBOT_OVERVIEW_TOTAL_SOURCES: 'stat-chatbot-total-sources',
  CHATBOT_OVERVIEW_TOTAL_CHUNKS: 'stat-chatbot-total-chunks',
  CHATBOT_OVERVIEW_RECENT_ACTIVITY: 'activity-chatbot-overview-recent',
  CHATBOT_OVERVIEW_QUICK_ACTIONS: 'actions-chatbot-overview-quick',
}

// ============================================================================
// SOURCES SELECTORS
// ============================================================================

export const SOURCES = {
  // Sources List
  SOURCES_LIST: 'list-sources',
  SOURCES_LIST_CONTAINER: 'container-sources-list',
  SOURCES_EMPTY_STATE: 'empty-sources-list',
  SOURCES_LOADING: 'loading-sources-list',
  SOURCES_ERROR_MESSAGE: 'error-sources-list',
  
  // Source Card
  SOURCE_CARD: 'card-source',
  SOURCE_CARD_NAME: 'name-source-card',
  SOURCE_CARD_TYPE: 'type-source-card',
  SOURCE_CARD_STATUS: 'status-source-card',
  SOURCE_CARD_CHUNK_COUNT: 'count-source-chunks',
  SOURCE_CARD_CREATED_AT: 'date-source-created',
  SOURCE_CARD_VIEW_BUTTON: 'btn-source-view',
  SOURCE_CARD_DELETE_BUTTON: 'btn-source-delete',
  SOURCE_CARD_EDIT_BUTTON: 'btn-source-edit',
  SOURCE_CARD_INSPECT_BUTTON: 'btn-source-inspect',
  SOURCE_CARD_REFRESH_BUTTON: 'btn-source-refresh',
  
  // Source Uploader
  SOURCE_UPLOADER: 'uploader-source',
  SOURCE_UPLOADER_DRAG_DROP: 'drag-drop-source-uploader',
  SOURCE_UPLOADER_FILE_INPUT: 'input-source-uploader-file',
  SOURCE_UPLOADER_URL_INPUT: 'input-source-uploader-url',
  SOURCE_UPLOADER_TEXT_INPUT: 'input-source-uploader-text',
  SOURCE_UPLOADER_SUBMIT_BUTTON: 'btn-source-uploader-submit',
  SOURCE_UPLOADER_CANCEL_BUTTON: 'btn-source-uploader-cancel',
  SOURCE_UPLOADER_PROGRESS: 'progress-source-uploader',
  SOURCE_UPLOADER_SUCCESS: 'success-source-uploader',
  SOURCE_UPLOADER_ERROR: 'error-source-uploader',
  
  // Source Options
  SOURCE_TEXT_OPTION: 'option-source-text',
  SOURCE_URL_OPTION: 'option-source-url',
  SOURCE_FILE_OPTION: 'option-source-file',
  SOURCE_PDF_OPTION: 'option-source-pdf',
  SOURCE_DOC_OPTION: 'option-source-doc',
  SOURCE_MARKDOWN_OPTION: 'option-source-markdown',
  SOURCE_CSV_OPTION: 'option-source-csv',
  SOURCE_JSON_OPTION: 'option-source-json',
  SOURCE_YOUTUBE_OPTION: 'option-source-youtube',
  SOURCE_WEBSITE_OPTION: 'option-source-website',
  
  // Source Detail
  SOURCE_DETAIL: 'detail-source',
  SOURCE_DETAIL_NAME: 'name-source-detail',
  SOURCE_DETAIL_TYPE: 'type-source-detail',
  SOURCE_DETAIL_STATUS: 'status-source-detail',
  SOURCE_DETAIL_CONTENT: 'content-source-detail',
  SOURCE_DETAIL_METADATA: 'metadata-source-detail',
  SOURCE_DETAIL_CHUNKS_COUNT: 'count-source-detail-chunks',
  
  // Chunk Inspector
  CHUNK_INSPECTOR: 'inspector-chunk',
  CHUNK_INSPECTOR_TITLE: 'title-chunk-inspector',
  CHUNK_INSPECTOR_SEARCH_INPUT: 'input-chunk-inspector-search',
  CHUNK_INSPECTOR_FILTER_SELECT: 'select-chunk-inspector-filter',
  CHUNK_INSPECTOR_SORT_SELECT: 'select-chunk-inspector-sort',
  CHUNK_INSPECTOR_LIST: 'list-chunks-inspector',
  CHUNK_INSPECTOR_PAGINATION: 'pagination-chunk-inspector',
  CHUNK_CARD: 'card-chunk',
  CHUNK_CARD_CONTENT: 'content-chunk-card',
  CHUNK_CARD_SCORE: 'score-chunk-card',
  CHUNK_CARD_INDEX: 'index-chunk-card',
  CHUNK_CARD_SELECT_BUTTON: 'btn-chunk-select',
  CHUNK_CARD_EXPAND_BUTTON: 'btn-chunk-expand',
}

// ============================================================================
// PLAYGROUND SELECTORS
// ============================================================================

export const PLAYGROUND = {
  // Playground Container
  PLAYGROUND_CONTAINER: 'container-playground',
  PLAYGROUND_HEADER: 'header-playground',
  PLAYGROUND_TITLE: 'title-playground',
  PLAYGROUND_DESCRIPTION: 'description-playground',
  
  // Chat Interface
  PLAYGROUND_CHAT_WINDOW: 'window-playground-chat',
  PLAYGROUND_CHAT_MESSAGES: 'messages-playground-chat',
  PLAYGROUND_CHAT_MESSAGE_USER: 'message-playground-chat-user',
  PLAYGROUND_CHAT_MESSAGE_ASSISTANT: 'message-playground-chat-assistant',
  PLAYGROUND_CHAT_MESSAGE_SYSTEM: 'message-playground-chat-system',
  PLAYGROUND_CHAT_INPUT_AREA: 'area-playground-chat-input',
  PLAYGROUND_MESSAGE_INPUT: 'input-playground-message',
  PLAYGROUND_MESSAGE_PLACEHOLDER: 'placeholder-playground-message',
  PLAYGROUND_SEND_BUTTON: 'btn-playground-send',
  PLAYGROUND_ATTACH_BUTTON: 'btn-playground-attach',
  PLAYGROUND_CLEAR_BUTTON: 'btn-playground-clear',
  
  // Chat Actions
  PLAYGROUND_OPEN_CHAT_BUTTON: 'btn-playground-open-chat',
  PLAYGROUND_CLOSE_CHAT_BUTTON: 'btn-playground-close-chat',
  PLAYGROUND_MINIMIZE_BUTTON: 'btn-playground-minimize',
  PLAYGROUND_MAXIMIZE_BUTTON: 'btn-playground-maximize',
  
  // Chat Messages
  PLAYGROUND_USER_MESSAGE: 'message-playground-user',
  PLAYGROUND_BOT_MESSAGE: 'message-playground-bot',
  PLAYGROUND_SOURCE_CITATION: 'citation-playground-source',
  PLAYGROUND_SOURCE_LINK: 'link-playground-source',
  
  // Configuration Panel
  PLAYGROUND_CONFIG_PANEL: 'panel-playground-config',
  PLAYGROUND_CONFIG_OPEN_BUTTON: 'btn-playground-config-open',
  PLAYGROUND_CONFIG_CLOSE_BUTTON: 'btn-playground-config-close',
  PLAYGROUND_MODEL_SELECT: 'select-playground-model',
  PLAYGROUND_TEMPERATURE_INPUT: 'input-playground-temperature',
  PLAYGROUND_MAX_TOKENS_INPUT: 'input-playground-max-tokens',
  PLAYGROUND_SYSTEM_PROMPT: 'input-playground-system-prompt',
  PLAYGROUND_CONFIG_SAVE_BUTTON: 'btn-playground-config-save',
  
  // Quick Actions
  PLAYGROUND_SUGGESTIONS: 'suggestions-playground',
  PLAYGROUND_SUGGESTION_BUTTON: 'btn-playground-suggestion',
}

// ============================================================================
// ANALYTICS SELECTORS
// ============================================================================

export const ANALYTICS = {
  // Analytics Container
  ANALYTICS_CONTAINER: 'container-analytics',
  ANALYTICS_HEADER: 'header-analytics',
  ANALYTICS_TITLE: 'title-analytics',
  ANALYTICS_DATE_RANGE: 'range-analytics-date',
  ANALYTICS_REFRESH_BUTTON: 'btn-analytics-refresh',
  ANALYTICS_EXPORT_BUTTON: 'btn-analytics-export',
  ANALYTICS_FILTER_SELECT: 'select-analytics-filter',
  
  // Charts
  ANALYTICS_MESSAGES_CHART: 'chart-analytics-messages',
  ANALYTICS_CONVERSATIONS_CHART: 'chart-analytics-conversations',
  ANALYTICS_USERS_CHART: 'chart-analytics-users',
  ANALYTICS_SOURCES_CHART: 'chart-analytics-sources',
  ANALYTICS_RETENTION_CHART: 'chart-analytics-retention',
  ANALYTICS_PERFORMANCE_CHART: 'chart-analytics-performance',
  
  // Stats Cards
  ANALYTICS_TOTAL_MESSAGES: 'stat-analytics-total-messages',
  ANALYTICS_TOTAL_CONVERSATIONS: 'stat-analytics-total-conversations',
  ANALYTICS_TOTAL_USERS: 'stat-analytics-total-users',
  ANALYTICS_AVG_RESPONSE_TIME: 'stat-analytics-avg-response-time',
  ANALYTICS_SATISFACTION_SCORE: 'stat-analytics-satisfaction-score',
  ANALYTICS_TOKEN_USAGE: 'stat-analytics-token-usage',
  
  // Tables
  ANALYTICS_CONVERSATIONS_TABLE: 'table-analytics-conversations',
  ANALYTICS_MESSAGES_TABLE: 'table-analytics-messages',
  ANALYTICS_USERS_TABLE: 'table-analytics-users',
  
  // Filters
  ANALYTICS_PERIOD_SELECT: 'select-analytics-period',
  ANALYTICS_BOT_SELECT: 'select-analytics-bot',
  ANALYTICS_CHANNEL_SELECT: 'select-analytics-channel',
}

// ============================================================================
// NAVIGATION SELECTORS
// ============================================================================

export const NAVIGATION = {
  // Sidebar
  SIDEBAR: 'nav-sidebar',
  SIDEBAR_COLLAPSE_BUTTON: 'btn-sidebar-collapse',
  SIDEBAR_EXPAND_BUTTON: 'btn-sidebar-expand',
  SIDEBAR_DASHBOARD_LINK: 'link-sidebar-dashboard',
  SIDEBAR_CHATBOTS_LINK: 'link-sidebar-chatbots',
  SIDEBAR_SOURCES_LINK: 'link-sidebar-sources',
  SIDEBAR_ANALYTICS_LINK: 'link-sidebar-analytics',
  SIDEBAR_SETTINGS_LINK: 'link-sidebar-settings',
  SIDEBAR_HELP_LINK: 'link-sidebar-help',
  SIDEBAR_LOGOUT_LINK: 'link-sidebar-logout',
  
  // Top Navigation
  NAVBAR: 'nav-navbar',
  NAVBAR_MENU_BUTTON: 'btn-navbar-menu',
  NAVBAR_SEARCH_INPUT: 'input-navbar-search',
  NAVBAR_NOTIFICATIONS_BUTTON: 'btn-navbar-notifications',
  NAVBAR_NOTIFICATIONS_DROPDOWN: 'dropdown-navbar-notifications',
  NAVBAR_USER_MENU_BUTTON: 'btn-navbar-user-menu',
  NAVBAR_USER_MENU_DROPDOWN: 'dropdown-navbar-user-menu',
  NAVBAR_USER_AVATAR: 'avatar-navbar-user',
  NAVBAR_USER_NAME: 'name-navbar-user',
  NAVBAR_USER_EMAIL: 'email-navbar-user',
  
  // Breadcrumbs
  BREADCRUMBS: 'nav-breadcrumbs',
  BREADCRUMB_ITEM: 'item-breadcrumb',
  BREADCRUMB_HOME: 'breadcrumb-home',
  
  // Mobile Navigation
  MOBILE_NAV_BOTTOM: 'nav-mobile-bottom',
  MOBILE_NAV_DRAWER: 'drawer-mobile-nav',
  MOBILE_NAV_OPEN_BUTTON: 'btn-mobile-nav-open',
  MOBILE_NAV_CLOSE_BUTTON: 'btn-mobile-nav-close',
}

// ============================================================================
// COMMON SELECTORS
// ============================================================================

export const COMMON = {
  // Loading
  LOADING_SPINNER: 'spinner-loading',
  LOADING_OVERLAY: 'overlay-loading',
  LOADING_SKELETON: 'skeleton-loading',
  LOADING_PROGRESS: 'progress-loading',
  
  // Empty States
  EMPTY_STATE: 'state-empty',
  EMPTY_STATE_TITLE: 'title-empty-state',
  EMPTY_STATE_DESCRIPTION: 'description-empty-state',
  EMPTY_STATE_ACTION_BUTTON: 'btn-empty-state-action',
  EMPTY_STATE_ICON: 'icon-empty-state',
  
  // Error States
  ERROR_STATE: 'state-error',
  ERROR_STATE_TITLE: 'title-error-state',
  ERROR_STATE_DESCRIPTION: 'description-error-state',
  ERROR_STATE_RETRY_BUTTON: 'btn-error-state-retry',
  ERROR_STATE_ICON: 'icon-error-state',
  ERROR_MESSAGE: 'message-error',
  ERROR_TOAST: 'toast-error',
  ERROR_ALERT: 'alert-error',
  
  // Success States
  SUCCESS_STATE: 'state-success',
  SUCCESS_STATE_TITLE: 'title-success-state',
  SUCCESS_STATE_DESCRIPTION: 'description-success-state',
  SUCCESS_MESSAGE: 'message-success',
  SUCCESS_TOAST: 'toast-success',
  SUCCESS_ALERT: 'alert-success',
  
  // Modal
  MODAL: 'modal',
  MODAL_OVERLAY: 'overlay-modal',
  MODAL_TITLE: 'title-modal',
  MODAL_DESCRIPTION: 'description-modal',
  MODAL_CONTENT: 'content-modal',
  MODAL_CLOSE_BUTTON: 'btn-modal-close',
  MODAL_CONFIRM_BUTTON: 'btn-modal-confirm',
  MODAL_CANCEL_BUTTON: 'btn-modal-cancel',
  MODAL_SUBMIT_BUTTON: 'btn-modal-submit',
  
  // Toast
  TOAST: 'toast',
  TOAST_CONTAINER: 'container-toast',
  TOAST_CLOSE_BUTTON: 'btn-toast-close',
  TOAST_TITLE: 'title-toast',
  TOAST_DESCRIPTION: 'description-toast',
  TOAST_SUCCESS: 'toast-success',
  TOAST_ERROR: 'toast-error',
  TOAST_WARNING: 'toast-warning',
  TOAST_INFO: 'toast-info',
  
  // Form Elements
  FORM: 'form',
  FORM_GROUP: 'group-form',
  FORM_LABEL: 'label-form',
  FORM_HELP_TEXT: 'text-form-help',
  FORM_ERROR_TEXT: 'text-form-error',
  FORM_REQUIRED_MARK: 'mark-form-required',
  
  // Buttons
  BUTTON_PRIMARY: 'btn-primary',
  BUTTON_SECONDARY: 'btn-secondary',
  BUTTON_TERTIARY: 'btn-tertiary',
  BUTTON_DANGER: 'btn-danger',
  BUTTON_SUCCESS: 'btn-success',
  BUTTON_WARNING: 'btn-warning',
  BUTTON_INFO: 'btn-info',
  BUTTON_GHOST: 'btn-ghost',
  BUTTON_LINK: 'btn-link',
  BUTTON_ICON: 'btn-icon',
  BUTTON_LOADING: 'btn-loading',
  BUTTON_DISABLED: 'btn-disabled',
  
  // Inputs
  INPUT_TEXT: 'input-text',
  INPUT_EMAIL: 'input-email',
  INPUT_PASSWORD: 'input-password',
  INPUT_NUMBER: 'input-number',
  INPUT_SEARCH: 'input-search',
  INPUT_URL: 'input-url',
  INPUT_TEL: 'input-tel',
  INPUT_DATE: 'input-date',
  INPUT_TIME: 'input-time',
  INPUT_DATETIME: 'input-datetime',
  INPUT_FILE: 'input-file',
  INPUT_TEXTAREA: 'textarea',
  
  // Select
  SELECT: 'select',
  SELECT_TRIGGER: 'trigger-select',
  SELECT_OPTION: 'option-select',
  SELECT_PLACEHOLDER: 'placeholder-select',
  SELECT_SEARCH_INPUT: 'input-select-search',
  SELECT_MULTIPLE_TAG: 'tag-select-multiple',
  
  // Checkbox
  CHECKBOX: 'checkbox',
  CHECKBOX_LABEL: 'label-checkbox',
  CHECKBOX_INPUT: 'input-checkbox',
  CHECKBOX_INDICATOR: 'indicator-checkbox',
  
  // Radio
  RADIO: 'radio',
  RADIO_GROUP: 'group-radio',
  RADIO_LABEL: 'label-radio',
  RADIO_INPUT: 'input-radio',
  RADIO_INDICATOR: 'indicator-radio',
  
  // Toggle
  TOGGLE: 'toggle',
  TOGGLE_THUMB: 'thumb-toggle',
  TOGGLE_TRACK: 'track-toggle',
  
  // Tabs
  TABS: 'tabs',
  TAB_LIST: 'list-tabs',
  TAB: 'tab',
  TAB_PANEL: 'panel-tab',
  TAB_ACTIVE: 'tab-active',
  TAB_DISABLED: 'tab-disabled',
  
  // Cards
  CARD: 'card',
  CARD_HEADER: 'header-card',
  CARD_TITLE: 'title-card',
  CARD_DESCRIPTION: 'description-card',
  CARD_CONTENT: 'content-card',
  CARD_FOOTER: 'footer-card',
  CARD_HOVER: 'card-hover',
  CARD_SELECTED: 'card-selected',
  
  // Tables
  TABLE: 'table',
  TABLE_HEADER: 'header-table',
  TABLE_BODY: 'body-table',
  TABLE_FOOTER: 'footer-table',
  TABLE_ROW: 'row-table',
  TABLE_CELL: 'cell-table',
  TABLE_HEAD: 'head-table',
  TABLE_DATA: 'data-table',
  TABLE_SORT_BUTTON: 'btn-table-sort',
  TABLE_FILTER_BUTTON: 'btn-table-filter',
  TABLE_SEARCH_INPUT: 'input-table-search',
  TABLE_PAGINATION: 'pagination-table',
  TABLE_EMPTY: 'empty-table',
  
  // Pagination
  PAGINATION: 'pagination',
  PAGINATION_PREV_BUTTON: 'btn-pagination-prev',
  PAGINATION_NEXT_BUTTON: 'btn-pagination-next',
  PAGINATION_PAGE_BUTTON: 'btn-pagination-page',
  PAGINATION_ACTIVE_PAGE: 'page-pagination-active',
  PAGINATION_ELLIPSIS: 'ellipsis-pagination',
  
  // Dropdown
  DROPDOWN: 'dropdown',
  DROPDOWN_TRIGGER: 'trigger-dropdown',
  DROPDOWN_MENU: 'menu-dropdown',
  DROPDOWN_ITEM: 'item-dropdown',
  DROPDOWN_DIVIDER: 'divider-dropdown',
  DROPDOWN_LABEL: 'label-dropdown',
  DROPDOWN_SEARCH_INPUT: 'input-dropdown-search',
  DROPDOWN_SELECTED_ITEM: 'item-dropdown-selected',
  
  // Tooltip
  TOOLTIP: 'tooltip',
  TOOLTIP_CONTENT: 'content-tooltip',
  TOOLTIP_TRIGGER: 'trigger-tooltip',
  
  // Avatar
  AVATAR: 'avatar',
  AVATAR_IMAGE: 'image-avatar',
  AVATAR_FALLBACK: 'fallback-avatar',
  AVATAR_SIZE_SMALL: 'avatar-small',
  AVATAR_SIZE_MEDIUM: 'avatar-medium',
  AVATAR_SIZE_LARGE: 'avatar-large',
  
  // Badge
  BADGE: 'badge',
  BADGE_PRIMARY: 'badge-primary',
  BADGE_SECONDARY: 'badge-secondary',
  BADGE_SUCCESS: 'badge-success',
  BADGE_WARNING: 'badge-warning',
  BADGE_ERROR: 'badge-error',
  BADGE_INFO: 'badge-info',
  
  // Tag
  TAG: 'tag',
  TAG_REMOVE_BUTTON: 'btn-tag-remove',
  TAG_ADD_BUTTON: 'btn-tag-add',
  
  // Progress
  PROGRESS: 'progress',
  PROGRESS_BAR: 'bar-progress',
  PROGRESS_VALUE: 'value-progress',
  PROGRESS_LABEL: 'label-progress',
  
  // Accordion
  ACCORDION: 'accordion',
  ACCORDION_ITEM: 'item-accordion',
  ACCORDION_TRIGGER: 'trigger-accordion',
  ACCORDION_CONTENT: 'content-accordion',
  ACCORDION_EXPANDED: 'accordion-expanded',
  ACCORDION_COLLAPSED: 'accordion-collapsed',
  
  // Alert
  ALERT: 'alert',
  ALERT_TITLE: 'title-alert',
  ALERT_DESCRIPTION: 'description-alert',
  ALERT_DISMISS_BUTTON: 'btn-alert-dismiss',
  ALERT_ICON: 'icon-alert',
  
  // Dialog
  DIALOG: 'dialog',
  DIALOG_TRIGGER: 'trigger-dialog',
  DIALOG_OVERLAY: 'overlay-dialog',
  DIALOG_CONTENT: 'content-dialog',
  DIALOG_TITLE: 'title-dialog',
  DIALOG_DESCRIPTION: 'description-dialog',
  DIALOG_CLOSE_BUTTON: 'btn-dialog-close',
  
  // Popover
  POPOVER: 'popover',
  POPOVER_TRIGGER: 'trigger-popover',
  POPOVER_CONTENT: 'content-popover',
  POPOVER_CLOSE_BUTTON: 'btn-popover-close',
  
  // Menu
  MENU: 'menu',
  MENU_ITEM: 'item-menu',
  MENU_SEPARATOR: 'separator-menu',
  MENU_LABEL: 'label-menu',
  
  // Context Menu
  CONTEXT_MENU: 'menu-context',
  CONTEXT_MENU_TRIGGER: 'trigger-context-menu',
  CONTEXT_MENU_ITEM: 'item-context-menu',
  CONTEXT_MENU_SEPARATOR: 'separator-context-menu',
  
  // User Menu
  USER_MENU: 'menu-user',
  USER_MENU_BUTTON: 'btn-user-menu',
  USER_MENU_DROPDOWN: 'dropdown-user-menu',
  USER_MENU_ITEM: 'item-user-menu',
  USER_MENU_AVATAR: 'avatar-user-menu',
  USER_MENU_NAME: 'name-user-menu',
  USER_MENU_EMAIL: 'email-user-menu',
  USER_MENU_PROFILE_LINK: 'link-user-menu-profile',
  USER_MENU_SETTINGS_LINK: 'link-user-menu-settings',
  USER_MENU_LOGOUT_BUTTON: 'btn-user-menu-logout',
  
  // Search
  SEARCH: 'search',
  SEARCH_INPUT: 'input-search',
  SEARCH_RESULTS: 'results-search',
  SEARCH_RESULT_ITEM: 'item-search-result',
  SEARCH_NO_RESULTS: 'no-results-search',
  SEARCH_CLEAR_BUTTON: 'btn-search-clear',
  SEARCH_SUBMIT_BUTTON: 'btn-search-submit',
  
  // Filter
  FILTER: 'filter',
  FILTER_BUTTON: 'btn-filter',
  FILTER_PANEL: 'panel-filter',
  FILTER_CHIP: 'chip-filter',
  FILTER_CLEAR_BUTTON: 'btn-filter-clear',
  FILTER_APPLY_BUTTON: 'btn-filter-apply',
  
  // Sort
  SORT: 'sort',
  SORT_BUTTON: 'btn-sort',
  SORT_MENU: 'menu-sort',
  SORT_OPTION: 'option-sort',
  SORT_ASC_ICON: 'icon-sort-asc',
  SORT_DESC_ICON: 'icon-sort-desc',
  
  // Organization
  ORGANIZATION_SELECT: 'select-organization',
  ORGANIZATION_AVATAR: 'avatar-organization',
  ORGANIZATION_NAME: 'name-organization',
  ORGANIZATION_SWITCH_BUTTON: 'btn-switch-organization',
  ORGANIZATION_CREATE_BUTTON: 'btn-create-organization',
  ORGANIZATION_EDIT_BUTTON: 'btn-edit-organization',
  
  // Workspace
  WORKSPACE_SELECT: 'select-workspace',
  WORKSPACE_AVATAR: 'avatar-workspace',
  WORKSPACE_NAME: 'name-workspace',
  WORKSPACE_SWITCH_BUTTON: 'btn-switch-workspace',
  WORKSPACE_CREATE_BUTTON: 'btn-create-workspace',
  WORKSPACE_EDIT_BUTTON: 'btn-edit-workspace',
}

// ============================================================================
// WIDGET SELECTORS
// ============================================================================

export const WIDGET = {
  // Widget Host
  WIDGET_HOST: 'host-widget',
  WIDGET_CONTAINER: 'container-widget',
  WIDGET_WRAPPER: 'wrapper-widget',
  
  // Widget Bubble
  WIDGET_BUBBLE: 'bubble-widget',
  WIDGET_BUBBLE_OPEN: 'bubble-widget-open',
  WIDGET_BUBBLE_CLOSE: 'bubble-widget-close',
  WIDGET_BUBBLE_ICON: 'icon-bubble-widget',
  WIDGET_BUBBLE_BADGE: 'badge-bubble-widget',
  
  // Widget Window
  WIDGET_WINDOW: 'window-widget',
  WIDGET_WINDOW_HEADER: 'header-widget-window',
  WIDGET_WINDOW_TITLE: 'title-widget-window',
  WIDGET_WINDOW_CLOSE_BUTTON: 'btn-widget-window-close',
  WIDGET_WINDOW_MINIMIZE_BUTTON: 'btn-widget-window-minimize',
  WIDGET_WINDOW_MAXIMIZE_BUTTON: 'btn-widget-window-maximize',
  WIDGET_WINDOW_DRAG_HANDLE: 'handle-widget-window-drag',
  
  // Widget Chat
  WIDGET_CHAT: 'chat-widget',
  WIDGET_CHAT_MESSAGES: 'messages-widget-chat',
  WIDGET_CHAT_MESSAGE_USER: 'message-widget-chat-user',
  WIDGET_CHAT_MESSAGE_BOT: 'message-widget-chat-bot',
  WIDGET_CHAT_INPUT_AREA: 'area-widget-chat-input',
  WIDGET_CHAT_INPUT: 'input-widget-chat',
  WIDGET_CHAT_SEND_BUTTON: 'btn-widget-chat-send',
  WIDGET_CHAT_ATTACH_BUTTON: 'btn-widget-chat-attach',
  WIDGET_CHAT_PLACEHOLDER: 'placeholder-widget-chat',
  
  // Widget Header
  WIDGET_HEADER: 'header-widget',
  WIDGET_HEADER_TITLE: 'title-widget-header',
  WIDGET_HEADER_SUBTITLE: 'subtitle-widget-header',
  WIDGET_HEADER_AVATAR: 'avatar-widget-header',
  WIDGET_HEADER_STATUS: 'status-widget-header',
  
  // Widget Footer
  WIDGET_FOOTER: 'footer-widget',
  WIDGET_FOOTER_POWERED_BY: 'text-widget-powered-by',
  WIDGET_FOOTER_LINK: 'link-widget-footer',
  
  // Widget Branding
  WIDGET_BRANDING: 'branding-widget',
  WIDGET_BRANDING_LOGO: 'logo-widget-branding',
  WIDGET_BRANDING_TEXT: 'text-widget-branding',
  WIDGET_BRANDING_LINK: 'link-widget-branding',
  
  // Widget Theme
  WIDGET_THEME_COLOR: 'color-widget-theme',
  WIDGET_THEME_FONT: 'font-widget-theme',
  WIDGET_THEME_POSITION: 'position-widget-theme',
}

// ============================================================================
// MOBILE SELECTORS
// ============================================================================

export const MOBILE = {
  // Mobile Navigation
  MOBILE_MENU_BUTTON: 'btn-mobile-menu',
  MOBILE_MENU_DRAWER: 'drawer-mobile-menu',
  MOBILE_MENU_CLOSE_BUTTON: 'btn-mobile-menu-close',
  MOBILE_MENU_OVERLAY: 'overlay-mobile-menu',
  
  // Mobile Bottom Navigation
  MOBILE_BOTTOM_NAV: 'nav-mobile-bottom',
  MOBILE_BOTTOM_NAV_ITEM: 'item-mobile-bottom-nav',
  MOBILE_BOTTOM_NAV_HOME: 'nav-mobile-bottom-home',
  MOBILE_BOTTOM_NAV_CHATBOTS: 'nav-mobile-bottom-chatbots',
  MOBILE_BOTTOM_NAV_SOURCES: 'nav-mobile-bottom-sources',
  MOBILE_BOTTOM_NAV_ANALYTICS: 'nav-mobile-bottom-analytics',
  MOBILE_BOTTOM_NAV_PROFILE: 'nav-mobile-bottom-profile',
  
  // Mobile Search
  MOBILE_SEARCH_BUTTON: 'btn-mobile-search',
  MOBILE_SEARCH_BAR: 'bar-mobile-search',
  MOBILE_SEARCH_CLOSE_BUTTON: 'btn-mobile-search-close',
  
  // Mobile Actions
  MOBILE_FAB: 'fab-mobile',
  MOBILE_FAB_CREATE: 'fab-mobile-create',
  MOBILE_ACTION_SHEET: 'sheet-mobile-action',
  MOBILE_ACTION_SHEET_OPTION: 'option-mobile-action-sheet',
  MOBILE_ACTION_SHEET_CANCEL: 'btn-mobile-action-sheet-cancel',
}

// ============================================================================
// EXPORTS
// ============================================================================

/**
 * All selectors organized by category
 */
export const SELECTORS = {
  PAGES,
  AUTH,
  CHATBOTS,
  SOURCES,
  PLAYGROUND,
  ANALYTICS,
  NAVIGATION,
  COMMON,
  WIDGET,
  MOBILE,
}

/**
 * Type for selector categories
 */
export type SelectorCategory = keyof typeof SELECTORS

/**
 * Get all selectors from a category
 */
export function getSelectorsByCategory(category: SelectorCategory): Record<string, string> {
  return SELECTORS[category] as Record<string, string>
}

/**
 * Get a specific selector
 */
export function getSelector(category: SelectorCategory, key: string): string | undefined {
  const categorySelectors = SELECTORS[category]
  if (!categorySelectors) return undefined
  return (categorySelectors as Record<string, string>)[key]
}
