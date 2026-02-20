# DataEase Embedding Demo

This directory contains demonstration pages and automated tests for multi-dimensional embedding in DataEase.

## Demo Pages

### Dashboard Embedding (`dashboard-embed.html`)

Demonstrates embedding a DataEase dashboard using iframe.

**Features:**
- Configure dashboard URL
- Set access token
- Choose theme (light/dark)
- Update embedding dynamically
- Real-time event logging
- Parent-child event communication

**Usage:**
1. Open `dashboard-embed.html` in a browser
2. Update the dashboard URL if needed (default: `http://localhost:5100/preview#/previewShow`)
3. Enter an access token if authentication is required
4. Click "Update Embedding" to apply changes
5. Monitor the Event Log for real-time communication events

**Supported Events:**
- `dataease-embedded-interactive` - Parameter updates from child
- `dataease-embedded-host` - Child initialization
- `ready` - Child fully ready
- `error` - Error notifications
- `user_interaction` - User actions (clicks, selections)
- `param_update` - Parameter changes

### Screen Embedding (`screen-embed.html`)

Demonstrates embedding a DataEase screen (data visualization).

**Features:**
- Configure screen URL
- Set access token
- Specify visualization ID (dvId)
- Real-time event logging

**Usage:**
1. Open `screen-embed.html` in a browser
2. Update the screen URL if needed (default: `http://localhost:5100/preview#/screenDataV`)
3. Enter access token and dvId if required
4. Click "Update Embedding" to apply changes
5. Monitor the Event Log for communication events

## Automated Tests

### Running Tests with Playwright

```bash
# Install dependencies if needed
cd core/core-frontend
npm install -D @playwright/test

# Run tests
npm run test:e2e

# Run tests in headless mode (CI/CD)
CI=true npm run test:e2e

# Run tests with specific pattern
npm run test:e2e embedding-verification
```

### Test Coverage

The automated tests verify:

1. **Dashboard Embedding**
   - Page loads correctly
   - Controls are visible and functional
   - Event logging works
   - Event listeners are initialized
   - Message receiving works

2. **Screen Embedding**
   - Page loads correctly
   - Screen-specific controls display
   - Parameter updates work

3. **Parameter Initialization**
   - Token parameters update iframe src
   - Multiple parameters are combined correctly
   - URL encoding works properly

4. **Event Communication**
   - `param_update` events are logged
   - `ready` events are logged
   - `error` events are logged
   - `user_interaction` events are logged
   - `init_ready` events are logged

5. **Iframe Functionality**
   - Iframe element is rendered
   - Correct attributes are set
   - Responsive sizing works

6. **Cross-Origin Communication**
   - Messages from child are received
   - Messages to child work correctly

7. **Responsive Design**
   - Layout adapts to different viewport sizes
   - Controls remain usable on mobile

8. **Accessibility**
   - All inputs have proper labels
   - Heading hierarchy is correct
   - Keyboard navigation works

9. **Error Detection**
   - Console errors are captured
   - Errors are displayed in log

## Manual Testing Checklist

When testing embedding functionality, verify:

### Dashboard Embedding
- [ ] Dashboard loads without errors
- [ ] Token parameter is accepted
- [ ] Theme parameter changes appearance
- [ ] Event log updates when child sends messages
- [ ] Parent can send parameter updates to child
- [ ] Child initialization event is received

### Screen Embedding
- [ ] Screen loads without errors
- [ ] dvId parameter is used correctly
- [ ] Token authentication works
- [ ] Events are logged properly

### Event Communication
- [ ] Parent → Child messages work
- [ ] Child → Parent messages work
- [ ] All event types are handled
- [ ] Message format matches specification

### Error Handling
- [ ] Invalid tokens show appropriate error
- [ ] Network errors are logged
- [ ] Expired tokens trigger refresh logic
- [ ] Console errors are displayed

### Cross-Origin
- [ ] Messages work across different origins
- [ ] Origin validation works correctly
- [ ] Security measures don't block legitimate messages

## Browser Compatibility

Test in the following browsers:
- [ ] Chrome/Chromium
- [ ] Firefox
- [ ] Safari
- [ ] Edge

## Troubleshooting

### Issue: Events not received
**Solution:** Check that:
- Iframe URL includes correct parameters
- Message event listener is attached before iframe loads
- `message` event uses correct event type

### Issue: Token not working
**Solution:** Verify:
- Token is valid and not expired
- Token is passed as URL parameter
- Backend accepts the token for embedding

### Issue: Styling issues
**Solution:** Check:
- Theme parameter is applied correctly
- CSS is not blocked by parent page
- Parent page allows iframe styling

### Issue: Console errors
**Solution:** Review:
- Network requests for resources
- JavaScript execution in iframe
- Cross-origin policies

## Development

### Running Dev Server

```bash
cd core/core-frontend
npm run dev
```

Then access demo pages at:
- `http://localhost:5173/embedding-demo/dashboard-embed.html`
- `http://localhost:5173/embedding-demo/screen-embed.html`

### API Endpoints

For testing, you'll need valid tokens from:
- `POST /api/embedded/iframe/init` - Initialize embedding
- `GET /api/embedded/token/args` - Get token arguments

See `embedded-demo-documentation.md` for detailed API usage.
