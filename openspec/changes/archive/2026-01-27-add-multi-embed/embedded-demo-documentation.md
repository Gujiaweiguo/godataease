# Embedded Demo/Docs

This document provides usage examples for multi-dimensional embedding in DataEase.

## Overview

DataEase supports embedding at multiple levels:
1. **Designer embedding** - Canvas designer (dvCanvas, dashboard editor)
2. **Full board/screen embedding** - Preview mode (dashboard/screen)
3. **Single chart embedding** - Individual chart view
4. **Dataset/Datasource embedding** - Dataset/datasource management
5. **Template embedding** - Template management
6. **Link embedding** - Link page
7. **Module-level page embedding** - With tree navigation

## Common Parameters

### URL Parameters (URL Query String or Form Data)

#### Basic Parameters
- `outerId`: Resource ID (dashboard ID, screen ID, chart ID)
- `dvId`: Data visualization ID (for screens)
- `chartId`: Chart ID (for chart embedding)
- `token`: Embedded access token
- `busiFlag`: Business type flag (dashboard, dataV, dataset, datasource)
- `opt`: Operation type (edit, copy, etc.)
- `pid`: Parent ID (for templates)
- `outerParams`: JSON string of additional parameters (Base64 encoded)
- `createType`: Template create type
- `templateParams`: Template parameters (Base64 encoded)
- `jumpInfoParam`: Jump info parameters
- `outerUrl`: Outer URL for redirect
- `datasourceId`: Datasource ID
- **Module-level pages**: Not applicable

### Event Types (Standardized)

#### Parent → Child Events
- `param_update`: Parameter updates from child to parent
- `user_interaction`: User interaction events (clicks, selections, etc.)
- `init_ready`: Child ready for initialization
- `ready`: Child fully initialized
- `error`: Child error notifications

#### Child → Parent Events
- `attach_params`: Parent provides parameters to child (DIV embed mode)
- `jump_to_target`: Jump to another target (chart-to-chart, screen-to-screen)

## Usage Examples

### Example 1: Dashboard Embedding (Iframe)

#### Parent HTML
```html
<!DOCTYPE html>
<html>
<head>
  <title>Embedded Dashboard Example</title>
</head>
<body>
  <h1>DataEase Dashboard Embedding</h1>

  <!-- Embed Dashboard with parameters -->
  <iframe
    id="de-dashboard-iframe"
    src="http://localhost:5100/preview#/previewShow"
    width="100%"
    height="800px"
    frameborder="0"
  ></iframe>

  <h2>Event Listener (Parent)</h2>
  <script>
    // Listen for events from embedded dashboard
    window.addEventListener('message', handleEmbeddingEvent)

    function handleEmbeddingEvent(event) {
      const msg = event.data
      console.log('Received message from child:', msg)

      switch (msg.type) {
        case 'dataease-embedded-interactive':
          // Handle param updates
          if (msg.args) {
            console.log('Parameter updated:', msg.args)
            // Update your UI accordingly
          }

        case 'dataease-embedded-host':
          // Handle initialization events
          if (msg.args) {
            console.log('Child initialized:', msg.args)
            // Show dashboard when ready
          }
      }
    }
  </script>

  <h2>Output</h2>
  <div id="output">
    <h3>Events received:</h3>
    <ul id="events"></ul>
  </div>
</body>
</html>
```

#### Child Code (DataEase Dashboard Preview)
The dashboard automatically handles parent communication through `useEmbeddedParentCommunication` composable.

### Example 2: Screen Embedding (Iframe)

