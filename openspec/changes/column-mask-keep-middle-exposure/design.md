## Overview

This slice exposes one additional existing backend column-mask capability through the permission center: keep-middle-three desensitization.

## Current Gap

- Backend runtime masking already supports `KeepMiddleThreeCharacters`.
- The permission-center frontend only offers `all`, `keep_ends`, and `custom`.
- The admin-service adapter that translates between frontend `maskRule` values and backend `DesensitizationRule` JSON also lacks keep-middle mapping.

## Decision

Add a single new frontend-facing `maskRule` value, `keep_middle`, and map it to the existing backend built-in rule `KeepMiddleThreeCharacters`.

## Scope

Included:
- permission center column-rule selector exposure
- admin-service encode/decode mapping
- focused backend and frontend regression tests

Excluded:
- additional built-in rules beyond keep-middle
- custom-rule UX redesign (`RetainMToN`, specialCharacter)
- row-permission operator/editor expansion
- any permission-center structural refactor

## Verification

- Backend tests prove encode/decode/page mapping for keep-middle.
- Frontend test proves the selector exposes keep-middle and submits `maskRule: 'keep_middle'`.
- Standard backend/frontend validation gates run before PR.
