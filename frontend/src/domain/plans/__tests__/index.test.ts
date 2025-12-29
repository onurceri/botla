import { describe, it, expect } from 'vitest';
import {
  PLAN_DISPLAY,
  normalizePlanCode,
  planCodeToTier,
  getPlanLabel,
  type PlanCode,
} from '../index';

describe('domain/plans', () => {
  describe('PLAN_DISPLAY', () => {
    it('should define display info for all plan codes', () => {
      const planCodes: PlanCode[] = ['free', 'pro', 'ultra'];
      
      planCodes.forEach((code) => {
        expect(PLAN_DISPLAY[code]).toBeDefined();
        expect(PLAN_DISPLAY[code].label).toBeDefined();
        expect(PLAN_DISPLAY[code].tier).toBeDefined();
      });
    });

    it('should have correct labels', () => {
      expect(PLAN_DISPLAY.free.label).toBe('FREE');
      expect(PLAN_DISPLAY.pro.label).toBe('PRO');
      expect(PLAN_DISPLAY.ultra.label).toBe('ULTRA');
    });
  });

  describe('normalizePlanCode', () => {
    it('should return "free" for null or undefined', () => {
      expect(normalizePlanCode(null)).toBe('free');
      expect(normalizePlanCode(undefined)).toBe('free');
    });

    it('should return "free" for empty string', () => {
      expect(normalizePlanCode('')).toBe('free');
    });

    it('should normalize exact plan codes', () => {
      expect(normalizePlanCode('free')).toBe('free');
      expect(normalizePlanCode('pro')).toBe('pro');
      expect(normalizePlanCode('ultra')).toBe('ultra');
    });

    it('should be case insensitive', () => {
      expect(normalizePlanCode('FREE')).toBe('free');
      expect(normalizePlanCode('Pro')).toBe('pro');
      expect(normalizePlanCode('ULTRA')).toBe('ultra');
    });

    it('should handle plan codes with suffixes', () => {
      expect(normalizePlanCode('pro_monthly')).toBe('pro');
      expect(normalizePlanCode('ultra-annual')).toBe('ultra');
    });

    it('should return "free" for unknown UUIDs', () => {
      expect(normalizePlanCode('550e8400-e29b-41d4-a716-446655440000')).toBe('free');
    });

    it('should map legacy/alternate names', () => {
      expect(normalizePlanCode('enterprise')).toBe('free'); // 'enterprise' without 'business' or 'ultra' falls back to free
    });

    it('should prioritize ultra over other matches', () => {
      expect(normalizePlanCode('ultra-pro')).toBe('ultra');
    });
  });

  describe('planCodeToTier', () => {
    it('should convert plan codes to display tiers', () => {
      expect(planCodeToTier('free')).toBe('free');
      expect(planCodeToTier('pro')).toBe('pro');
      expect(planCodeToTier('ultra')).toBe('ultra');
    });

    it('should handle unknown codes as free tier', () => {
      expect(planCodeToTier('unknown')).toBe('free');
    });
  });

  describe('getPlanLabel', () => {
    it('should return correct labels', () => {
      expect(getPlanLabel('free')).toBe('FREE');
      expect(getPlanLabel('pro')).toBe('PRO');
      expect(getPlanLabel('ultra')).toBe('ULTRA');
    });

    it('should handle mapped codes', () => {
       // Only standard codes are supported now
       expect(getPlanLabel('pro-monthly')).toBe('PRO');
    });

    it('should handle unknown codes', () => {
      expect(getPlanLabel('unknown')).toBe('FREE');
    });
  });
});