#### Parent HTML
```html
<!DOCTYPE html>
<html>
<head>
  <title>Embedded Screen Example</title>
</head>
<body>
  <h1>DataEase Screen Embedding</h1>

  <!-- Embed Screen with parameters -->
  <iframe
    id="de-screen-iframe"
    src="http://localhost:5100/previewShow"
    width="100%"
    height="800px"
    frameborder="0"
  ></iframe>

  <h2>Event Listener (Parent)</h2>
  <script>
    window.addEventListener('message', handleEmbeddingEvent)

    function handleEmbeddingEvent(event) {
      const msg = event.data
      console.log('Received message from child:', msg)

      switch (msg.type) {
        case 'dataease-embedded-interactive':
          // Handle interactions
          if (msg.args) {
            console.log('Interaction:', msg.args)
          }

        case 'dataease-embedded-host':
          // Handle lifecycle events
          if (msg.args) {
            if (msg.args.includes('init_ready')) {
              console.log('Screen initialized')
            }
          }
      }
    }
  </script>

  <h2>Output</h2>
  <div id="output">
    <h3>Events received:</h3>
    <ul id="events"></ul>
  </div>
</body>
</html>
```

### Example 3: Single Chart Embedding (Iframe)

#### Parent HTML
```html
<!DOCTYPE html>
<html>
<head>
  <title>Embedded Chart Example</title>
</head>
<body>
  <h1>DataEase Chart Embedding</h1>

  <!-- Embed single chart -->
  <iframe
    id="de-chart-iframe"
    src="http://localhost:5100/chart#/chartView"
    width="800px"
    height="600px"
    frameborder="0"
  ></iframe>

  <h2>Event Listener (Parent)</h2>
  <script>
    window.addEventListener('message', handleEmbeddingEvent)

    function handleEmbeddingEvent(event) {
      const msg = event.data
      console.log('Received message from child:', msg)

      switch (msg.type) {
        case 'dataease-embedded-interactive':
          // Chart-specific interactions
          if (msg.args) {
            console.log('Chart interaction:', msg.args)
          }
      }
    }
  </script>

  <h2>Output</h2>
  <div id="output">
    <h3>Events received:</h3>
    <ul id="event-log"></ul>
  </div>
</body>
</html>
```

### Example 4: Dataset Embedding (Iframe)

#### Parent HTML
```html
<!DOCTYPE html>
<html>
<head>
  <title>Embedded Dataset Example</title>
</head>
<body>
  <h1>DataEase Dataset Embedding</h1>

  <!-- Embed dataset management -->
  <iframe
    id="de-dataset-iframe"
    src="http://localhost:5100/dataset#/dataset-embedded"
    width="100%"
    height="600px"
    frameborder="0"
  ></iframe>

  <h2>Event Listener (Parent)</h2>
  <script>
    window.addEventListener('message', handleEmbeddingEvent)

    function handleEmbeddingEvent(event) {
      const msg = event.data
      console.log('Received message from child:', msg)

      switch (msg.type) {
        case 'dataease-embedded-interactive':
          // Dataset interactions
          if (msg.args) {
            console.log('Dataset interaction:', msg.args)
          }
      }
    }
  </script>

  <h2>Output</h2>
  <div id="output">
    <h3>Events received:</h3>
    <ul id="event-log"></ul>
  </div>
</body>
</html>
```

## Event Protocol Details

### Initialization Flow

1. **Parent loads iframe** with token
2. **Child initializes** and checks token validity
3. **Child sends `INIT_READY` event** with resource ID
4. **Parent can send initialization parameters back** if needed
5. **Child sends `READY` event** when fully loaded
6. **Child validates token periodically**
7. **Parent displays warnings** when token is expiring

### Parameter Passing

1. **To Child**: Via `ATTACH_PARAMS` event (DIV embed)
   - Parent sends parameters as Base64 JSON string
   - Example: `{ "param1": "value1", "param2": "value2" }`

2. **From Child**: Via `PARAM_UPDATE` event
   - Child sends parameter updates
   - Example: `{ "param1": "newValue" }`

3. **Interaction**: Via `INTERACTION` event
   - User clicks, selections, filters, etc.
   - Example: `{ "interactionType": "click", "target": "element-id" }`

### Token Lifecycle

