# Verinode — Authentication & Authorization (RBAC)

*Design + justification for who can access what. Implemented in `services/internal/auth`
and enforced in the API gateway middleware.*

## The question: who gets the admin panel?

**Decision: the platform admin console is restricted to `super_admin` + `admin` only.
Registered companies do NOT get the platform console — they get a *scoped org view*
(the buyer/seller portals) that shows only their own data.**

### Why not give companies an admin panel?

The admin console's functions are **platform trust-domain** functions:

- **Participant approval / KYB-KYC / credit limits** — a compliance decision Verinode makes
  *about* a company; a company approving itself is nonsensical and a control failure.
- **Surveillance** (wash-trade, related-party, concentration detection) — must be
  independent of the parties being surveilled. Exposing it to companies would let a
  manipulator see exactly what trips a flag.
- **Cross-company audit** — spans multiple companies' trades; showing it to any one company
  leaks competitors' RFQs, prices and counterparties, and breaks the "independent benchmark
  administrator" claim the index depends on.

So companies get **self-service scoped to their own entity**, never the platform console.
This mirrors the PRD's separation of internal Ops/Admin + Compliance from external Buyer/Seller,
and the architecture doc's trust-domain separation.

## Role model

| Role | Scope | Can |
|---|---|---|
| `super_admin` | platform | everything, incl. user/role management |
| `admin` | platform | the `/admin` console: participant approval, surveillance, claims, cross-company audit, KYC/credit |
| `company_admin` | one company | manage that company's users; full buyer/seller actions **for that company only** |
| `trader` | one company | create RFQs, respond to quotes, within their company |
| `viewer` | one company | read-only, within their company |

`super_admin` + `admin` are **platform staff** (`Role.IsPlatformStaff()`). The other three are
**company-scoped** and carry a `company_id` in their token; the platform console is closed to them.

## Enforcement

- **Auth**: `POST /v1/auth/login`, `POST /v1/auth/register` (company self-registration →
  `company_admin`), `GET /v1/auth/me`. Sessions are HS256 JWTs (12h) delivered as a Bearer
  token and an HttpOnly `vn_session` cookie.
- **Passwords**: PBKDF2-HMAC-SHA256, 120k iterations, per-user salt, constant-time compare
  (stdlib only — swappable for the OIDC/SAML vendor named in the architecture docs; the
  middleware boundary is unchanged).
- **Gateway RBAC middleware** rejects any request to a platform-privileged surface
  (`/v1/admin/*`, `/v1/surveillance/*`, `/v1/internal/*`, and KYC/credit mutations) that is
  not carrying a valid platform-staff token: `401` when unauthenticated, `403` when a company
  principal attempts it.

## Seeded demo principals (dev only)

| Role | Email | Password |
|---|---|---|
| super_admin | root@verinode.io | verinode-super-admin |
| admin | compliance@verinode.io | verinode-admin |
| company_admin (buyer) | admin@anthropic-spv.example | demo-company-admin |
| trader (buyer) | trader@anthropic-spv.example | demo-trader |
| company_admin (seller) | ops@lambda-labs.example | demo-seller-admin |

Rotate `AUTH_JWT_SECRET` in production; the seeded store is replaced by the Postgres-backed
store + OIDC/SAML vendor for real deployments.
