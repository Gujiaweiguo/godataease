# api-compatibility-bridge Specification

## Purpose
Define compatibility bridge requirements that map legacy endpoints to Go backend behavior without breaking existing clients.
## Requirements
### Requirement: Java Route Compatibility Bridge
The Go backend SHALL provide compatibility route mappings for Java-era API prefixes during migration.

#### Scenario: Call Java-style route prefix
- **WHEN** a client calls a Java-style route that is listed in the migration compatibility matrix
- **THEN** the Go backend routes the request to the equivalent Go capability
- **AND** the business behavior matches the canonical Go route behavior

### Requirement: Response Contract Parity
The Go backend SHALL keep response contracts compatible with Java client expectations for migrated endpoints.

#### Scenario: Return successful response
- **WHEN** a compatibility route request succeeds
- **THEN** the response uses `code/data/msg` structure
- **AND** the response semantics for empty data remain consistent with Java conventions

#### Scenario: Return failed response
- **WHEN** a compatibility route request fails validation or authorization
- **THEN** the response returns mapped error code and message compatible with Java client handling

### Requirement: Alias Consistency Verification
The migration process SHALL verify functional parity between canonical and compatibility routes.

#### Scenario: Verify alias parity
- **WHEN** regression suite executes for both canonical Go route and compatibility alias
- **THEN** both routes return equivalent status, code, and payload semantics

### Requirement: CI Contract Diff Gate for Critical Java/Go APIs
The migration pipeline SHALL run a CI contract diff gate for critical Java/Go APIs before merge and release.

#### Scenario: Run gate on pull request
- **WHEN** a pull request changes Go backend code or compatibility routing behavior
- **THEN** CI MUST execute the Java/Go contract diff job for whitelisted critical APIs
- **AND** the pull request MUST be blocked when the gate result is failed

#### Scenario: Run gate on protected branch
- **WHEN** code is merged into protected branches
- **THEN** CI SHALL rerun the contract diff gate and persist the result for audit

### Requirement: Whitelisted Critical API Set Governance
The system SHALL maintain a versioned whitelist of critical APIs included in the contract diff gate.

#### Scenario: Maintain whitelist metadata
- **WHEN** an API is added to or removed from the whitelist
- **THEN** the change MUST include endpoint, HTTP method, owner, and blocking level metadata
- **AND** the change MUST be reviewable in version control

#### Scenario: Limit gate scope to governed APIs
- **WHEN** contract diff gate executes
- **THEN** only APIs in the approved whitelist are considered for pass/fail evaluation

### Requirement: Failure Threshold Policy and Report Archival
The contract diff gate SHALL enforce explicit failure thresholds and archive diff reports for each run.

#### Scenario: Enforce threshold-based failure
- **WHEN** diff results violate configured thresholds (overall parity, required APIs, or blocking-level differences)
- **THEN** CI MUST return a failed status and provide categorized failure reasons

#### Scenario: Archive gate reports
- **WHEN** the contract diff gate completes
- **THEN** CI SHALL archive machine-readable and human-readable reports as build artifacts
- **AND** archived reports MUST allow tracing by commit or pull request identifier

#### Scenario: Update threshold policy
- **WHEN** threshold policy is updated for migration phase changes
- **THEN** the policy change MUST be versioned and reviewable in repository history
- **AND** gate evaluation MUST use the latest approved threshold version

### Requirement: Contract Diff Runtime Engine Execution
The system SHALL provide an executable contract diff runtime engine that compares Java and Go API responses based on a governed whitelist.

#### Scenario: Load and validate whitelist before execution
- **WHEN** the runtime engine starts with a whitelist file
- **THEN** it MUST validate required fields (`path`, `method`, `owner`, `priority`, `blockingLevel`) for every API entry
- **AND** it MUST stop with a deterministic configuration error if any required field is missing or invalid

#### Scenario: Execute comparisons in bounded parallel mode
- **WHEN** the runtime engine processes the whitelist API set
- **THEN** it MUST execute Java/Go requests with configurable bounded concurrency
- **AND** it MUST produce deterministic aggregated results independent of request completion order

### Requirement: Stable Timeout and Retry Semantics
The runtime engine SHALL apply stable timeout and retry behavior for transient request failures.