1. **Initialization**:
   - `TokenManager.initialize()` - Validates token and sets up lifecycle
   - Auto-refresh every 5 minutes
   - Expiry warning 5 minutes before expiry

2. **Validation**:
   - `validateToken()` - Checks if token is valid
   - `extractTokenExpiryTime()` - Gets expiry time from JWT token
   - `needsTokenRefresh()` - Checks if refresh is needed

3. **Usage in Component**:
   ```typescript
   const tokenLifecycle = useTokenLifecycle()
   const expiryTime = tokenLifecycle.getCurrentTokenInfo()?.expiryTime
   
   // Add expiry warning
   if (expiryTime && tokenLifecycle.needsRefresh(window.location.origin, 5)) {
     console.warn(`Token will expire in 5 minutes`)
   }
   
   // Initialize token lifecycle
   await tokenLifecycle.initialize(
     token,
     window.location.origin,
     {
       refreshEnabled: true,
       tokenType: 'iframe',
       resourceId: dvId
     }
   )
   ```

### Error Handling

1. **Invalid Token**: Child sends `ERROR` event
2. **Expired Token**: Token validation returns expired
3. **Network Error**: Initialization API fails
4. **Parent displays**: Console warnings and error messages

### Module-Level Page Embedding (Enhancement)

**Routes**: `/module-dataset`, `/module-datasource`
**Components**: Module pages with tree navigation
**Features**:
- Left sidebar tree navigation
- Dynamic loading based on tree selection
- Parameter passing to parent
- Token management integration

## Best Practices

1. **Always validate origins**: Use `validateEmbeddedOrigin()` before processing events
2. **Use structured event types**: Use `EmbeddingEventType` enum for type safety
3. **Handle errors gracefully**: Send `ERROR` events with clear error messages
4. **Document parameters**: Provide clear documentation of required vs optional parameters
5. **Test thoroughly**: Test in multiple browsers and scenarios
6. **Monitor token expiry**: Display warnings before token expires
7. **Use auto-refresh**: Enable token auto-refresh for long-lived sessions

## Troubleshooting

### Issue: Events not received
**Check**:
- Is iframe src correct?
- Are parameters being passed correctly?
- Is browser console open for debugging?

**Solution**:
- Add `console.log()` statements throughout component lifecycle
- Verify token is being initialized
- Check parent frame communication is working

### Issue: Token expiry warnings too frequent
**Check**:
- Is expiry time calculation correct?
- Is warning threshold appropriate?

**Solution**:
- Adjust `warningThresholdMinutes` in token lifecycle utilities
- Only warn once per session (track last warning time)

### Issue: Parameters not updating
**Check**:
- Are event handlers calling `setParam()`?
- Is store state being updated?

**Solution**:
- Verify `setParam()` updates both `params` object and individual keys
- Check store getters return updated values

### Issue: Component stuck on initialization
**Check**:
- Is initialization hanging?
- Are there console errors?

**Solution**:
- Check token validation result
- Verify `embeddedInitIframeApi()` is being called
- Check origin resolution is working

## API Reference

### Token Initialization
- `embeddedInitIframeApi()` - Initialize iframe with token
- `embeddedGetTokenArgsApi()` - Get token arguments
- `embedded/domainListApi()` - Get origin allowlist

### Event Type Registry
- `EmbeddingEventType` enum in `/events/embedding/types.ts`

### Communication Composable
- `useEmbeddedParentCommunication()` in `/hooks/event/useEmbeddedParentCommunication.ts`

### Token Manager
- `TokenManager` class in `/services/TokenManager.ts`

### Token Lifecycle Composable
- `useTokenLifecycle()` in `/hooks/embedded/useTokenLifecycle.ts`

### Token Utils
- `isTokenExpiringSoon()` - Check if token is expiring soon
- `extractTokenExpiryTime()` - Extract expiry from JWT token
- `needsTokenRefresh()` - Check if refresh is needed
