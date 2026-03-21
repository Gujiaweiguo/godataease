## Why

The hardening change `harden-recovered-route-semantics` closed datasource forbidden coverage at the permission-middleware boundary, but it explicitly blocked runtime repair for datasource list aliases.

The reason is structural: the live datasource list endpoints do not carry a stable resource identifier, so they cannot safely reuse `CheckDatasourceView()` without changing API meaning. Continuing that work inside a hardening change would blur the line between verification hardening and permission-model redesign.

This change isolates the design problem so it can be solved deliberately: what identity or filtering model should datasource list routes use when the system wants forbidden semantics to be explicit at runtime?

## What Changes

- Define the permission and API model for datasource list routes where callers currently lack a stable per-resource identity.
- Compare viable implementation strategies for datasource list authorization and semantics.
- Select one design direction and break it into executable implementation tasks.

## Capabilities

### Modified Capabilities
- `datasource-management`: clarify runtime authorization semantics for datasource list endpoints and their compatibility aliases

## Impact

- **Backend Go**: datasource route design, permission middleware strategy, and possibly datasource service/repository filtering behavior
- **Frontend**: datasource list caller expectations for forbidden vs filtered vs empty results
- **Verification**: future route-level and permission-semantics coverage for datasource list aliases
