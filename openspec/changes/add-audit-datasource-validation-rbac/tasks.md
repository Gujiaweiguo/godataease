## 1. Audit export/download authorization

- [x] 1.1 Identify the concrete audit menu path already governing the audit feature surface and document it in the implementation notes.
- [x] 1.2 Attach menu-based authorization to `POST /api/audit/export` and `GET /api/audit/download` without changing the existing JWT or rate-limit protections.
- [x] 1.3 Extend backend route/handler tests to verify authorized audit export/download requests still succeed and unauthorized authenticated requests receive `403`.

## 2. Datasource validation authorization

- [x] 2.1 Apply datasource authorization checks to canonical validation routes so `GET /api/ds/validate/:id` requires datasource authority and `POST /api/ds/validate` enforces the draft-validation fallback path when no stable datasource ID exists.
- [x] 2.2 Apply the same datasource validation authorization semantics to compatibility aliases under `/api/datasource/*` and `/de2api/datasource/*`.
- [x] 2.3 Extend backend route/handler tests to verify allowed validation requests still succeed and unauthorized authenticated requests receive `403` across canonical and compatibility validation routes.

## 3. Verification and rollout safety

- [x] 3.1 Run targeted backend tests covering audit and datasource validation authorization allow/deny behavior and fix regressions.
- [x] 3.2 Run repository-level backend validation (`make test`, affected compatibility checks, and any route authorization gates) before opening the implementation PR.
