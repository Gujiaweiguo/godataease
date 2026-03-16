## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: Locale-Aware Compatibility Navigation Fallback
Compatibility navigation endpoints SHALL apply the governed locale fallback policy consistently whenever request locale input is missing, unsupported, or incomplete.

#### Scenario: Compatibility menu endpoints fall back deterministically
- **WHEN** a client calls `/api/roleRouter/query` or `/api/auth/menuResource` without a supported request locale
- **THEN** both endpoints MUST resolve the same effective locale for that request context
- **AND** both endpoints MUST return navigation titles localized with the same fallback result
