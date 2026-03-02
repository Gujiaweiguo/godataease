# E2E Tests

End-to-end tests for DataEase frontend using [Playwright](https://playwright.dev/).

## Structure

```
e2e/
├── auth/              # Authentication tests
│   └── login.spec.ts  # Login flow tests
├── chart/             # Chart editor tests
│   └── chart.spec.ts  # Chart editor smoke tests
├── datasource/        # Datasource management tests
│   └── datasource.spec.ts
├── embedding/         # Embedding parameter tests
│   └── embedding.spec.ts
├── interactive/       # Interactive tree tests
│   └── interactive.spec.ts
├── map/               # Map chart tests
│   └── map.spec.ts
└── smoke.spec.ts      # Basic smoke tests
```

## Running Tests

### Prerequisites

1. Install dependencies:
   ```bash
   npm install
   ```

2. Install Playwright browsers:
   ```bash
   npx playwright install
   ```

### Local Development

```bash
# Run all E2E tests (starts dev server automatically)
npm run e2e

# Run tests with UI mode
npm run e2e:ui

# Run tests in debug mode
npm run e2e:debug

# View test report
npm run e2e:report
```

### Against Running Server

```bash
# Run against localhost:8080
E2E_BASE_URL=http://localhost:8080 npm run e2e
```

### CI

E2E tests can be run in CI with backend service:

1. **Manual trigger**: Go to Actions → Frontend CI → Run workflow → Enable "Run E2E tests"

2. **Automatic (when backend available)**: The `e2e` job will run when backend service is configured in CI.

Note: Tests marked with `test.fixme` require backend service and will be skipped without it.

## Test Categories

| Category | Description | Backend Required |
|----------|-------------|------------------|
| Smoke | Basic page load tests | No |
| Auth | Login/logout tests | Partial |
| Chart | Chart editor functionality | Yes |
| Map | Map chart types (flow, heat, symbolic) | Yes |
| Embedding | Embedded mode parameters | Yes |
| Interactive | Resource tree navigation | Yes |
| Datasource | Datasource management | Yes |
## Test Credentials

Tests use environment variables for credentials:

- `E2E_USERNAME` - Default: `admin`
- `E2E_PASSWORD` - Default: `DataEase123456`

## Writing Tests

```typescript
import { expect, test } from '@playwright/test'

test('should do something', async ({ page }) => {
  await page.goto('/')
  await expect(page.locator('h1')).toBeVisible()
})
```

## Best Practices

1. Use data-testid attributes for reliable selectors
2. Keep tests independent - each test should work in isolation
3. Use `test.beforeEach` for common setup (e.g., login)
4. Avoid hard-coded waits - use Playwright's auto-waiting