#### Scenario: Retry transient failures
- **WHEN** an API request fails due to transient timeout or connection issues
- **THEN** the engine MUST retry according to configured retry count and timeout policy
- **AND** each exhausted retry MUST be recorded with failure category and final error detail

#### Scenario: Preserve deterministic final status
- **WHEN** retries are exhausted for an API entry
- **THEN** the engine MUST emit a deterministic final failure result for that entry
- **AND** the gate decision MUST use the post-retry final result only

### Requirement: Structured Diff and Failure Classification Output
The runtime engine SHALL output structured diffs and classified failures for CI gate consumption.

#### Scenario: Emit multi-dimensional diff
- **WHEN** Java and Go responses are both available
- **THEN** the output MUST include diff dimensions for `status`, `code`, `msg`, `payload schema`, and `payload value`
- **AND** each API result MUST include `passed` status and categorized failure reason when failed

#### Scenario: Produce gate-ready reports and exit semantics
- **WHEN** diff execution finishes
- **THEN** the engine MUST write both machine-readable JSON and human-readable Markdown reports
- **AND** it MUST return non-zero exit code if blocking-level failures violate threshold policy, otherwise return zero

### Requirement: Interface-Level Baseline Fixture Governance
The contract diff system SHALL manage baseline fixtures at interface level with deterministic structure, naming, and metadata.

#### Scenario: Resolve baseline fixture by API identity
- **WHEN** a contract diff run targets a specific API (`path + method`)
- **THEN** the system MUST resolve exactly one baseline fixture by deterministic path and naming rules
- **AND** the resolved fixture MUST contain required metadata for version and ownership traceability

#### Scenario: Reject malformed baseline fixtures
- **WHEN** a baseline fixture is missing required fields or has invalid metadata
- **THEN** the validation step MUST fail with deterministic configuration errors
- **AND** the gate MUST stop before performing parity comparison

### Requirement: Incremental Baseline Refresh Workflow
The system SHALL support incremental baseline refresh with dry-run preview and controlled apply mode.

#### Scenario: Preview incremental changes in dry-run mode
- **WHEN** refresh is executed in dry-run mode
- **THEN** no baseline files MUST be modified
- **AND** a diff preview report MUST be generated for review

#### Scenario: Apply incremental refresh for scoped APIs only
- **WHEN** refresh is executed in apply mode with a scoped API set
- **THEN** only targeted baseline fixtures MUST be updated
- **AND** non-target baseline fixtures MUST remain unchanged

### Requirement: Review, Rollback, and Drift Control Policy
The baseline mechanism SHALL enforce review gates, rollback capability, and drift safeguards to reduce false positives/negatives.

#### Scenario: Enforce review before baseline adoption
- **WHEN** baseline fixture changes are proposed
- **THEN** required reviewers (owner/tech lead/QA) MUST approve before adoption
- **AND** unapproved baseline updates MUST NOT be used to bypass gate failures

#### Scenario: Roll back to last stable baseline
- **WHEN** baseline updates cause regression or unacceptable drift
- **THEN** the system MUST support rollback to the last stable baseline version
- **AND** rollback evidence (trigger, operator, result) MUST be archived for audit

### Requirement: Unauthorized vs Forbidden Semantic Parity
The contract diff system SHALL enforce semantic parity for authentication and authorization failures between Java and Go APIs.

#### Scenario: Distinguish unauthenticated and unauthorized access
- **WHEN** a request is sent without valid authentication credentials
- **THEN** both Java and Go responses MUST return unauthenticated semantics (`401` + expected error contract)
- **AND** responses MUST NOT be misclassified as authorization failures (`403`)

#### Scenario: Enforce forbidden semantics for insufficient privileges
- **WHEN** an authenticated principal lacks required permission
- **THEN** both Java and Go responses MUST return forbidden semantics (`403` + expected error contract)
- **AND** the gate MUST fail if status or error semantics diverge

### Requirement: Row-Level Access and Column Masking Parity
The contract diff system SHALL validate negative security parity for row-level access control and sensitive column masking.

#### Scenario: Block unauthorized row visibility
- **WHEN** a low-privilege principal queries data outside authorized scope
- **THEN** Java and Go responses MUST both hide unauthorized rows consistently
- **AND** any unauthorized row exposure MUST be classified as blocking failure

