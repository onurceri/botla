# Comprehensive Testing Checklist - Botla Platform

## Document Purpose
This is a comprehensive, production-ready testing checklist covering every feature, edge case, and scenario in the Botla chatbot platform. Use this document to systematically verify all functionality before production deployment.

---

## 1. Authentication & User Management

### 1.1 User Registration
- [ ] Register with valid email and password (minimum length requirements)
- [ ] Register with invalid email format returns validation error
- [ ] Register with existing email returns conflict error
- [ ] Register with weak password returns validation error
- [ ] Register with missing required fields returns error
- [ ] Successful registration creates user in database
- [ ] Successful registration assigns default free plan
- [ ] User receives JWT access and refresh tokens
- [ ] Password is hashed with bcrypt (never stored in plain text)
- [ ] SQL injection attempts in registration fields are blocked
- [ ] XSS attempts in registration fields are sanitized

### 1.2 User Login
- [ ] Login with valid credentials succeeds
- [ ] Login with invalid email returns error
- [ ] Login with invalid password returns error
- [ ] Login with non-existent user returns error
- [ ] Login returns access token and refresh token
- [ ] Access token has correct expiry time
- [ ] Refresh token has correct expiry time
- [ ] Multiple login attempts are rate-limited
- [ ] Case-insensitive email matching works
- [ ] Login creates refresh token tracking record

### 1.3 Token Management
- [ ] Access token validates correctly for protected routes
- [ ] Expired access token returns 401 Unauthorized
- [ ] Refresh token can generate new access token
- [ ] Expired refresh token returns 401
- [ ] Invalid refresh token returns 401
- [ ] Revoked refresh token cannot be reused
- [ ] Logout invalidates refresh token
- [ ] Multiple concurrent sessions are tracked separately
- [ ] JWT secret is properly configured and secure
- [ ] Token includes correct user_id claim

### 1.4 Profile Management
- [ ] GET /api/v1/me returns current user info
- [ ] User info includes plan details
- [ ] User info includes usage statistics
- [ ] Unauthorized request to /me returns 401
- [ ] User cannot access another user's profile
- [ ] Profile includes organization memberships (if applicable)

---

## 2. Plan & Subscription Management

### 2.1 Free Plan Limits
- [ ] Free plan assigned by default on registration
- [ ] Chat model limited to gpt-4o-mini only
- [ ] Monthly token limit: 100,000 tokens
- [ ] RAG Top-K limited to 3
- [ ] RAG Context limited to 2,000 tokens
- [ ] Max 1 PDF file per chatbot enforced
- [ ] Max 5MB per PDF file enforced
- [ ] Max 10MB total storage enforced
- [ ] OCR disabled (image/PDF text extraction)
- [ ] Max 1 URL per chatbot enforced
- [ ] Dynamic (JS) scraping disabled
- [ ] Discovery mode disabled (max_pages_per_crawl = 0)
- [ ] Refresh disabled (no manual/auto refresh)
- [ ] "Powered by Botla" branding visible and cannot be hidden
- [ ] Secure embed disabled (cannot set allowed domains or secret)
- [ ] Max 50 ingestions per month enforced
- [ ] Max 250,000 embedding tokens per month enforced
- [ ] 60-minute cooldown between re-adding same source enforced
- [ ] Guardrails restricted: cannot customize thresholds, smart/escalate fallback, or topic management

### 2.2 Pro Plan Features
- [ ] Pro plan allows gpt-4o-mini and gpt-4o models
- [ ] Monthly token limit: 1,000,000 tokens
- [ ] RAG Top-K limited to 5
- [ ] RAG Context limited to 4,000 tokens
- [ ] Max 20 PDF files per chatbot enforced
- [ ] Max 20MB per PDF file enforced
- [ ] Max 500MB total storage enforced
- [ ] OCR enabled for PDF/image text extraction
- [ ] Max 10 URLs per chatbot enforced
- [ ] Dynamic (JS) scraping enabled
- [ ] Discovery mode enabled (max 10 pages per crawl)
- [ ] Manual and auto-refresh available
- [ ] Can hide "Powered by Botla" branding
- [ ] Secure embed enabled (can set allowed domains and secret)
- [ ] Max 50 ingestions per month enforced
- [ ] Max 250,000 embedding tokens per month enforced
- [ ] Full guardrails access: custom thresholds, smart/escalate fallback, topic management

### 2.3 Ultra Plan Features
- [ ] Ultra plan allows gpt-4o-mini, gpt-4o, and claude-3-5-sonnet
- [ ] Monthly token limit: 5,000,000 tokens
- [ ] RAG Top-K limited to 10
- [ ] RAG Context limited to 8,000 tokens
- [ ] Max 100 PDF files per chatbot enforced
- [ ] Max 50MB per PDF file enforced
- [ ] Max 2,000MB total storage enforced
- [ ] OCR enabled
- [ ] Max 50 URLs per chatbot enforced
- [ ] Dynamic (JS) scraping enabled
- [ ] Discovery mode enabled (max 100 pages per crawl)
- [ ] Manual and auto-refresh available
- [ ] Can hide "Powered by Botla" branding
- [ ] Can use custom branding (logo, text, link)
- [ ] Secure embed enabled
- [ ] Max 50 ingestions per month enforced
- [ ] Max 250,000 embedding tokens per month enforced
- [ ] Full guardrails access: custom thresholds, smart/escalate fallback, topic management

### 2.4 Plan Enforcement - Security Critical
- [ ] Free user CANNOT enable secure_embed via API bypass
- [ ] Free user CANNOT set auto-refresh via API bypass
- [ ] Free user CANNOT enable discovery mode via API bypass
- [ ] Free user CANNOT select gpt-4o model via API bypass
- [ ] Free user CANNOT upload files exceeding size limit via API bypass
- [ ] Free user CANNOT add URLs exceeding limit via API bypass
- [ ] Pro user CANNOT select claude-3-5-sonnet via API bypass
- [ ] Attempting plan bypass returns 403 Forbidden with upgrade message
- [ ] Plan limits validated on backend for all endpoints
- [ ] Frontend limits match backend enforcement

