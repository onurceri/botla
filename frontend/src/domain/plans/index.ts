/**
 * Plan types and utilities for the frontend domain layer.
 * Centralizes plan-related business logic.
 * 
 * NOTE: These limits are client-side fallbacks. 
 * The authoritative source of truth is the backend database (plans table).
 * These should be kept in sync with seed migrations (e.g., 000035_fix_plan_features.up.sql).
 */

export type PlanCode = 'free' | 'pro' | 'ultra';

// UI display variants (used in badges, etc.)
export type PlanTier = 'free' | 'pro' | 'ultra';

export interface PlanFeatures {
  customBranding: boolean;
  analytics: boolean; // Not explicitly in migration, assuming true for paid
  handoff: boolean; // can_use_escalate_fallback
  apiAccess: boolean; // Not explicitly in migration, assuming true for paid
  refresh: boolean;
  secureEmbed: boolean;
  dynamicScraping: boolean;
  ocr: boolean;
  smartFallback: boolean;
  topicManagement: boolean;
}

export interface PlanLimits {
  maxChatbots: number;
  maxSources: number; // Not direct match, using approximation or omitting
  maxURLs: number;    // max_urls_per_bot
  maxPDFs: number;    // max_files_per_bot
  maxFileSizeMB: number;
  maxMonthlyTokens: number;
  maxTextLength: number;
  maxPagesPerCrawl: number;
  maxMonthlyRefreshes: number;
  features: PlanFeatures;
}

/**
 * Default plan limits configuration.
 * These serve as fallbacks when server-side limits are not available.
 * Values derived from migration 000035_fix_plan_features.up.sql
 */
export const PLAN_LIMITS: Record<PlanCode, PlanLimits> = {
  free: {
    maxChatbots: 1,
    maxSources: 5, // Approximate based on files+urls
    maxURLs: 1,
    maxPDFs: 1,
    maxFileSizeMB: 5,
    maxMonthlyTokens: 100000,
    maxTextLength: 400000,
    maxPagesPerCrawl: 5,
    maxMonthlyRefreshes: 0,
    features: {
      customBranding: false,
      analytics: false,
      handoff: false,
      apiAccess: false,
      refresh: false,
      secureEmbed: false,
      dynamicScraping: false,
      ocr: false,
      smartFallback: true, // Enabled in free per migration
      topicManagement: false,
    },
  },
  pro: {
    maxChatbots: 10,
    maxSources: 50,
    maxURLs: 10,
    maxPDFs: 20,
    maxFileSizeMB: 20,
    maxMonthlyTokens: 1000000,
    maxTextLength: 400000,
    maxPagesPerCrawl: 50,
    maxMonthlyRefreshes: 5,
    features: {
      customBranding: false, // can_custom_branding is false, can_hide_branding is true
      analytics: true,
      handoff: false,
      apiAccess: true,
      refresh: true,
      secureEmbed: true,
      dynamicScraping: true,
      ocr: true,
      smartFallback: true,
      topicManagement: true,
    },
  },
  ultra: {
    maxChatbots: 100,
    maxSources: 200,
    maxURLs: 50,
    maxPDFs: 100,
    maxFileSizeMB: 50,
    maxMonthlyTokens: 5000000,
    maxTextLength: 400000,
    maxPagesPerCrawl: 200,
    maxMonthlyRefreshes: 100,
    features: {
      customBranding: true,
      analytics: true,
      handoff: true,
      apiAccess: true,
      refresh: true,
      secureEmbed: true,
      dynamicScraping: true,
      ocr: true,
      smartFallback: true,
      topicManagement: true,
    },
  },
};

/**
 * Plan display configuration for UI components
 */
export const PLAN_DISPLAY: Record<PlanCode, { label: string; tier: PlanTier }> = {
  free: { label: 'FREE', tier: 'free' },
  pro: { label: 'PRO', tier: 'pro' },
  ultra: { label: 'ULTRA', tier: 'ultra' },
};

/**
 * Get plan limits for a given plan code.
 * Falls back to 'free' for unknown plan codes.
 */
export function getPlanLimits(plan: string): PlanLimits {
  const planCode = normalizePlanCode(plan);
  return PLAN_LIMITS[planCode];
}

/**
 * Normalize a plan string to a valid PlanCode.
 * Handles UUIDs, variations, and unknown values.
 */
export function normalizePlanCode(planId: string | null | undefined): PlanCode {
  if (!planId) return 'free';
  const lower = planId.toLowerCase();
  
  if (lower === 'ultra' || lower.includes('ultra')) return 'ultra';
  if (lower === 'pro' || lower.includes('pro')) return 'pro';
  if (lower === 'free' || lower.includes('free')) return 'free';
  
  // UUID or unknown - default to free
  return 'free';
}

/**
 * Convert PlanCode to display PlanTier for UI components.
 */
export function planCodeToTier(planCode: string): PlanTier {
  const normalized = normalizePlanCode(planCode);
  return PLAN_DISPLAY[normalized].tier;
}

/**
 * Get display label for a plan.
 */
export function getPlanLabel(planCode: string): string {
  const normalized = normalizePlanCode(planCode);
  return PLAN_DISPLAY[normalized].label;
}

// Re-export limits functions
export * from './limits';
