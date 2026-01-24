# Change: Add multi-dimensional embedding

## Why
Customers need to embed DataEase content at multiple levels (designer, full dashboards/screens, module pages, and single charts) with consistent parameter passing and cross-system interaction.

## What Changes
- Define a single embedding capability that covers designer embedding, view-only board embedding, module-level page embedding, and single chart embedding.
- Require bidirectional parameter passing for dashboards/screens and single charts.
- Standardize token-based embedding initialization and origin allowlisting.
- Document expected entry points for iframe and DIV embedding.

## Impact
- Affected specs: embedded-bi (new)
- Affected code: core-frontend embedded store/router/init, permissions embedded API, share/public-link integration, whitelist handling