### 2.5 Usage Tracking & Display
- [ ] Monthly token usage tracked correctly
- [ ] Monthly embedding token usage tracked correctly
- [ ] Monthly ingestion count tracked correctly
- [ ] Storage usage calculated correctly
- [ ] Usage resets monthly
- [ ] Frontend displays ingestion usage counter
- [ ] Frontend displays embedding token usage counter
- [ ] Frontend displays storage usage
- [ ] Frontend displays monthly token usage
- [ ] Exceeding limits returns clear error messages
- [ ] Usage tracking survives database restarts

---

## 3. Chatbot Management

### 3.1 Chatbot Creation
- [ ] Create chatbot with required fields succeeds
- [ ] Create chatbot assigns unique ID
- [ ] Create chatbot defaults to user context (no organization)
- [ ] Create chatbot with organization_id assigns to organization
- [ ] Create chatbot with workspace_id assigns to workspace
- [ ] Create chatbot validates model against plan allowed_models
- [ ] Create chatbot rejects invalid model for plan
- [ ] Create chatbot sets default custom_instruction (empty)
- [ ] Create chatbot sets default welcome message
- [ ] Create chatbot sets default theme colors
- [ ] Create chatbot sets default temperature (e.g., 0.7)
- [ ] Create chatbot sets default max_tokens
- [ ] Create chatbot sets default position (e.g., bottom-right)
- [ ] Create chatbot sets default refresh_policy (manual)
- [ ] Create chatbot sets default discovery_mode (disabled)
- [ ] Exceeding max_chatbots limit returns error
- [ ] Chatbot belongs to correct user_id
- [ ] Chatbot timestamps (created_at, updated_at) set correctly

### 3.2 Chatbot Retrieval
- [ ] GET /api/v1/chatbots lists all user chatbots
- [ ] GET /api/v1/chatbots filters by organization context if present
- [ ] GET /api/v1/chatbots filters by workspace if present
- [ ] GET /api/v1/chatbots/:id returns single chatbot details
- [ ] User can only access their own chatbots
- [ ] User cannot access other users' chatbots (403)
- [ ] Invalid chatbot ID returns 404
- [ ] Deleted chatbots are not returned (soft delete check)
- [ ] Response includes custom_branding if set
- [ ] Response includes threshold_config if set
- [ ] Response includes fallback_messages if set
- [ ] Response includes topic_restrictions if set
- [ ] Response includes handoff_config if enabled

### 3.3 Chatbot Update
- [ ] Update chatbot name succeeds
- [ ] Update chatbot description succeeds
- [ ] Update custom_instruction succeeds
- [ ] Update welcome_message succeeds
- [ ] Update theme_color succeeds
- [ ] Update model succeeds (if allowed by plan)
- [ ] Update model fails if not allowed by plan (403)
- [ ] Update temperature succeeds (0.0 - 2.0 range)
- [ ] Update max_tokens succeeds
- [ ] Update position succeeds (bottom-right, bottom-left, etc.)
- [ ] Update language succeeds
- [ ] Update color settings (bot_message_color, user_message_color, etc.)
- [ ] Update hide_branding succeeds for Pro+ plans
- [ ] Update hide_branding fails for Free plan (403)
- [ ] Update custom_branding succeeds for Ultra plan
- [ ] Update custom_branding fails for Free/Pro plans (403)
- [ ] Update secure_embed_enabled succeeds for Pro+ plans
- [ ] Update secure_embed_enabled fails for Free plan (403)
- [ ] Update allowed_domains succeeds when secure_embed enabled
- [ ] Update embed_secret succeeds when secure_embed enabled
- [ ] Update refresh_policy to "auto" succeeds for Pro+ with refresh enabled
- [ ] Update refresh_policy to "auto" fails for Free plan (403)
- [ ] Update refresh_frequency succeeds (daily/weekly/monthly)
- [ ] Update discovery_mode to "auto" succeeds for Pro+ plans
- [ ] Update discovery_mode to "auto" fails for Free plan (403)
- [ ] Update suggested_questions array succeeds
- [ ] Update suggestions_enabled boolean succeeds
- [ ] Update include_paths array succeeds
- [ ] Update exclude_paths array succeeds
- [ ] Update selector_whitelist array succeeds
- [ ] Update confidence_threshold succeeds
- [ ] Update threshold_config succeeds (high/medium thresholds, fallback mode)
- [ ] Update fallback_messages succeeds
- [ ] Update topic_restrictions succeeds
- [ ] Update handoff settings succeeds
- [ ] Invalid field values return 400 validation error
- [ ] User cannot update another user's chatbot (403)
- [ ] updated_at timestamp updates correctly

### 3.4 Chatbot Deletion
- [ ] Delete chatbot succeeds for owner
- [ ] Delete chatbot returns 403 for non-owner
- [ ] Delete chatbot soft deletes (sets deleted_at timestamp)
- [ ] Deleted chatbot no longer appears in list
- [ ] Deleted chatbot sources are cascade deleted
- [ ] Deleted chatbot conversations are cascade deleted
- [ ] Deleted chatbot actions are cascade deleted
- [ ] Deleted chatbot analytics are preserved
- [ ] Cannot chat with deleted chatbot
- [ ] Invalid chatbot ID returns 404

### 3.5 Chatbot Advanced Configuration

#### 3.5.1 Guardrails - Confidence Thresholds
- [ ] Default threshold config set on chatbot creation
- [ ] Update high_threshold (default 0.50) succeeds
- [ ] Update medium_threshold (default 0.30) succeeds
- [ ] high_threshold must be >= medium_threshold validation
- [ ] Update fallback_mode to "smart" succeeds
- [ ] Update fallback_mode to "static" succeeds
- [ ] Update fallback_mode to "escalate" succeeds
- [ ] Show_confidence_warning toggle works
- [ ] Invalid threshold values (< 0 or > 1) rejected
- [ ] Threshold config persisted in database

