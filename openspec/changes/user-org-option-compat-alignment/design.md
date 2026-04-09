## Overview

This change resolves a route-binding semantic drift in `RegisterCompatibilityBridgeRoutes`: `/user/org/option` must always expose user-option semantics and must not switch to organization-list semantics based on optional handler presence.

## Current Problem

- In compatibility bridge user route registration, `/user/org/option` is conditionally bound:
  - `org.ListOrgs` when `org != nil`
  - `user.GetUserOptions` when `org == nil`
- This conditional branch yields two incompatible payload semantics for the same endpoint.

## Decision

Bind `/user/org/option` unconditionally to `user.GetUserOptions` whenever `user` routes are registered.

### Non-goals

- No `/org/mounted` payload-shape redesign in this change.
- No broad compat-route cleanup or middleware policy changes.

## Test Strategy

Add focused compatibility bridge tests that prove:

1. `/user/org/option` returns user-option payload shape when both user and org handlers are present.
2. `/user/org/option` still returns user-option payload shape when org handler is absent.

The tests seed minimal sqlite-backed user/org data and assert response JSON shape to distinguish user options (`userId`/`username` style fields) from org list (`orgId`/`orgName` fields).

## Risk and Rollback

- Risk is low and isolated to one legacy route registration branch.
- Rollback is a single-route binding revert and test rollback if required.
