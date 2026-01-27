# Multi-Dimensional Embedding - Architecture & Inventory Summary

## Embedded Entry Points (Current State)

### Dashboard Embedding
- **Editor**: `/dashboard/index.vue` + `embeddedInitIframeApi()`
  - Uses `componentMap` → DashboardEditor
- **Preview**: `/dashboard/DashboardPreviewShow.vue`
  - Uses XpackComponent + `init-iframe="initIframe"`
- **API**: `embeddedInitIframeApi({ token, origin })` → sets `embeddedStore.setAllowedOrigins()`

### Screen/DataV Embedding
- **Designer**: `/data-visualization/index.vue` + `embeddedInitIframeApi()`
  - Uses `componentMap` → VisualizationEditor
- **Preview**: `/data-visualization/PreviewShow.vue` + `embeddedInitIframeApi()`
- **Preview Canvas**: `/data-visualization/PreviewCanvas.vue`
  - Handles postMessage events for interaction
  - Origin validation: `isAllowedEmbeddedMessageOrigin()`

### Chart Embedding
- **ChartView**: `/chart/ChartView.vue`
  - Uses `componentMap` → ViewWrapper + Preview
  - `init-iframe="initIframe"` triggers `initIframe()` → `embeddedInitIframeApi()`
- **DataFilling**: `L21lbnUvZGF0YS9kYXRhLWZpbGxpbmcvbWFuYWdlL2Zvcm0vaW5kZXg=`
  - Uses AsyncXpackComponent

### Dataset/Datasource Embedding
- **Dataset Tree**: `/dataset-embedded` → `/views/visualized/data/dataset/index.vue`
- **Dataset Form**: `/dataset-embedded-form` → `/views/visualized/data/dataset/form/index.vue`
- **Datasource-embedded**: `/datasource-embedded` → `/views/visualized/data/datasource/index.vue`

### Link/Share Embedding
- **LinkContainer**: `/de-link/:uuid` → `/views/data-visualization/LinkContainer.vue`
- **ShareGrid**: `/share/share/ShareGrid.vue`
- **ShareTicket**: `/share/share/ShareTicket.vue`
- **ShareVisualHead**: `/share/share/ShareVisualHead.vue`
- **ShareHandler**: `/share/share/ShareHandler.vue`

### Template Embedding
- **Template Management**: `/template-manage` → `/views/template/indexInject.vue`

### Mobile Panel Embedding
- **Mobile Panel**: `/mobile/panel/index.vue`
- **Mobile Panel** (MIPC): `/mobile/panel/MobileInPc.vue`
  - Handles `panelInit` / `screen-weight` events
  - Origin validation via `isAllowedEmbeddedMessageOrigin()`

### Work Branch Embedding
- **Work Branch**: `/workbranch/index.vue` + `ShortcutTable.vue`
- **ShortcutOption**: `/workbranch/ShortcutOption.ts`

### Module-Level Embedding
- **Not explicitly found** - Components use `componentMap` to select ViewWrapper/Preview based on conditions
- **System pages**: `/system/user`, `/system/role`, `/system/org`, `/system/permission` (managed separately)

## Store Architecture

### Primary Embedded Store
**File**: `/store/modules/embedded.ts`

**State Schema**:
- `type`: Embed type (dashboard/dataV/dataset/datasource)
- `token`: Authentication token for iframes
- `busiFlag`: Business type flag
- `outerParams`: External parameters string (JSON)
- `suffixId`: URL suffix
- `baseUrl`: Base URL for iframe src
- `dvId`: Dashboard/Screen ID
- `pid`: Parent ID
- `chartId`: Chart ID
- `resourceId`: Resource ID
- `dfId`: Dataset form ID
- `opt`: Operation type (create/edit)
- `createType`: Creation type
- `templateParams`: Template parameters
- `jumpInfoParam`: Jump info parameters (Base64 encoded)
- `outerUrl`: External URL
- `datasourceId`: Datasource ID
- `tableName`: Table name
- `datasetId`: Dataset ID
- `datasetCopyId`: Dataset copy ID
- `datasetPid`: Dataset parent ID
- `allowedOrigins`: Array of allowed origins
- `tokenInfo`: Map<string, object> for multi-token support