#### Scenario: Enforce sensitive column masking consistency
- **WHEN** a principal without sensitive-field permission accesses protected columns
- **THEN** Java and Go responses MUST apply equivalent masking semantics
- **AND** unmasked sensitive values in either side MUST fail the gate

### Requirement: Export and Download Authorization Negative Suite
The contract diff system SHALL validate authorization controls for export and download flows under negative security scenarios.

#### Scenario: Reject expired or invalid token download
- **WHEN** export/download APIs are called with expired or invalid credentials
- **THEN** Java and Go responses MUST reject requests with equivalent security semantics
- **AND** successful access in either side MUST be treated as blocking regression

#### Scenario: Reject cross-user or cross-tenant artifact access
- **WHEN** a principal tries to download artifacts owned by another principal or tenant
- **THEN** Java and Go responses MUST deny access consistently
- **AND** the gate MUST archive scenario evidence and fail on any bypass

### Requirement: Required Gate Interface Policy
The API compatibility bridge SHALL define and govern a required-gate interface set that must pass compatibility gates before release.

#### Scenario: Maintain required-gate interface set
- **WHEN** interfaces are classified as required-gate scope
- **THEN** the scope MUST be versioned, reviewable, and traceable to ownership
- **AND** scope changes MUST follow explicit approval rules

#### Scenario: Prevent release with unmet required-gate interfaces
- **WHEN** at least one required-gate interface fails or is not evaluated
- **THEN** the release process MUST be blocked
- **AND** the system MUST provide actionable failure evidence per interface

### Requirement: Mandatory Pre-Release Gate Enforcement
The system SHALL enforce pre-release compatibility gate checks as a mandatory release condition.

#### Scenario: Block release on gate failure
- **WHEN** pre-release gate execution returns failed status for required interfaces
- **THEN** release MUST NOT proceed to publish stage
- **AND** bypass through manual override MUST be denied unless approved exception is active

#### Scenario: Require gate evidence for release approval
- **WHEN** a release candidate requests approval
- **THEN** gate evidence MUST include result status, evaluated interface scope, and execution timestamp
- **AND** missing evidence MUST be treated as release blocking

### Requirement: Exception Approval and Waiver Expiration Governance
The system SHALL enforce exception approval and time-bounded waiver governance for required-gate policy.

#### Scenario: Enforce exception approval before waiver activation
- **WHEN** a team requests temporary waiver for required-gate failure
- **THEN** waiver MUST remain inactive until required approvers complete approval with rationale
- **AND** unapproved waiver MUST NOT unblock release

#### Scenario: Enforce waiver expiry and revalidation
- **WHEN** an approved waiver reaches expiry time
- **THEN** waiver MUST automatically expire and stop unblocking release
- **AND** release MUST require renewed approval or fully passing gate results

### Requirement: Frontend Compatibility Endpoints
The system SHALL provide frontend compatibility endpoints to support Java-to-Go migration and SHALL preserve locale-aware contract semantics for navigation metadata returned to migrated frontend clients.

#### Scenario: Role router query endpoint
- **WHEN** GET request to `/api/roleRouter/query`
- **THEN** returns route configuration with system menu structure
- **AND** localized menu titles in `meta.title` MUST use the effective locale for that request instead of fixed-language labels or untranslated i18n keys

#### Scenario: Menu resource endpoint
- **WHEN** GET request to `/api/auth/menuResource`
- **THEN** returns menu tree with items containing path and meta fields
- **AND** localized menu titles in `meta.title` MUST use the effective locale for that request instead of fixed-language labels or untranslated i18n keys

#### Scenario: Interactive tree endpoint
- **WHEN** POST request to `/api/dataVisualization/interactiveTree` with JSON body
- **THEN** returns visualization tree structure or empty object

#### Scenario: AI base URL endpoint
- **WHEN** GET request to `/api/aiBase/findTargetUrl`
- **THEN** returns empty map or AI configuration

#### Scenario: Xpack component endpoint
- **WHEN** GET request to `/api/xpackComponent/content/:id`
- **THEN** returns HTTP 501 (Not Implemented) as enterprise feature

