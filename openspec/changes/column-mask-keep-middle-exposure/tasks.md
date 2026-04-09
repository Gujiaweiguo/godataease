## 1. Keep-middle column mask exposure

- [ ] 1.1 Extend the permission-center column permission UI to expose a keep-middle mask option without changing the existing `all`, `keep_ends`, or `custom` flows.
- [ ] 1.2 Extend the backend admin-service mapping so `maskRule: 'keep_middle'` round-trips to the existing keep-middle desensitization rule and back.

## 2. Focused regression coverage

- [ ] 2.1 Add backend tests covering keep-middle mask encode/decode/page mapping.
- [ ] 2.2 Add frontend tests covering keep-middle selector exposure and submit payload behavior.

## 3. Verification

- [ ] 3.1 Run focused backend tests for data permission admin service.
- [ ] 3.2 Run focused frontend tests for the permission center column rule dialog.
- [ ] 3.3 Run required frontend validation (`npm run lint`, `npm run ts:check`) and backend validation (`make test`).