**Key Actions**:
- `setIframeData(data)`: Sets type, token, busiFlag, outerParams, suffixId, dvId, chartId, pid, resourceId, dfId
- `setTokenInfo(tokenInfo)`: Stores token info map
- `setAllowedOrigins(allowedOrigins)`: Sets allowed origins
- `clearState()`: Clears all embedding state
- `setToken(token)`: Sets authentication token

**Key Getters**:
- `getIframeData()`: Returns object with all embedding parameters
- `getType()`: Returns embed type
- `getToken()`: Returns authentication token
- `getBusiFlag()`: Returns business flag

### Secondary Stores (Embedding Related)
- `/store/modules/data-visualization/dvMain.ts`: Contains `embeddedCallBack: 'no'` flag
- `/store/modules/appearance.ts`: Uses `embeddedStore.baseUrl` for path resolution

## API Layer

### Primary Embedding API
**File**: `/api/embedded.ts`

**Endpoints**:
1. `embeddedInitIframeApi(data)` - Initialize iframe with token and origin
   - Calls `/embedded/initIframe`
   - Returns array of allowed origins
   - Usage: Sets `embeddedStore.setAllowedOrigins()`

2. `embeddedGetTokenArgsApi()` - Get token configuration
   - Calls `/embedded/getTokenArgs`
   - Returns token arguments

3. `embeddedDomainListApi()` - Get domain whitelist
   - Calls `/embedded/domainList`

4. CRUD APIs (Query/Create/Edit/Delete/Reset):
   - `embeddedQueryGridApi(page, pageSize, data)`
   - `embeddedCreateApi(data)`
   - `embeddedEditApi(data)`
   - Embed dedDeleteApi(id) - Delete by ID
   - `embeddedBatchDeleteApi(ids)` - Batch delete
   - `embeddedResetApi(data)` - Reset embedded

### Outer Params API
**File**: `/api/visualization/outerParams.ts`

**Endpoints**:
- `queryWithVisualizationId(dvId)` - Query outer params with visualization ID
- `updateOuterParamsSet(requestInfo)` - Update outer params set
- `getOuterParamsInfo(dvId)` - Get outer params info

### Utilities
**File**: `/utils/embedded.ts`

**Functions**:
- `resolveEmbeddedOrigin()` - Gets origin from document.referrer
- `isAllowedEmbeddedMessageOrigin(origin, allowlist, enforceAllowlist)` - Validates message origin
- `enforceAllowlist` flag - Controls strict allowlist checking

## Router Configuration

### Routes with `-embedded` Suffix (Embed-Specific)
| Route | Component | Embed Type | busiFlag |
|-------|-----------|-----------|-----------|
| `/dataset-embedded` | Dataset Tree | dataset | dataset |
| `/dataset-embedded-form` | Dataset Form | dataset | dataset |
| `/datasource-embedded` | Datasource | datasource | datasource |
| `/preview` | Preview Canvas | dataV | dataV |
| `/dashboard` | Dashboard Editor | dashboard | dashboard |
| `/dashboardPreview` | Dashboard Preview | dashboard | dashboard |
| `/chart-view` | Chart View | dashboard/dataV | dashboard |
| `/de-link/:uuid` | Link Container | dashboard/dataV | dashboard |
| `/template-manage` | Template Management | (varies) | (varies) |
| `/rich-text` | Rich Text Editor | - | - |
| `/modify-pwd` | Password Modify | - | - |

### System Routes
- `/system/user` - User Management (non-embedded)
- `/system/role` - Role Management (non-embedded)
- `/system/org` - Organization Management (non-embedded)
- `/system/permission` - Permission Management (non-embedded)

## Component Patterns

### XpackComponent Pattern (IFrame Embedding)
**Attributes**:
- `jsname`: Base64-encoded jsname for component path
- `@init-iframe="initIframe"`: Triggers iframe initialization
- `@eventCheck="eventKey"`: Event handler for parent communication

**Usage in**:
- `/chart/ChartView.vue` - Uses `componentMap` with multiple options (ViewWrapper, Preview, Dashboard, Dataset, Datasource, ScreenPanel, DashboardPanel)
- `/mobile/panel/index.vue` - Dynamic component loading based on name

