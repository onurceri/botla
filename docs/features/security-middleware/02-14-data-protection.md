> Extracted from: security-middleware.md — Section: 2.14 Data Protection

## 2.14 Data Protection

- [ ] Passwords are hashed with bcrypt (cost factor ≥ 12)
- [ ] Refresh tokens are stored as SHA-256 hashes (not plaintext)
- [ ] Soft delete preserves user data for GDPR compliance (30-day window)
- [ ] Hard delete permanently removes data after soft-delete window expires
- [ ] Privacy request workflow: download, rectify, delete, restrict
- [ ] Admin actions are logged to audit log (who, what, when, IP)
- [ ] Audit log is append-only (no deletion or modification)
- [ ] Audit log includes: actor ID, action type, target ID, metadata, timestamp, source IP