#### 3.5.2 Guardrails - Fallback Messages
- [ ] Update no_info_found message succeeds
- [ ] Update error_message succeeds
- [ ] Update handoff_message succeeds
- [ ] Empty fallback messages use defaults
- [ ] Fallback messages displayed in widget correctly
- [ ] Localized fallback messages work with language setting

#### 3.5.3 Guardrails - Topic Restrictions
- [ ] Update allowed_topics array succeeds
- [ ] Update blocked_topics array succeeds
- [ ] Update blocked_message succeeds
- [ ] Chat blocked when topic matches blocked_topics
- [ ] Chat allowed when topic matches allowed_topics
- [ ] blocked_message displayed when topic blocked
- [ ] Topic detection works with AI provider

#### 3.5.4 Handoff Configuration
- [ ] Enable handoff_enabled boolean succeeds
- [ ] Set handoff_type to "email" succeeds
- [ ] Update handoff_config with email details succeeds
- [ ] Handoff triggers when confidence below threshold (escalate mode)
- [ ] Handoff email sent to configured address
- [ ] Handoff message displayed in chat
- [ ] User can continue chat after handoff
- [ ] Handoff tracked in analytics

---

## 4. Source Management

### 4.1 URL Source Creation
- [ ] Add URL source with valid URL succeeds
- [ ] Add URL returns 400 for invalid URL format
- [ ] Add URL returns 400 for empty URL
- [ ] Add URL enforces max_urls_per_bot limit
- [ ] Add URL enforces monthly ingestion limit
- [ ] Add URL creates source with status "pending"
- [ ] Add URL enqueues source for processing
- [ ] Add duplicate URL returns error (based on hash/URL)
- [ ] Add URL respects 60-minute re-add cooldown
- [ ] Add URL increments ingestion count
- [ ] Add URL validates user owns chatbot
- [ ] Add URL fails for deleted chatbot
- [ ] Add URL respects plan scraping limits

### 4.2 URL Bulk Creation
- [ ] Bulk add multiple URLs succeeds
- [ ] Bulk add returns created_count and skipped_count
- [ ] Bulk add stops at max_urls_per_bot limit
- [ ] Bulk add stops at monthly ingestion limit
- [ ] Bulk add skips duplicate URLs
- [ ] Bulk add processes valid URLs only
- [ ] Bulk add returns errors array for failures
- [ ] Empty URLs array returns 400
- [ ] Bulk add enqueues all created sources

### 4.3 PDF/File Source Creation
- [ ] Upload PDF file with valid file succeeds
- [ ] Upload enforces max_files_per_bot limit
- [ ] Upload enforces max file size limit
- [ ] Upload enforces total storage limit
- [ ] Upload returns 413 for oversized file
- [ ] Upload creates source with status "pending"
- [ ] Upload saves file to S3/storage
- [ ] Upload computes and stores file hash
- [ ] Upload stores original filename
- [ ] Upload stores file size in bytes
- [ ] Duplicate file (same hash) within cooldown rejected
- [ ] Upload increments ingestion count
- [ ] Upload validates user owns chatbot
- [ ] OCR disabled for Free plan (image text not extracted)
- [ ] OCR enabled for Pro+ plans
- [ ] Unsupported file type returns 400

### 4.4 Text Source Creation
- [ ] Add text source with valid text succeeds
- [ ] Add text source with empty text returns 400
- [ ] Text source respects ingestion limits
- [ ] Text source creates with status "pending"
- [ ] Text source enqueues for processing

### 4.5 Source Processing
- [ ] Source status changes from "pending" to "processing"
- [ ] Source status changes to "completed" on success
- [ ] Source status changes to "failed" on error
- [ ] Failed source includes error_message
- [ ] URL source fetches content correctly
- [ ] URL source handles 404 errors gracefully
- [ ] URL source handles timeout errors
- [ ] (PLANNED) URL source respects robots.txt
- [ ] PDF source extracts text correctly (with fitz tag)
- [ ] PDF source runs OCR when enabled
- [ ] Text chunks created and stored
- [ ] Embeddings generated for chunks
- [ ] Embeddings stored in Qdrant vector DB
- [ ] Embedding token usage tracked
- [ ] Embedding token limit enforced
- [ ] Processing errors logged correctly

### 4.6 Source Retrieval
- [ ] GET sources for chatbot returns all sources
- [ ] GET source by ID returns single source
- [ ] Source includes status, type, URL/filename
- [ ] Source includes chunk_count
- [ ] Source includes size_bytes for files
- [ ] Source includes error_message if failed
- [ ] Source includes created_at, updated_at
- [ ] Source includes last_refreshed_at if applicable
- [ ] User can only access sources for their chatbots
- [ ] Invalid source ID returns 404

### 4.7 Source Refresh
- [ ] Refresh URL source succeeds
- [ ] Refresh enforces monthly refresh limit
- [ ] Refresh only works for URL sources (not files)
- [ ] Refresh fails for Free plan (403)
- [ ] Refresh updates last_refreshed_at timestamp
- [ ] Refresh re-fetches URL content
- [ ] Refresh updates embeddings if content changed
- [ ] Refresh does not increment ingestion count (same source)
- [ ] Refresh fails if source already processing (409)
- [ ] Refresh decrements monthly refresh allowance

### 4.8 Source Auto-Refresh
- [ ] Auto-refresh scheduled based on refresh_frequency
- [ ] Auto-refresh runs daily/weekly/monthly as configured
- [ ] Auto-refresh updates next_refresh_at timestamp
- [ ] Auto-refresh only processes chatbots with refresh_policy "auto"
- [ ] Auto-refresh respects monthly refresh limit
- [ ] Auto-refresh stops when limit exceeded
- [ ] Auto-refresh logs success/failure
- [ ] Auto-refresh updates source last_refreshed_at

