> Extracted from: chat-operations.md — Section: 1.3 Public Chatbot Config

## 1.3 Public Chatbot Config

- [ ] GET /api/v1/public/chatbots/{id}/config returns chatbot appearance settings
- [ ] Response includes `theme`, `primaryColor`, `position`, `title`, `description`
- [ ] Response includes `suggestions` (array of starter prompts)
- [ ] Response includes `branding` settings (logo URL, company name, hideBranding flag)
- [ ] Non-public chatbot returns 404
- [ ] Response is cached with appropriate Cache-Control headers
- [ ] Config endpoint does not leak internal fields (API keys, system prompts)