#### Scenario: Xpack plugin static info endpoint
- **WHEN** GET request to `/api/xpackComponent/pluginStaticInfo/:id`
- **THEN** returns HTTP 501 (Not Implemented) as enterprise feature

#### Scenario: WebSocket info endpoint
- **WHEN** GET request to `/api/websocket/info`
- **THEN** returns HTTP 501 (Not Implemented) or connection info

### Requirement: Dynamic Compatibility Menu Responses
Compatibility endpoints for frontend menu loading SHALL be generated from the same authorization-backed menu service as canonical menu APIs.

#### Scenario: Compatibility menuResource uses dynamic authorization source
- **WHEN** a client calls `GET /api/auth/menuResource`
- **THEN** the response menu tree is generated from persisted menu data filtered by role authorization
- **AND** the endpoint does not return a hardcoded static menu list

#### Scenario: Compatibility roleRouter uses dynamic authorization source
- **WHEN** a client calls `GET /api/roleRouter/query`
- **THEN** router records are generated from authorized persisted menus
- **AND** route metadata remains contract-compatible with existing frontend route parser

### Requirement: Compatibility and Canonical Parity for Menu Authorization
The system SHALL keep menu authorization semantics equivalent between compatibility and canonical endpoints.

#### Scenario: Same user receives consistent menu visibility
- **WHEN** the same authenticated user requests both canonical and compatibility menu endpoints
- **THEN** the visible menu scope is consistent across endpoints
- **AND** differences are limited to contract shape fields required by each endpoint

### Requirement: Locale-Aware Compatibility Navigation Fallback
Compatibility navigation endpoints SHALL apply the governed locale fallback policy consistently whenever request locale input is missing, unsupported, or incomplete.

#### Scenario: Compatibility menu endpoints fall back deterministically
- **WHEN** a client calls `/api/roleRouter/query` or `/api/auth/menuResource` without a supported request locale
- **THEN** both endpoints MUST resolve the same effective locale for that request context
- **AND** both endpoints MUST return navigation titles localized with the same fallback result

### Requirement: Placeholder Success Prohibition for Compatibility Endpoints
Migration-scoped compatibility endpoints SHALL NOT return placeholder success responses when core behavior is not implemented.

#### Scenario: Reject placeholder success semantics
- **WHEN** a compatibility endpoint lacks required business implementation
- **THEN** the endpoint MUST return explicit non-success semantics (for example deterministic unavailable/error status)
- **AND** MUST NOT return `code=000000` with empty placeholder payload to simulate parity

### Requirement: Compatibility Runtime-to-Matrix Status Consistency
Compatibility route status metadata SHALL remain consistent with observed runtime behavior.

#### Scenario: Keep matrix status synchronized with runtime behavior
- **WHEN** endpoint behavior changes between `full/partial/stub/missing`
- **THEN** migration matrix and whitelist metadata MUST be updated in the same change set
- **AND** changes MUST include evidence references to tests or contract-diff outputs

#### Scenario: Block merge on status drift
- **WHEN** CI detects status drift between route behavior and governed metadata
- **THEN** the pipeline MUST fail with actionable drift diagnostics
- **AND** merge MUST be blocked until metadata and behavior are aligned

### Requirement: Time-Bounded Stub Waiver Governance
Temporary compatibility stubs SHALL require explicit, time-bounded waiver governance.

#### Scenario: Use approved waiver for temporary stubs
- **WHEN** a team keeps an endpoint in stub status for a release window
- **THEN** the stub MUST reference approved owner, reason, and expiry metadata
- **AND** expired waiver MUST stop unblocking release gates automatically



### Requirement: Permission Compatibility Endpoint Family Coverage
The compatibility bridge SHALL expose permission-management endpoints required by frontend role/menu administration flows.

#### Scenario: Permission endpoints are routable through compatibility prefix
- **WHEN** clients call legacy permission endpoints under compatibility path (for example `/de2api/auth/menuPermission`)
- **THEN** Go backend MUST route the request to implemented handlers instead of returning `404`
- **AND** response semantics remain compatible with frontend caller expectations

### Requirement: System Role Path Compatibility
The compatibility bridge SHALL support legacy `/system/role/*` API paths used by frontend admin pages.

