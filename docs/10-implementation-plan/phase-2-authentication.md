# NEO Bank — Implementation Plan — Phase 2: Authentication

_Part of [10-implementation-plan.md](../10-implementation-plan.md)._

## Purpose

Implement user registration, login, token refresh, and logout. Phase 2 establishes the authentication layer that protects every subsequent banking operation.

**Week:** 2-3

## Deliverables

| Deliverable | Description |
|-------------|-------------|
| Backend endpoints | `POST /register`, `POST /login`, `POST /refresh`, `POST /logout` |
| Frontend screens | Register, Login |
| JWT tokens | RS256-signed access (15 min) and refresh (7 days) tokens |
| Password hashing | Argon2id with secure parameters |
| API Gateway JWT enforcement | Deny-by-default middleware on the gateway: validate the RS256 access token on every `/api/v1/*` route except the public auth routes (`/register`, `/login`, `/refresh`, `/logout`), then forward the token to the upstream service |
| Refresh token storage | Hashed refresh tokens in Cassandra |
| Tests | Unit and integration tests for all auth endpoints and screens |

## Tasks

- [ ] Implement password hashing with Argon2id (recommended parameters).
- [ ] Implement JWT token generation (RS256, 15-minute expiry).
- [ ] Implement JWT token validation middleware for protected endpoints (the single authentication middleware owned by Phase 2 and reused by every later phase).
- [ ] Implement API Gateway JWT enforcement: deny-by-default middleware that validates the RS256 access token on every `/api/v1/*` route except the public auth routes (`POST /register`, `/login`, `/refresh`, `/logout`), returns `401 UNAUTHENTICATED`/`TOKEN_EXPIRED` in the standard error envelope, and forwards the (validated) `Authorization` header to the upstream service. Downstream services still validate independently (defence in depth).
- [ ] Implement `POST /register` endpoint (email validation, password strength, duplicate check).
- [ ] Implement `POST /login` endpoint (credential validation, issue access + refresh tokens).
- [ ] Implement `POST /refresh` endpoint (validate refresh token, issue new access token).
- [ ] Implement `POST /logout` endpoint (invalidate refresh token).
- [ ] Implement user repository (CQL queries: insert, find by email, find by ID).
- [ ] Implement refresh token repository (CQL queries: insert, find, invalidate).
- [ ] Write unit tests for auth service (hashing, token generation/validation).
- [ ] Write integration tests for auth endpoints with real Cassandra.
- [ ] Implement Register screen (frontend, form validation, error states).
- [ ] Implement Login screen (frontend, form validation, error states).
- [ ] Implement token refresh logic in the frontend API client.
- [ ] Write frontend tests for auth screens (Vitest).

## Acceptance Criteria

- User can register with a valid email and strong password.
- Duplicate email registration is rejected without revealing account existence.
- User can log in with valid credentials and receive access + refresh tokens.
- Invalid credentials return a generic error (no account enumeration).
- Access tokens expire after 15 minutes; refresh tokens after 7 days.
- Refresh tokens can be used to obtain new access tokens.
- Logout invalidates refresh tokens (subsequent refresh attempts rejected).
- Passwords hashed with Argon2id (never stored in plain text).
- Tokens validated on every protected endpoint (401 on missing/invalid) and at the gateway before the request reaches a service.
- A protected `/api/v1/*` route called without a valid JWT is rejected by the gateway with `401` and never reaches the upstream service; the public auth routes remain reachable without a token.
- Authentication completes within 300 ms (p95).
- Tests pass (unit, integration, frontend).

## Dependencies

Phase 1 (data model for users and refresh_tokens tables).

## Testing

- Unit tests: password hashing, JWT generation/validation, email validation.
- Integration tests: register → login → refresh → logout flow with real Cassandra.
- Frontend tests: Register/Login screens with validation and error states.
- Security tests: no account enumeration on duplicate email or invalid credentials.
- Negative authorisation tests: protected gateway route without a token, with an expired token, with a tampered signature, and with a token for another account — each must be rejected (`404` preferred over `403` for another account's resources).

## Risks & Considerations

- **Account enumeration:** registration and login errors must not reveal whether an email exists.
- **Token storage:** access/refresh tokens kept in memory, not `localStorage`, to reduce XSS risk.
- **Refresh token rotation:** consider rotating refresh tokens on each use for stronger security.

## Cross-references

- Features: [`FEATURE-01`](../02-features/FEATURE-01-Account-Registration.md), [`FEATURE-02`](../02-features/FEATURE-02-Authentication.md)
- User stories: [`US-001`](../03-user-stories/US-001.md), [`US-002`](../03-user-stories/US-002.md)
- Security rules: `AGENTS.md`
- Backend auth: [`docs/08-tech-stack/backend/jwt.md`](../08-tech-stack/backend/jwt.md)
- Definition of Done: [`definition-of-done.md`](./definition-of-done.md)
- Planned test: `test_phase_2_authentication` (verifies auth deliverables are met)