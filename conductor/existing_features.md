# Existing Features Audit

This document serves as a comprehensive inventory of the features currently implemented in the Botla-co project. It is used to avoid redundancy when planning new tracks.

## 1. Authentication & Multi-Tenancy
- **User Management:** Registration, Login (JWT), Password Hashing, Profile Management.
- **Organization Support:** Multi-tenant architecture allowing users to belong to organizations.
- **Security:** Refresh tokens, session management.

## 2. Chatbot Management & Configuration
- **Core CRUD:** Create, Read, Update, Delete chatbots.
- **Customization:**
    - **Branding:** Custom colors (primary/secondary), launcher icons, logos, widget titles.
    - **Instructions:** System prompts and custom instructions to define persona.
    - **Models:** Selection of underlying AI models (e.g., GPT-4, GPT-3.5).
- **Guardrails:** Configuration to prevent specific topics or behaviors.
- **Actions:** Definition of "tools" or actions the chatbot can take.
    - **Smart Naming:** LLM-powered automatic generation of API-compatible tool names from user-friendly descriptions.
- **Localization:** Backend and frontend support for multiple languages (English, Turkish).

## 3. RAG (Retrieval-Augmented Generation) & Ingestion
- **Data Sources:**
    - **PDF:** Upload and processing of PDF documents.
    - **Web Scraping:** Ingestion via Sitemap or Single URL crawling.
    - **Text:** Direct raw text input.
- **Processing:** Text chunking, embedding generation (Qdrant), and metadata tracking.
- **Search:** Semantic search for relevant context during chat.
- **Suggestions:** Automated generation of "Suggested Questions" from ingested content using LLMs.
- **Management:**
    - **Manual Refresh:** Ability to manually trigger a re-sync for URL sources.
    - **Bulk Ingestion:** Bulk creation of sources via URL lists.
    - **Pending URLs:** Approval workflow for discovered URLs.

## 4. Chat Interface & Widget
- **Embeddable Widget:** A standalone, customizable chat widget for external websites.
- **Real-time Chat:** Streaming responses from the LLM.
- **Message Features:**
    - **Sources:** Display of citations/sources used for answers.
    - **Feedback:** Thumbs up/down feedback on messages.
    - **Suggested Questions:** Display of interactive conversation starters.
- **Playground:** A dashboard-integrated chat interface for testing bots.

## 5. Analytics & Observability
- **Dashboard:** Visualization of key metrics.
- **Metrics Tracked:**
    - Message volume.
    - Token consumption.
    - User interactions.
- **Feedback Analysis:** Tracking of user feedback (thumbs up/down).

## 6. Advanced Features
- **Handoff:** Logic to transfer conversation to a human agent, including email capture.
- **Plan Enforcement:**
    - Tiered limits (Free, Pro, Ultra).
    - Rate limiting.
    - Feature gating (e.g., Secure Embed, remove branding).
- **Secure Embed:** Domain whitelisting and security settings for the widget.