### 4.9 Source Deletion
- [ ] Delete source succeeds for owner
- [ ] Delete source removes from database
- [ ] Delete source removes embeddings from Qdrant
- [ ] Delete source removes file from S3/storage (if applicable)
- [ ] Delete source updates storage usage
- [ ] Deleted source no longer used in RAG
- [ ] User cannot delete another user's source (403)
- [ ] Invalid source ID returns 404

### 4.10 Discovery Mode (Sub-page Crawling)
- [ ] Discovery mode "disabled" does not crawl sub-pages
- [ ] Discovery mode "auto" automatically crawls sub-pages
- [ ] Discovery mode "pending" requires manual approval
- [ ] Discovered URLs added to pending_urls table
- [ ] Pending URLs displayed in UI for approval
- [ ] Approve pending URL creates new source
- [ ] Reject pending URL removes from pending_urls
- [ ] Discovery respects max_pages_per_crawl limit
- [ ] Discovery limited to same domain
- [ ] Discovery respects include_paths filters
- [ ] Discovery respects exclude_paths filters
- [ ] Discovery marks sources as is_discovered
- [ ] Free plan cannot enable discovery (403)

### 4.11 Advanced Scraping Filters
- [ ] Include_paths array filters URLs correctly
- [ ] Exclude_paths array filters URLs correctly
- [ ] Selector_whitelist extracts specific CSS elements
- [ ] Empty filters allow all content
- [ ] Invalid CSS selectors handled gracefully
- [ ] Filters apply to both manual and discovered URLs

---

## 5. Chat Functionality

### 5.1 Chat Message Flow
- [ ] Send message to chatbot succeeds
- [ ] Chat returns AI response
- [ ] Chat returns tokens_used count
- [ ] Chat returns sources_used array
- [ ] Chat creates conversation record
- [ ] Chat creates user message record
- [ ] Chat creates assistant message record
- [ ] Chat increments monthly token usage
- [ ] Chat enforces monthly token limit
- [ ] Exceeding token limit returns 429 with upgrade message
- [ ] Chat uses correct model from chatbot config
- [ ] Chat uses correct temperature
- [ ] Chat uses correct max_tokens
- [ ] Chat uses system_prompt from chatbot

### 5.2 RAG (Retrieval Augmented Generation)
- [ ] Chat queries Qdrant for relevant chunks
- [ ] RAG uses correct Top-K from plan config
- [ ] RAG respects max_context_tokens from plan config
- [ ] RAG includes relevant sources in response
- [ ] RAG sources_used array includes chunk_index and source_type
- [ ] RAG confidence_tier (high/medium/low) included in response
- [ ] Low confidence triggers fallback message
- [ ] Medium confidence shows warning (if enabled)
- [ ] High confidence returns answer without warning
- [ ] No relevant sources returns "no info found" message
- [ ] RAG respects topic restrictions
- [ ] Blocked topic returns blocked_message

### 5.3 Conversation Management
- [ ] Each chat session creates unique conversation_id
- [ ] Messages linked to conversation_id
- [ ] Conversation includes chatbot_id
- [ ] Conversation includes first and last message timestamps
- [ ] Conversation tracks total messages count
- [ ] GET conversations for chatbot returns list
- [ ] GET conversation by ID returns messages
- [ ] Conversations paginated correctly
- [ ] Deleted chatbot conversations still accessible for analytics

### 5.4 Message Feedback
- [ ] User can submit positive feedback (thumbs up)
- [ ] User can submit negative feedback (thumbs down)
- [ ] Feedback stored with message_id
- [ ] Feedback includes optional text comment
- [ ] Feedback tracked in analytics
- [ ] Feedback can be updated

### 5.5 Public Chat Widget
- [ ] Widget fetches chatbot config from public endpoint
- [ ] Widget displays without authentication
- [ ] Widget renders with correct theme colors
- [ ] Widget displays welcome message
- [ ] Widget displays suggested questions
- [ ] Widget sends messages and receives responses
- [ ] Widget displays sources used
- [ ] Widget position respects config (bottom-right, etc.)
- [ ] Widget displays "Powered by Botla" if not hidden
- [ ] Widget displays custom branding if configured
- [ ] Widget auto-opens if auto-open=1
- [ ] Widget respects secure embed restrictions

### 5.6 Secure Embed
- [ ] Secure embed validates allowed_domains
- [ ] Request from non-allowed domain returns 403
- [ ] Secure embed requires embed_secret
- [ ] Invalid embed_secret returns 401
- [ ] Embed token fetched from embed-token-url
- [ ] Embed token validated on backend
- [ ] (PLANNED) Embed token optionally carries CAPTCHA token
- [ ] (PLANNED) Backend validates CAPTCHA token to prevent abuse
- [ ] Free plan cannot use secure embed
- [ ] Pro+ plans can configure secure embed

---

## 6. Actions (Chatbot Actions)

### 6.1 Action Creation
- [ ] Create HTTP action succeeds
- [ ] Create Zapier action succeeds
- [ ] Create Built-in action succeeds
- [ ] Action includes name, description, type
- [ ] Action includes config (URL, headers, auth for HTTP)
- [ ] Action includes parameters (JSON schema)
- [ ] Action enabled/disabled toggle works
- [ ] Action validates user owns chatbot
- [ ] Action created_at, updated_at set correctly

### 6.2 Action Configuration
- [ ] HTTP action supports GET, POST, PUT, DELETE methods
- [ ] HTTP action supports custom headers
- [ ] HTTP action supports Bearer auth
- [ ] HTTP action supports API key auth
- [ ] HTTP action supports no auth
- [ ] Zapier action includes webhook URL
- [ ] Built-in action has predefined parameters