#### Scenario: Create/update/delete role through legacy system path
- **WHEN** a client calls `/de2api/system/role/create`, `/de2api/system/role/update`, or `/de2api/system/role/delete/:roleId`
- **THEN** the backend MUST map these requests to canonical Go role operations
- **AND** return Java-compatible response envelope

### Requirement: Core BI Compatibility Gate Scope
The compatibility bridge SHALL define datasource, dataset, dashboard, and big-screen endpoints as a governed critical-flow scope for migration release readiness.

#### Scenario: Govern core BI routes as required verification scope
- **WHEN** compatibility gate scope is prepared for release or merge validation
- **THEN** the scope MUST identify the canonical route, compatibility alias, owner, and blocking level for each in-scope BI endpoint family
- **AND** the governed scope MUST be reviewable in version control with evidence references

#### Scenario: Block release on missing core BI verification
- **WHEN** any governed datasource, dataset, dashboard, or big-screen endpoint is not evaluated or fails required parity checks
- **THEN** release readiness MUST be treated as failed
- **AND** the system MUST provide actionable evidence for the missing or failing route family

### Requirement: Core BI Compatibility Endpoints Must Not Be Stub-Success
Compatibility endpoints that participate in the critical BI flow SHALL return implemented behavior or explicit non-success semantics.

#### Scenario: Reject placeholder success for critical BI compatibility route
- **WHEN** an in-scope BI compatibility endpoint lacks required business implementation
- **THEN** the endpoint MUST return deterministic non-success behavior
- **AND** MUST NOT return `code=000000` with placeholder payload to simulate parity

#### Scenario: Keep alias and canonical route semantics aligned for critical BI flows
- **WHEN** both canonical and compatibility forms of an in-scope BI route are invoked with equivalent inputs
- **THEN** status, envelope semantics, and business result shape MUST remain aligned
- **AND** divergence MUST be treated as a governed compatibility regression

### Requirement: Interactive Tree Compatibility Status Promotion
The compatibility bridge SHALL treat `dataVisualization/interactiveTree` as a governed parity endpoint once real resource-tree behavior is implemented.

#### Scenario: Promote interactive tree from partial to full
- **WHEN** interactive tree behavior is backed by implementation and regression evidence
- **THEN** governed whitelist or matrix metadata MUST be updated from `partial` to `full`
- **AND** the metadata update MUST reference the evidence proving parity scope completion

#### Scenario: Block governance promotion without parity evidence
- **WHEN** interactive tree still depends on synthetic placeholders or lacks regression evidence
- **THEN** governance metadata MUST remain non-full
- **AND** release documentation MUST NOT claim complete parity for the endpoint

### Requirement: Dataset and Datasource Interactive Governance Consistency
The compatibility and governance model SHALL describe dataset and datasource interactive discovery in a way that is consistent with the interactive aggregate view used by the frontend.

#### Scenario: Interactive aggregate governance covers dataset and datasource discovery
- **WHEN** interactive discovery behavior is evaluated for release readiness
- **THEN** dataset and datasource discovery paths used by the aggregate interactive loader MUST be documented with implementation evidence and governed status
- **AND** the documented status MUST match the actual runtime loading path used by the frontend

### Requirement: Admin-Domain Compatibility Policy Must Distinguish Canonical and Legacy Contracts
The system MUST explicitly govern which admin-domain routes are canonical, which are compatibility aliases, and which are scheduled for retirement.

#### Scenario: Maintainer reviews an admin-domain route family
- **WHEN** user, role, organization, or permission endpoints are evaluated for alignment
- **THEN** the workflow MUST classify each route family as canonical, compatibility alias, or deprecated
- **AND** the classification MUST be reviewable before semantic parity is declared complete

### Requirement: Missing Compatibility Paths Must Be Classified Before Release
The system MUST classify missing or unimplemented compatibility paths as governed contract gaps.

#### Scenario: Frontend references legacy admin-domain path
- **WHEN** a governed frontend flow still references a legacy path with no working backend behavior
- **THEN** the system MUST classify the issue as a compatibility gap
- **AND** release readiness MUST depend on either a working alias, a frontend migration, or an approved retirement decision
