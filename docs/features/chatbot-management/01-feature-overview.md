> Extracted from: chatbot-management.md — Section 1

# Chatbot Management — Feature Testing Document

# 1. Feature Overview

Chatbot Management covers the full lifecycle of AI chatbot instances on the botla.app platform. Users can create, configure, update, and delete chatbots. Each chatbot has **61 fields** organized into 9 categories: Identity, Personality, Appearance, Security, Discovery, Refresh, Guardrails, Handoff, Branding, and Suggestions. Updates are granular via domain-specific `PUT` endpoints. The system enforces multi-tenancy (workspace isolation), plan-based feature gating, and input validation.
