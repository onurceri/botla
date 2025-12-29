/**
 * Plan types and utilities for the frontend domain layer.
 * 
 * For plan limits and features, use the usePlans() hook:
 * import { usePlans, usePlanByCode } from '@/hooks/queries/usePlans'
 */

export type PlanCode = 'free' | 'pro' | 'ultra';

// UI display variants (used in badges, etc.)
export type PlanTier = 'free' | 'pro' | 'ultra';

/**
 * Plan display configuration for UI components
 */
export const PLAN_DISPLAY: Record<PlanCode, { label: string; tier: PlanTier }> = {
  free: { label: 'FREE', tier: 'free' },
  pro: { label: 'PRO', tier: 'pro' },
  ultra: { label: 'ULTRA', tier: 'ultra' },
};

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
