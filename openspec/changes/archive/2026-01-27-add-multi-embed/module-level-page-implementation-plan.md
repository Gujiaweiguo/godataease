# Phase 1.2: Module-Level Page Embedding - Implementation

## Overview

This document details the implementation plan for extending embedded routing to support module-level page embedding with tree navigation.

## Requirements

Based on the `embedding-architecture.md` and `bidirectional-parameter-passing-plan.md`, module-level page embedding requires:

1. **Module-Level Page Routing**: Support `/module-*` routes for embedding datasource, dataset, and other module-level pages
2. **Tree Navigation**: Left-side tree navigation for browsing module content
3. **Parameter Passing**: Standardized bidirectional parameter passing using `useEmbeddedParentCommunication()`
4. **Backward Compatibility**: Maintain support for existing embed modes

## Current State

### Existing Routes
From `router/index.ts`, we have existing embedding routes:
- `/dvCanvas` - Screen canvas (designer mode)
- `/dashboard` - Dashboard canvas (designer mode)
- `/previewShow` - Screen preview
- `/dashboardPreview` - Dashboard preview
- `/chart` - Chart view (wrapper)
- `/preview` - Screen preview canvas
- `/dataset-embedded` - Dataset embedding
- `/dataset-embedded-form` - Dataset form embedding
- `/datasource-embedded` - Datasource embedding

### Existing Module-Level Components
From `embedding-inventory.md`:
- Dataset: `/views/visualized/data/dataset/index.vue` - Already has embedding
- Datasource: `/views/visualized/data/datasource/index.vue` - Already has embedding

## Implementation Plan

### 1. Router Configuration

**New Routes to Add:**
```typescript
{
  path: '/module-dataset',
  name: 'module-dataset',
  hidden: true,
  meta: {},
  component: () => import('@/views/visualized/data/dataset/ModulePageWithTree.vue')
},
{
  path: '/module-datasource',
  name: 'module-datasource',
  hidden: true,
  meta: {},
  component: () => import('@/views/visualized/datasource/ModulePageWithTree.vue')
}
```

### 2. Module Page Component Structure

**Component Layout:**
- Left sidebar: Tree navigation (data structure)
- Main content: Data table/form component
- Top bar: Title and actions

**Component Features:**
- Tree navigation for browsing module hierarchy
- Bidirectional parameter passing integration
- Embed mode detection and initialization
- Read-only access control (based on embed parameters)

### 3. Store Enhancements

**Add to embedded store:**
```typescript
interface AppState {
  // ... existing fields
  moduleMode: boolean  // NEW: Track if in module-level page mode
  moduleType: string    // NEW: Track module type (dataset/datasource/etc)
  treeSelection: Record<string, any>  // NEW: Selected tree node
}

actions: {
  setModuleMode(mode: boolean): void
  setModuleType(type: string): void
  setTreeSelection(selection: Record<string, any>): void
}
```

### 4. Tree Navigation Component

**New Component: `/components/navigation/ModuleTree.vue`**

Features:
- Recursive tree rendering
- Click/double-click handlers
- Selection state management
- Expand/collapse functionality
- Search/filter support

## Tasks

### Phase 1.2.1: Requirements Analysis
- [x] Analyze module-level page requirements from architecture docs
- [ ] Identify existing components that can be reused
- [ ] Determine module types to support (initially: dataset, datasource)

### Phase 1.2.2: Router Extension
- [ ] Add `/module-dataset` route
- [ ] Add `/module-datasource` route
- [ ] Test routing in embedded mode

### Phase 1.2.3: Component Development
- [ ] Create `ModuleTree.vue` navigation component
- [ ] Create `/views/visualized/data/dataset/ModulePageWithTree.vue`
- [ ] Create `/views/visualized/data/datasource/ModulePageWithTree.vue`

### Phase 1.2.4: Store Integration
- [ ] Add module mode state to embedded store
- [ ] Add module type state to embedded store
- [ ] Add tree selection state to embedded store

### Phase 1.2.5: Testing
- [ ] Test module page embedding in iframe mode
- [ ] Test tree navigation interactions
- [ ] Test parameter passing between parent and child
- [ ] Manual verification with embedded demo

## Backward Compatibility

To ensure smooth migration:
1. Keep existing `/dataset-embedded` and `/datasource-embedded` routes functional
2. New `/module-*` routes as parallel options
3. Document migration path in code comments
4. Use same embedded store and APIs

## Success Criteria

- [ ] Module-level pages accessible via `/module-*` routes
- [ ] Tree navigation functional for dataset and datasource
- [ ] Bidirectional parameter passing working
- [ ] Embed initialization flow consistent with dashboard/screen
- [ ] Documentation updated with module-level examples
- [ ] All tests passing