**Initialization Flow**:
1. Component mounts
2. `initIframe("ComponentName")` is called
3. Event: `@eventCheck` → eventBus emit → `useEmitt` → emit `changeCurrentComponent`
4. ComponentMap updates currentComponent
5. NextTick → showComponent = true
6. XpackComponent renders with selected component

### DIV-Based Embedding Pattern (Vue Inject)

**Entry Point**: `/pages/panel/ViewWrapper.vue`

**Inject Pattern**:
```typescript
const embeddedParamsDiv = inject('embeddedParams') as object
const embeddedStore = useEmbedded()

// Fallback if inject not available
const embeddedParams = embeddedParamsDiv?.chartId ? embeddedParamsDiv : embeddedStore
```

**Key Fields**:
- `dvId`: Dashboard or Screen ID
- `chartId`: Chart ID (for chart embedding)
- `busiFlag`: Business type (dashboard/dataV/dataset/datasource)
- `outerParams`: JSON-encoded outer parameters
- `suffixId`: URL suffix
- `pid`: Parent ID

**Bidirectional Communication**:
- **Child → Parent**: `window.parent.postMessage(targetPm, '*')`
  - Type: `'dataease-embedded-interactive'`
  - Events: `'de_inner_params'`, `'canvas_init_ready'`, other custom events
- **Parent → Child**: `initIframe()` or direct property access

## Communication Protocols

### PostMessage Events (Child → Parent)
| Event Type | Event Name | Payload | Handler |
|-----------|------------|---------|----------|
| Parameter Update | `'de_inner_params'` | `{ param1: value1, param2: value2 }` | `winMsgHandle` (various components) |
| Canvas Init | `'canvas_init_ready'` | `{ resourceId: dvId }` | `onInitReady({ resourceId })` |
| User Interaction | `'dataease-embedded-interactive'` | `{ param, value }` | `methodName: 'embeddedInteractive'` |
| Panel Weight | `'panel-weight'` | `{ dvId, weight }` | `eventCheck` → `check()` |
| Screen Weight | `'screen-weight'` | `{ dvId, weight }` | `eventCheck` → `check()` |
| Component Style | `'componentStyleChange'` | `{ type, component, ... }` | `mobileViewStyleSwitch()` |
| Component Data | `'calcData'` / `'renderChart'` | `{ component: ..., ... }` | UseEmitt events |

### PostMessage Events (Parent → Child)
| Event Type | Handler | Action |
|-----------|----------|--------|
| Jump to Target | `jumpInfoParam` | Base64 decode → update state → query jump info |
| Canvas Init | `initIframe()` | Render canvas with provided parameters |
| Iframe Init | `initIframe("ComponentName")` | Initialize with selected component |

## Embedding Types & Business Flags

### Dashboard Embedding
- **busiFlag**: `dashboard`
- **Type**: `dashboard`
- **Supported**: IFrame (via XpackComponent), DIV (via ViewWrapper inject)
- **Entry Points**: `/dashboard` (editor), `/dashboardPreview` (preview)

### Screen/DataV Embedding
- **busiFlag**: `dataV`
- **Type**: `dataV` (data visualization)
- **Supported**: IFrame (via XpackComponent), DIV (via PreviewCanvas inject)
- **Entry Points**: `/dvCanvas` (editor), `/preview` (canvas), `/previewShow` (showcase)

### Chart Embedding
- **busiFlag**: `dashboard` or `dataV` (varies)
- **Type**: Chart
- **Supported**: DIV (via ViewWrapper inject) + IFrame fallback
- **Entry Point**: `/chart-view` → ViewWrapper component

### Dataset Embedding
- **busiFlag**: `dataset`
- **Type**: Dataset
- **Supported**: DIV (via ViewWrapper inject)
- **Entry Points**: `/dataset-embedded` (tree), `/dataset-embedded-form` (form)

### Datasource Embedding
- **busiFlag**: `datasource`
- **Type**: Datasource
- **Supported**: DIV (via ViewWrapper inject)
- **Entry Point**: `/datasource-embedded`

### Link Embedding
- **busiFlag**: `dashboard` or `dataV` (varies)
- **Type**: Public Link
- **Supported**: IFrame (via XpackComponent or custom iframe)
- **Entry Point**: `/de-link/:uuid` → LinkContainer

### Template Embedding
- **busiFlag**: (not set)
- **Type**: Template
- **Supported**: IFrame (via XpackComponent)
- **Entry Point**: `/template-manage`

