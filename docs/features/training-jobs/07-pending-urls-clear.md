> Extracted from: training-jobs.md — Section: 2.6 Pending URLs — Clear All

# 2.6 Pending URLs — Clear All

- [ ] `DELETE /api/v1/chatbots/{id}/pending-urls/clear` deletes all pending URLs for the chatbot
- [ ] Response includes `cleared_count`
- [ ] No-op when no pending URLs exist (returns cleared_count=0)
