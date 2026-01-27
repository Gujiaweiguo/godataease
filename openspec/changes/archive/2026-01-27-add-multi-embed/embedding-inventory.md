# OpenSpec Change: Add multi-dimensional embedding

## Inventory of Current Embedding Entry Points

This document provides a comprehensive inventory of current embedding entry points in the DataEase frontend, mapped to:
- Dashboards
- Screens
- Data sources
- Datasets
- Charts
- Designers

### 1. Dashboard Entry Points

| Entry Type | View | Component | API | Store Method | Notes |
|-----------|------|-----------|-------|----------|-------------|--------|
| Dashboard | `/dashboard/index.vue` | `XpackComponent` + `init-iframe="initIframe"` | `embeddedInitIframeApi` | Dashboard mode via embedded store |
| Dashboard Preview | `/dashboard/DashboardPreviewShow.vue` | `XpackComponent` + `init-iframe="initIframe"` | `embeddedInitIframeApi` | Dashboard preview with configuration panel |

### 2. Screen Entry Points

| Entry Type | View | Component | API | Event Check Key | Notes |
|-----------|------|-----------|-------|----------|--------------------------|
| Screen | `/data-visualization/PreviewShow.vue` | `XpackComponent` + `init-iframe="initIframe"` | `screen-weight` | Screen embedding with weight configuration |
| Panel Weight | `/mobile/panel/index.vue` | `XpackComponent` | `panel-weight` | Panel configuration events |

### 3. Chart Entry Points