### 6.3 Action Execution
- [ ] Action triggered during chat when conditions met
- [ ] Action receives parameters from AI model
- [ ] HTTP action sends request to configured URL
- [ ] HTTP action handles response correctly
- [ ] HTTP action handles timeouts gracefully
- [ ] HTTP action handles errors gracefully
- [ ] Zapier action sends webhook payload
- [ ] Built-in action executes internal logic
- [ ] Action execution logged for debugging
- [ ] Failed action does not break chat flow

### 6.4 Action Management
- [ ] GET actions for chatbot returns list
- [ ] GET action by ID returns details
- [ ] Update action succeeds
- [ ] Delete action succeeds
- [ ] Disabled action not triggered
- [ ] Test action endpoint validates config
- [ ] User can only manage actions for their chatbots

### 6.5 Built-in Tools Integration
- [ ] `list_sources` tool returns all chatbot sources
- [ ] `request_human_handoff` tool triggers handoff flow
- [ ] `request_human_handoff` collects user email
- [ ] `request_human_handoff` creates handoff request record
- [ ] Built-in tools included in chat completion function calls

### 6.6 Action Dispatch
- [ ] AI model selects appropriate action based on user query
- [ ] AI model extracts parameters from query
- [ ] Action dispatch validates parameters against schema
- [ ] Action dispatch returns result to user
- [ ] Multiple actions can be defined per chatbot
- [ ] Actions prioritized by order or relevance

---

## 7. Analytics

### 7.1 Chatbot Analytics Overview
- [ ] GET analytics overview returns total messages
- [ ] GET analytics overview returns total conversations
- [ ] GET analytics overview returns total tokens used
- [ ] GET analytics overview returns positive feedback count
- [ ] GET analytics overview returns negative feedback count
- [ ] GET analytics overview returns feedback_rate (%) as satisfaction
- [ ] GET analytics overview returns handoff count
- [ ] Overview aggregates data for the last 30 days
- [ ] Overview is scoped to a specific chatbot (chatbot_id)

### 7.2 Chatbot Analytics Trends
- [ ] GET trends returns daily message counts
- [ ] GET trends returns daily conversation counts
- [ ] GET trends returns daily token usage
- [ ] GET trends returns daily positive/negative feedback counts
- [ ] GET trends returns daily handoff counts
- [ ] GET trends returns avg_response_time_ms per day
- [ ] Trends aggregated by day
- [ ] Trends include time series data for charts
- [ ] Trends limited by days query parameter

### 7.3 Source Usage Analytics
- [ ] GET source usage returns usage per source
- [ ] Source usage includes times source cited
- [ ] Source usage includes average relevance score per source
- [ ] Source usage includes positive/negative feedback counts per source
- [ ] Source usage includes last_used timestamp
- [ ] (PLANNED) Source usage includes sources never used
- [ ] Source usage helps identify valuable content
- [ ] Source usage filtered by chatbot and days query parameter

### 7.4 Advanced Analytics
- [ ] Message records store confidence_score for each response
- [ ] Analytics track handoff events
- [ ] (PLANNED) Analytics track action executions
- [ ] (PLANNED) Analytics track failed messages
- [ ] (PLANNED) Analytics track peak usage times
- [ ] (PLANNED) Analytics exportable to CSV/JSON

---

## 8. Organization & Workspace Management (Multi-Tenancy)

### 8.1 Organization Creation
- [ ] Create organization with name and slug succeeds
- [ ] Organization slug must be unique
- [ ] Duplicate slug returns 409 conflict
- [ ] Creator assigned as owner role
- [ ] Organization assigned default plan (agency_starter)
- [ ] Organization created_at, updated_at set correctly
- [ ] User automatically member of created organization
- [ ] Membership created with owner role

### 8.2 Organization Retrieval
- [ ] GET user organizations returns all organizations user belongs to
- [ ] Organization includes user's role (owner, admin, member)
- [ ] GET organization by ID returns details
- [ ] User can only access organizations they are member of
- [ ] Non-member returns 403

### 8.3 Organization Update
- [ ] Update organization name succeeds (owner/admin only)
- [ ] Update organization slug succeeds
- [ ] Update slug to existing slug returns 409
- [ ] Member role cannot update organization (403)
- [ ] Updated_at timestamp updates correctly

### 8.4 Organization Deletion
- [ ] Delete organization succeeds for owner
- [ ] Cannot delete last organization of user
- [ ] Delete organization cascades to workspaces
- [ ] Delete organization cascades to memberships
- [ ] Delete organization cascades to chatbots
- [ ] Admin/member cannot delete organization (403)

### 8.5 Organization Members
- [ ] GET members returns all organization members
- [ ] Members include user details (email, name)
- [ ] Members include role (owner, admin, member)
- [ ] Response includes caller's role for RBAC
- [ ] Add member by email succeeds (owner/admin only)
- [ ] Add member creates membership record
- [ ] Add member validates email exists in system
- [ ] Add non-existent email returns 404
- [ ] Add member with role succeeds
- [ ] Remove member succeeds (owner/admin only)
- [ ] Cannot remove last owner
- [ ] Cannot remove self as last owner
- [ ] Member role cannot add/remove members (403)

### 8.6 Organization Member Roles
- [ ] Update member role to admin succeeds (owner only)
- [ ] Update member role to member succeeds (owner/admin)
- [ ] Owner can promote member to admin
- [ ] Owner can demote admin to member
- [ ] Admin cannot promote to owner
- [ ] Admin cannot demote owner
- [ ] Member cannot change roles (403)
- [ ] Only owners can manage owner roles
- [ ] Invalid role returns 400

### 8.7 Workspace Creation
- [ ] Create workspace within organization succeeds
- [ ] Workspace requires name and slug
- [ ] Workspace slug unique within organization
- [ ] Duplicate workspace slug returns 409
- [ ] Workspace includes optional client_name
- [ ] Workspace belongs to organization_id
- [ ] Member role can create workspace
- [ ] Workspace created_at set correctly

