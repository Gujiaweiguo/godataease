## 1. Domain Layer

- [x] 1.1 Create Watermark domain object (`internal/domain/visualization/watermark.go`)
- [x] 1.2 Define WatermarkSaveRequest struct
- [x] 1.3 Define default watermark settings JSON structure

## 2. Repository Layer

- [x] 2.1 Create WatermarkRepository interface
- [x] 2.2 Implement FindLatest method
- [x] 2.3 Implement SaveDefault method with upsert
- [ ] 2.4 Add repository integration test (deferred)

## 3. Service Layer

- [x] 3.1 Create WatermarkService struct
- [x] 3.2 Implement Find method with default fallback
- [x] 3.3 Implement Save method
- [x] 3.4 Add service unit test for Find
- [x] 3.5 Add service unit test for Save
- [ ] 3.6 Add service integration test (deferred)

## 4. HTTP Handler

- [x] 4.1 Create WatermarkHandler struct
- [x] 4.2 Implement Find handler
- [x] 4.3 Implement Save handler
- [x] 4.4 Implement RegisterWatermarkRoutes function
- [x] 4.5 Add handler unit test for Find
- [x] 4.6 Add handler unit test for Save

## 5. Router Integration

- [x] 5.1 Register watermark routes in router.go
- [x] 5.2 Initialize watermark dependencies in wireDependencies

## 6. Testing & Validation

- [x] 6.1 Run unit tests: `go test ./internal/service/... ./internal/transport/http/...`
- [ ] 6.2 Run integration tests: `go test -tags=integration ./internal/...` (deferred)
- [x] 6.3 Run build: `make build`
- [x] 6.4 Run linter: `golangci-lint run`

## 7. Documentation

- [ ] 7.1 Add API documentation comments (deferred)
- [ ] 7.2 Update contract whitelist if needed (deferred)