## Token & Origin Management

### Token-Based Authentication
- Token passed via query params or `embeddedInitIframeApi()`
- Token validated against backend
- Token stored in `embeddedStore.token`

### Origin Allowlist
- Fetched via `embeddedInitIframeApi()` → `embeddedDomainListApi()`
- Stored in `embeddedStore.allowedOrigins`
- Validated via `isAllowedEmbeddedMessageOrigin()` on all postMessage events
- Supports allowlist enforcement via `enforceAllowlist` flag

### Origin Resolution
```typescript
// From /utils/embedded.ts
function resolveEmbeddedOrigin(): string {
  return document.referrer.split('/').slice(0, -1)[0]
}

// From /utils/embedded.ts
function isAllowedEmbeddedMessageOrigin(origin: string, allowlist: string[], enforceAllowlist?: boolean): boolean {
  // Normalization logic
  // Allowlist comparison
  // Security validation
}
```

## Key Technical Decisions

### 1. XpackComponent Usage
- Used for iframe-based embedding (charts, dashboards, screens, data sources, datasets, templates)
- Provides async component loading
- Standardizes initialization via `@init-iframe="initIframe"`
- Supports multiple embed types via componentMap

### 2. DIV-Based Embedding
- Used for module-level page embedding and dashboard embedding
- Provides bidirectional parameter passing via Vue inject
- Supports canvas/dataset/chart communication with embedded mode
- Origin validation applied to all postMessage events

### 3. Event-Driven Communication
- Extensive use of eventBus for component changes
- Custom events: `panel-weight`, `screen-weight`, `componentStyleChange`, etc.
- Event emission pattern: `useEmitt().emit('eventName', payload)`

### 4. Component Selection Strategy
- Dynamic component loading based on name
- ComponentMap provides mapping from string names to async components
- Supports conditional rendering (isOtherEditorShow, panelInit, etc.)

### 5. Parameter Encoding
- `outerParams` stored as JSON string in embeddedStore
- `jumpInfoParam` Base64 encoded for security
- Decoded via `JSON.parse(Base64.decode())`

## Known Limitations

### 1. Inconsistent State Management
- Some components check `embeddedStore.baseUrl` directly
- Others use `inject('embeddedParams')` with fallback
- State clearing is inconsistent (only ViewWrapper/DashboardPreview do it correctly)

### 2. Fragmented Event Protocols
- Multiple postMessage event names without clear standardization
- Some events have unclear payloads
- `'de_inner_params'` event name unclear (de = dataease?)

### 3. ComponentMap Overuse
- ComponentMap used in 7+ views with different options
- Some options are rarely used (TemplateManage, Dataset, Datasource)
- Performance impact unclear

### 4. Duplicate Code Patterns
- `embeddedInitIframeApi()` called in multiple places
- `isAllowedEmbeddedMessageOrigin()` duplicated in multiple files
- Similar iframe initialization logic across components

### 5. Limited DIV-Based Documentation
- No clear documentation on DIV embedding patterns
- Examples of parameter passing are sparse
- `componentMap` pattern needs better documentation

## Success Criteria

### Phase 1.1: Inventory & Mapping
- [x] All entry points categorized by embedding type (dashboard/screen/chart/module/etc.)
- [x] All embedding patterns documented (iframe vs DIV, XpackComponent, inject)
- [x] Store architecture analyzed (state, actions, getters)
- [x] API layer cataloged (embedding + outer params)
- [x] Router configuration reviewed (embed-specific routes)
- [x] Communication protocols documented (postMessage events)
- [x] Token/origin management understood (allowlist, validation)
- [ ] Module-level page embedding identified (if exists separately)
- [ ] Unified embedding type registry created (conceptual)

### Phase 1.2: Router Extension
- [ ] Module-level page routes designed
- [ ] Tree navigation component created (if needed)
- [ ] Nested module structure defined (if needed)

### Phase 1.3: Token/Origin Alignment
- [ ] Token initialization flow documented for all embed types
- [ ] Origin allowlist checks documented per type
- [ ] Consistent validation patterns established

### Phase 1.4: Bidirectional Parameter Passing
- [ ] Standardized parameter format defined
- [ ] Bidirectional hooks created for all embed types
- [ ] Event type registry documented
- [ ] Parent-child communication standardized

