# embedded-bi Specification

## Purpose
This capability defines the multi-dimensional embedding framework for DataEase content integration. It supports embedding dashboards, screens, module pages, and individual charts into third-party systems using iframe or DIV containers. The framework includes token-based authentication, origin validation, and bidirectional parameter passing for seamless cross-system interaction and data synchronization.
## Requirements
### Requirement: Embedded Application Registration
The system SHALL allow administrators to create embedded applications with an app id, app secret, and allowed origin list.

#### Scenario: Registering an embedded application
- **WHEN** an administrator submits an embedded application name and origin list
- **THEN** the system stores the application and provides an app id and secret

### Requirement: Token-Based Embedding Initialization
The system SHALL support initializing embedded content using a JWT-style embedded token generated from app id, app secret, and a DataEase user account.

#### Scenario: Generating an embedded token
- **WHEN** a caller provides app id, app secret, and a valid user account
- **THEN** the system returns an embedded token for use in iframe or DIV embedding

### Requirement: Designer Embedding
The system SHALL support embedding the dashboard and data screen designers in a third-party system with edit capability.

#### Scenario: Embedding a dashboard designer
- **WHEN** the host initializes an embedded designer session with a valid token and resource id
- **THEN** the dashboard designer loads and allows edits within the embedded container

### Requirement: Board Embedding
The system SHALL support embedding completed dashboards and data screens with interactive features including linkage, jump, drill, and filter components.

#### Scenario: Embedding a dashboard with interactivity
- **WHEN** a host embeds a dashboard with a valid token and resource id
- **THEN** the dashboard renders and interactive actions operate as in the native product

### Requirement: Module-Level Embedding
The system SHALL support embedding module pages for datasources, datasets, dashboards, and data screens, including the left-side tree navigation.

#### Scenario: Embedding a dataset module page
- **WHEN** a host embeds the dataset module with a valid token and entry route
- **THEN** the dataset module loads with tree navigation available

### Requirement: Single Chart Embedding
The system SHALL support embedding a single chart resource as a standalone view within a host system.

#### Scenario: Embedding a chart
- **WHEN** a host embeds a chart with a valid token and chart id
- **THEN** the chart renders inside the host container

### Requirement: Bidirectional Parameter Passing
The system SHALL support bidirectional parameter passing between the host system and embedded dashboards/screens/charts.

#### Scenario: Host passes filter parameters to embedded content
- **WHEN** the host supplies external parameters during initialization
- **THEN** the embedded content applies those parameters to filter data

#### Scenario: Embedded content sends interaction parameters to host
- **WHEN** a user interacts with an embedded component configured to send parameters
- **THEN** the embedded content posts a message with the interaction payload to the host

### Requirement: Iframe and DIV Embedding Entry Points
The system SHALL document and maintain iframe and DIV embedding entry points for supported resource types.

#### Scenario: Using iframe embedding
- **WHEN** a host uses the iframe embedding entry point with a valid token and init parameters
- **THEN** the embedded content initializes via postMessage and renders in the iframe

#### Scenario: Using DIV embedding
- **WHEN** a host initializes a DIV embedding container with the embedded JS module
- **THEN** the embedded content renders inside the specified DOM element

### Requirement: Origin Validation for Embedding Messages
The system SHALL validate the origin of embedding initialization and bidirectional messages against the configured allowlist.

#### Scenario: Rejecting untrusted origins
- **WHEN** a message is received from an origin not in the allowlist
- **THEN** the message is ignored and the embedded session remains uninitialized

