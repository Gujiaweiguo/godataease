## Overview

This slice migrates the menu authorization tab in the unified permission center from compatibility menu-permission save/query APIs to canonical role-menu APIs for role-scoped operations.

## Current State

- Frontend menu tab currently uses:
  - `menuPerApi` -> `/auth/menuPermission`
  - `menuPerSaveApi` -> `/auth/saveMenuPer`
- Canonical endpoints already exist and are protected under authenticated API routes:
  - `GET /roleMenu/auth/:roleId`
  - `POST /roleMenu/auth`

## Decision

1. For role-selected loading in `MenuPermission.vue`, use `roleMenuAuthApi(roleId)`.
2. For save action, use `roleMenuAuthSaveApi({ roleId, menuIds })`.
3. Keep no-role initial tree loading via existing menu-tree query to avoid broad UX changes.

## Scope

Included:
- One frontend view migration (`MenuPermission.vue`)
- One focused frontend unit test

Excluded:
- Removing compatibility routes from backend
- Full migration of all non-permission-center menu auth callers
- Menu/resource/row/column broader refactors

## Verification

- Frontend unit test verifies role load/save paths call canonical role-menu APIs.
- Existing frontend lint and ts-check remain green.
