import { describe, it, expect } from 'vitest';
import {
  canCreateChatbot,
  canAddSource,
  canAddURL,
  canAddPDF,
  canUploadFile,
  canAddText,
  canRefreshSource,
  canUseHandoff,
  canCustomizeBranding,
  hasAnalyticsAccess,
  hasApiAccess,
  canUseOCR,
  canUseDynamicScraping,
  canUseSecureEmbed,
  getPlanFeatures,
  formatRemaining,
  getLimitStatusMessage,
  type LimitCheckResult,
} from '../limits';

describe('domain/plans/limits', () => {
  describe('canCreateChatbot', () => {
    it('should allow creating chatbot when under limit', () => {
      const result = canCreateChatbot('free', 0);
      expect(result.allowed).toBe(true);
      expect(result.current).toBe(0);
      expect(result.limit).toBe(1);
      expect(result.remaining).toBe(1);
    });

    it('should not allow creating chatbot when at limit', () => {
      const result = canCreateChatbot('free', 1);
      expect(result.allowed).toBe(false);
      expect(result.remaining).toBe(0);
    });

    it('should handle unlimited plans', () => {
      // Ultra has high limits but strictly speaking not -1 for most things now
      // simulating unlimited behavior for the test utility if it encounters -1
      // but let's test specific ultra limit
      const result = canCreateChatbot('ultra', 100);
      expect(result.allowed).toBe(false); // Ultra limit is 100, if we have 100 we can't create more
      expect(result.limit).toBe(100);
    });
  });

  describe('canAddSource', () => {
    it('should allow adding source when under limit', () => {
      const result = canAddSource('free', 3);
      expect(result.allowed).toBe(true);
      expect(result.remaining).toBe(2);
    });

    it('should not allow adding source when at limit', () => {
      const result = canAddSource('free', 5);
      expect(result.allowed).toBe(false);
    });

    it('should have higher limits for pro plan', () => {
      const freeResult = canAddSource('free', 10);
      const proResult = canAddSource('pro', 10);
      expect(freeResult.allowed).toBe(false);
      expect(proResult.allowed).toBe(true);
    });
  });

  describe('canAddURL', () => {
    it('should check URL limits', () => {
      const result = canAddURL('free', 0);
      expect(result.allowed).toBe(true);
      expect(result.limit).toBe(1); // Free plan has 1 URL limit
    });

    it('should deny when at URL limit', () => {
      const result = canAddURL('free', 1);
      expect(result.allowed).toBe(false);
    });
  });

  describe('canAddPDF', () => {
    it('should check PDF limits', () => {
      const result = canAddPDF('free', 0);
      expect(result.allowed).toBe(true);
    });

    it('should deny when at PDF limit', () => {
      const result = canAddPDF('free', 1); // Free plan has 1 PDF limit
      expect(result.allowed).toBe(false);
    });
  });

  describe('canUploadFile', () => {
    it('should allow file within size limit', () => {
      const result = canUploadFile('free', 3);
      expect(result.allowed).toBe(true);
      expect(result.remaining).toBe(2);
    });

    it('should deny file exceeding size limit', () => {
      const result = canUploadFile('free', 10);
      expect(result.allowed).toBe(false);
    });

    it('should handle large file size in ultra', () => {
      const result = canUploadFile('ultra', 45); // Limit is 50
      expect(result.allowed).toBe(true);
    });
  });

  describe('canAddText', () => {
    it('should allow text within length limit', () => {
      const result = canAddText('free', 1000);
      expect(result.allowed).toBe(true);
    });

    it('should deny text exceeding length limit', () => {
      const result = canAddText('free', 500000); // Limit is 400000
      expect(result.allowed).toBe(false);
    });
  });

  describe('feature checks', () => {
    describe('canRefreshSource', () => {
      it('should return false for free plan', () => {
        expect(canRefreshSource('free')).toBe(false);
      });

      it('should return true for pro plan', () => {
        expect(canRefreshSource('pro')).toBe(true);
      });

      it('should return true for ultra plan', () => {
        expect(canRefreshSource('ultra')).toBe(true);
      });
    });

    describe('canUseHandoff', () => {
      it('should return false for free and pro plans', () => {
        expect(canUseHandoff('free')).toBe(false);
        expect(canUseHandoff('pro')).toBe(false);
      });

      it('should return true for ultra plan', () => {
        expect(canUseHandoff('ultra')).toBe(true);
      });
    });

    describe('canCustomizeBranding', () => {
      it('should return false for free and pro plans', () => {
        expect(canCustomizeBranding('free')).toBe(false);
        expect(canCustomizeBranding('pro')).toBe(false);
      });

      it('should return true for ultra plan', () => {
        expect(canCustomizeBranding('ultra')).toBe(true);
      });
    });

    describe('hasAnalyticsAccess', () => {
      it('should return false for free plan', () => {
        expect(hasAnalyticsAccess('free')).toBe(false);
      });

      it('should return true for pro and above', () => {
        expect(hasAnalyticsAccess('pro')).toBe(true);
        expect(hasAnalyticsAccess('ultra')).toBe(true);
      });
    });

    describe('hasApiAccess', () => {
      it('should return false for free', () => {
        expect(hasApiAccess('free')).toBe(false);
      });

      it('should return true for pro', () => {
        expect(hasApiAccess('pro')).toBe(true);
      });
    });

    describe('canUseOCR', () => {
      it('should return false for free', () => {
        expect(canUseOCR('free')).toBe(false);
      });

      it('should return true for pro', () => {
        expect(canUseOCR('pro')).toBe(true);
      });
    });

    describe('canUseDynamicScraping', () => {
      it('should return false for free', () => {
        expect(canUseDynamicScraping('free')).toBe(false);
      });

      it('should return true for pro', () => {
        expect(canUseDynamicScraping('pro')).toBe(true);
      });
    });

    describe('canUseSecureEmbed', () => {
      it('should return false for free', () => {
        expect(canUseSecureEmbed('free')).toBe(false);
      });

      it('should return true for pro', () => {
        expect(canUseSecureEmbed('pro')).toBe(true);
      });
    });
  });

  describe('getPlanFeatures', () => {
    it('should return all features for a plan', () => {
      const features = getPlanFeatures('ultra');
      expect(features.customBranding).toBe(true);
      expect(features.analytics).toBe(true);
      expect(features.handoff).toBe(true);
    });

    it('should return restricted features for free plan', () => {
      const features = getPlanFeatures('free');
      expect(features.customBranding).toBe(false);
      expect(features.analytics).toBe(false);
      expect(features.handoff).toBe(false);
    });
  });

  describe('formatRemaining', () => {
    it('should format unlimited result', () => {
      const result: LimitCheckResult = {
        allowed: true,
        current: 10,
        limit: -1,
        remaining: Infinity,
        isUnlimited: true,
      };
      expect(formatRemaining(result)).toBe('Sınırsız');
    });

    it('should format zero remaining', () => {
      const result: LimitCheckResult = {
        allowed: false,
        current: 5,
        limit: 5,
        remaining: 0,
        isUnlimited: false,
      };
      expect(formatRemaining(result)).toBe('Limit doldu');
    });

    it('should format remaining count', () => {
      const result: LimitCheckResult = {
        allowed: true,
        current: 3,
        limit: 5,
        remaining: 2,
        isUnlimited: false,
      };
      expect(formatRemaining(result)).toBe('2 kaldı');
    });
  });

  describe('getLimitStatusMessage', () => {
    it('should return unlimited message', () => {
      const result: LimitCheckResult = {
        allowed: true,
        current: 10,
        limit: -1,
        remaining: Infinity,
        isUnlimited: true,
      };
      expect(getLimitStatusMessage(result, 'chatbot')).toBe('Sınırsız chatbot ekleyebilirsiniz');
    });

    it('should return limit reached message', () => {
      const result: LimitCheckResult = {
        allowed: false,
        current: 1,
        limit: 1,
        remaining: 0,
        isUnlimited: false,
      };
      expect(getLimitStatusMessage(result, 'chatbot')).toBe('chatbot limitinize ulaştınız (1)');
    });

    it('should return remaining count message', () => {
      const result: LimitCheckResult = {
        allowed: true,
        current: 2,
        limit: 5,
        remaining: 3,
        isUnlimited: false,
      };
      expect(getLimitStatusMessage(result, 'kaynak')).toBe('3/5 kaynak ekleyebilirsiniz');
    });
  });
});
