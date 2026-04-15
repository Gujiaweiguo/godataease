## 1. Backend: Add Canonical Handler Methods

- [ ] 1.1 Add `Save` method to `DatasetHandler` in `dataset_handler.go`: parse request via `parseDatasetWriteRequest(c, true)`, call `h.service.Save(req)`, return result via `response.Success`/`response.Error`
- [ ] 1.2 Add `Create` method to `DatasetHandler` in `dataset_handler.go`: parse request via `parseDatasetWriteRequest(c, true)`, call `h.service.Create(req)`, return result
- [ ] 1.3 Add `Rename` method to `DatasetHandler` in `dataset_handler.go`: parse request via `parseDatasetWriteRequest(c, true)`, validate `req.ID > 0`, call `h.service.Rename(req.ID, req.Name)`, return result
- [ ] 1.4 Add `Move` method to `DatasetHandler` in `dataset_handler.go`: parse request via `parseDatasetWriteRequest(c, false)`, validate `req.ID > 0`, extract PID (default 0), call `h.service.Move(req.ID, pid)`, return result
- [ ] 1.5 Add `Delete` method to `DatasetHandler` in `dataset_handler.go`: parse `:id` path param as int64, call `h.service.Delete(id)`, return nil result in success envelope
- [ ] 1.6 Add `PerDelete` method to `DatasetHandler` in `dataset_handler.go`: parse `:id` path param as int64, call `h.service.PerDelete(id)`, return result in success envelope

## 2. Backend: Register Canonical Routes

- [ ] 2.1 Add 6 route registrations in `registerDatasetRoutes()` in `router.go`: `POST /save`, `POST /create`, `POST /rename`, `POST /move`, `POST /delete/:id`, `POST /perDelete/:id` mapped to the new `DatasetHandler` methods
- [ ] 2.2 Update the standalone `RegisterDatasetRoutes` function in `dataset_handler.go` to also register the 6 new routes for consistency
- [ ] 2.3 Run `make test` in `apps/backend-go` to verify compilation and existing tests pass

## 3. Frontend: Update API Paths

- [ ] 3.1 Update 6 URL paths in `apps/frontend/src/api/dataset.ts`: change `/datasetTree/save` → `/dataset/save`, `/datasetTree/create` → `/dataset/create`, `/datasetTree/rename` → `/dataset/rename`, `/datasetTree/move` → `/dataset/move`, `/datasetTree/delete/${id}` → `/dataset/delete/${id}`, `/datasetTree/perDelete/${id}` → `/dataset/perDelete/${id}`
- [ ] 3.2 Run `npm run lint` and `npm run ts:check` in `apps/frontend` to verify no regressions

## 4. Verification

- [ ] 4.1 Start backend (`make run-local`) and verify all 6 new canonical endpoints respond correctly (curl smoke test)
- [ ] 4.2 Verify frontend dev server proxies requests to new paths without errors
