package api

// Error codes for API responses.
// Frontend handles localization - backend returns only codes.
// Group: HTTP standard errors
const (
	ErrCodeBadRequest          = "BAD_REQUEST"
	ErrCodeUnauthorized        = "UNAUTHORIZED"
	ErrCodeForbidden           = "FORBIDDEN"
	ErrCodeNotFound            = "NOT_FOUND"
	ErrCodeConflict            = "CONFLICT"
	ErrCodeTooManyRequests     = "TOO_MANY_REQUESTS"
	ErrCodePaymentRequired     = "PAYMENT_REQUIRED"
	ErrCodeInternalError       = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable  = "SERVICE_UNAVAILABLE"
	ErrCodeMethodNotAllowed    = "METHOD_NOT_ALLOWED"
	ErrCodeRequestEntityTooBig = "REQUEST_ENTITY_TOO_LARGE"
	ErrCodeGone                = "GONE"
)

// Group: Authentication errors
const (
	ErrEmailRequired           = "ERR_EMAIL_REQUIRED"
	ErrPasswordRequired        = "ERR_PASSWORD_REQUIRED"
	ErrEmailAndPasswordReq     = "ERR_EMAIL_AND_PASSWORD_REQUIRED" //nolint:gosec
	ErrInvalidEmailFormat      = "ERR_INVALID_EMAIL_FORMAT"
	ErrPasswordTooShort        = "ERR_PASSWORD_TOO_SHORT"
	ErrPasswordWeak            = "ERR_PASSWORD_WEAK"
	ErrEmailExists             = "ERR_EMAIL_EXISTS"
	ErrInvalidCredentials      = "ERR_INVALID_CREDENTIALS" //nolint:gosec
	ErrInvalidRequestBody      = "ERR_INVALID_REQUEST_BODY"
	ErrDatabaseError           = "ERR_DATABASE_ERROR"
	ErrFailedToHashPassword    = "ERR_FAILED_TO_HASH_PASSWORD" //nolint:gosec
	ErrFailedToCreateUser      = "ERR_FAILED_TO_CREATE_USER"
	ErrInvalidIDFormat         = "ERR_INVALID_ID_FORMAT"
	ErrMissingID               = "ERR_MISSING_ID"
)

// Group: Chatbot/Source errors
const (
	ErrMonthlyTokensExceeded   = "ERR_MONTHLY_TOKENS_EXCEEDED" //nolint:gosec
	ErrPdfLimitReached         = "ERR_PDF_LIMIT_REACHED"
	ErrFileTooLarge            = "ERR_FILE_TOO_LARGE"
	ErrReaddCooldownActive     = "ERR_READD_COOLDOWN_ACTIVE"
	ErrDuplicateURL            = "ERR_DUPLICATE_URL"
	ErrOnlyURLRefresh          = "ERR_ONLY_URL_REFRESH"
	ErrSourceAlreadyProcessing = "ERR_SOURCE_ALREADY_PROCESSING"
	ErrPlanRefreshUnavailable  = "ERR_PLAN_REFRESH_UNAVAILABLE"
	ErrMonthlyRefreshExceeded  = "ERR_MONTHLY_REFRESH_EXCEEDED"
	ErrRefreshCooldownActive   = "ERR_REFRESH_COOLDOWN_ACTIVE"
	ErrNoURLsProvided          = "ERR_NO_URLS_PROVIDED"
	ErrURLLimitReached         = "ERR_URL_LIMIT_REACHED"
	ErrMonthlyIngestionExceeded = "ERR_MONTHLY_INGESTION_EXCEEDED"
	ErrSitemapParseFailed      = "ERR_SITEMAP_PARSE_FAILED"
	ErrMaxChatbotsExceeded     = "ERR_MAX_CHATBOTS_EXCEEDED"
	ErrTextTooLong             = "ERR_TEXT_TOO_LONG"
	ErrDuplicateContent        = "ERR_DUPLICATE_CONTENT"
)

// Group: Action errors
const (
	ErrNameAndActionTypeRequired = "ERR_NAME_AND_ACTION_TYPE_REQUIRED"
	ErrInvalidStatus             = "ERR_INVALID_STATUS"
)

// Group: Handoff errors
const (
	ErrHandoffExists      = "ERR_HANDOFF_EXISTS"
	ErrHandoffNotFound    = "ERR_HANDOFF_NOT_FOUND"
	ErrHandoffExpired     = "ERR_HANDOFF_EXPIRED"
	ErrHandoffClosed      = "ERR_HANDOFF_CLOSED"
	ErrHandoffRateLimited = "ERR_HANDOFF_RATE_LIMITED"
	ErrHandoffNotEnabled  = "ERR_HANDOFF_NOT_ENABLED"
)

// Group: Privacy/Export errors
const (
	ErrMissingRequestID     = "ERR_MISSING_REQUEST_ID"
	ErrMissingExportID      = "ERR_MISSING_EXPORT_ID"
	ErrMissingUserID        = "ERR_MISSING_USER_ID"
	ErrPrivacyRequestNotFound = "ERR_PRIVACY_REQUEST_NOT_FOUND"
	ErrNotExportRequest     = "ERR_NOT_EXPORT_REQUEST"
	ErrExportNotReady       = "ERR_EXPORT_NOT_READY"
	ErrExportURLMissing     = "ERR_EXPORT_URL_MISSING"
	ErrExportExpired        = "ERR_EXPORT_EXPIRED"
	ErrStorageNotConfigured = "ERR_STORAGE_NOT_CONFIGURED"
	ErrFailedToGenerateURL  = "ERR_FAILED_TO_GENERATE_URL"
	ErrFailedToDownload     = "ERR_FAILED_TO_DOWNLOAD"
	ErrInvalidAction        = "ERR_INVALID_ACTION"
)

// Group: Admin errors
const (
	ErrFailedToFetchStats      = "ERR_FAILED_TO_FETCH_STATS"
	ErrFailedToListUsers       = "ERR_FAILED_TO_LIST_USERS"
	ErrUserNotFound            = "ERR_USER_NOT_FOUND"
	ErrFailedToGetUser         = "ERR_FAILED_TO_GET_USER"
	ErrFailedToUpdateUser      = "ERR_FAILED_TO_UPDATE_USER"
	ErrFailedToListOrgs        = "ERR_FAILED_TO_LIST_ORGS"
	ErrOrgNotFound             = "ERR_ORG_NOT_FOUND"
	ErrFailedToGetOrg          = "ERR_FAILED_TO_GET_ORG"
	ErrMissingJobID            = "ERR_MISSING_JOB_ID"
	ErrFailedToFetchJobs       = "ERR_FAILED_TO_FETCH_JOBS"
	ErrFailedToRetryJob        = "ERR_FAILED_TO_RETRY_JOB"
	ErrFailedToDeleteJob       = "ERR_FAILED_TO_DELETE_JOB"
)

// Group: Workspace/Organization errors
const (
	ErrNotWorkspaceMember  = "ERR_NOT_WORKSPACE_MEMBER"
	ErrNotOrgMember        = "ERR_NOT_ORG_MEMBER"
	ErrWorkspaceCheckError = "ERR_WORKSPACE_CHECK_ERROR"
	ErrMembershipCheckError = "ERR_MEMBERSHIP_CHECK_ERROR"
	ErrGetPlanError        = "ERR_GET_PLAN_ERROR"
	ErrCreateBotError      = "ERR_CREATE_BOT_ERROR"
)