| Entry Type | View | Component | API | Data Path | Notes |
|-----------|------|-----------|-------|----------|---------------------|--------|
| Data Filling | `/chart/ChartView.vue` | `XpackComponent` + `init-iframe="initIframe"` | `embeddedInitIframeApi` | `dataFillingPath` (L21lbnUvZGF0YS9kYXRhLWZpbGxpbmcvbWFuYWdlL2Zvcm0vaW5kZXg=') |

### 4. Data Source Entry Points

| Entry Type | View | Notes |
|-----------|------|--------|
| Data Source Tree | `/common/DeResourceTree.vue` | Resource management for embedding |

### 5. Dataset Entry Points

| Entry Type | View | Notes |
|-----------|------|--------|
| Dataset Management | `/views/visualized/data/dataset/` | Dataset management for embedding |

### 6. Designer Entry Points

| Entry Type | View | Component | Notes |
|-----------|------|-----------|--------|
| Designer | `/views/data-visualization/index.vue` | `VisualizationEditor` | Designer editor with preview |
| Designer Preview | `/views/data-visualization/PreviewCanvas.vue` | Canvas preview component |

### 7. Module-Level Page Entry Points

| Entry Type | View | Notes |
|-----------|------|--------|
| Work Branch | `/views/workbranch/` | Work branch operations |

## Key Findings

### 1. Embedding Store Architecture

The `/store/modules/embedded.ts` store provides:
- Token management for embedded iframes
- Allowed origins (allowedOrigins)
- Embed type selection (create/read/edit)
- Resource IDs for various entities (dvId, chartId, resourceId, etc.)
- Parameter management for different embed types
- Token info mapping for multiple tokens

### 2. Iframe-based Embedding

**Current Implementation**:
- `initIframe="initIframe"` on `XpackComponent` triggers iframe initialization
- Calls `embeddedInitIframeApi` with token and origin
- Sets iframe style (full height/width)

**Embed Types**:
- `DataFilling`: L21lbnUvZGF0YS9kYXRhLWZpbGxpbmcvbWFuYWdlL2Zvcm0vaW5kZXg=`
- `Dashboard`: `dvId` based dashboard paths
- `Designer`: Designer-specific paths

### 3. XpackComponent Usage

The `XpackComponent` provides:
- Async component loading for external systems
- Message passing via `@init-iframe` attribute
- Event handling with `@init-iframe` and `@eventCheck` hooks
- Support for multiple embed types (DataFilling, Dashboard, PanelWeight, Screen, etc.)

### 4. Event-Driven Communication

**Event Types**:
- `panel-weight`: Panel weight configuration changes
- `screen-weight`: Screen weight configuration changes
- `componentStyleChange`: Component style changes
- `updateTitle`: Title updates
- `calcData`: Data calculation updates
- `renderChart`: Chart rendering events

### 5. Router Configuration

Current routing structure supports:
- `/dvCanvas`: Designer canvas
- `/dashboard`: Dashboard view
- `/chart`: Chart view
- `/previewShow`: Preview showcase

## Next Steps

### Phase 1. Implementation

1.1 Extend routing to support module-level page embedding with tree navigation
   - Add routes for module-level pages
   - Implement tree navigation component

1.2 Implement bidirectional parameter passing hooks for dashboard/screen/chart embedding
   - Standardize parameter format and callbacks
   - Add whitelist configuration support

1.3 Extend embedded store to support DIV-based embedding
   - Add DIV embedding entry points to domain list
   - Implement DIV style configuration
   - Support DIV resize events

1.4 Update embedded demo/docs to reflect multi-dimensional embedding
   - Document parameter formats for each embed type
   - Add examples for dashboard/screen/chart/module-level page embedding

### Phase 2: Validation

2.1 Add/extend tests for embedded token generation and origin allowlist
2.2 Add/extend tests for embedding parameter initialization and callback messaging
2.3 Manual verification with embedded demo for dashboards, screens, module pages, and single charts

### Phase 3: Documentation

3.1 Update user guide for multi-dimensional embedding
3.2 Add API documentation (Knife4j)
3.3 Create migration guide for upgrading existing embed types

## Open Questions

1. Should we prioritize iframe-based or DIV-based embedding for new embed types?
2. Should we create a unified embed API that abstracts both iframe and DIV approaches?
3. Do we need to support legacy embed types while introducing new ones?
4. What is the target date for this feature?
5. Should we create a migration path for existing embedded integrations?

## Architecture Decisions

### Decision: Use XpackComponent for Unified Embedding

**Rationale**: The existing `XpackComponent` provides:
- Standardized initialization via `@init-iframe`
- Event-driven communication system
- Support for multiple embed types (DataFilling, Dashboard, PanelWeight, Screen)
- Async component loading pattern

**Alternatives Considered**:
- Create separate iframe and DIV components
- Use message passing via postMessage events
- Use inline iframe src for static content

**Trade-offs**:
- XpackComponent: External dependency, but provides rich features
- Custom component: More control, but higher maintenance cost

**Migration Plan**: Gradual migration to custom components for critical embed types

### Risks / Mitigation

| Risk | Mitigation |
|------|------------|
| Breaking existing embed integrations | Create extension hooks for backwards compatibility |
| Token-based iframe init may become deprecated | Document migration path for new embed types |
| Multiple token formats in use | Standardize on token format and origin allowlist |

## Implementation Priority

1. **High**: Inventory and mapping (THIS TASK)
2. **High**: Bidirectional parameter passing
3. **High**: Module-level page routing
4. **Medium**: DIV embedding capability
5. **Medium**: Update embedded demo/docs
6. **Low**: Unified embed API design

## Related Specs

- `embedded-bi` (new spec for this change)

## Success Criteria

- [ ] 1.1 Complete inventory of all embedding entry points mapped to target types
- [ ] 1.2 Document current iframe/DIV embedding patterns per entry point
- [ ] 1.3 Create routing map for all embed entry points
- [ ] 2.1 Implement unified parameter passing hooks for all embed types
- [ ] 2.2 Add/extend tests for parameter passing
- [ ] 2.3 Manual verification with embedded demo
- [ ] 3.1 Update user guide with multi-dimensional embedding instructions
- [ ] 3.2 Add API documentation for unified embedding
- [ ] 3.3 Create migration guide
