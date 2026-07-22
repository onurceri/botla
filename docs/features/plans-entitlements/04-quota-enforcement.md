> Extracted from: plans-entitlements.md — Section: 1.4 Quota Enforcement

## 1.4 Quota Enforcement

- [ ] Token quota: `QuotaEnforcer.ReserveTokens` reserves estimated tokens before chat; `AdjustTokens` corrects after completion; `RefundTokens` on error
- [ ] Chatbot count: checked at chatbot creation time against `MaxChatbots`
- [ ] File limits: per-bot (`FilesMaxFilesPerBot`) and total (`FilesMaxFilesTotal`, `FilesTotalStorageMB`)
- [ ] Rate limits: Redis-backed sliding window per plan (`pkg/ratelimit` + `pkg/middleware`)
- [ ] Cooldown: `MinReAddCooldownMinutes` prevents rapid re-addition of same source