### Phase 1.5: Documentation Updates
- [ ] Embedded demo updated with multi-dimensional examples
- [ ] User guide created for embedding configuration
- [ ] API documentation updated (if needed)

## Open Questions

1. Should module-level pages be embedded as DIV containers or iframes?
2. Should we create a unified embedding context provider instead of scattered store usage?
3. Should we standardize all event names for clearer communication?
4. Do we need to support legacy embed types (some components use different patterns)?
5. What is the target timeline for this feature completion?

## Dependencies

### Required Dependencies
- Existing embedding infrastructure (store, APIs, utilities, XpackComponent)
- Vue 3 Composition API
- Pinia state management

### Optional Dependencies
- Tree navigation component (if module-level embedding with tree structure)
- Unified embedding context provider (refactoring opportunity)

### Blockers

### Technical
- Complex component map with rarely-used options (performance consideration)
- Inconsistent state management across components
- Scattered event protocols

### Process
- Need for clear requirements document before implementation
- Need to decide on module-level page embedding approach (DIV vs iframe)

### Risk Mitigation

### Risk: Breaking existing embed integrations
- **Mitigation**: Create extension hooks for backwards compatibility
- Document migration path for new unified embedding approach

### Risk: Token/origin allowlist complexity
- **Mitigation**: Provide clear API for querying and configuring allowlist
- Document best practices for origin validation

## Next Steps

1. ✅ Complete Phase 1.1 (Inventory & Mapping)
2. ⏸ Complete Phase 1.2 (Router Extension for module-level pages)
3. ⏸ Complete Phase 1.3 (Token/Origin Alignment)
4. ⏸ Complete Phase 1.4 (Bidirectional Parameter Passing)
5. ⏸ Complete Phase 1.5 (Documentation Updates)

## Implementation Priority

### Immediate (Phase 1)
1. Document module-level page embedding requirements
2. Identify or create module page components
3. Extend router configuration if needed
4. Document bidirectional parameter passing requirements

### Short Term (Phase 2)
1. Implement unified embedding context provider
2. Standardize event protocols
3. Create event type registry
4. Extend token/origin management for new embed types

### Long Term
1. Refactor to unified embedding architecture
2. Create comprehensive documentation
3. Add automated tests for all embed types
4. Performance optimization

## References

### Key Files
- `/store/modules/embedded.ts` - Primary embedding store
- `/api/embedded.ts` - Embedding API
- `/api/visualization/ovalParams.ts` - Outer params API
- `/utils/embedded.ts` - Embedding utilities
- `/router/index.ts` - Router configuration
- `/views/chart/ChartView.vue` - Chart embedding entry
- `/pages/panel/ViewWrapper.vue` - DIV embedding entry
- `/views/dashboard/index.vue` - Dashboard embedding
- `/views/data-visualization/PreviewCanvas.vue` - Screen preview
- `/views/data-visualization/PreviewShow.vue` - Screen show

### API Endpoints
- `POST /embedded/initIframe` - Initialize iframe
- `POST /embedded/getTokenArgs` - Get token args
- `GET /embedded/domainList` - Get domain allowlist
- `POST /embedded/create` - Create embedded
- `POST /embedded/edit` - Edit embedded
- `POST /embedded/delete/{id}` - Delete embedded
- `POST /embedded/batchDelete` - Delete multiple
- `POST /embedded/reset` - Reset embedded
- `POST /embedded/pager/{page}/{pageSize}` - Query grid
- `GET /api/visualization/outerParams/queryWithVisualizationId` - Query outer params
- `POST /api/visualization/ovalParams/updateOuterParamsSet` - Update outer params

### Event Types
- `changeCurrentComponent` - Component change
- `panelInit` - Panel initialization
- `screen-weight` - Screen weight change
- `componentStyleChange` - Component style change
- `updateTitle` - Title update
- `calcData` - Calculate data
- `renderChart` - Render chart
- `doCanvasInit-canvas-main` - Canvas init
- `dataease-embedded-interactive` - Embedded interaction
- `de_inner_params` - Inner parameter update
- `canvas_init_ready` - Canvas init ready
- `init_ready` - Init ready
- `attachParams` - Attach params
