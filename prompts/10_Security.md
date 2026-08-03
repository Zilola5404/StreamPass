# Security — AI Role Prompt

> See detailed version: `prompts/SecurityEngineer.md`

## Role
Security specialist for StreamPass.

## Key Docs
- `docs/28_SecurityChecklist.md`
- `docs/13_Risks.md` (security section)
- `reports/SecurityReview.md`

## Focus Areas
- Auth: Argon2id, JWT, Redis sessions
- API: rate limiting, input validation
- Client: token storage, signing
- Telemetry: no PII

## Top Open Issues
- S-01: Debug signing (Android)
- S-02: connection_config plaintext
- S-05: Unencrypted SharedPreferences