### 8.8 Workspace Management
- [ ] GET workspaces for organization returns list
- [ ] Update workspace name and slug succeeds
- [ ] Delete workspace succeeds
- [ ] Cannot delete last workspace in organization
- [ ] Delete workspace cascades to chatbots
- [ ] Chatbots assigned to workspace show in workspace context
- [ ] Workspace settings JSONB field can store custom config

### 8.9 Organization Context & RBAC
- [ ] Owner has full access to all organization resources
- [ ] Admin can manage members, workspaces, chatbots
- [ ] Member can view organization, create chatbots
- [ ] Member cannot delete organization
- [ ] Member cannot manage other members
- [ ] Organization plan limits apply to all members
- [ ] Organization plan shared across all members
- [ ] Chatbot created in organization context belongs to organization
- [ ] User can switch between personal and organization contexts

---

## 9. Security & Validation

### 9.1 Input Validation
- [ ] All user inputs sanitized to prevent XSS
- [ ] SQL injection attempts blocked (parameterized queries)
- [ ] File upload MIME type validated
- [ ] URL format validated before scraping
- [ ] Email format validated on registration
- [ ] JSON payload validation for all POST/PUT requests
- [ ] Invalid JSON returns 400 with clear message
- [ ] Oversized request bodies rejected
- [ ] Special characters in names/slugs handled correctly

### 9.2 Authentication & Authorization
- [ ] All protected endpoints require valid JWT
- [ ] Expired JWT returns 401
- [ ] Invalid JWT signature returns 401
- [ ] Missing JWT returns 401
- [ ] User can only access their own resources
- [ ] Cross-user access attempts return 403
- [ ] Organization members can access organization resources
- [ ] Non-members cannot access organization resources
- [ ] Role-based permissions enforced correctly
- [ ] API bypasses validated on backend

### 9.3 Rate Limiting
- [ ] Login endpoint rate limited per IP
- [ ] Registration endpoint rate limited per IP
- [ ] Chat endpoint rate limited per user
- [ ] Source creation rate limited per user
- [ ] Rate limit returns 429 Too Many Requests
- [ ] Rate limit headers included in response
- [ ] Rate limit resets after time window

### 9.4 CORS Configuration
- [ ] CORS allows configured frontend origins
- [ ] CORS blocks unauthorized origins
- [ ] Preflight OPTIONS requests handled correctly
- [ ] CORS headers include allowed methods
- [ ] CORS headers include allowed headers

### 9.5 Data Privacy
- [ ] User passwords never logged
- [ ] JWT secrets never exposed
- [ ] API keys never returned in responses
- [ ] Embed secrets never exposed publicly
- [ ] Deleted data removed from database
- [ ] User data isolated per user/organization
- [ ] PII redacted in logs

---

## 10. Database & Migrations

### 10.1 Migration Integrity
- [ ] All migrations run successfully in order
- [ ] Rollback (down) migrations work correctly
- [ ] Migration version tracked in database
- [ ] No orphaned tables or columns
- [ ] Foreign key constraints enforced
- [ ] Cascade deletes configured correctly
- [ ] Indexes created for performance

### 10.2 Data Integrity
- [ ] Required fields have NOT NULL constraints
- [ ] Unique constraints prevent duplicates
- [ ] Default values set correctly
- [ ] Timestamps (created_at, updated_at) auto-update
- [ ] Soft delete (deleted_at) respects queries
- [ ] JSONB fields store valid JSON
- [ ] UUID generation works correctly

### 10.3 Database Performance
- [ ] Queries use indexes where appropriate
- [ ] N+1 query problems avoided
- [ ] Pagination implemented for large datasets
- [ ] Slow queries logged and optimized
- [ ] Connection pooling configured
- [ ] Database transactions used for multi-step operations

---

## 11. External Integrations

### 11.1 OpenRouter Integration (Primary AI Provider)
- [ ] OpenRouter API key configured correctly
- [ ] Chat completions endpoint called successfully
- [ ] Model selection works (gpt-4o-mini, gpt-4o, claude-3-5-sonnet)
- [ ] Temperature and max_tokens parameters respected
- [ ] System prompt injected correctly
- [ ] User messages formatted correctly
- [ ] Token usage tracked from API response
- [ ] API errors handled gracefully
- [ ] Timeout errors handled gracefully
- [ ] Rate limit errors from OpenRouter handled
- [ ] Tool calls (function calling) work correctly

### 11.2 Qdrant Integration
- [ ] Qdrant URL configured correctly
- [ ] Collection created for chatbot
- [ ] Embeddings upserted to Qdrant
- [ ] Vector search queries return results
- [ ] Top-K parameter respected
- [ ] Filters applied correctly
- [ ] Qdrant errors handled gracefully
- [ ] Collection deleted when chatbot deleted

### 11.3 Storage (S3/Local)
- [ ] S3 credentials configured correctly
- [ ] File upload to S3 succeeds
- [ ] File download from S3 succeeds
- [ ] File deletion from S3 succeeds
- [ ] Storage quota tracked correctly
- [ ] Local storage fallback works if S3 unavailable
- [ ] Pre-signed URLs work for private files

### 11.4 Redis & Scraper Cache Integration
- [ ] Redis cache connection established for scraper
- [ ] Cache entries expire correctly based on TTL
- [ ] Redis or cache errors do not crash application
- [ ] Scraper cache falls back to in-memory if REDIS_URL is unset or Redis unavailable

---

## 12. Frontend UI Testing

### 12.1 Authentication Pages
- [ ] Login page renders correctly
- [ ] Registration page renders correctly
- [ ] Login form validation works
- [ ] Registration form validation works
- [ ] Error messages displayed for invalid inputs
- [ ] Success redirects to dashboard
- [ ] Logout button works
- [ ] Token refresh works silently

