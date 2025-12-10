# Comprehensive Feature & Plan Documentation

This document details the features available for each plan based on a deep analysis of the Backend (BE) database migrations/models and Frontend (FE) codebase. It also highlights discrepancies where features are implemented in one but not the other, or where security gaps exist.

## 1. Plan Features Overview

### **Free Plan**
*   **Chat:**
    *   **Models:** `gpt-4o-mini`
    *   **Tokens:** 100,000 monthly tokens
    *   **RAG:** Top-K: 3, Context: 2,000 tokens
*   **Files:**
    *   **Limits:** Max 1 file/bot, 5 files total, 5MB max size per file, 10MB total storage.
    *   **OCR:** Disabled (Image/PDF text extraction).
*   **Scraping:**
    *   **Limits:** Max 1 URL/bot.
    *   **Capabilities:** Dynamic (JS) scraping DISABLED. Sub-page crawling (Discovery) DISABLED (Max pages/crawl: 0).
*   **Refresh:** Disabled (Manual & Auto).
*   **Branding:** "Powered by Botla" visible.
*   **Secure Embed:** Disabled (Frontend enforces, Backend does NOT).

### **Pro Plan**
*   **Chat:**
    *   **Models:** `gpt-4o-mini`, `gpt-4o`
    *   **Tokens:** 1,000,000 monthly tokens
    *   **RAG:** Top-K: 5, Context: 4,000 tokens
*   **Files:**
    *   **Limits:** Max 20 files/bot, 100 files total, 20MB max size per file, 500MB total storage.
    *   **OCR:** Enabled.
*   **Scraping:**
    *   **Limits:** Max 10 URLs/bot.
    *   **Capabilities:** Dynamic (JS) scraping ENABLED. Sub-page crawling ENABLED (Max 10 pages/crawl).
*   **Refresh:**
    *   **Allowance:** 5 refreshes/month.
    *   **Auto-Refresh:** Available (Frontend shows option).
*   **Branding:** Option to **Hide** "Powered by Botla".
*   **Secure Embed:** Enabled (Allowed Domains & Secret).

### **Ultra Plan**
*   **Chat:**
    *   **Models:** `gpt-4o-mini`, `gpt-4o`, `claude-3-5-sonnet`
    *   **Tokens:** 5,000,000 monthly tokens
    *   **RAG:** Top-K: 10, Context: 8,000 tokens
*   **Files:**
    *   **Limits:** Max 100 files/bot, 1000 files total, 50MB max size per file, 2,000MB total storage.
    *   **OCR:** Enabled.
*   **Scraping:**
    *   **Limits:** Max 50 URLs/bot.
    *   **Capabilities:** Dynamic (JS) scraping ENABLED. Sub-page crawling ENABLED (Max 100 pages/crawl).
*   **Refresh:**
    *   **Allowance:** 10 refreshes/month.
*   **Branding:** Option to **Hide** branding AND **Custom Branding** (Logo/Link).
*   **Secure Embed:** Enabled.

### **Common Limits (All Plans)**
*   **Ingestion Rate:** Max 50 ingestions/month (Enforced in BE, **Hidden in FE**).
*   **Embedding Tokens:** Max 250,000 embedding tokens/month (Enforced in BE, **Hidden in FE**).
*   **Re-add Cooldown:** 60 minutes between re-adding the same source.

---

## 2. Ungated / Universal Features
*These features are present in the codebase but currently have no plan-based restrictions in either BE or FE.*

*   **Guardrails:** Confidence Threshold, Fallback Messages, Topic Restrictions.
*   **Handoff:** Human Handoff via Email.
*   **Chatbot Actions:** HTTP Request, Zapier, Built-in actions.
*   **Advanced Filtering:** Path Filters (Include/Exclude), CSS Selector Whitelist.

---

## 3. Critical Discrepancies & Gaps

### **A. Security Gaps (Frontend Gating without Backend Enforcement)**
Users on lower plans can bypass these limits by calling the API directly.

1.  **Secure Embed (`secure_embed_enabled`):**
    *   **FE:** Hidden for Free plan.
    *   **BE:** `UpdateChatbot` handler accepts this field without checking the plan.
2.  **Auto Refresh (`refresh_policy`):**
    *   **FE:** UI might hide/disable it based on plan (implied).
    *   **BE:** `UpdateChatbot` handler accepts `auto` policy without checking if `plan.Config.Refresh.Enabled` is true.
3.  **Discovery Mode (`discovery_mode`):**
    *   **FE:** UI disables "Auto"/"Pending" modes for Free plan.
    *   **BE:** `UpdateChatbot` handler accepts any mode without checking `plan.Config.Scraping.MaxPagesPerCrawl`.

### **B. Missing Frontend Information**
1.  **Ingestion & Embedding Limits:**
    *   **BE:** Enforces monthly limits on ingestions and embedding tokens (Migrations 000005/000006).
    *   **FE:** `PlanPage.tsx` does **not** display these counters or limits, leading to potential user confusion if they hit them.
2.  **Branding Config:**
    *   **FE:** `PlanPage.tsx` does not list Branding capabilities in the plan details card (though it is used in `ChatbotDetailPage`).

### **C. Potential Backend Bug (Ultra Plan)**
*(Correction: Upon closer inspection of the full migration file, the `allowed_models` field IS present for Ultra. This was a false alarm based on initial truncated search results.)*

## 4. Recommendations for Testing
1.  **Test API Bypasses:** Try to enable `secure_embed_enabled` or `refresh_policy='auto'` on a Free plan user via `curl`/Postman to confirm the security gap.
2.  **Check Ingestion Limits:** hit the 50 ingestion limit and verify if the FE shows a meaningful error (since it doesn't show the limit usage).
