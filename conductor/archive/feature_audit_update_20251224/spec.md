# Specification: Comprehensive Feature Audit & Documentation Update

## 1. Overview
This track involves a deep-dive audit of the entire Botla-co codebase and database schema to produce a fully detailed, "documentation style" update of `conductor/existing_features.md`. The primary goal is to capture every existing feature, configuration, limitation, and behavior with 100% accuracy, ensuring no "small feature" is overlooked. This document will serve as the single source of truth for the system's current capabilities.

## 2. Objectives
- **Complete Inventory:** Identify and document every user-facing feature, configuration option, and system behavior.
- **Verification:** Validate all documented features against the actual Go backend code, React frontend logic, and PostgreSQL database schema/migrations.
- **Detailed Granularity:** Go beyond high-level summaries to include specific values (e.g., exact model names available per plan, default values, character limits).

## 3. Scope of Analysis
The investigation will cover, but is not limited to:
- **Plans & Entitlements:**
    - Exact breakdown of limits (files, storage, message quotas).
    - **Plan Comparison:** Explicitly document the differences between plan tiers (Free, Pro, Ultra) regarding feature access, limits, and model availability.
    - **LLM Model Availability:** A definitive matrix of which AI models (e.g., GPT-3.5-turbo, GPT-4o) are accessible to which plan tiers, derived strictly from backend configuration logic.
- **Chatbot Configuration:**
    - All branding options (colors, icons, localization).
    - Behavior settings (prompts, temperature, guardrails).
    - Action/Tool capabilities and logging details.
- **RAG & Data Ingestion:**
    - Supported file types and strict validation rules.
    - Crawler behaviors (sitemap vs. single URL), depth limits, and refresh schedules.
    - Suggestion generation logic.
- **Widget & Interface:**
    - Embed configuration parameters.
    - Security settings (domain whitelisting, secure mode).
    - End-user features (feedback, sources display, message history).

## 4. Deliverables
- **Updated `conductor/existing_features.md`:** A comprehensive rewrite of the existing file using a "Full Documentation" style. This will include detailed descriptions, definitive lists, and specific constraints for every feature found in the code.

## 5. Out of Scope
- Implementing new features.
- Fixing bugs discovered during the audit (though they should be noted).
- Refactoring code (unless strictly necessary for understanding).