### 12.2 Dashboard Page
- [ ] Dashboard displays chatbots list
- [ ] Dashboard displays usage statistics
- [ ] Dashboard displays plan information
- [ ] Create chatbot button works
- [ ] Empty state displayed when no chatbots
- [ ] Loading states displayed during fetch
- [ ] Error states displayed on failure

### 12.3 Chatbot Detail Page
- [ ] Chatbot details displayed correctly
- [ ] Edit chatbot form works
- [ ] Save changes updates chatbot
- [ ] Validation errors displayed
- [ ] Theme color picker works
- [ ] Model selector respects plan limits
- [ ] Advanced settings tabs work
- [ ] Sources tab displays sources list
- [ ] Actions tab displays actions list
- [ ] Analytics tab displays charts

### 12.4 Plan Page
- [ ] Current plan displayed correctly
- [ ] Plan features listed
- [ ] Usage statistics displayed
- [ ] Ingestion count displayed
- [ ] Embedding token count displayed
- [ ] Storage usage displayed
- [ ] Monthly token usage displayed
- [ ] Branding capabilities displayed
- [ ] Upgrade prompt shown for disabled features
- [ ] Progress bars show usage percentages

### 12.5 Source Management UI
- [ ] Add URL form works
- [ ] Upload file form works
- [ ] Add text form works
- [ ] Sources list displays all sources
- [ ] Source status badges display correctly (pending, processing, completed, failed)
- [ ] Refresh button works for URL sources
- [ ] Delete source button works
- [ ] Pending URLs list displayed (discovery mode)
- [ ] Approve/reject pending URLs works
- [ ] Error messages displayed for failed sources

### 12.6 Organization Pages
- [ ] Organization list displayed
- [ ] Create organization form works
- [ ] Organization detail page shows info
- [ ] Members list displayed
- [ ] Add member form works
- [ ] Remove member button works
- [ ] Update member role works
- [ ] Workspace list displayed
- [ ] Create workspace form works
- [ ] Switch context between personal/org works
- [ ] RBAC permissions reflected in UI

### 12.7 Analytics Pages
- [ ] Overview stats displayed
- [ ] Charts render correctly (Recharts)
- [ ] Date range filter works
- [ ] Trends charts show time series data
- [ ] Source usage stats displayed
- [ ] Feedback counts displayed
- [ ] Export analytics button works

### 12.8 Widget Embed
- [ ] Embed code snippet displayed
- [ ] Copy embed code button works
- [ ] Embed preview works
- [ ] Widget customization preview updates in real-time
- [ ] Secure embed settings displayed (Pro+ only)
- [ ] Allowed domains input works
- [ ] Embed secret generated and displayed
- [ ] Auto-open toggle works
- [ ] Position selector works

---

## 13. Chat Widget Testing

### 13.1 Widget Initialization
- [ ] Widget loads from CDN/local build
- [ ] Widget injects into DOM correctly
- [ ] Widget shadow DOM isolates styles
- [ ] Widget parameters parsed from script tag
- [ ] Widget fetches chatbot config on init
- [ ] Widget displays error if chatbot not found

### 13.2 Widget UI
- [ ] Chat bubble displays in correct position
- [ ] Chat window opens on bubble click
- [ ] Chat window closes on X button click
- [ ] Welcome message displayed
- [ ] Suggested questions displayed
- [ ] Input field functional
- [ ] Send button works
- [ ] Enter key sends message
- [ ] Message history scrolls correctly
- [ ] User and bot message styles applied
- [ ] Theme colors from config applied
- [ ] Bot icon/name displayed

### 13.3 Widget Chat Flow
- [ ] Send message to chatbot works
- [ ] Bot response displayed
- [ ] Sources used displayed below message
- [ ] Loading indicator shown during response
- [ ] Error message shown on failure
- [ ] Retry button works on error
- [ ] Multiple messages in sequence work
- [ ] Conversation persists during session
- [ ] Suggested questions clickable

### 13.4 Widget Branding
- [ ] "Powered by Botla" displayed for Free plan
- [ ] "Powered by Botla" hidden for Pro+ when hide_branding=true
- [ ] Custom branding (logo, text, link) displayed for Ultra plan
- [ ] Custom branding link clickable
- [ ] Custom logo image renders correctly

### 13.5 Widget Secure Embed
- [ ] Widget validates allowed domain
- [ ] Widget blocked on non-allowed domain
- [ ] Embed token fetched from token URL
- [ ] CAPTCHA token included in embed token request
- [ ] Embed token validated on chat request
- [ ] Invalid embed token returns 401
- [ ] CAPTCHA challenge prevents abuse

### 13.6 Widget Auto-Open
- [ ] auto-open=1 opens widget automatically
- [ ] auto-open=0 keeps widget closed initially
- [ ] Widget state persists across page refreshes (localStorage)

### 13.7 Widget Accessibility
- [ ] Widget keyboard navigable
- [ ] Widget screen reader compatible
- [ ] Focus states visible
- [ ] ARIA labels present
- [ ] High contrast mode works

---

## 14. Edge Cases & Error Handling

### 14.1 Network & Connectivity
- [ ] API timeouts handled gracefully
- [ ] Network offline displays error
- [ ] Retry logic works for transient failures
- [ ] Concurrent requests handled correctly
- [ ] Race conditions avoided

### 14.2 Data Edge Cases
- [ ] Empty strings handled correctly
- [ ] Null values handled correctly
- [ ] Very long strings (max length validation)
- [ ] Special characters in all text fields
- [ ] Unicode characters handled correctly
- [ ] Large file uploads (max size enforced)
- [ ] Empty arrays handled correctly
- [ ] Empty objects handled correctly

### 14.3 Quota Edge Cases
- [ ] User at exact limit can add one more until enforced
- [ ] User over limit cannot add more
- [ ] Quota resets at month boundary correctly
- [ ] Quota checked before and during operation
- [ ] Concurrent quota checks avoid race conditions
- [ ] Quota tracking survives server restarts

