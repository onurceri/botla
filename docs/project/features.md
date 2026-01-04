# Existing Features Audit (Comprehensive)

This document provides a detailed, documentation-style inventory of the features implemented in botla-app. It serves as the definitive reference for the project's current capabilities.

## 1. Authentication & Multi-Tenancy
- **User Management:**
    - Secure registration and login using JWT-based authentication.
    - Password hashing using Argon2.
    - Profile management (Avatar, Full Name, Preferred Language).
    - Email verification workflow.
- **Organization & Workspace Support:**
    - Multi-tenant architecture allowing users to belong to organizations.
    - Workspace-level resource isolation.
- **Security:**
    - Refresh token rotation with secure hashing and revocation support.
    - Role-based access control (RBAC) at the organization and workspace levels.

## 2. Plan & Entitlement System
The system enforces strict limits and feature gates based on three primary tiers: **Free**, **Pro**, and **Ultra**.

### 2.1. Feature Matrix
| Feature | Free | Pro | Ultra |
| :--- | :--- | :--- | :--- |
| **Max Chatbots** | 1 | 10 | 100 |
| **Default AI Model** | `gpt-4o-mini` | `gpt-4o` | `gpt-4o` |
| **Allowed Models** | `gpt-4o-mini` | `gpt-4o-mini`, `gpt-4o` | `gpt-4o-mini`, `gpt-4o`, `gpt-5` |
| **Monthly Tokens** | 100,000 | 1,000,000 | 5,000,000 |
| **OCR (File Reading)** | No | Yes | Yes |
| **Dynamic Scraping (JS)** | No | Yes | Yes |
| **Guardrails Customization** | No | Yes | Yes |
| **Smart Fallback** | No | Yes | Yes |
| **Escalate to Human** | No | No | Yes |
| **Custom Branding** | No | No | Yes |
| **Max Suggested Questions** | 3 | 6 | 10 |

### 2.2. Ingestion & Storage Limits
| Limit | Free | Pro | Ultra |
| :--- | :--- | :--- | :--- |
| **Max Files per Bot** | 1 | 20 | 100 |
| **Max Files Total** | 5 | 100 | 1,000 |
| **Max File Size (MB)** | 5 | 20 | 50 |
| **Total Storage (MB)** | 10 | 500 | 2,000 |
| **Scraping Max URLs** | 1 | 10 | 50 |
| **Scraping Max Pages/Crawl** | 5 | 50 | 200 |

### 2.3. Rate Limits
| Limit | Free | Pro | Ultra |
| :--- | :--- | :--- | :--- |
| **Global RPM** | 100 | 500 | 2,000 |
| **Chat RPM** | 30 | 100 | 500 |
| **Sources API RPM** | 10 | 30 | 100 |

## 3. Chatbot Configuration & "Intelligence"
- **Identity & Persona:**
    - **Custom Instructions:** User-editable instructions to define bot persona and behavior.
    - **System Prompt Generation:** Dynamically built from name, instructions, and capability summaries.
    - **Localization:** Support for English (`en-US`) and Turkish (`tr-TR`) with localized system messages and fallbacks.
- **Smart Actions (Tools):**
    - **Custom Tool Creation:** Ability to define external API calls or internal functions the bot can execute.
    - **LLM-Powered Naming:** Automatic generation of API-compatible "Tool Names" from user descriptions.
    - **Execution Logging:** Detailed history of every tool execution (request, response, status, duration).
- **Advanced Configuration:**
    - **Path Filters:** Domain-specific include/exclude paths for scrapers.
    - **Selector Whitelist:** Target specific CSS selectors for clean text extraction.
    - **Discovery Modes:** `auto` (auto-ingest sub-pages), `pending` (manual approval workflow), `disabled`.
    - **Refresh Policies:** `manual` or `auto` (Daily, Weekly, Monthly) with refresh tracking and cooldowns.

## 4. RAG (Retrieval-Augmented Generation) Pipeline
- **Ingestion Engine:**
    - **Multi-Format Support:** PDF, Web URLs, Sitemap crawling, and raw text.
    - **OCR Support:** Optical Character Recognition for image-based PDFs (Pro+).
    - **Smart Chunking:** Sentence-aware splitting (512 tokens) with ~15% tail overlap for context preservation.
    - **Metadata Extraction:** Automated extraction of "Capability Summaries" and "Suggested Questions" from every source.
- **Search & Retrieval:**
    - **Tiered Confidence Search:** Categorizes results as **High**, **Medium**, or **Low** based on vector similarity scores.
    - **Confidence Injection:** Injects uncertainty warnings into the LLM prompt for "Medium" confidence results.
    - **1-Level Depth Crawling:** Intelligent discovery limits to prevent infinite scraping loops.
- **Fallback Mechanisms:**
    - **Static Fallback:** Standard "I don't know" message.
    - **Smart Fallback (Pro+):** Uses AI to redirect users based on chatbot "Capabilities" when no direct answer is found.
    - **Escalate Fallback (Ultra):** Directs the user to a human handoff workflow.

## 5. Chat Interface & Widget Experience
- **Interactive Experience:**
    - **Real-time Streaming:** Smooth message delivery via Server-Sent Events (SSE).
    - **Markdown Support:** Full rendering of rich text, lists, and links.
    - **Citation Display:** Citations showing exactly which sources were used for a specific response.
- **Engagement Features:**
    - **Suggested Questions:** A carousel of interactive questions generated from ingested knowledge.
    - **Feedback System:** Thumbs up/down with detailed analytics tracking.
    - **Human Handoff:** In-chat email capture form for transferring queries to support agents.
- **Customization & Branding:**
    - **Full UI Customization:** Colors (Primary, Secondary, Background, Header), rounded corners, and positions.
    - **Branding Removal (Ultra):** Ability to hide "Powered by Botla" branding.
    - **Custom Launcher:** Custom icons and display names for the bot.

## 6. Analytics & Monitoring
- **Performance Metrics:** Tracking of message volume, token consumption (embedding and chat), and user engagement.
- **Feedback Analysis:** Aggregated reporting of positive/negative feedback trends.
- **Conversation Logs:** Full history of all user interactions with search results and tool usage details.
- **Secure Embeds:** Domain whitelisting and secret key validation for widget security.
