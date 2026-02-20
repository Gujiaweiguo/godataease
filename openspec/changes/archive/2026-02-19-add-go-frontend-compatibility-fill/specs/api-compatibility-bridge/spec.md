# API Compatibility Bridge Spec Delta

## ADDED Requirements

### Requirement: Frontend Compatibility Endpoints
The system SHALL provide frontend compatibility endpoints to support Java-to-Go migration.

#### Scenario: Role router query endpoint
- **WHEN** GET request to `/api/roleRouter/query`
- **THEN** returns route configuration with system menu structure

#### Scenario: Menu resource endpoint
- **WHEN** GET request to `/api/auth/menuResource`
- **THEN** returns menu tree with items containing path and meta fields

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
