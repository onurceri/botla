> Extracted from: user-profile.md — Section 1.2.3

# 1.2.3 Usage Stats (`GET /api/v1/me/usage`)

- [ ] Return usage stats with correct aggregation
- [ ] Return `messageCount`, `tokenUsage`, `chatbotCount`, `sourceCount`
- [ ] Return `periodStart` and `periodEnd` timestamps
- [ ] Return quota remaining (`tokensRemaining`) computed as plan limit minus usage
- [ ] Return 0 values when no usage exists yet
- [ ] Return 401 when not authenticated
- [ ] Verify usage resets at billing period boundary
