/**
 * Plan types and utilities for the frontend domain layer.
 * Centralizes plan-related business logic.
 */

export type PlanCode = 'free' | 'starter' | 'pro' | 'enterprise';

// UI display variants (used in badges, etc.)
export type PlanTier = 'free' | 'pro' | 'business' | 'ultra';

export interface PlanFeatures {
  customBranding: boolean;
  analytics: boolean;
  handoff: boolean;
  apiAccess: boolean;
  refresh: boolean;
  secureEmbed: boolean;
  dynamicScraping: boolean;
  ocr: boolean;
  smartFallback: boolean;
  topicManagement: boolean;
}

export interface PlanLimits {
  maxChatbots: number;
  maxSources: number;
  maxURLs: number;
  maxPDFs: number;
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
 */
export const PLAN_LIMITS: Record<PlanCode, PlanLimits> = {
  free: {
    maxChatbots: 1,
    maxSources: 5,
    maxURLs: 3,
    maxPDFs: 2,
    maxFileSizeMB: 5,
    maxMonthlyTokens: 10000,
    maxTextLength: 5000,
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
      smartFallback: false,
      topicManagement: false,
    },
  },
  starter: {
    maxChatbots: 3,
    maxSources: 15,
    maxURLs: 10,
    maxPDFs: 5,
    maxFileSizeMB: 10,
    maxMonthlyTokens: 50000,
    maxTextLength: 15000,
    maxPagesPerCrawl: 20,
    maxMonthlyRefreshes: 5,
    features: {
      customBranding: false,
      analytics: true,
      handoff: false,
      apiAccess: false,
      refresh: true,
      secureEmbed: false,
      dynamicScraping: false,
      ocr: false,
      smartFallback: false,
      topicManagement: false,
    },
  },
  pro: {
    maxChatbots: 10,
    maxSources: 50,
    maxURLs: 30,
    maxPDFs: 20,
    maxFileSizeMB: 25,
    maxMonthlyTokens: 200000,
    maxTextLength: 50000,
    maxPagesPerCrawl: 100,
    maxMonthlyRefreshes: 30,
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
  enterprise: {
    maxChatbots: -1, // unlimited
    maxSources: -1,
    maxURLs: -1,
    maxPDFs: -1,
    maxFileSizeMB: 100,
    maxMonthlyTokens: -1,
    maxTextLength: -1,
    maxPagesPerCrawl: -1,
    maxMonthlyRefreshes: -1,
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
  starter: { label: 'STARTER', tier: 'pro' },
  pro: { label: 'PRO', tier: 'business' },
  enterprise: { label: 'ENTERPRISE', tier: 'ultra' },
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
  
  if (lower === 'enterprise' || lower.includes('enterprise')) return 'enterprise';
  if (lower === 'pro' || lower.includes('pro')) return 'pro';
  if (lower === 'starter' || lower.includes('starter')) return 'starter';
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
