/**
 * Plan limit checking functions.
 * Provides utilities to check if a user can perform actions based on their plan.
 */

import { PlanLimits, getPlanLimits, type PlanCode } from './index';

/**
 * Result of a limit check
 */
export interface LimitCheckResult {
  allowed: boolean;
  current: number;
  limit: number;
  remaining: number;
  isUnlimited: boolean;
}

/**
 * Check if user can create another chatbot.
 */
export function canCreateChatbot(plan: string, currentCount: number): LimitCheckResult {
  const limits = getPlanLimits(plan);
  return checkLimit(currentCount, limits.maxChatbots);
}

/**
 * Check if user can add another source to a chatbot.
 */
export function canAddSource(plan: string, currentCount: number): LimitCheckResult {
  const limits = getPlanLimits(plan);
  return checkLimit(currentCount, limits.maxSources);
}

/**
 * Check if user can add another URL source.
 */
export function canAddURL(plan: string, currentCount: number): LimitCheckResult {
  const limits = getPlanLimits(plan);
  return checkLimit(currentCount, limits.maxURLs);
}

/**
 * Check if user can add another PDF.
 */
export function canAddPDF(plan: string, currentCount: number): LimitCheckResult {
  const limits = getPlanLimits(plan);
  return checkLimit(currentCount, limits.maxPDFs);
}

/**
 * Check if file size is within plan limits.
 */
export function canUploadFile(plan: string, fileSizeMB: number): LimitCheckResult {
  const limits = getPlanLimits(plan);
  const isUnlimited = limits.maxFileSizeMB === -1;
  return {
    allowed: isUnlimited || fileSizeMB <= limits.maxFileSizeMB,
    current: fileSizeMB,
    limit: limits.maxFileSizeMB,
    remaining: isUnlimited ? Infinity : Math.max(0, limits.maxFileSizeMB - fileSizeMB),
    isUnlimited,
  };
}

/**
 * Check if text length is within plan limits.
 */
export function canAddText(plan: string, textLength: number): LimitCheckResult {
  const limits = getPlanLimits(plan);
  const isUnlimited = limits.maxTextLength === -1;
  return {
    allowed: isUnlimited || textLength <= limits.maxTextLength,
    current: textLength,
    limit: limits.maxTextLength,
    remaining: isUnlimited ? Infinity : Math.max(0, limits.maxTextLength - textLength),
    isUnlimited,
  };
}

/**
 * Check if user can refresh sources (plan feature).
 */
export function canRefreshSource(plan: string): boolean {
  const limits = getPlanLimits(plan);
  return limits.features.refresh;
}

/**
 * Check if user can use handoff feature.
 */
export function canUseHandoff(plan: string): boolean {
  const limits = getPlanLimits(plan);
  return limits.features.handoff;
}

/**
 * Check if user can use custom branding.
 */
export function canCustomizeBranding(plan: string): boolean {
  const limits = getPlanLimits(plan);
  return limits.features.customBranding;
}

/**
 * Check if user has access to analytics.
 */
export function hasAnalyticsAccess(plan: string): boolean {
  const limits = getPlanLimits(plan);
  return limits.features.analytics;
}

/**
 * Check if user has API access.
 */
export function hasApiAccess(plan: string): boolean {
  const limits = getPlanLimits(plan);
  return limits.features.apiAccess;
}

/**
 * Check if user can use OCR features.
 */
export function canUseOCR(plan: string): boolean {
  const limits = getPlanLimits(plan);
  return limits.features.ocr;
}

/**
 * Check if user can use dynamic scraping.
 */
export function canUseDynamicScraping(plan: string): boolean {
  const limits = getPlanLimits(plan);
  return limits.features.dynamicScraping;
}

/**
 * Check if user can use secure embed.
 */
export function canUseSecureEmbed(plan: string): boolean {
  const limits = getPlanLimits(plan);
  return limits.features.secureEmbed;
}

/**
 * Get all feature flags for a plan.
 */
export function getPlanFeatures(plan: string): PlanLimits['features'] {
  const limits = getPlanLimits(plan);
  return limits.features;
}

/**
 * Internal helper to check a numeric limit.
 */
function checkLimit(current: number, limit: number): LimitCheckResult {
  const isUnlimited = limit === -1;
  return {
    allowed: isUnlimited || current < limit,
    current,
    limit,
    remaining: isUnlimited ? Infinity : Math.max(0, limit - current),
    isUnlimited,
  };
}

/**
 * Format remaining count for display.
 * Returns "Unlimited" for unlimited plans, or "X remaining" for limited plans.
 */
export function formatRemaining(result: LimitCheckResult): string {
  if (result.isUnlimited) return 'Sınırsız';
  if (result.remaining === 0) return 'Limit doldu';
  return `${result.remaining} kaldı`;
}

/**
 * Get a descriptive message for limit status.
 */
export function getLimitStatusMessage(result: LimitCheckResult, itemName: string): string {
  if (result.isUnlimited) {
    return `Sınırsız ${itemName} ekleyebilirsiniz`;
  }
  if (!result.allowed) {
    return `${itemName} limitinize ulaştınız (${result.limit})`;
  }
  return `${result.remaining}/${result.limit} ${itemName} ekleyebilirsiniz`;
}