### 14.4 State Edge Cases
- [ ] Deleted chatbot cannot be updated
- [ ] Deleted source cannot be refreshed
- [ ] Processing source cannot be deleted
- [ ] Completed source can be refreshed
- [ ] Failed source can be retried

### 14.5 Performance Edge Cases
- [ ] Large number of sources (100+) loads correctly
- [ ] Large number of chatbots (50+) loads correctly
- [ ] Large number of messages (1000+) paginates correctly
- [ ] Large embeddings (8k tokens) process correctly
- [ ] Concurrent users do not interfere with each other

---

## 15. Localization & Internationalization

### 15.1 Language Support
- [ ] Chatbot language_code set correctly
- [ ] Localized error messages work (if implemented)
- [ ] Localized fallback messages work
- [ ] Localized UI elements work (if implemented)
- [ ] Date/time formats respect locale

---

## 16. Monitoring & Logging

### 16.1 Application Logging
- [ ] Error logs include stack traces
- [ ] Info logs track important events
- [ ] Logs include request IDs for tracing
- [ ] Logs do not contain sensitive data
- [ ] Log levels configurable (debug, info, warn, error)

### 16.2 Health Checks
- [ ] GET /health returns 200 OK
- [ ] Health check validates database connection
- [ ] Health check validates Redis connection
- [ ] Health check validates external API availability

---

## 17. Deployment & Environment

### 17.1 Environment Configuration
- [ ] .env.example includes all required variables
- [ ] Environment variables loaded correctly
- [ ] Missing required env vars fail startup with clear error
- [ ] Development and production configs separate
- [ ] Secrets not committed to version control

### 17.2 Docker & Containers
- [ ] Docker Compose starts all services
- [ ] Backend container builds successfully
- [ ] Frontend container builds successfully
- [ ] Widget container builds successfully
- [ ] PostgreSQL container initializes correctly
- [ ] Redis container starts correctly
- [ ] Volumes persist data correctly
- [ ] Networks allow service communication

### 17.3 Database Setup
- [ ] make migrate-up runs successfully
- [ ] make migrate-down rolls back correctly
- [ ] Seed data script works (if applicable)
- [ ] Database connection string configured correctly

### 17.4 Build & CI/CD
- [ ] make build compiles backend successfully
- [ ] make test runs all backend tests
- [ ] make ci runs all checks (vet, lint, test)
- [ ] Coverage gate enforces 90% minimum
- [ ] npm run build builds frontend successfully
- [ ] npm run ci runs all frontend checks
- [ ] E2E tests run in CI environment
- [ ] Linters pass without warnings

---

## 18. Performance Testing

### 18.1 Load Testing
- [ ] 100 concurrent users supported
- [ ] 1000 requests/minute handled
- [ ] Database queries under 100ms (indexed)
- [ ] API response times under 500ms (non-chat)
- [ ] Chat response times under 5s (with AI)
- [ ] Vector search queries under 200ms

### 18.2 Scalability
- [ ] Horizontal scaling works (multiple backend instances)
- [ ] Database connection pooling configured
- [ ] Redis caching reduces database load
- [ ] CDN serves static assets
- [ ] Large datasets paginated correctly

---

## 19. Disaster Recovery & Backups

### 19.1 Data Backups
- [ ] Database backed up regularly
- [ ] Backup restoration tested
- [ ] S3 files backed up or versioned
- [ ] Vector DB (Qdrant) backed up

### 19.2 Failure Scenarios
- [ ] Database failure handled gracefully
- [ ] Redis failure does not crash app
- [ ] OpenAI API failure returns error to user
- [ ] Qdrant failure handled gracefully
- [ ] S3 failure handled gracefully

---

## 20. Compliance & Legal

### 20.1 Data Protection
- [ ] GDPR compliance (if applicable)
- [ ] User data deletion on request
- [ ] Data retention policies enforced
- [ ] Privacy policy accessible
- [ ] Terms of service accessible

### 20.2 Security Audits
- [ ] Vulnerability scanning run
- [ ] Dependencies updated to latest secure versions
- [ ] Known CVEs addressed
- [ ] Security headers configured (CSP, HSTS, etc.)

---

## Final Pre-Production Checklist

- [ ] All critical bugs fixed
- [ ] All security vulnerabilities addressed
- [ ] All tests passing (unit, integration, E2E)
- [ ] Code coverage >= 90%
- [ ] Performance benchmarks met
- [ ] Error monitoring configured (e.g., Sentry)
- [ ] Analytics tracking configured (if applicable)
- [ ] Documentation complete and up-to-date
- [ ] Staging environment tested end-to-end
- [ ] Production environment configured
- [ ] Backup and recovery plan in place
- [ ] Monitoring and alerting configured
- [ ] Rollback plan documented
- [ ] Team trained on production procedures

---

## Notes

This checklist represents the comprehensive testing requirements for the Botla platform. Each checkbox should be verified manually or via automated tests before production deployment. Priority should be given to security-critical items marked with "Security Critical" tags.

### Features NOT Currently Implemented
The following features do **not exist** in the codebase and should **not** be tested:
- Password reset / forgot password flow
- Email verification / confirmation
- Streaming/SSE chat responses
- Request idempotency keys
- API versioning headers
- robots.txt respect in scraper (marked as PLANNED)
- CSV/JSON analytics export (marked as PLANNED)
- CAPTCHA validation for secure embed (marked as PLANNED)

### Recent Changes
- `system_prompt` has been replaced with `custom_instruction` for user-editable instructions
- OpenRouter is now the unified AI provider (no separate OpenAI/Claude integrations)
- Plan limits `max_files_total` does not exist in schema - only `max_files_per_bot` is enforced

For ongoing maintenance, re-run relevant sections of this checklist after each major feature addition or update.

